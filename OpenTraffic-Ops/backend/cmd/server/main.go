package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"opentraffic-ops-backend/internal/config"
	"opentraffic-ops-backend/internal/middleware"
	"opentraffic-ops-backend/internal/router"
	"opentraffic-ops-backend/internal/service"
	"opentraffic-ops-backend/pkg/response"
	"opentraffic-ops-backend/pkg/static"
)

func init() {
	// 设置全局时区为东八区（不依赖系统 zoneinfo 文件）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}
	time.Local = loc
}

func main() {
	flag.Parse()

	// 加载配置（固定路径：~/.opentraffic-ops/opentraffic-ops-config.yaml）
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := config.InitLogger(&cfg.Log); err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}

	zap.L().Info("Starting OpenTraffic Ops Server",
		zap.String("version", config.Version),
		zap.String("build_time", config.BuildTime),
		zap.String("go_version", config.GoVersion),
	)

	// 初始化数据库
	db, err := config.InitDatabase(&cfg.Datasource)
	if err != nil {
		zap.L().Fatal("Failed to init database", zap.Error(err))
	}
	_ = db

	// 初始化Redis
	if err := config.InitRedis(&cfg.Redis); err != nil {
		zap.L().Fatal("Failed to init redis", zap.Error(err))
	}

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.Server.Mode == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建Gin引擎
	r := gin.New()

	// 注册全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.XSS())

	// 健康检查接口
	r.GET("/ping", func(c *gin.Context) {
		response.Success(c, "pong")
	})

	// 静态文件服务（用于文件上传下载）
	r.Static("/profile", cfg.App.Profile)

	// 注册路由
	services := router.SetupRouter(r, db)

	// SPA fallback: serve frontend static files for unmatched routes
	r.NoRoute(gin.WrapH(static.Handler()))

	// 启动定时任务调度器
	scheduler := service.NewScheduler(db, services.HostInfoService, services.HostHealthService)
	scheduler.Start()

	// 启动HTTP服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		zap.L().Info("HTTP server started", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("Shutting down server...")

	// 关闭定时任务调度器
	if scheduler != nil {
		scheduler.Stop()
	}

	// 关闭HTTP服务
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("Server forced to shutdown", zap.Error(err))
	}

	// 关闭资源
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}
	if config.RedisPlatform != nil {
		config.RedisPlatform.Close()
	}
	if config.RedisEdge != nil {
		config.RedisEdge.Close()
	}

	zap.L().Info("Server stopped")
}
