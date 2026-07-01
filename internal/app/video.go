package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type VideoOutput struct {
	URL          string `json:"url,omitempty"`
	CoverURL     string `json:"cover_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	PreviewURL   string `json:"preview_url,omitempty"`
}

func (s *Server) videoGeneration(c *gin.Context) {
	req, imageURLs, err := s.parseEditRequest(c)
	if err != nil {
		openAIError(c, http.StatusBadRequest, err.Error())
		return
	}
	s.runVideoTask(c, req, imageURLs)
}

func (s *Server) runVideoTask(c *gin.Context, req ImageRequest, imageURLs []string) {
	spec, err := modelSpecByID(req.Model)
	if err != nil {
		openAIError(c, http.StatusBadRequest, err.Error())
		return
	}
	if spec.Media != MediaVideo {
		openAIError(c, http.StatusBadRequest, fmt.Sprintf("model %q is an image model", req.Model))
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		openAIError(c, http.StatusBadRequest, "prompt is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.RequestTimeout)
	defer cancel()

	started := time.Now()
	ok := false
	defer func() {
		s.metrics.Record("video", ok, time.Since(started))
		s.metrics.Record(spec.Family, ok, time.Since(started))
	}()

	var items []VideoOutput
	err = s.pool.Run(ctx, func(ctx context.Context) error {
		var runErr error
		items, runErr = s.runOneVideoTask(ctx, req, imageURLs, spec)
		return runErr
	})
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, ErrQueueFull) {
			status = http.StatusTooManyRequests
		}
		openAIError(c, status, err.Error())
		return
	}
	ok = true
	c.JSON(http.StatusOK, gin.H{"created": time.Now().Unix(), "data": items})
}

func (s *Server) runOneVideoTask(ctx context.Context, req ImageRequest, imageURLs []string, spec ModelSpec) ([]VideoOutput, error) {
	clientTaskID := req.ClientTaskID
	if clientTaskID == "" {
		clientTaskID = fmt.Sprintf("video-%d", time.Now().UnixNano())
	}
	taskReq := BananaSubmitRequest{
		Prompt:       req.Prompt,
		AspectRatio:  normalizeVideoAspectRatio(firstNonEmpty(req.AspectRatio, req.Size)),
		Resolution:   spec.Resolution,
		Duration:     normalizeVideoDuration(req.Duration),
		WebhookURL:   req.WebhookURL,
		ClientTaskID: clientTaskID,
	}

	path := spec.TextEndpoint
	if spec.StartEndEndpoint != "" && len(imageURLs) >= 2 && req.FirstFrameURL == "" && req.LastFrameURL == "" {
		req.FirstFrameURL = imageURLs[0]
		req.LastFrameURL = imageURLs[1]
	}
	if req.FirstFrameURL != "" || req.LastFrameURL != "" {
		if spec.StartEndEndpoint == "" {
			return nil, errors.New("start/end frame video is not supported by this model")
		}
		taskReq.FirstFrameURL = req.FirstFrameURL
		taskReq.LastFrameURL = req.LastFrameURL
		path = spec.StartEndEndpoint
	} else if len(imageURLs) > 0 {
		taskReq.ImageURLs = imageURLs
		path = spec.ImageEndpoint
	}

	if path == spec.ImageEndpoint && len(taskReq.ImageURLs) == 0 {
		return nil, errors.New("imageUrls is required for image-to-video")
	}
	if path == spec.StartEndEndpoint && (taskReq.FirstFrameURL == "" || taskReq.LastFrameURL == "") {
		return nil, errors.New("firstFrameUrl and lastFrameUrl are required for start/end video")
	}

	task, err := s.client.Submit(ctx, path, taskReq)
	if err != nil {
		return nil, err
	}
	waitCtx, cancel := context.WithTimeout(context.Background(), s.cfg.RequestTimeout)
	defer cancel()
	done, err := s.client.Wait(waitCtx, task.TaskID)
	if err != nil {
		return nil, err
	}
	outputs := make([]VideoOutput, 0, len(done.Results))
	for _, result := range done.Results {
		if url := result.VideoURLValue(); url != "" {
			outputs = append(outputs, VideoOutput{
				URL:          url,
				CoverURL:     result.CoverURLValue(),
				ThumbnailURL: result.ThumbnailURLValue(),
				PreviewURL:   result.PreviewURLValue(),
			})
		}
	}
	if len(outputs) == 0 {
		return nil, errors.New("banana returned no video urls")
	}
	return outputs, nil
}
