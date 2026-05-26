package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"rtm-server/internal/dto"
	"rtm-server/internal/middleware"
	"rtm-server/internal/service"
	"rtm-server/pkg/response"
)

// AlarmRecordHandler 告警记录处理器
type AlarmRecordHandler struct {
	alarmRecordService *service.AlarmRecordService
	operLogService     *service.OperLogService
}

// NewAlarmRecordHandler 创建告警记录处理器
func NewAlarmRecordHandler(alarmRecordService *service.AlarmRecordService, operLogService ...*service.OperLogService) *AlarmRecordHandler {
	h := &AlarmRecordHandler{alarmRecordService: alarmRecordService}
	if len(operLogService) > 0 {
		h.operLogService = operLogService[0]
	}
	return h
}

// RegisterRoutes 注册告警记录路由
func (h *AlarmRecordHandler) RegisterRoutes(r *gin.RouterGroup) {
	record := r.Group("/rtm/alarmRecord")
	{
		record.GET("/list", h.List)
		record.GET("/:id", h.GetInfo)
		record.GET("/unreadCount", h.UnreadCount)
		if h.operLogService != nil {
			record.PUT("/ack/:id", middleware.OperLog(middleware.DefaultOperLogConfig("告警记录", middleware.BusinessTypeUpdate), h.operLogService), h.Acknowledge)
			record.PUT("/batchAck", middleware.OperLog(middleware.DefaultOperLogConfig("告警记录", middleware.BusinessTypeUpdate), h.operLogService), h.BatchAcknowledge)
		} else {
			record.PUT("/ack/:id", h.Acknowledge)
			record.PUT("/batchAck", h.BatchAcknowledge)
		}
	}
}

// List 告警记录列表
func (h *AlarmRecordHandler) List(c *gin.Context) {
	var query dto.AlarmRecordQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	records, total, err := h.alarmRecordService.List(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessPage(c, total, records)
}

// GetInfo 告警记录详情
func (h *AlarmRecordHandler) GetInfo(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.Error(c, "记录ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "记录ID格式错误")
		return
	}

	record, err := h.alarmRecordService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "告警记录不存在")
		return
	}

	response.Success(c, record)
}

// Acknowledge 确认告警
func (h *AlarmRecordHandler) Acknowledge(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.Error(c, "记录ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "记录ID格式错误")
		return
	}

	if err := h.alarmRecordService.Acknowledge(c.Request.Context(), id); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// BatchAcknowledge 批量确认
func (h *AlarmRecordHandler) BatchAcknowledge(c *gin.Context) {
	var req dto.AlarmRecordBatchAckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if err := h.alarmRecordService.BatchAcknowledge(c.Request.Context(), req.IDs); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UnreadCount 未读告警数量
func (h *AlarmRecordHandler) UnreadCount(c *gin.Context) {
	count, err := h.alarmRecordService.CountUnread(c.Request.Context())
	if err != nil {
		response.Error(c, err.Error())
		return
	}
	response.Success(c, count)
}
