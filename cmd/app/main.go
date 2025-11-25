// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Main Application Entry Point"
//   Timestamp: "2025-11-25T13:51:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python main entry from web.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S, Clean Architecture"
//   Quality_Check: "Graceful shutdown, signal handling, component initialization"
// }}

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imhuimie/let-monitor-go/internal/config"
	"github.com/imhuimie/let-monitor-go/internal/database"
	"github.com/imhuimie/let-monitor-go/internal/monitor"
	"github.com/imhuimie/let-monitor-go/internal/server"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Setup logging
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(log.InfoLevel)

	log.Info("启动 Let-Monitor-Go...")

	// Load .env file first (before reading any environment variables)
	if err := godotenv.Load("data/.env"); err != nil {
		log.Warnf("无法加载 data/.env 文件: %v (将使用系统环境变量或默认值)", err)
	}

	// Load environment variables
	dbType := config.GetEnv("DB_TYPE", "sqlite")
	mongoHost := config.GetEnv("MONGO_HOST", "mongodb://localhost:27017/")
	sqlitePath := config.GetEnv("SQLITE_PATH", "data/forum_monitor.db")
	accessToken := config.GetEnv("ACCESS_TOKEN", "default_token")
	port := config.GetEnv("PORT", "5556")

	// Initialize configuration manager
	cfgMgr := config.NewManager("data/config.json")
	if err := cfgMgr.Load(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// Connect to database based on type
	var connectionString string
	switch dbType {
	case "sqlite":
		connectionString = sqlitePath
		log.Infof("使用 SQLite 数据库: %s", sqlitePath)
	case "mongodb":
		connectionString = mongoHost
		log.Infof("使用 MongoDB 数据库: %s", mongoHost)
	default:
		log.Fatalf("不支持的数据库类型: %s", dbType)
	}

	db, err := database.NewDatabase(database.DatabaseType(dbType), connectionString)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Disconnect()

	// Create forum monitor
	mon, err := monitor.NewForumMonitor(cfgMgr, db)
	if err != nil {
		log.Fatalf("创建监控器失败: %v", err)
	}

	// Start monitoring in background
	mon.Start()
	defer mon.Stop()

	// Create and start web server
	srv := server.NewServer(cfgMgr, mon, accessToken, port)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("收到关闭信号，正在优雅关闭...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("服务器关闭错误: %v", err)
	}

	log.Info("应用已关闭")
}
