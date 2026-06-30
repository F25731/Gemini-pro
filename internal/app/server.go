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
	cfg    Config
	client *BananaClient
	pool   *WorkerPool
}

func NewServer(cfg Config) *Server {
	client := NewBananaClient(cfg)
	return &Server{cfg: cfg, client: client, pool: NewWorkerPool(cfg.MaxWorkers, cfg.MaxQueue)}
}

func (s *Server) Router() http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(cors())
	router.GET("/health", s.health)
	router.GET("/v1/models", s.openAIAuth(), s.models)
	router.POST("/v1/images/generations", s.openAIAuth(), s.imageGeneration)
	router.POST("/v1/images/edits", s.openAIAuth(), s.imageEdit)

	admin := router.Group("/api/admin", s.adminAuth())
	admin.GET("/status", s.adminStatus)
	admin.GET("/config", s.adminConfig)

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
	router.GET("/admin", func(c *gin.Context) { c.FileFromFS("index.html", http.FS(dist)) })
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
	c.JSON(http.StatusOK, gin.H{
		"publicBaseUrl":     s.cfg.PublicBaseURL,
		"bananaBaseUrl":     s.cfg.BananaBaseURL,
		"modelPrefix":       s.cfg.ModelPrefix,
		"maxWorkers":        s.cfg.MaxWorkers,
		"maxQueue":          s.cfg.MaxQueue,
		"pollIntervalMs":    s.cfg.PollInterval.Milliseconds(),
		"requestTimeoutSec": int(s.cfg.RequestTimeout.Seconds()),
		"returnB64JSON":     s.cfg.ReturnB64JSON,
		"models":            s.cfg.Models(),
	})
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
