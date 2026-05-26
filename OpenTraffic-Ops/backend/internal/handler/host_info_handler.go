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

// HostInfoHandler 主机处理器
type HostInfoHandler struct {
	hostInfoService *service.HostInfoService
	operLogService      *service.OperLogService
}

// NewHostInfoHandler 创建主机处理器
func NewHostInfoHandler(hostInfoService *service.HostInfoService, operLogService ...*service.OperLogService) *HostInfoHandler {
	h := &HostInfoHandler{hostInfoService: hostInfoService}
	if len(operLogService) > 0 {
		h.operLogService = operLogService[0]
	}
	return h
}

// RegisterRoutes 注册主机管理路由
func (h *HostInfoHandler) RegisterRoutes(r *gin.RouterGroup) {
	host := r.Group("/rtm/hostInfo")
	{
		host.GET("/list", h.List)
		host.GET("/:id", h.GetInfo)
		if h.operLogService != nil {
			host.POST("", middleware.OperLog(middleware.DefaultOperLogConfig("主机管理", middleware.BusinessTypeInsert), h.operLogService), h.Create)
			host.PUT("", middleware.OperLog(middleware.DefaultOperLogConfig("主机管理", middleware.BusinessTypeUpdate), h.operLogService), h.Update)
			host.DELETE("/:ids", middleware.OperLog(middleware.DefaultOperLogConfig("主机管理", middleware.BusinessTypeDelete), h.operLogService), h.Delete)
		} else {
			host.POST("", h.Create)
			host.PUT("", h.Update)
			host.DELETE("/:ids", h.Delete)
		}
		host.GET("/selectHostInfoTreeNode", h.SelectHostInfoTreeNode)
	}
}

// List 主机信息列表
func (h *HostInfoHandler) List(c *gin.Context) {
	var query dto.HostInfoQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	hosts, total, err := h.hostInfoService.List(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessPage(c, total, hosts)
}

// GetInfo 主机详情
func (h *HostInfoHandler) GetInfo(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.Error(c, "主机ID不能为空")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "主机ID格式错误")
		return
	}

	host, err := h.hostInfoService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, "主机不存在")
		return
	}

	response.Success(c, host)
}

// Create 新增主机信息（注册）
func (h *HostInfoHandler) Create(c *gin.Context) {
	var req dto.HostInfoCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if err := h.hostInfoService.Create(c.Request.Context(), &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Update 修改主机信息
func (h *HostInfoHandler) Update(c *gin.Context) {
	var req dto.HostInfoUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if err := h.hostInfoService.Update(c.Request.Context(), &req); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除主机信息
func (h *HostInfoHandler) Delete(c *gin.Context) {
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
		response.Error(c, "主机ID不能为空")
		return
	}

	if err := h.hostInfoService.DeleteByIDs(c.Request.Context(), ids); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// SelectHostInfoTreeNode 查询主机信息树形节点
func (h *HostInfoHandler) SelectHostInfoTreeNode(c *gin.Context) {
	list, err := h.hostInfoService.SelectHostInfoTreeNode(c.Request.Context())
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, list)
}
