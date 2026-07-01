package app

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type jsonHeartbeat struct {
	c       *gin.Context
	stop    chan struct{}
	done    chan struct{}
	once    sync.Once
	started atomic.Bool
}

func startJSONHeartbeat(c *gin.Context, interval time.Duration) *jsonHeartbeat {
	if interval <= 0 {
		return nil
	}
	h := &jsonHeartbeat{
		c:    c,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go h.run(interval)
	return h
}

func (h *jsonHeartbeat) run(interval time.Duration) {
	defer close(h.done)
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-h.stop:
			return
		case <-timer.C:
			if !h.writeBlank() {
				return
			}
			timer.Reset(interval)
		}
	}
}

func (h *jsonHeartbeat) writeBlank() bool {
	if !h.started.Load() {
		header := h.c.Writer.Header()
		header.Set("Content-Type", "application/json; charset=utf-8")
		header.Set("Cache-Control", "no-cache")
		header.Set("X-Accel-Buffering", "no")
		h.c.Status(http.StatusOK)
		h.started.Store(true)
	}
	if _, err := h.c.Writer.Write([]byte("\n")); err != nil {
		return false
	}
	if flusher, ok := h.c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
	return true
}

func (h *jsonHeartbeat) stopAndStarted() bool {
	if h == nil {
		return false
	}
	h.once.Do(func() {
		close(h.stop)
	})
	<-h.done
	return h.started.Load()
}

func respondJSON(c *gin.Context, h *jsonHeartbeat, status int, payload any) {
	if h == nil || !h.stopAndStarted() {
		c.JSON(status, payload)
		return
	}
	body, err := json.Marshal(payload)
	if err != nil {
		body, _ = json.Marshal(openAIErrorPayload("failed to encode response"))
	}
	_, _ = c.Writer.Write(body)
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

