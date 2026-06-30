package app

import (
	"crypto/subtle"
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
		if s.cfg.AdminToken == "" && (s.cfg.AdminUsername == "" || s.cfg.AdminPassword == "") {
			c.Next()
			return
		}
		if cookie, err := c.Cookie("banana_admin_session"); err == nil && cookie != "" && s.cfg.AdminToken != "" && constantEqual(cookie, s.cfg.AdminToken) {
			c.Next()
			return
		}
		token := bearer(c.GetHeader("Authorization"))
		if token == "" {
			token = strings.TrimSpace(c.Query("token"))
		}
		if s.cfg.AdminToken == "" || !constantEqual(token, s.cfg.AdminToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid admin token"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Server) adminLogin(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	if s.cfg.AdminUsername == "" || s.cfg.AdminPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "admin username/password not configured"})
		return
	}
	if !constantEqual(input.Username, s.cfg.AdminUsername) || !constantEqual(input.Password, s.cfg.AdminPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid username or password"})
		return
	}
	token := s.cfg.AdminToken
	if token == "" {
		token = s.cfg.AdminUsername + ":" + s.cfg.AdminPassword
	}
	c.SetCookie("banana_admin_session", token, 86400*30, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func bearer(header string) string {
	value := strings.TrimSpace(header)
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return value
}

func constantEqual(a string, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
