package controller

import (
	"net/http"
	"rtm-initialization-backend/internal/middleware"
	"rtm-initialization-backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeployController 部署控制器
type DeployController struct {
	deployService *service.DeployService
}

// NewDeployController 创建部署控制器
func NewDeployController(deployService *service.DeployService) *DeployController {
	return &DeployController{
		deployService: deployService,
	}
}

// DeployBinary 部署二进制文件
func (ctrl *DeployController) DeployBinary(c *gin.Context) {
	var req service.DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := ctrl.deployService.Deploy(&req, middleware.GetUsername(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

// UndeployBinary 卸载二进制文件
func (ctrl *DeployController) UndeployBinary(c *gin.Context) {
	var req service.UndeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ctrl.deployService.Undeploy(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service undeployed successfully"})
}

// ListDeployRecords 获取部署记录列表
func (ctrl *DeployController) ListDeployRecords(c *gin.Context) {
	serverID := c.Query("server_id")

	records, err := ctrl.deployService.ListRecords(serverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list deploy records"})
		return
	}

	c.JSON(http.StatusOK, records)
}

// GetDeployRecord 获取部署记录详情
func (ctrl *DeployController) GetDeployRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid record ID"})
		return
	}

	record, err := ctrl.deployService.GetRecord(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deploy record not found"})
		return
	}

	c.JSON(http.StatusOK, record)
}
