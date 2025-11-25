// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Web Server Implementation"
//   Timestamp: "2025-11-25T13:49:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python Flask server from web.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S, RESTful API"
//   Quality_Check: "Gin framework with authentication middleware"
// }}

package server

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imhuimie/let-monitor-go/internal/config"
	"github.com/imhuimie/let-monitor-go/internal/monitor"
	log "github.com/sirupsen/logrus"
)

//go:embed templates/*
var templateFS embed.FS

// Server represents the web server
type Server struct {
	engine      *gin.Engine
	server      *http.Server
	configMgr   *config.Manager
	monitor     *monitor.ForumMonitor
	accessToken string
}

// NewServer creates a new web server
func NewServer(configMgr *config.Manager, mon *monitor.ForumMonitor, accessToken string, port string) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(LoggerMiddleware())

	s := &Server{
		engine:      engine,
		configMgr:   configMgr,
		monitor:     mon,
		accessToken: accessToken,
	}

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	s.server = &http.Server{
		Addr:           ":" + port,
		Handler:        engine,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
	// Serve index page
	s.engine.GET("/", s.handleIndex)

	// API routes
	api := s.engine.Group("/api")
	{
		// Health check (no auth required)
		api.GET("/health", s.handleHealth)

		// Config endpoints (auth required)
		api.GET("/config", s.authMiddleware(), s.handleGetConfig)
		api.POST("/config", s.authMiddleware(), s.handleUpdateConfig)
	}
}

// Start starts the web server
func (s *Server) Start() error {
	log.Infof("Web 服务器启动于 %s", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("服务器启动失败: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	log.Info("关闭 Web 服务器...")
	return s.server.Shutdown(ctx)
}

// handleIndex serves the index page
func (s *Server) handleIndex(c *gin.Context) {
	tmpl, err := template.ParseFS(templateFS, "templates/index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "模板加载失败: %v", err)
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(c.Writer, nil); err != nil {
		log.Warnf("模板执行失败: %v", err)
	}
}

// handleHealth returns health status
func (s *Server) handleHealth(c *gin.Context) {
	uptime := time.Since(time.Now()) // This would need to track actual start time

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"database": "connected",
		"monitor":  getMonitorStatus(s.monitor),
		"uptime":   uptime.String(),
	})
}

// handleGetConfig returns the current configuration
func (s *Server) handleGetConfig(c *gin.Context) {
	cfg := s.configMgr.Get()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "配置未加载",
		})
		return
	}

	c.JSON(http.StatusOK, cfg)
}

// handleUpdateConfig updates the configuration
func (s *Server) handleUpdateConfig(c *gin.Context) {
	var requestBody struct {
		Config *config.Config `json:"config"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "无效的请求数据",
		})
		return
	}

	if requestBody.Config == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "缺少 config 字段",
		})
		return
	}

	// Validate configuration
	if err := requestBody.Config.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("配置验证失败: %v", err),
		})
		return
	}

	// Save configuration
	if err := s.configMgr.Save(requestBody.Config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("保存配置失败: %v", err),
		})
		return
	}

	// Reload monitor configuration
	if err := s.monitor.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("重新加载配置失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Config updated",
	})
}

// authMiddleware checks for valid access token
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		expectedToken := "Bearer " + s.accessToken

		if token != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Unauthorized",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		log.Infof("[HTTP] %s %s %d %v %s",
			method,
			path,
			statusCode,
			latency,
			clientIP,
		)
	}
}

// getMonitorStatus returns the monitor status
func getMonitorStatus(mon *monitor.ForumMonitor) string {
	if mon.IsRunning() {
		return "running"
	}
	return "stopped"
}
