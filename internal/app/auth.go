package app

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) openAIAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.cfg.WrapperAPIKey == "" {
			c.Next()
			return
		}
		if bearer(c.GetHeader("Authorization")) != s.cfg.WrapperAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"message": "invalid api key", "type": "invalid_request_error"}})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) adminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.cfg.AdminToken == "" {
			c.Next()
			return
		}
		token := bearer(c.GetHeader("Authorization"))
		if token == "" {
			token = strings.TrimSpace(c.Query("token"))
		}
		if token != s.cfg.AdminToken {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid admin token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func bearer(header string) string {
	value := strings.TrimSpace(header)
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return value
}
