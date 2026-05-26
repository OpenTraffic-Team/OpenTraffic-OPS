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

// AlarmRuleHandler 告警规则处理器
type AlarmRuleHandler struct {
	alarmRuleService *service.AlarmRuleService
	operLogService   *service.OperLogService
}

// NewAlarmRuleHandler 创建告警规则处理器
func NewAlarmRuleHandler(alarmRuleService *service.AlarmRuleService, operLogService ...*service.OperLogService) *AlarmRuleHandler {
	h := &AlarmRuleHandler{alarmRuleService: alarmRuleService}
	if len(operLogService) > 0 {
		h.operLogService = operLogService[0]
	}
	return h
}

// RegisterRoutes 注册告警规则路由
func (h *AlarmRuleHandler) RegisterRoutes(r *gin.RouterGroup) {
	rule := r.Group("/rtm/alarmRule")
	{
		rule.GET("/list", h.List)
		rule.GET("/:id", h.GetInfo)
		if h.operLogService != nil {
			rule.POST("", middleware.OperLog(middleware.DefaultOperLogConfig("告警规则", middleware.BusinessTypeInsert), h.operLogService), h.Create)
			rule.PUT("", middleware.OperLog(middleware.DefaultOperLogConfig("告警规则", middleware.BusinessTypeUpdate), h.operLogService), h.Update)
			rule.DELETE("/:ids", middleware.OperLog(middleware.DefaultOperLogConfig("告警规则", middleware.BusinessTypeDelete), h.operLogService), h.Delete)
			rule.PUT("/status", middleware.OperLog(middleware.DefaultOperLogConfig("告警规则", middleware.BusinessTypeUpdate), h.operLogService), h.UpdateStatus)
		} else {
			rule.POST("", h.Create)
			rule.PUT("", h.Update)
			rule.DELETE("/:ids", h.Delete)
			rule.PUT("/status", h.UpdateStatus)
		}
	}
}

// List 告警规则列表
func (h *AlarmRuleHandler) List(c *gin.Context) {
	var query dto.AlarmRuleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	rules, total, err := h.alarmRuleService.List(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessPage(c, total, rules)
}

// GetInfo 告警规则详情
func (h *AlarmRuleHandler) GetInfo(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.Error(c, "规则ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "规则ID格式错误")
		return
	}

	rule, err := h.alarmRuleService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "告警规则不存在")
		return
	}

	response.Success(c, rule)
}

// Create 新增告警规则
func (h *AlarmRuleHandler) Create(c *gin.Context) {
	var req dto.AlarmRuleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	createBy := middleware.GetUsername(c)
	if err := h.alarmRuleService.Create(c.Request.Context(), &req, createBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Update 修改告警规则
func (h *AlarmRuleHandler) Update(c *gin.Context) {
	var req dto.AlarmRuleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	updateBy := middleware.GetUsername(c)
	if err := h.alarmRuleService.Update(c.Request.Context(), &req, updateBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除告警规则
func (h *AlarmRuleHandler) Delete(c *gin.Context) {
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
		response.Error(c, "规则ID不能为空")
		return
	}

	if err := h.alarmRuleService.DeleteByIDs(c.Request.Context(), ids); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateStatus 更新状态
func (h *AlarmRuleHandler) UpdateStatus(c *gin.Context) {
	var req dto.AlarmRuleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	updateBy := middleware.GetUsername(c)
	if err := h.alarmRuleService.UpdateStatus(c.Request.Context(), &req, updateBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
