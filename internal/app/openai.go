package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ImageRequest struct {
	Model          string   `json:"model"`
	Prompt         string   `json:"prompt"`
	Size           string   `json:"size"`
	Quality        string   `json:"quality"`
	AspectRatio    string   `json:"aspectRatio"`
	Resolution     string   `json:"resolution"`
	N              int      `json:"n"`
	ResponseFormat string   `json:"response_format"`
	ImageURL       string   `json:"image_url"`
	ImageURLs      []string `json:"imageUrls"`
	WebhookURL     string   `json:"webhookUrl"`
	ClientTaskID   string   `json:"clientTaskId"`
}

type ImageOutput struct {
	URL     string `json:"url,omitempty"`
	B64JSON string `json:"b64_json,omitempty"`
}

type completionProbeRequest struct {
	Model string `json:"model"`
}

func (s *Server) chatCompletions(c *gin.Context) {
	var req completionProbeRequest
	_ = c.ShouldBindJSON(&req)
	if _, err := modelSpecByID(req.Model); err != nil {
		openAIError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      "chatcmpl-banana-probe",
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   req.Model,
		"choices": []gin.H{{
			"index": 0,
			"message": gin.H{
				"role":    "assistant",
				"content": "banana pro image wrapper is reachable",
			},
			"finish_reason": "stop",
		}},
		"usage": gin.H{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
	})
}

func (s *Server) completions(c *gin.Context) {
	var req completionProbeRequest
	_ = c.ShouldBindJSON(&req)
	if _, err := modelSpecByID(req.Model); err != nil {
		openAIError(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":      "cmpl-banana-probe",
		"object":  "text_completion",
		"created": time.Now().Unix(),
		"model":   req.Model,
		"choices": []gin.H{{
			"index":         0,
			"text":          "banana pro image wrapper is reachable",
			"finish_reason": "stop",
		}},
		"usage": gin.H{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
	})
}

func (s *Server) imageGeneration(c *gin.Context) {
	var req ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		openAIError(c, http.StatusBadRequest, "invalid json body")
		return
	}
	if len(req.ImageURLs) > 0 || req.ImageURL != "" {
		imageURLs := append([]string{}, req.ImageURLs...)
		if req.ImageURL != "" {
			imageURLs = append(imageURLs, req.ImageURL)
		}
		s.runImageTask(c, req, imageURLs, true)
		return
	}
	s.runImageTask(c, req, nil, false)
}

func (s *Server) imageEdit(c *gin.Context) {
	req, imageURLs, err := s.parseEditRequest(c)
	if err != nil {
		openAIError(c, http.StatusBadRequest, err.Error())
		return
	}
	s.runImageTask(c, req, imageURLs, true)
}

func (s *Server) runImageTask(c *gin.Context, req ImageRequest, imageURLs []string, isEdit bool) {
	spec, err := modelSpecByID(req.Model)
	if err != nil {
		openAIError(c, http.StatusBadRequest, err.Error())
		return
	}
	if spec.Media != MediaImage {
		openAIError(c, http.StatusBadRequest, fmt.Sprintf("model %q is not an image model", req.Model))
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		openAIError(c, http.StatusBadRequest, "prompt is required")
		return
	}
	count := req.N
	if count <= 0 {
		count = 1
	}
	if count > 8 {
		count = 8
	}

	heartbeat := startJSONHeartbeat(c, s.cfg.HeartbeatInterval)
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.RequestTimeout)
	defer cancel()

	started := time.Now()
	ok := false
	defer func() {
		s.metrics.Record("image", ok, time.Since(started))
		s.metrics.Record(spec.Family, ok, time.Since(started))
	}()

	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, count)
	outputs := make([]ImageOutput, 0, count)
	for i := 0; i < count; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.pool.Run(ctx, func(ctx context.Context) error {
				items, err := s.runOneImageTask(ctx, req, imageURLs, spec, isEdit, i)
				if err != nil {
					return err
				}
				mu.Lock()
				outputs = append(outputs, items...)
				mu.Unlock()
				return nil
			})
			if err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	if err := <-errCh; err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, ErrQueueFull) {
			status = http.StatusTooManyRequests
		}
		respondJSON(c, heartbeat, status, openAIErrorPayload(err.Error()))
		return
	}
	ok = true
	respondJSON(c, heartbeat, http.StatusOK, gin.H{"created": time.Now().Unix(), "data": outputs})
}

func (s *Server) runOneImageTask(ctx context.Context, req ImageRequest, imageURLs []string, spec ModelSpec, isEdit bool, index int) ([]ImageOutput, error) {
	clientTaskID := req.ClientTaskID
	if clientTaskID == "" {
		clientTaskID = fmt.Sprintf("banana-%d-%d", time.Now().UnixNano(), index)
	}
	taskReq := BananaSubmitRequest{
		Prompt:       req.Prompt,
		ImageURLs:    imageURLs,
		AspectRatio:  normalizeImageAspectRatio(firstNonEmpty(req.AspectRatio, req.Size)),
		Resolution:   spec.Resolution,
		WebhookURL:   req.WebhookURL,
		ClientTaskID: clientTaskID,
	}
	var task BananaTask
	var err error
	if isEdit {
		if len(taskReq.ImageURLs) == 0 {
			return nil, errors.New("image is required for image edit")
		}
		task, err = s.client.Submit(ctx, spec.ImageEndpoint, taskReq)
	} else {
		task, err = s.client.Submit(ctx, spec.TextEndpoint, taskReq)
	}
	if err != nil {
		return nil, err
	}
	waitCtx, cancel := context.WithTimeout(context.Background(), s.cfg.RequestTimeout)
	defer cancel()
	done, err := s.client.Wait(waitCtx, task.TaskID)
	if err != nil {
		return nil, err
	}
	return s.outputsFromResults(waitCtx, done.Results, req.ResponseFormat)
}

