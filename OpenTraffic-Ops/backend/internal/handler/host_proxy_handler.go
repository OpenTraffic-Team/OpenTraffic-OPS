package handler

import (
	"github.com/gin-gonic/gin"

	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/service"
	"opentraffic-ops-backend/pkg/response"
)

// HostProxyHandler HostProxy 上报接口处理器
type HostProxyHandler struct {
	hostProxyService *service.HostProxyService
}

// NewHostProxyHandler 创建 HostProxy 处理器
func NewHostProxyHandler(hostProxyService *service.HostProxyService) *HostProxyHandler {
	return &HostProxyHandler{hostProxyService: hostProxyService}
}

// RegisterRoutes 注册 HostProxy 公开路由（无需认证）
func (h *HostProxyHandler) RegisterRoutes(r *gin.RouterGroup) {
	proxy := r.Group("/api/v1/proxy")
	{
		proxy.POST("/heartbeat", h.Heartbeat)
		proxy.POST("/register", h.Register)
		proxy.POST("/poll", h.Poll)
		proxy.POST("/ack", h.Ack)
	}
}

// Heartbeat 心跳上报（合并健康度数据）
func (h *HostProxyHandler) Heartbeat(c *gin.Context) {
	var req dto.HostProxyHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if err := h.hostProxyService.Heartbeat(c.Request.Context(), &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Register HostProxy 首次注册
func (h *HostProxyHandler) Register(c *gin.Context) {
	var req dto.HostProxyRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	resp, err := h.hostProxyService.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// Poll 轮询待执行指令
func (h *HostProxyHandler) Poll(c *gin.Context) {
	var req dto.HostProxyPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 兼容 query 形式
		if err2 := c.ShouldBindQuery(&req); err2 != nil {
			response.Error(c, "请求参数错误")
			return
		}
	}

	commands, err := h.hostProxyService.Poll(c.Request.Context(), req.IP)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, &dto.HostProxyPollResponse{Commands: commands})
}

// Ack 指令执行结果上报
func (h *HostProxyHandler) Ack(c *gin.Context) {
	var req dto.HostProxyCommandAckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if err := h.hostProxyService.AckCommand(c.Request.Context(), &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
