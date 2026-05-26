package main

import (
	"fmt"
	"log"
	"rtm-initialization-backend/internal/controller"
	"rtm-initialization-backend/internal/middleware"
	"rtm-initialization-backend/internal/repository"
	"rtm-initialization-backend/internal/service"
	"rtm-initialization-backend/pkg/config"
	"rtm-initialization-backend/pkg/crypto"
	"rtm-initialization-backend/pkg/static"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := repository.InitDatabase(cfg.Database.DataDir); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 创建加密器
	encryptor := crypto.NewEncryptor(cfg.EncryptionKey)

	// 初始化服务
	authService := service.NewAuthService(cfg.JWT.Secret, cfg.JWT.ExpireHours)
	componentService, err := service.NewComponentService(encryptor)
	if err != nil {
		log.Fatalf("Failed to create component service: %v", err)
	}
	defer componentService.Close()

	monitorService, err := service.NewMonitorService()
	if err != nil {
		log.Fatalf("Failed to create monitor service: %v", err)
	}
	defer monitorService.Close()

	serverService := service.NewServerService(encryptor)
	deployService := service.NewDeployService(serverService)

	// 创建控制器
	authController := controller.NewAuthController(authService)
	componentController := controller.NewComponentController(componentService)
	monitorController := controller.NewMonitorController(monitorService)
	serverController := controller.NewServerController(serverService)
	deployController := controller.NewDeployController(deployService)

	// 创建Gin引擎
	if cfg.Server.Host == "0.0.0.0" || cfg.Server.Host == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.MaxMultipartMemory = 2 << 30 // 约 2GB

	// 添加中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "rtm-initialization-backend",
		})
	})

	// API路由组
	api := r.Group("/api")
	{
		// 认证路由（不需要JWT验证）
		auth := api.Group("/auth")
		{
			auth.POST("/login", authController.Login)
			auth.POST("/logout", authController.Logout)
		}

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware(authService))
		{
			// 用户路由
			users := authenticated.Group("/users")
			{
				users.GET("/profile", authController.GetProfile)
			}

			// 组件路由
			components := authenticated.Group("/components")
			{
				components.GET("", componentController.ListComponents)
				components.GET("/catalog", componentController.GetComponentCatalog)
				components.GET("/:id", componentController.GetComponent)
				components.POST("", componentController.InstallComponent)
				components.DELETE("/:id", componentController.UninstallComponent)
				components.POST("/:id/start", componentController.StartComponent)
				components.POST("/:id/stop", componentController.StopComponent)
				components.POST("/:id/restart", componentController.RestartComponent)
				components.GET("/:id/logs", componentController.GetComponentLogs)
				components.GET("/:id/logs/stream", componentController.StreamLogs)
				components.GET("/:id/stats", componentController.GetComponentStats)
				components.PUT("/:id/config", componentController.UpdateComponentConfig)
			}

			// 监控路由
			monitor := authenticated.Group("/monitor")
			{
				monitor.GET("/overview", monitorController.GetOverview)
				monitor.GET("/components", monitorController.GetComponentDetails)
				monitor.GET("/realtime", monitorController.WebSocketRealtime)
			}

			// 服务器路由
			servers := authenticated.Group("/servers")
			{
				servers.GET("", serverController.ListServers)
				servers.POST("", serverController.CreateServer)
				servers.GET("/:id", serverController.GetServer)
				servers.PUT("/:id", serverController.UpdateServer)
				servers.DELETE("/:id", serverController.DeleteServer)
				servers.POST("/:id/test", serverController.TestServerConnection)
				servers.GET("/:id/proxy-config", serverController.GetProxyConfig)
				servers.PUT("/:id/proxy-config", serverController.UpdateProxyConfig)
				servers.GET("/:id/configs/:software", serverController.GetSoftwareConfig)
				servers.PUT("/:id/configs/:software", serverController.UpdateSoftwareConfig)
				servers.GET("/configs/:software/default", serverController.GetDefaultSoftwareConfig)
				servers.GET("/:id/services/:software/status", serverController.GetServiceStatus)
				servers.POST("/:id/services/:software/start", serverController.StartService)
				servers.POST("/:id/services/:software/stop", serverController.StopService)
				servers.POST("/:id/services/:software/restart", serverController.RestartService)
			}

			// 部署路由
			deploy := authenticated.Group("/deploy")
			{
				deploy.POST("", deployController.DeployBinary)
				deploy.POST("/undeploy", deployController.UndeployBinary)
				deploy.GET("/records", deployController.ListDeployRecords)
				deploy.GET("/records/:id", deployController.GetDeployRecord)
			}

		}
	}

	// 兜底：为前端 SPA 提供静态文件服务（非 API/health 路径）
	r.NoRoute(gin.WrapH(static.Handler()))

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting RTM Initialization Backend Server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
