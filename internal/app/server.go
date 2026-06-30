package app

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var webFS embed.FS

type Server struct {
	cfg     Config
	runtime *RuntimeStore
	client  *BananaClient
	pool    *WorkerPool
}

func NewServer(cfg Config) *Server {
	runtime := NewRuntimeStore(cfg.RuntimeConfigPath, RuntimeConfig{BananaAPIKey: cfg.BananaAPIKey})
	client := NewBananaClient(cfg, runtime)
	return &Server{cfg: cfg, runtime: runtime, client: client, pool: NewWorkerPool(cfg.MaxWorkers, cfg.MaxQueue)}
}

func (s *Server) Router() http.Handler {
	router := gin.New()
	router.RedirectTrailingSlash = false
	router.Use(gin.Recovery())
	router.Use(cors())
	router.GET("/health", s.health)
	router.GET("/v1/models", s.openAIAuth(), s.models)
	router.POST("/v1/images/generations", s.openAIAuth(), s.imageGeneration)
	router.POST("/v1/images/edits", s.openAIAuth(), s.imageEdit)
	router.POST("/api/admin/login", s.adminLogin)

	admin := router.Group("/api/admin", s.adminAuth())
	admin.GET("/status", s.adminStatus)
	admin.GET("/config", s.adminConfig)
	admin.POST("/config", s.adminSaveConfig)

	s.mountAdmin(router)
	return router
}

func (s *Server) mountAdmin(router *gin.Engine) {
	dist, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		return
	}
	assets, err := fs.Sub(webFS, "web/dist/assets")
	if err == nil {
		router.StaticFS("/admin/assets", http.FS(assets))
	router.StaticFS("/assets", http.FS(assets))
	}
	router.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/admin") })
	adminPage := func(c *gin.Context) {
		body, err := webFS.ReadFile("web/dist/index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"message": "admin page not found"}})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", body)
	}
	router.GET("/admin", adminPage)
	router.GET("/admin/", adminPage)
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/admin") {
			c.FileFromFS("index.html", http.FS(dist))
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"message": "not found"}})
	})
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "time": time.Now().Unix()})
}

func (s *Server) models(c *gin.Context) {
	data := make([]gin.H, 0, len(s.cfg.Models()))
	for _, model := range s.cfg.Models() {
		data = append(data, gin.H{"id": model, "object": "model", "created": 0, "owned_by": "banana-pro-wrapper"})
	}
	c.JSON(http.StatusOK, gin.H{"object": "list", "data": data})
}

func (s *Server) adminStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"pool":   s.pool.Stats(),
		"models": s.cfg.Models(),
	})
}

func (s *Server) adminConfig(c *gin.Context) {
	runtime := s.runtime.Get()
	c.JSON(http.StatusOK, gin.H{
		"publicBaseUrl":     s.cfg.PublicBaseURL,
		"bananaBaseUrl":     s.cfg.BananaBaseURL,
		"bananaApiKeySet":   runtime.BananaAPIKey != "",
		"bananaApiKeyHint":  maskKey(runtime.BananaAPIKey),
		"modelPrefix":       s.cfg.ModelPrefix,
		"maxWorkers":        s.cfg.MaxWorkers,
		"maxQueue":          s.cfg.MaxQueue,
		"pollIntervalMs":    s.cfg.PollInterval.Milliseconds(),
		"requestTimeoutSec": int(s.cfg.RequestTimeout.Seconds()),
		"returnB64JSON":     s.cfg.ReturnB64JSON,
		"models":            s.cfg.Models(),
	})
}

func (s *Server) adminSaveConfig(c *gin.Context) {
	var input struct {
		BananaAPIKey string `json:"bananaApiKey"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	if strings.TrimSpace(input.BananaAPIKey) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "banana api key is required"})
		return
	}
	if err := s.runtime.Save(RuntimeConfig{BananaAPIKey: input.BananaAPIKey}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "bananaApiKeyHint": maskKey(input.BananaAPIKey)})
}

func maskKey(key string) string {
	key = strings.TrimSpace(key)
	if len(key) <= 12 {
		if key == "" {
			return ""
		}
		return "set"
	}
	return key[:8] + "..." + key[len(key)-4:]
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
