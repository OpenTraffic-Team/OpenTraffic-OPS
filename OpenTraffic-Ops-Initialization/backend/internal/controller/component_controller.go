package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"rtm-initialization-backend/internal/middleware"
	"rtm-initialization-backend/internal/model"
	"rtm-initialization-backend/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

// ComponentController 组件控制器
type ComponentController struct {
	componentService *service.ComponentService
}

// NewComponentController 创建组件控制器
func NewComponentController(componentService *service.ComponentService) *ComponentController {
	return &ComponentController{
		componentService: componentService,
	}
}

// InstallComponentForm 安装组件表单请求
type InstallComponentForm struct {
	Name        string              `form:"name" binding:"required"`
	Type        model.ComponentType `form:"type" binding:"required,oneof=postgresql redis"`
	Image       string              `form:"image"`
	Version     string              `form:"version"`
	Config      string              `form:"config"`
	ImageSource string              `form:"image_source" binding:"omitempty,oneof=pull upload embedded"`
}

// ListComponents 获取组件列表
// @Summary 获取组件列表
// @Description 获取所有已安装的组件列表
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Component
// @Router /api/components [get]
func (ctrl *ComponentController) ListComponents(c *gin.Context) {
	components, err := ctrl.componentService.ListComponents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list components"})
		return
	}

	c.JSON(http.StatusOK, components)
}

// GetComponentCatalog 获取组件目录
// @Summary 获取组件目录
// @Description 获取所有内置组件及其实时 Docker 状态
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.ComponentCatalogItemWithStatus
// @Router /api/components/catalog [get]
func (ctrl *ComponentController) GetComponentCatalog(c *gin.Context) {
	catalog, err := ctrl.componentService.GetComponentCatalog(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, catalog)
}

// GetComponent 获取组件详情
// @Summary 获取组件详情
// @Description 根据ID获取组件详细信息
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "组件ID"
// @Success 200 {object} model.Component
// @Failure 404 {object} ErrorResponse
// @Router /api/components/{id} [get]
func (ctrl *ComponentController) GetComponent(c *gin.Context) {
	id := c.Param("id")
	component, err := ctrl.componentService.GetComponent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Component not found"})
		return
	}

	c.JSON(http.StatusOK, component)
}

// InstallComponent 安装组件
// @Summary 安装组件
// @Description 安装一个新的组件（支持拉取镜像或上传镜像包）
// @Tags 组件管理
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param name formData string true "组件名称"
// @Param type formData string true "组件类型"
// @Param image formData string false "Docker镜像地址"
// @Param version formData string false "版本"
// @Param config formData string false "配置JSON字符串"
// @Param image_source formData string false "镜像来源: pull 或 upload"
// @Param image_file formData file false "镜像tar包"
// @Success 201 {object} model.Component
// @Failure 400 {object} ErrorResponse
// @Router /api/components [post]
func (ctrl *ComponentController) InstallComponent(c *gin.Context) {
	var form InstallComponentForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageSource := form.ImageSource
	if imageSource == "" {
		imageSource = "pull"
	}

	if imageSource == "pull" && form.Image == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image is required when image_source is pull"})
		return
	}

	if imageSource == "embedded" {
		// embedded 模式下 image 可空，后端自动补全默认值
		if form.Image == "" {
			form.Image = ""
		}
	}

	var config model.ComponentConfig
	if form.Config != "" {
		if err := json.Unmarshal([]byte(form.Config), &config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid config json: " + err.Error()})
			return
		}
	}

	var imageFilePath string
	if imageSource == "upload" {
		fileHeader, err := c.FormFile("image_file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image_file is required when image_source is upload"})
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open image_file: " + err.Error()})
			return
		}
		defer file.Close()

		tempFile, err := os.CreateTemp("", "image-*.tar")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file: " + err.Error()})
			return
		}
		defer tempFile.Close()

		if _, err := io.Copy(tempFile, file); err != nil {
			os.Remove(tempFile.Name())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image_file: " + err.Error()})
			return
		}
		imageFilePath = tempFile.Name()
	}

	component, err := ctrl.componentService.InstallComponent(c.Request.Context(), &service.InstallComponentRequest{
		Name:          form.Name,
		Type:          form.Type,
		Image:         form.Image,
		Version:       form.Version,
		Config:        config,
		ImageSource:   imageSource,
		ImageFilePath: imageFilePath,
		UserName:      middleware.GetUsername(c),
	})

	if err != nil {
		if imageFilePath != "" {
			os.Remove(imageFilePath)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, component)
}

