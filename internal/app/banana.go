package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type BananaClient struct {
	cfg        Config
	httpClient *http.Client
}

type BananaSubmitRequest struct {
	Prompt       string   `json:"prompt"`
	ImageURLs    []string `json:"imageUrls,omitempty"`
	AspectRatio  string   `json:"aspectRatio,omitempty"`
	Resolution   string   `json:"resolution"`
	WebhookURL   string   `json:"webhookUrl,omitempty"`
	ClientTaskID string   `json:"clientTaskId,omitempty"`
}

type BananaTask struct {
	TaskID       string         `json:"taskId"`
	Status       string         `json:"status"`
	ErrorCode    string         `json:"errorCode"`
	ErrorMessage string         `json:"errorMessage"`
	FailedReason map[string]any `json:"failedReason"`
	Results      []BananaResult `json:"results"`
	ClientID      string         `json:"clientId"`
	PromptTips   string         `json:"promptTips"`
}

type BananaResult struct {
	URL        string `json:"url"`
	OutputType string `json:"outputType"`
	Text       string `json:"text"`
}

func NewBananaClient(cfg Config) *BananaClient {
	return &BananaClient{cfg: cfg, httpClient: &http.Client{Timeout: cfg.BananaHTTPTimeout}}
}

func (c *BananaClient) SubmitTextToImage(ctx context.Context, req BananaSubmitRequest) (BananaTask, error) {
	return c.postTask(ctx, "/v1/banana_pro/text-to-image", req)
}

func (c *BananaClient) SubmitImageToImage(ctx context.Context, req BananaSubmitRequest) (BananaTask, error) {
	return c.postTask(ctx, "/v1/banana_pro/image-to-image", req)
}

func (c *BananaClient) Wait(ctx context.Context, taskID string) (BananaTask, error) {
	ticker := time.NewTicker(c.cfg.PollInterval)
	defer ticker.Stop()
	for {
		task, err := c.Query(ctx, taskID)
		if err != nil {
			return BananaTask{}, err
		}
		switch strings.ToUpper(task.Status) {
		case "SUCCESS":
			if len(task.Results) == 0 {
				return BananaTask{}, errors.New("banana task succeeded without results")
			}
			return task, nil
		case "FAILED", "TIMEOUT", "CANCELED":
			return BananaTask{}, fmt.Errorf("banana task %s: %s", task.Status, bananaError(task))
		}
		select {
		case <-ctx.Done():
			return BananaTask{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (c *BananaClient) Query(ctx context.Context, taskID string) (BananaTask, error) {
	var task BananaTask
	err := c.postJSON(ctx, "/v1/query", map[string]string{"taskId": taskID}, &task)
	return task, err
}

func (c *BananaClient) Upload(ctx context.Context, fileName string, body io.Reader) (string, error) {
	var payload bytes.Buffer
	writer := multipart.NewWriter(&payload)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, body); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BananaBaseURL+"/v1/media/upload/binary", &payload)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+c.cfg.BananaAPIKey)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("upload failed: %s", readBodyMessage(responseBody))
	}
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			DownloadURL string `json:"download_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return "", err
	}
	if result.Code != 0 || result.Data.DownloadURL == "" {
		return "", fmt.Errorf("upload failed: %s", result.Message)
	}
	return result.Data.DownloadURL, nil
}

func (c *BananaClient) postTask(ctx context.Context, path string, req BananaSubmitRequest) (BananaTask, error) {
	var task BananaTask
	err := c.postJSON(ctx, path, req, &task)
	if err != nil {
		return BananaTask{}, err
	}
	if task.TaskID == "" {
		return BananaTask{}, errors.New("banana response missing taskId")
	}
	return task, nil
}

func (c *BananaClient) postJSON(ctx context.Context, path string, input any, output any) error {
	body, _ := json.Marshal(input)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BananaBaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.cfg.BananaAPIKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("banana http %d: %s", response.StatusCode, readBodyMessage(responseBody))
	}
	if err := json.Unmarshal(responseBody, output); err != nil {
		return err
	}
	return nil
}

func bananaError(task BananaTask) string {
	if task.ErrorMessage != "" {
		return task.ErrorMessage
	}
	if len(task.FailedReason) > 0 {
		body, _ := json.Marshal(task.FailedReason)
		return string(body)
	}
	if task.ErrorCode != "" {
		return task.ErrorCode
	}
	return "task failed"
}

func readBodyMessage(body []byte) string {
	var payload map[string]any
	if json.Unmarshal(body, &payload) == nil {
		for _, key := range []string{"message", "msg", "errorMessage"} {
			if value, ok := payload[key].(string); ok && value != "" {
				return value
			}
		}
		if value, ok := payload["error"].(map[string]any); ok {
			if message, ok := value["message"].(string); ok && message != "" {
				return message
			}
		}
	}
	return strings.TrimSpace(string(body))
}
