package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"rtm-server/internal/dto"
	"rtm-server/internal/middleware"
	"rtm-server/internal/service"
	"rtm-server/pkg/response"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
	operLogService  *service.OperLogService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService, operLogService ...*service.OperLogService) *UserHandler {
	h := &UserHandler{userService: userService}
	if len(operLogService) > 0 {
		h.operLogService = operLogService[0]
	}
	return h
}

// RegisterRoutes 注册用户管理路由
func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	user := r.Group("/system/user")
	{
		user.GET("/list", h.List)
		user.GET("/:userId", h.GetInfo)
		if h.operLogService != nil {
			user.POST("", middleware.OperLog(middleware.DefaultOperLogConfig("用户管理", middleware.BusinessTypeInsert), h.operLogService), h.Create)
			user.PUT("", middleware.OperLog(middleware.DefaultOperLogConfig("用户管理", middleware.BusinessTypeUpdate), h.operLogService), h.Update)
			user.DELETE("/:userIds", middleware.OperLog(middleware.DefaultOperLogConfig("用户管理", middleware.BusinessTypeDelete), h.operLogService), h.Delete)
			user.PUT("/resetPwd", middleware.OperLog(middleware.DefaultOperLogConfig("用户管理", middleware.BusinessTypeUpdate), h.operLogService), h.ResetPwd)
			user.PUT("/changeStatus", middleware.OperLog(middleware.DefaultOperLogConfig("用户管理", middleware.BusinessTypeUpdate), h.operLogService), h.ChangeStatus)
		} else {
			user.POST("", h.Create)
			user.PUT("", h.Update)
			user.DELETE("/:userIds", h.Delete)
			user.PUT("/resetPwd", h.ResetPwd)
			user.PUT("/changeStatus", h.ChangeStatus)
		}
	}
}

// List 用户列表
func (h *UserHandler) List(c *gin.Context) {
	var query dto.UserQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	users, total, err := h.userService.List(c.Request.Context(), &query)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessPage(c, total, users)
}

// GetInfo 用户详情
func (h *UserHandler) GetInfo(c *gin.Context) {
	userIDStr := c.Param("userId")
	if userIDStr == "" {
		response.Error(c, "用户ID不能为空")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, "用户ID格式错误")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, "用户不存在")
		return
	}

	response.Success(c, user)
}

// Create 新增用户
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	createBy := middleware.GetUsername(c)
	if err := h.userService.Create(c.Request.Context(), &req, createBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Update 修改用户
func (h *UserHandler) Update(c *gin.Context) {
	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	updateBy := middleware.GetUsername(c)
	if err := h.userService.Update(c.Request.Context(), &req, updateBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	userIDsStr := c.Param("userIds")
	strs := strings.Split(userIDsStr, ",")
	var userIDs []int64
	for _, s := range strs {
		id, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			userIDs = append(userIDs, id)
		}
	}

	if len(userIDs) == 0 {
		response.Error(c, "用户ID不能为空")
		return
	}

	if err := h.userService.Delete(c.Request.Context(), userIDs); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ResetPwd 重置密码
func (h *UserHandler) ResetPwd(c *gin.Context) {
	var req dto.ResetPwdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if err := h.userService.ResetPwd(c.Request.Context(), req.UserID, req.Password); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// ChangeStatus 修改状态
func (h *UserHandler) ChangeStatus(c *gin.Context) {
	var req dto.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	if err := h.userService.ChangeStatus(c.Request.Context(), req.UserID, req.Status); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
