package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/config"
	"opentraffic-ops-backend/internal/handler"
	"opentraffic-ops-backend/internal/middleware"
	"opentraffic-ops-backend/internal/service"
)

// RouterServices 路由创建的所有服务集合
type RouterServices struct {
	HostInfoService   *service.HostInfoService
	HostHealthService *service.HostHealthService
}

// SetupRouter 设置路由
func SetupRouter(r *gin.Engine, db *gorm.DB) *RouterServices {
	// 创建日志服务
	operLogService := service.NewOperLogService(db)
	loginLogService := service.NewLoginLogService(db)

	// 创建服务
	authService := service.NewAuthService(db, config.RedisPlatform, loginLogService)

	// 创建处理器
	authHandler := handler.NewAuthHandler(authService)

	// 公开路由组（无需认证）
	public := r.Group("")
	authHandler.RegisterRoutes(public)

	// 需要认证的路由组
	auth := r.Group("")
	auth.Use(middleware.JWTAuth())
	authHandler.RegisterAuthRoutes(auth)

	// 系统管理路由

	// 用户管理
	userService := service.NewUserService(db)
	userHandler := handler.NewUserHandler(userService, operLogService)
	userHandler.RegisterRoutes(auth)

	// 主机信息管理
	hostInfoService := service.NewHostInfoService(db)
	hostInfoHandler := handler.NewHostInfoHandler(hostInfoService, operLogService)
	hostInfoHandler.RegisterRoutes(auth)

	// 主机健康度服务
	hostHealthService := service.NewHostHealthService(db)
	hostHealthHandler := handler.NewHostHealthHandler(hostHealthService)
	hostHealthHandler.RegisterRoutes(auth)

	// HostProxy 上报服务（无需认证，公开路由）
	hostProxyService := service.NewHostProxyService(hostInfoService, hostHealthService)
	hostProxyHandler := handler.NewHostProxyHandler(hostProxyService)
	hostProxyHandler.RegisterRoutes(public)

	// HostProxy WebSocket连接（无需认证）
	public.GET("/api/v1/proxy/ws", handler.HandleHostProxyWS)

	// 前端终端WebSocket（需要认证）—— 单独注册，避免 auth 组的 JWTAuth() 拦截（WS 通过 query param 传 token）
	r.GET("/ws/terminal", middleware.WSAuth(), handler.HandleTerminalWS)

	// 首页统计
	homeService := service.NewHomeService(db)
	homeHandler := handler.NewHomeHandler(homeService)
	homeHandler.RegisterRoutes(auth)

	// 操作日志管理
	operLogHandler := handler.NewOperLogHandler(operLogService)
	operLogHandler.RegisterRoutes(auth)

	// 登录日志管理
	loginLogHandler := handler.NewLoginLogHandler(loginLogService)
	loginLogHandler.RegisterRoutes(auth)

	// 远程文件操作
	remoteHandler := handler.NewRemoteHandler()
	remoteHandler.RegisterRoutes(auth)

	// Agent 外部 API 代理（需要认证）
	agentControlHandler := handler.NewAgentControlHandler(config.GlobalConfig.Agent.Control)
	agentControlHandler.RegisterRoutes(auth)

	// 感知 Agent 外部 API 代理（需要认证）
	agentPerceiveHandler := handler.NewAgentPerceiveHandler(config.GlobalConfig.Agent.Perceive)
	agentPerceiveHandler.RegisterRoutes(auth)

	// 告警模块路由
	alarmChannelService := service.NewAlarmChannelService(db)
	alarmChannelHandler := handler.NewAlarmChannelHandler(alarmChannelService, operLogService)
	alarmChannelHandler.RegisterRoutes(auth)

	alarmRuleService := service.NewAlarmRuleService(db)
	alarmRuleHandler := handler.NewAlarmRuleHandler(alarmRuleService, operLogService)
	alarmRuleHandler.RegisterRoutes(auth)

	alarmRecordService := service.NewAlarmRecordService(db)
	alarmRecordHandler := handler.NewAlarmRecordHandler(alarmRecordService, operLogService)
	alarmRecordHandler.RegisterRoutes(auth)

	// Agent 聊天会话模块
	chatService := service.NewChatService(db)
	chatHandler := handler.NewChatHandler(chatService, operLogService)
	chatHandler.RegisterRoutes(auth)

	return &RouterServices{
		HostInfoService:   hostInfoService,
		HostHealthService: hostHealthService,
	}
}
