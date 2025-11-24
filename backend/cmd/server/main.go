package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-knowledge-app/internal/api"
	"ai-knowledge-app/internal/config"
	"ai-knowledge-app/pkg/database"
	"ai-knowledge-app/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title AI知识库查询API
// @version 1.0
// @description 一个基于Go和Gin的AI知识库查询应用后端API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	if err := logger.InitLogger(&cfg.Log); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger.GetLogger().Info("Starting AI Knowledge Application...")

	// 初始化数据库
	if err := database.InitDatabase(&cfg.Database); err != nil {
		logger.GetLogger().WithField("error", err).Fatal("Failed to initialize database")
	}

	// 自动迁移数据库
	if err := database.AutoMigrate(); err != nil {
		logger.GetLogger().WithField("error", err).Fatal("Failed to migrate database")
	}

	// 创建种子数据
	if err := database.SeedDatabase(); err != nil {
		logger.GetLogger().WithField("error", err).Fatal("Failed to seed database")
	}

	// 创建路由器
	router := api.NewRouter(cfg)
	engine := router.SetupRoutes()

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: engine,
	}

	// 设置服务器配置
	if gin.Mode() == gin.ReleaseMode {
		server.ReadTimeout = 10 * time.Second
		server.WriteTimeout = 10 * time.Second
		server.IdleTimeout = 60 * time.Second
	}

	// 启动服务器的goroutine
	go func() {
		logger.GetLogger().Infof("Server starting on %s", server.Addr)

		var err error
		if cfg.Server.Host == "localhost" || cfg.Server.Host == "127.0.0.1" {
			// 开发环境使用HTTP
			err = server.ListenAndServe()
		} else {
			// 生产环境可以考虑HTTPS（需要配置证书）
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.GetLogger().WithField("error", err).Fatal("Server failed to start")
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.GetLogger().Info("Shutting down server...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.GetLogger().WithField("error", err).Error("Server forced to shutdown")
	}

	// 关闭数据库连接
	if err := database.CloseDatabase(); err != nil {
		logger.GetLogger().WithField("error", err).Error("Failed to close database")
	}

	logger.GetLogger().Info("Server exited")
}