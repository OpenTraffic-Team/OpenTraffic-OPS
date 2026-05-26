package controller

import (
	"net/http"
	"opentraffic-ops-init-backend/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// MonitorController 监控控制器
type MonitorController struct {
	monitorService *service.MonitorService
}

// NewMonitorController 创建监控控制器
func NewMonitorController(monitorService *service.MonitorService) *MonitorController {
	return &MonitorController{
		monitorService: monitorService,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
}

// GetOverview 获取总览信息
// @Summary 获取系统总览
// @Description 获取系统整体运行状态统计
// @Tags 监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} service.Overview
// @Router /api/monitor/overview [get]
func (ctrl *MonitorController) GetOverview(c *gin.Context) {
	overview, err := ctrl.monitorService.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get overview"})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetComponentDetails 获取所有组件详情
// @Summary 获取所有组件详情
// @Description 获取所有组件的详细运行信息
// @Tags 监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} service.ComponentDetail
// @Router /api/monitor/components [get]
func (ctrl *MonitorController) GetComponentDetails(c *gin.Context) {
	details, err := ctrl.monitorService.GetAllComponentDetails(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get component details"})
		return
	}

	c.JSON(http.StatusOK, details)
}

// WebSocketRealtime 实时监控WebSocket
// @Summary 实时监控WebSocket
// @Description 通过WebSocket推送实时监控数据
// @Tags 监控
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/monitor/realtime [get]
func (ctrl *MonitorController) WebSocketRealtime(c *gin.Context) {
	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 定时发送实时数据
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			// 获取组件详情
			details, err := ctrl.monitorService.GetAllComponentDetails(c.Request.Context())
			if err != nil {
				continue
			}

			// 发送数据到客户端
			if err := conn.WriteJSON(details); err != nil {
				return
			}
		}
	}
}
