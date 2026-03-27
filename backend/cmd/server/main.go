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
	"github.com/JoyMod/ManboTV/backend/internal/repository/redis"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

// 默认API站点配置
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

	logger.Info("服务启动中",
		zap.String("mode", cfg.Server.Mode),
		zap.String("version", "1.0.0"),
	)

	// 连接Redis
	redisClient, err := redis.NewClient(&cfg.Redis, logger)
	if err != nil {
		logger.Error("连接Redis失败", zap.Error(err))
	} else {
		defer redisClient.Close()
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(&cfg.CORS))
	r.Use(middleware.Logger(logger))

	// 初始化服务
	searchService := service.NewSearchService(&cfg.Search, &cfg.HTTPClient, logger)
	imageService, err := service.NewImageService(&cfg.ImageProxy, &cfg.HTTPClient, logger)
	if err != nil {
		logger.Fatal("初始化图片服务失败", zap.Error(err))
	}
	detailService := service.NewDetailService(&cfg.Search, &cfg.HTTPClient, logger)
	suggestionService := service.NewSuggestionService(logger)
	m3u8Service := service.NewM3U8ProxyService(&cfg.HTTPClient, logger)
	segmentService := service.NewSegmentProxyService(&cfg.HTTPClient, logger)

	var storageService model.StorageService
	var adminStorageService model.AdminStorageService
	if redisClient != nil {
		storageService = service.NewStorageService(redisClient, logger)
		adminStorageService = service.NewAdminStorageService(redisClient, logger)
		logger.Info("使用Redis存储")
	} else {
		logger.Fatal("Admin功能需要Redis存储，请配置Redis")
	}

	liveService := service.NewLiveService(adminStorageService, logger)

	// 从环境变量获取管理员账号
	ownerUser := os.Getenv("USERNAME")
	ownerPass := os.Getenv("PASSWORD")
	if ownerUser == "" {
		ownerUser = "admin"
		logger.Warn("未设置 USERNAME 环境变量，使用默认值: admin")
	}
	if ownerPass == "" {
		ownerPass = "admin"
		logger.Warn("未设置 PASSWORD 环境变量，使用默认值")
	}

	// 初始化处理器
	searchHandler := handler.NewSearchHandler(searchService, logger, defaultSites, adminStorageService, ownerUser)
	searchBootstrapHandler := handler.NewSearchBootstrapHandler(
		searchService,
		suggestionService,
		storageService,
		logger,
		defaultSites,
		adminStorageService,
		ownerUser,
	)
	imageHandler := handler.NewImageHandler(imageService, logger)
	favoriteHandler := handler.NewFavoriteHandler(storageService, logger)
	favoritesBootstrapHandler := handler.NewFavoritesBootstrapHandler(storageService, logger)
	recordHandler := handler.NewRecordHandler(storageService, logger)
	detailHandler := handler.NewDetailHandler(detailService, logger, defaultSites, adminStorageService, ownerUser)
	playBootstrapHandler := handler.NewPlayBootstrapHandler(
		detailService,
		searchService,
		storageService,
		logger,
		defaultSites,
		adminStorageService,
		ownerUser,
	)
	proxyHandler := handler.NewProxyHandler(m3u8Service, segmentService, logger)
	authHandler := handler.NewAuthHandler(
		cfg.Auth.CookieName,
		cfg.Auth.TokenExpireHours*time.Hour,
		cfg.Auth.JWTSecret,
		ownerUser,
		ownerPass,
		adminStorageService,
		logger,
	)
	searchHistoryHandler := handler.NewSearchHistoryHandler(storageService, logger)
	skipConfigHandler := handler.NewSkipConfigHandler(storageService, logger)
	doubanHandler := handler.NewDoubanHandler(logger)
	posterHandler := handler.NewPosterHandler(
		searchService,
		logger,
		defaultSites,
		adminStorageService,
		ownerUser,
		redisClient,
	)
	liveHandler := handler.NewLiveHandler(liveService, logger)
	legacySystemHandler := handler.NewLegacySystemHandler(adminStorageService, logger)
	browseBootstrapHandler := handler.NewBrowseBootstrapHandler(logger)

	// Admin Handlers
	adminHandler := handler.NewAdminHandler(
		storageService,
		adminStorageService,
		ownerUser,
		ownerPass,
		logger,
	)
	adminLegacyHandler := handler.NewAdminLegacyHandler(adminHandler, adminStorageService, logger)

	// 认证中间件配置 (用于 /api/v1)
	authMiddleware := middleware.AuthMiddleware(&middleware.AuthConfig{
		CookieName: cfg.Auth.CookieName,
		JWTSecret:  cfg.Auth.JWTSecret,
		OwnerPass:  ownerPass,
		SkipPaths: []string{
			"/api/v1/health",
			"/api/v1/auth/login",
			"/api/v1/search",
			"/api/v1/search/one",
			"/api/v1/search/sites",
			"/api/v1/search/suggestions",
			"/api/v1/detail",
			"/api/v1/details",
			"/api/v1/image",
			"/api/v1/image/header",
			"/api/v1/proxy",
			"/api/v1/douban",
		},
		Logger: logger,
	})

	// 认证中间件配置 (用于旧版 /api)
	authMiddlewareLegacy := middleware.AuthMiddleware(&middleware.AuthConfig{
		CookieName: cfg.Auth.CookieName,
		JWTSecret:  cfg.Auth.JWTSecret,
		OwnerPass:  ownerPass,
		SkipPaths: []string{
			"/api/login",
			"/api/search",
			"/api/search/one",
			"/api/search/suggestions",
			"/api/detail",
			"/api/image-proxy",
			"/api/douban",
			"/api/proxy",
		},
		Logger: logger,
	})

	// ========== 新版 API (/api/v1) ==========
	apiV1 := r.Group("/api/v1")
	{
		// 公开接口
		apiV1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"version":   "1.0.0",
				"timestamp": time.Now().Unix(),
			})
		})

		// 认证相关
		apiV1.POST("/auth/login", authHandler.Login)
		apiV1.POST("/auth/logout", authHandler.Logout)

		// 搜索相关 (公开)
		apiV1.GET("/search", searchHandler.Search)
		apiV1.GET("/search/one", searchHandler.SearchSingle)
		apiV1.GET("/search/sites", searchHandler.GetSites)
		apiV1.GET("/search/suggestions", func(c *gin.Context) {
			query := c.Query("q")
			suggestions, _ := suggestionService.GetSuggestions(c.Request.Context(), query)
			c.JSON(http.StatusOK, model.Success(suggestions))
		})

		// 详情相关 (公开)
		apiV1.GET("/detail", detailHandler.GetDetail)
		apiV1.GET("/details", detailHandler.GetDetails)

		// 图片代理 (公开)
		apiV1.GET("/image", imageHandler.Proxy)
		apiV1.GET("/image/header", imageHandler.ProxyWithHeader)
		apiV1.GET("/poster/recover", posterHandler.Recover)

		// 视频代理 (公开)
		apiV1.GET("/proxy/m3u8", proxyHandler.ProxyM3U8)
		apiV1.GET("/proxy/segment", proxyHandler.ProxySegment)
		apiV1.GET("/proxy/key", proxyHandler.ProxyKey)
		apiV1.GET("/proxy/logo", proxyHandler.ProxyLogo)

		// 豆瓣相关 (公开)
		apiV1.GET("/douban", doubanHandler.Search)
		apiV1.GET("/douban/recommends", doubanHandler.GetRecommends)
		apiV1.GET("/douban/categories", doubanHandler.GetCategories)
		apiV1.GET("/browse/bootstrap", browseBootstrapHandler.GetBootstrap)

		// 需要认证的接口
		authorized := apiV1.Group("/")
		authorized.Use(authMiddleware)
		{
			// 认证相关
			authorized.GET("/auth/me", authHandler.GetCurrentUser)
			authorized.PUT("/auth/password", authHandler.ChangePassword)

			// 收藏相关
			authorized.GET("/favorites", favoriteHandler.GetFavorites)
			authorized.POST("/favorites", favoriteHandler.AddFavorite)
			authorized.DELETE("/favorites/:key", favoriteHandler.DeleteFavorite)
			authorized.GET("/favorites/bootstrap", favoritesBootstrapHandler.GetBootstrap)
			authorized.GET("/play/bootstrap", playBootstrapHandler.GetBootstrap)
			authorized.GET("/search/bootstrap", searchBootstrapHandler.GetBootstrap)

			// 播放记录相关
			authorized.GET("/playrecords", recordHandler.GetRecords)
			authorized.POST("/playrecords", recordHandler.SaveRecord)
			authorized.DELETE("/playrecords/:key", recordHandler.DeleteRecord)

			// 搜索历史相关
			authorized.GET("/searchhistory", searchHistoryHandler.GetHistory)
			authorized.POST("/searchhistory", searchHistoryHandler.AddHistory)
			authorized.DELETE("/searchhistory", searchHistoryHandler.DeleteHistory)

			// 跳过配置相关
			authorized.GET("/skipconfigs", skipConfigHandler.GetConfig)
			authorized.POST("/skipconfigs", skipConfigHandler.SetConfig)
			authorized.DELETE("/skipconfigs", skipConfigHandler.DeleteConfig)

			// 直播相关
			authorized.GET("/live/sources", liveHandler.GetSources)
			authorized.GET("/live/channels", liveHandler.GetChannels)
			authorized.GET("/live/epg", liveHandler.GetEPG)
			authorized.POST("/live/precheck", liveHandler.Precheck)

			// 管理后台 (需要管理员权限)
			admin := authorized.Group("/admin")
			admin.Use(middleware.RequireAdmin())
			{
				// 配置管理
				admin.GET("/config", adminHandler.GetConfig)
				admin.PUT("/config", adminHandler.UpdateConfig)

				// 用户管理
				admin.GET("/users", adminHandler.GetUsers)
				admin.POST("/users", adminHandler.CreateUser)
				admin.PUT("/users/:username", adminHandler.UpdateUser)
				admin.DELETE("/users/:username", adminHandler.DeleteUser)
				admin.PUT("/users/:username/password", adminHandler.ChangeUserPassword)

				// 站点（视频源）管理
				admin.GET("/sites", adminHandler.GetSites)
				admin.PUT("/sites/:key", adminHandler.UpdateSite)
				admin.DELETE("/sites/:key", adminHandler.DeleteSite)

				// 统计数据
				admin.GET("/data-status", adminHandler.GetDataStatus)

				// 数据迁移
				admin.GET("/data/export", adminHandler.ExportData)
				admin.POST("/data/import", adminHandler.ImportData)
			}
		}
	}

	// ========== 旧版 API 兼容层 (/api) ==========
	// 这些路由与前端的调用路径完全兼容
	apiLegacy := r.Group("/api")
	{
		// 公开接口
		apiLegacy.POST("/login", authHandler.Login)
		apiLegacy.POST("/logout", authHandler.Logout)

		// 搜索相关 (公开)
		apiLegacy.GET("/search", searchHandler.SearchLegacy)
		apiLegacy.GET("/search/one", searchHandler.SearchSingleLegacy)
		apiLegacy.GET("/search/resources", searchHandler.SearchResourcesLegacy)
		apiLegacy.GET("/search/ws", searchHandler.SearchStreamLegacy)
		apiLegacy.GET("/search/suggestions", func(c *gin.Context) {
			query := c.Query("q")
			suggestions, _ := suggestionService.GetSuggestions(c.Request.Context(), query)
			items := make([]gin.H, 0, len(suggestions))
			for _, item := range suggestions {
				items = append(items, gin.H{"text": item, "type": "related", "score": 1})
			}
			c.JSON(http.StatusOK, gin.H{"suggestions": items})
		})

		// 详情相关 (公开)
		apiLegacy.GET("/detail", detailHandler.GetDetailLegacy)
		apiLegacy.GET("/details", detailHandler.GetDetails)

		// 图片代理 (公开) - 注意路径不同
		apiLegacy.GET("/image-proxy", imageHandler.Proxy)
		apiLegacy.GET("/image", imageHandler.Proxy)
		apiLegacy.GET("/image/header", imageHandler.ProxyWithHeader)
		apiLegacy.GET("/poster/recover", posterHandler.RecoverLegacy)

		// 视频代理 (公开)
		apiLegacy.GET("/proxy/m3u8", proxyHandler.ProxyM3U8)
		apiLegacy.GET("/proxy/segment", proxyHandler.ProxySegment)
		apiLegacy.GET("/proxy/key", proxyHandler.ProxyKey)
		apiLegacy.GET("/proxy/logo", proxyHandler.ProxyLogo)

		// 豆瓣相关 (公开)
		apiLegacy.GET("/douban", doubanHandler.Search)
		apiLegacy.GET("/douban/recommends", doubanHandler.GetRecommends)
		apiLegacy.GET("/douban/categories", doubanHandler.GetCategories)

		// 其他公开接口
		apiLegacy.GET("/server-config", legacySystemHandler.GetServerConfig)
		apiLegacy.GET("/home", legacySystemHandler.GetHome)
		apiLegacy.GET("/browse/bootstrap", browseBootstrapHandler.GetBootstrapLegacy)
		apiLegacy.GET("/bangumi/calendar", legacySystemHandler.GetBangumiCalendar)
		apiLegacy.GET("/cron", legacySystemHandler.RunCron)

		// 需要认证的接口
		authorizedLegacy := apiLegacy.Group("/")
		authorizedLegacy.Use(authMiddlewareLegacy)
		{
			// 密码修改
			authorizedLegacy.POST("/change-password", authHandler.ChangePassword)

			// 收藏相关
			authorizedLegacy.GET("/favorites", favoriteHandler.GetFavoritesLegacy)
			authorizedLegacy.POST("/favorites", favoriteHandler.AddFavoriteLegacy)
			authorizedLegacy.DELETE("/favorites", favoriteHandler.DeleteFavoriteLegacy)
			authorizedLegacy.GET("/favorites/bootstrap", favoritesBootstrapHandler.GetBootstrapLegacy)
			authorizedLegacy.GET("/search/bootstrap", searchBootstrapHandler.GetBootstrapLegacy)
			authorizedLegacy.GET("/play/bootstrap", playBootstrapHandler.GetBootstrapLegacy)

			// 播放记录相关
			authorizedLegacy.GET("/playrecords", recordHandler.GetRecordsLegacy)
			authorizedLegacy.POST("/playrecords", recordHandler.SaveRecordLegacy)
			authorizedLegacy.DELETE("/playrecords", recordHandler.DeleteRecordLegacy)

			// 搜索历史相关
			authorizedLegacy.GET("/searchhistory", searchHistoryHandler.GetHistoryLegacy)
			authorizedLegacy.POST("/searchhistory", searchHistoryHandler.AddHistoryLegacy)
			authorizedLegacy.DELETE("/searchhistory", searchHistoryHandler.DeleteHistoryLegacy)

			// 跳过配置相关
			authorizedLegacy.GET("/skipconfigs", skipConfigHandler.GetConfigLegacy)
			authorizedLegacy.POST("/skipconfigs", skipConfigHandler.SetConfigLegacy)
			authorizedLegacy.DELETE("/skipconfigs", skipConfigHandler.DeleteConfigLegacy)

			// 直播相关
			authorizedLegacy.GET("/live/sources", liveHandler.GetSourcesLegacy)
			authorizedLegacy.GET("/live/channels", liveHandler.GetChannelsLegacy)
			authorizedLegacy.GET("/live/epg", liveHandler.GetEPGLegacy)
			authorizedLegacy.GET("/live/precheck", liveHandler.PrecheckLegacy)
			authorizedLegacy.POST("/live/precheck", liveHandler.Precheck)

			// 管理后台 (兼容旧版路径)
			adminLegacy := authorizedLegacy.Group("/admin")
			adminLegacy.Use(middleware.RequireAdmin())
			{
				// 管理配置（兼容前端新管理页）
				adminLegacy.GET("/config", adminLegacyHandler.HandleConfigGet)
				adminLegacy.PUT("/config", adminLegacyHandler.HandleConfigPut)

				// 兼容前端新管理页的 users/sites REST 路径
				adminLegacy.GET("/users", adminLegacyHandler.HandleUsersGet)
				adminLegacy.POST("/users", adminLegacyHandler.HandleUsersCreate)
				adminLegacy.PUT("/users/:username", adminLegacyHandler.HandleUsersUpdate)
				adminLegacy.DELETE("/users/:username", adminLegacyHandler.HandleUsersDelete)
				adminLegacy.GET("/sites", adminLegacyHandler.HandleSitesGet)
				adminLegacy.PUT("/sites/:key", adminLegacyHandler.HandleSitesUpdate)
				adminLegacy.DELETE("/sites/:key", adminLegacyHandler.HandleSitesDelete)

				// 用户管理
				adminLegacy.POST("/user", adminLegacyHandler.HandleUser)

				// 资源站（视频源）管理
				adminLegacy.POST("/source", adminLegacyHandler.HandleSource)
				adminLegacy.GET("/source/validate", adminLegacyHandler.HandleSourceValidate)

				// 分类管理
				adminLegacy.POST("/category", adminLegacyHandler.HandleCategory)

				// 站点配置
				adminLegacy.GET("/site", adminLegacyHandler.HandleGetSiteConfig)
				adminLegacy.POST("/site", adminLegacyHandler.HandleSite)

				// 配置订阅
				adminLegacy.POST("/config_subscription/fetch", adminLegacyHandler.HandleFetchSubscription)

				// 配置文件
				adminLegacy.POST("/config_file", adminLegacyHandler.HandleConfigFile)

				// 直播源管理
				adminLegacy.POST("/live", adminLegacyHandler.HandleLive)
				adminLegacy.POST("/live/refresh", adminLegacyHandler.HandleLiveRefresh)
				adminLegacy.GET("/reset", adminLegacyHandler.HandleReset)

				// 数据迁移
				adminLegacy.GET("/data_migration/export", adminLegacyHandler.HandleDataExport)
				adminLegacy.POST("/data_migration/import", adminLegacyHandler.HandleDataImport)
			}
		}
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
