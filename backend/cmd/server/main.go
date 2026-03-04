// cmd/server/main.go
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/config"
	"github.com/JoyMod/ManboTV/backend/internal/handler"
	"github.com/JoyMod/ManboTV/backend/internal/middleware"
	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

// 硬编码的 API 站点配置 (后续应从配置文件读取)
var defaultSites = []model.ApiSite{
	{Key: "example1", API: "https://api1.example.com/api.php", Name: "示例源1"},
	{Key: "example2", API: "https://api2.example.com/api.php", Name: "示例源2"},
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger, err := config.InitLogger(&cfg.Log)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer config.Sync()

	logger.Info("服务启动中", zap.String("mode", cfg.Server.Mode))

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(&cfg.CORS))
	r.Use(middleware.Logger(logger))

	// 初始化服务
	searchService := service.NewSearchService(&cfg.Search, &cfg.HTTPClient, logger)

	// 初始化处理器
	searchHandler := handler.NewSearchHandler(searchService, logger, defaultSites)

	// 注册路由
	apiV1 := r.Group("/api/v1")
	{
		// 搜索相关
		apiV1.GET("/search", searchHandler.Search)
		apiV1.GET("/search/one", searchHandler.SearchSingle)
		apiV1.GET("/search/sites", searchHandler.GetSites)

		// 健康检查
		apiV1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
			})
		})
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	logger.Info("服务已启动", zap.String("addr", addr))

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("服务正在关闭...")

	// 优雅关闭
	if err := srv.Close(); err != nil {
		logger.Error("服务关闭失败", zap.Error(err))
	}

	logger.Info("服务已退出")
}
