package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"rtm-server/internal/dto"
	"rtm-server/internal/service"
	"rtm-server/pkg/response"
)

// LoginLogHandler 登录日志处理器
type LoginLogHandler struct {
	service *service.LoginLogService
}

// NewLoginLogHandler 创建登录日志处理器
func NewLoginLogHandler(service *service.LoginLogService) *LoginLogHandler {
	return &LoginLogHandler{service: service}
}

// RegisterRoutes 注册登录日志管理路由
func (h *LoginLogHandler) RegisterRoutes(r *gin.RouterGroup) {
	logininfor := r.Group("/monitor/logininfor")
	{
		logininfor.GET("/list", h.List)
		logininfor.GET("/:infoId", h.GetInfo)
		logininfor.DELETE("/:infoIds", h.Delete)
		logininfor.DELETE("/clean", h.Clean)
	}
}

// List 登录日志列表
func (h *LoginLogHandler) List(c *gin.Context) {
	var query dto.LoginLogQuery
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

// GetInfo 登录日志详情
func (h *LoginLogHandler) GetInfo(c *gin.Context) {
	idStr := c.Param("infoId")
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

// Delete 删除登录日志
func (h *LoginLogHandler) Delete(c *gin.Context) {
	idsStr := c.Param("infoIds")
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

// Clean 清空登录日志
func (h *LoginLogHandler) Clean(c *gin.Context) {
	if err := h.service.Clean(c.Request.Context()); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