func (s *Server) parseEditRequest(c *gin.Context) (ImageRequest, []string, error) {
	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		var req ImageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			return req, nil, err
		}
		urls := append([]string{}, req.ImageURLs...)
		if req.ImageURL != "" {
			urls = append(urls, req.ImageURL)
		}
		return req, urls, nil
	}

	_, params, err := mime.ParseMediaType(contentType)
	if err != nil || !strings.HasPrefix(contentType, "multipart/form-data") {
		return ImageRequest{}, nil, errors.New("multipart/form-data or json body is required")
	}
	reader := multipart.NewReader(c.Request.Body, params["boundary"])
	form, err := reader.ReadForm(64 << 20)
	if err != nil {
		return ImageRequest{}, nil, err
	}
	defer form.RemoveAll()
	req := ImageRequest{
		Model:          formValue(form, "model"),
		Prompt:         formValue(form, "prompt"),
		Size:           formValue(form, "size"),
		Quality:        formValue(form, "quality"),
		ResponseFormat: formValue(form, "response_format"),
		WebhookURL:     formValue(form, "webhookUrl"),
		ClientTaskID:   formValue(form, "clientTaskId"),
		AspectRatio:    formValue(form, "aspectRatio"),
		Resolution:     formValue(form, "resolution"),
	}
	if n, _ := strconv.Atoi(formValue(form, "n")); n > 0 {
		req.N = n
	}
	var urls []string
	for _, key := range []string{"image_url", "imageUrls"} {
		for _, value := range form.Value[key] {
			if strings.TrimSpace(value) != "" {
				urls = append(urls, strings.TrimSpace(value))
			}
		}
	}
	for _, files := range form.File {
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				return req, nil, err
			}
			url, err := s.client.Upload(c.Request.Context(), safeFileName(fileHeader.Filename), file)
			_ = file.Close()
			if err != nil {
				return req, nil, err
			}
			urls = append(urls, url)
		}
	}
	return req, urls, nil
}

func (s *Server) outputsFromResults(ctx context.Context, results []BananaResult, format string) ([]ImageOutput, error) {
	outputs := make([]ImageOutput, 0, len(results))
	wantB64 := strings.EqualFold(format, "b64_json") || s.cfg.ReturnB64JSON
	for _, result := range results {
		url := result.ImageURLValue()
		if url == "" {
			continue
		}
		if !wantB64 {
			outputs = append(outputs, ImageOutput{URL: url})
			continue
		}
		b64, err := downloadB64(ctx, s.client.httpClient, url)
		if err != nil {
			outputs = append(outputs, ImageOutput{URL: url})
			continue
		}
		outputs = append(outputs, ImageOutput{B64JSON: b64})
	}
	if len(outputs) == 0 {
		return nil, errors.New("banana returned no image urls")
	}
	return outputs, nil
}

func (result BananaResult) ImageURLValue() string {
	for _, value := range []string{result.ImageURL, result.ImageURLAlt, result.URL, result.DownloadURL} {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func normalizeImageAspectRatio(size string) string {
	value := strings.TrimSpace(size)
	if value == "" || strings.EqualFold(value, "auto") {
		return ""
	}
	allowed := map[string]bool{"1:1": true, "16:9": true, "9:16": true, "4:3": true, "3:4": true, "3:2": true, "2:3": true, "5:4": true, "4:5": true, "21:9": true}
	if allowed[value] {
		return value
	}
	parts := strings.Split(strings.ToLower(value), "x")
	if len(parts) != 2 {
		return ""
	}
	w, _ := strconv.Atoi(parts[0])
	h, _ := strconv.Atoi(parts[1])
	if w <= 0 || h <= 0 {
		return ""
	}
	ratio := float64(w) / float64(h)
	candidates := []struct {
		label string
		value float64
	}{
		{"1:1", 1}, {"16:9", 16.0 / 9}, {"9:16", 9.0 / 16}, {"4:3", 4.0 / 3}, {"3:4", 3.0 / 4}, {"3:2", 1.5}, {"2:3", 2.0 / 3}, {"5:4", 1.25}, {"4:5", 0.8}, {"21:9", 21.0 / 9},
	}
	best := candidates[0]
	bestDiff := abs(ratio - best.value)
	for _, item := range candidates[1:] {
		if diff := abs(ratio - item.value); diff < bestDiff {
			best = item
			bestDiff = diff
		}
	}
	return best.label
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func downloadB64(ctx context.Context, client *http.Client, url string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("download result failed: %d", response.StatusCode)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, io.LimitReader(response.Body, 50<<20)); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func formValue(form *multipart.Form, key string) string {
	if values := form.Value[key]; len(values) > 0 {
		return strings.TrimSpace(values[0])
	}
	return ""
}

func safeFileName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == "" {
		return "image.png"
	}
	return base
}

func openAIError(c *gin.Context, status int, message string) {
	c.JSON(status, openAIErrorPayload(message))
}

func openAIErrorPayload(message string) gin.H {
	return gin.H{"error": gin.H{"message": message, "type": "invalid_request_error", "code": "upstream_failed"}}
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
