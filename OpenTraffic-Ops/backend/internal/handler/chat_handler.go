package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"opentraffic-ops-backend/internal/constant"
	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/middleware"
	"opentraffic-ops-backend/internal/service"
	"opentraffic-ops-backend/pkg/response"
)

// normalizeSessionType 规范化 sessionType；空或非法值兜底为 control，保持与历史前端兼容
func normalizeSessionType(s string) string {
	switch strings.TrimSpace(s) {
	case constant.ChatSessionTypePerceive:
		return constant.ChatSessionTypePerceive
	default:
		return constant.ChatSessionTypeControl
	}
}

// ChatHandler Agent 聊天会话处理器
type ChatHandler struct {
	chatService    *service.ChatService
	operLogService *service.OperLogService
}

// NewChatHandler 创建聊天会话处理器
func NewChatHandler(chatService *service.ChatService, operLogService ...*service.OperLogService) *ChatHandler {
	h := &ChatHandler{chatService: chatService}
	if len(operLogService) > 0 {
		h.operLogService = operLogService[0]
	}
	return h
}

// RegisterRoutes 注册聊天会话路由
func (h *ChatHandler) RegisterRoutes(r *gin.RouterGroup) {
	ch := r.Group("/rtm/chatSession")
	{
		ch.GET("/list", h.List)
		ch.GET("/:id", h.GetInfo)
		ch.POST("/turn", h.SaveTurn)
		if h.operLogService != nil {
			ch.PUT("", middleware.OperLog(middleware.DefaultOperLogConfig("Agent聊天会话", middleware.BusinessTypeUpdate), h.operLogService), h.Rename)
			ch.DELETE("/:ids", middleware.OperLog(middleware.DefaultOperLogConfig("Agent聊天会话", middleware.BusinessTypeDelete), h.operLogService), h.Delete)
		} else {
			ch.PUT("", h.Rename)
			ch.DELETE("/:ids", h.Delete)
		}
	}
}

// List 当前用户的会话分页列表
func (h *ChatHandler) List(c *gin.Context) {
	var query dto.ChatSessionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	sessionType := normalizeSessionType(query.SessionType)
	sessions, total, err := h.chatService.ListSessions(c.Request.Context(), userID, sessionType, &query)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.SuccessPage(c, total, sessions)
}

// GetInfo 会话详情（含消息）
func (h *ChatHandler) GetInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, "会话ID格式错误")
		return
	}

	userID := middleware.GetUserID(c)
	detail, err := h.chatService.GetSessionDetail(c.Request.Context(), userID, id)
	if err != nil {
		response.Error(c, "会话不存在")
		return
	}

	response.Success(c, detail)
}

// SaveTurn 保存一轮对话（user + assistant）
func (h *ChatHandler) SaveTurn(c *gin.Context) {
	var req dto.ChatTurnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)
	sessionType := normalizeSessionType(req.SessionType)
	resp, err := h.chatService.SaveTurn(c.Request.Context(), userID, username, sessionType, &req)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// Rename 重命名会话
func (h *ChatHandler) Rename(c *gin.Context) {
	var req dto.ChatSessionRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	updateBy := middleware.GetUsername(c)
	if err := h.chatService.Rename(c.Request.Context(), userID, &req, updateBy); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 批量删除会话
func (h *ChatHandler) Delete(c *gin.Context) {
	idsStr := c.Param("ids")
	strs := strings.Split(idsStr, ",")
	var ids []int64
	for _, s := range strs {
		id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		response.Error(c, "会话ID不能为空")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.chatService.DeleteByIDs(c.Request.Context(), userID, ids); err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, nil)
}