// UninstallComponent 卸载组件
// @Summary 卸载组件
// @Description 卸载指定的组件
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "组件ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /api/components/{id} [delete]
func (ctrl *ComponentController) UninstallComponent(c *gin.Context) {
	id := c.Param("id")

	err := ctrl.componentService.UninstallComponent(c.Request.Context(), id, middleware.GetUsername(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component uninstalled successfully"})
}

// StartComponent 启动组件
// @Summary 启动组件
// @Description 启动已停止的组件
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "组件ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /api/components/{id}/start [post]
func (ctrl *ComponentController) StartComponent(c *gin.Context) {
	id := c.Param("id")

	err := ctrl.componentService.StartComponent(c.Request.Context(), id, middleware.GetUsername(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component started successfully"})
}

// StopComponent 停止组件
// @Summary 停止组件
// @Description 停止运行中的组件
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "组件ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /api/components/{id}/stop [post]
func (ctrl *ComponentController) StopComponent(c *gin.Context) {
	id := c.Param("id")

	err := ctrl.componentService.StopComponent(c.Request.Context(), id, middleware.GetUsername(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component stopped successfully"})
}

// RestartComponent 重启组件
// @Summary 重启组件
// @Description 重启指定的组件
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "组件ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /api/components/{id}/restart [post]
func (ctrl *ComponentController) RestartComponent(c *gin.Context) {
	id := c.Param("id")

	err := ctrl.componentService.RestartComponent(c.Request.Context(), id, middleware.GetUsername(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Component restarted successfully"})
}

// GetComponentLogs 获取组件日志
// @Summary 获取组件日志
// @Description 获取组件的运行日志
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "组件ID"
// @Param tail query string false "日志行数" default("100")
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /api/components/{id}/logs [get]
func (ctrl *ComponentController) GetComponentLogs(c *gin.Context) {
	id := c.Param("id")
	tail := c.DefaultQuery("tail", "100")

	logs, err := ctrl.componentService.GetComponentLogs(c.Request.Context(), id, tail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// GetComponentStats 获取组件统计信息
// @Summary 获取组件统计信息
// @Description 获取组件的资源使用统计
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "组件ID"
// @Success 200 {object} service.ComponentStats
// @Failure 404 {object} ErrorResponse
// @Router /api/components/{id}/stats [get]
func (ctrl *ComponentController) GetComponentStats(c *gin.Context) {
	id := c.Param("id")

	stats, err := ctrl.componentService.GetComponentStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateComponentConfig 更新组件配置
// @Summary 更新组件配置
// @Description 更新组件的配置信息
// @Tags 组件管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "组件ID"
// @Param request body map[string]interface{} true "配置信息"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /api/components/{id}/config [put]
func (ctrl *ComponentController) UpdateComponentConfig(c *gin.Context) {
	id := c.Param("id")

	var config model.ComponentConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ctrl.componentService.UpdateComponentConfig(id, config, middleware.GetUsername(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})
}

// StreamLogs 流式传输日志
// @Summary 流式传输组件日志
// @Description 通过SSE流式传输组件日志
// @Tags 组件管理
// @Accept json
// @Produce text/event-stream
// @Security BearerAuth
// @Param id path string true "组件ID"
// @Router /api/components/{id}/logs/stream [get]
func (ctrl *ComponentController) StreamLogs(c *gin.Context) {
	id := c.Param("id")

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 创建一个通道来发送日志
	logChan := make(chan string)
	defer close(logChan)

	// 模拟日志流（实际应用中应该从Docker持续读取）
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for i := 0; i < 30; i++ { // 发送30秒的日志
			select {
			case <-c.Request.Context().Done():
				return
			case <-ticker.C:
				logs, err := ctrl.componentService.GetComponentLogs(c.Request.Context(), id, "10")
				if err == nil {
					logChan <- logs
				}
			}
		}
	}()

	// 发送日志到客户端
	for log := range logChan {
		select {
		case <-c.Request.Context().Done():
			return
		default:
			c.SSEvent("message", log)
			c.Writer.Flush()
		}
	}
}
