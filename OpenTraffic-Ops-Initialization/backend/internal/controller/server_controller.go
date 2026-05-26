package controller

import (
	"net/http"
	"opentraffic-ops-init-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// ServerController 服务器控制器
type ServerController struct {
	serverService *service.ServerService
}

// NewServerController 创建服务器控制器
func NewServerController(serverService *service.ServerService) *ServerController {
	return &ServerController{
		serverService: serverService,
	}
}

// ListServers 获取服务器列表
func (ctrl *ServerController) ListServers(c *gin.Context) {
	servers, err := ctrl.serverService.ListServers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list servers"})
		return
	}
	c.JSON(http.StatusOK, servers)
}

// GetServer 获取服务器详情
func (ctrl *ServerController) GetServer(c *gin.Context) {
	id := c.Param("id")
	server, err := ctrl.serverService.GetServer(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Server not found"})
		return
	}
	c.JSON(http.StatusOK, server)
}

// CreateServer 创建服务器
func (ctrl *ServerController) CreateServer(c *gin.Context) {
	var req service.CreateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := ctrl.serverService.CreateServer(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, server)
}

// UpdateServer 更新服务器
func (ctrl *ServerController) UpdateServer(c *gin.Context) {
	id := c.Param("id")

	var req service.UpdateServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	server, err := ctrl.serverService.UpdateServer(id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, server)
}

// DeleteServer 删除服务器
func (ctrl *ServerController) DeleteServer(c *gin.Context) {
	id := c.Param("id")

	if err := ctrl.serverService.DeleteServer(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Server deleted successfully"})
}

// TestServerConnection 测试SSH连接
func (ctrl *ServerController) TestServerConnection(c *gin.Context) {
	id := c.Param("id")

	if err := ctrl.serverService.TestConnection(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
}

// GetProxyConfig 获取opentraffic-ops-proxy配置
func (ctrl *ServerController) GetProxyConfig(c *gin.Context) {
	id := c.Param("id")

	content, err := ctrl.serverService.GetProxyConfig(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// UpdateProxyConfig 更新opentraffic-ops-proxy配置
func (ctrl *ServerController) UpdateProxyConfig(c *gin.Context) {
	id := c.Param("id")

	var req service.UpdateProxyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.serverService.UpdateProxyConfig(id, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})
}

// GetSoftwareConfig 获取指定软件的配置
func (ctrl *ServerController) GetSoftwareConfig(c *gin.Context) {
	id := c.Param("id")
	software := c.Param("software")

	content, err := ctrl.serverService.GetSoftwareConfig(id, software)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// GetDefaultSoftwareConfig 获取指定软件的默认配置（嵌入资源）
func (ctrl *ServerController) GetDefaultSoftwareConfig(c *gin.Context) {
	software := c.Param("software")

	content, err := ctrl.serverService.GetDefaultSoftwareConfig(software)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// UpdateSoftwareConfig 更新指定软件的配置
func (ctrl *ServerController) UpdateSoftwareConfig(c *gin.Context) {
	id := c.Param("id")
	software := c.Param("software")

	var req service.UpdateProxyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.serverService.UpdateSoftwareConfig(id, software, req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})
}

// GetServiceStatus 获取指定服务的运行状态
func (ctrl *ServerController) GetServiceStatus(c *gin.Context) {
	id := c.Param("id")
	software := c.Param("software")

	status, err := ctrl.serverService.GetServiceStatus(id, software)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// StartService 启动指定服务
func (ctrl *ServerController) StartService(c *gin.Context) {
	id := c.Param("id")
	software := c.Param("software")

	if err := ctrl.serverService.StartService(id, software); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service started successfully"})
}

// StopService 停止指定服务
func (ctrl *ServerController) StopService(c *gin.Context) {
	id := c.Param("id")
	software := c.Param("software")

	if err := ctrl.serverService.StopService(id, software); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service stopped successfully"})
}

// RestartService 重启指定服务
func (ctrl *ServerController) RestartService(c *gin.Context) {
	id := c.Param("id")
	software := c.Param("software")

	if err := ctrl.serverService.RestartService(id, software); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service restarted successfully"})
}
