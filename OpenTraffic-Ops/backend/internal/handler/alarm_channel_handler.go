package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/middleware"
	"opentraffic-ops-backend/internal/service"
	"opentraffic-ops-backend/pkg/response"
)

// AlarmChannelHandler 告警通道处理器
type AlarmChannelHandler struct {
	alarmChannelService *service.AlarmChannelService
	operLogService      *service.OperLogService
}

// NewAlarmChannelHandler 创建告警通道处理器
func NewAlarmChannelHandler(alarmChannelService *service.AlarmChannelService, operLogService ...*service.OperLogService) *AlarmChannelHandler {
	h := &AlarmChannelHandler{alarmChannelService: alarmChannelService}
	if len(operLogService) > 0 {
		h.operLogService = operLogService[0]
	}
	return h
}

// RegisterRoutes 注册告警通道路由
func (h *AlarmChannelHandler) RegisterRoutes(r *gin.RouterGroup) {
	ch := r.Group("/rtm/alarmChannel")
	{
		ch.GET("/list", h.List)
		ch.GET("/all", h.FindAll)
		ch.GET("/:id", h.GetInfo)
		if h.operLogService != nil {
			ch.POST("", middleware.OperLog(middleware.DefaultOperLogConfig("告警通道", middleware.BusinessTypeInsert), h.operLogService), h.Create)
			ch.PUT("", middleware.OperLog(middleware.DefaultOperLogConfig("告警通道", middleware.BusinessTypeUpdate), h.operLogService), h.Update)
			ch.DELETE("/:ids", middleware.OperLog(middleware.DefaultOperLogConfig("告警通道", middleware.BusinessTypeDelete), h.operLogService), h.Delete)
			ch.PUT("/status", middleware.OperLog(middleware.DefaultOperLogConfig("告警通道", middleware.BusinessTypeUpdate), h.operLogService), h.UpdateStatus)
			ch.PUT("/setDefault", middleware.OperLog(middleware.DefaultOperLogConfig("告警通道", middleware.BusinessTypeUpdate), h.operLogService), h.SetDefault)
		} else {
			ch.POST("", h.Create)
			ch.PUT("", h.Update)
			ch.DELETE("/:ids", h.Delete)
			ch.PUT("/status", h.UpdateStatus)
			ch.PUT("/setDefault", h.SetDefault)
		}
	}
}

// List 告警通道列表
func (h *AlarmChannelHandler) List(c *gin.Context) {
	var query dto.AlarmChannelQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	channels, total, err := h.alarmChannelService.List(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessPage(c, total, channels)
}

// FindAll 查询所有通道
func (h *AlarmChannelHandler) FindAll(c *gin.Context) {
	channels, err := h.alarmChannelService.FindAll(c.Request.Context())
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, channels)
}

// GetInfo 告警通道详情
func (h *AlarmChannelHandler) GetInfo(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.Error(c, "通道ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "通道ID格式错误")
		return
	}

	channel, err := h.alarmChannelService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "告警通道不存在")
		return
	}

	response.Success(c, channel)
}

// Create 新增告警通道
func (h *AlarmChannelHandler) Create(c *gin.Context) {
	var req dto.AlarmChannelCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	createBy := middleware.GetUsername(c)
	if err := h.alarmChannelService.Create(c.Request.Context(), &req, createBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Update 修改告警通道
func (h *AlarmChannelHandler) Update(c *gin.Context) {
	var req dto.AlarmChannelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	updateBy := middleware.GetUsername(c)
	if err := h.alarmChannelService.Update(c.Request.Context(), &req, updateBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除告警通道
func (h *AlarmChannelHandler) Delete(c *gin.Context) {
	idsStr := c.Param("ids")
	strs := strings.Split(idsStr, ",")
	var ids []int64
	for _, s := range strs {
		id, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		response.Error(c, "通道ID不能为空")
		return
	}

	if err := h.alarmChannelService.DeleteByIDs(c.Request.Context(), ids); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateStatus 更新状态
func (h *AlarmChannelHandler) UpdateStatus(c *gin.Context) {
	var req dto.AlarmChannelStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	updateBy := middleware.GetUsername(c)
	if err := h.alarmChannelService.UpdateStatus(c.Request.Context(), &req, updateBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// SetDefault 设置默认通道
func (h *AlarmChannelHandler) SetDefault(c *gin.Context) {
	var req dto.AlarmChannelSetDefaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	updateBy := middleware.GetUsername(c)
	if err := h.alarmChannelService.SetDefault(c.Request.Context(), &req, updateBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
