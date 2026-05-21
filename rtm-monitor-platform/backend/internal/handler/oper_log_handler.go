package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rtm-server/internal/dto"
	"rtm-server/internal/service"
	"rtm-server/pkg/response"
)

// OperLogHandler 操作日志处理器
type OperLogHandler struct {
	service *service.OperLogService
}

// NewOperLogHandler 创建操作日志处理器
func NewOperLogHandler(service *service.OperLogService) *OperLogHandler {
	return &OperLogHandler{service: service}
}

// RegisterRoutes 注册操作日志管理路由
func (h *OperLogHandler) RegisterRoutes(r *gin.RouterGroup) {
	operLog := r.Group("/monitor/operlog")
	{
		operLog.GET("/list", h.List)
		operLog.GET("/:operId", h.GetInfo)
		operLog.DELETE("/:operIds", h.Delete)
		operLog.DELETE("/clean", h.Clean)
	}
}

// List 操作日志列表
func (h *OperLogHandler) List(c *gin.Context) {
	var query dto.OperLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	list, total, err := h.service.List(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessPage(c, total, list)
}

// GetInfo 操作日志详情
func (h *OperLogHandler) GetInfo(c *gin.Context) {
	idStr := c.Param("operId")
	if idStr == "" {
		response.Error(c, "日志ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "日志ID格式错误")
		return
	}

	log, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "日志不存在")
		return
	}

	response.Success(c, log)
}

// Delete 删除操作日志
func (h *OperLogHandler) Delete(c *gin.Context) {
	idsStr := c.Param("operIds")
	strs := strings.Split(idsStr, ",")
	var ids []int64
	for _, s := range strs {
		id, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}

	if len(ids) == 0 {
		response.Error(c, "日志ID不能为空")
		return
	}

	if err := h.service.DeleteByIDs(c.Request.Context(), ids); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Clean 清空操作日志
func (h *OperLogHandler) Clean(c *gin.Context) {
	if err := h.service.Clean(c.Request.Context()); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
