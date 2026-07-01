package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) queryTask(c *gin.Context) {
	var req struct {
		TaskID string `json:"taskId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TaskID == "" {
		openAIError(c, http.StatusBadRequest, "taskId is required")
		return
	}
	task, err := s.client.Query(c.Request.Context(), req.TaskID)
	if err != nil {
		openAIError(c, http.StatusBadGateway, err.Error())
		return
	}
	data := make([]gin.H, 0, len(task.Results))
	for _, result := range task.Results {
		if url := result.ResultURLValue(); url != "" {
			item := gin.H{"url": url}
			if result.OutputType != "" {
				item["outputType"] = result.OutputType
			}
			if result.Text != "" {
				item["text"] = result.Text
			}
			data = append(data, item)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"id":           task.TaskID,
		"object":       "generation.task",
		"created":      time.Now().Unix(),
		"taskId":       task.TaskID,
		"status":       task.Status,
		"errorCode":    task.ErrorCode,
		"errorMessage": task.ErrorMessage,
		"failedReason": task.FailedReason,
		"results":      task.Results,
		"data":         data,
		"clientId":     task.ClientID,
		"promptTips":   task.PromptTips,
	})
}
