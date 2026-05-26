package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"rtm-server/internal/service"
	"rtm-server/pkg/response"
)

// HostHealthHandler 主机健康度处理器
type HostHealthHandler struct {
	hostHealthService *service.HostHealthService
}

// NewHostHealthHandler 创建主机健康度处理器
func NewHostHealthHandler(hostHealthService *service.HostHealthService) *HostHealthHandler {
	return &HostHealthHandler{hostHealthService: hostHealthService}
}

// RegisterRoutes 注册主机监控路由
func (h *HostHealthHandler) RegisterRoutes(r *gin.RouterGroup) {
	host := r.Group("/rtm/hostMon")
	{
		host.GET("/list", h.List)
		host.GET("/hostHistory", h.HostHistory)
	}
}

// List 主机监控列表（当前最新数据）
func (h *HostHealthHandler) List(c *gin.Context) {
	list, err := h.hostHealthService.SelectHostInfoVoList(c.Request.Context())
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"msg":     "查询成功",
		"data":    list,
		"total":   len(list),
		"rows":    list,
	})
}

// HostHistory 主机历史监控数据
func (h *HostHealthHandler) HostHistory(c *gin.Context) {
	ip := c.Query("ip")
	queryLevel := c.DefaultQuery("queryLevel", "1")
	queryDate := c.Query("queryDate")
	queryHour := c.Query("queryHour")

	if ip == "" {
		response.Error(c, "IP不能为空")
		return
	}

	result := h.hostHealthService.GetHostMonHistoryData(c.Request.Context(), ip, queryLevel, queryDate, queryHour)
	response.Success(c, result)
}
