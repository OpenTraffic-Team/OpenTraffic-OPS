package handler

import (
	"github.com/gin-gonic/gin"

	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/middleware"
	"opentraffic-ops-backend/internal/service"
	"opentraffic-ops-backend/internal/utils"
	"opentraffic-ops-backend/pkg/crypto"
	"opentraffic-ops-backend/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRoutes 注册认证路由（公开路由）
func (h *AuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/login", h.Login)
	r.GET("/captchaImage", h.CaptchaImage)
	r.GET("/getPublicKey", h.GetPublicKey)
}

// RegisterAuthRoutes 注册需要认证的路由
func (h *AuthHandler) RegisterAuthRoutes(r *gin.RouterGroup) {
	r.POST("/logout", h.Logout)
	r.GET("/getInfo", h.GetInfo)
	r.GET("/getRouters", h.GetRouters)
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "请求参数错误")
		return
	}

	clientIP := utils.GetClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	token, err := h.authService.Login(c.Request.Context(), &req, clientIP, userAgent)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token": token,
	})
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	uuid := middleware.GetUUID(c)
	if uuid != "" {
		h.authService.Logout(uuid)
	}
	response.Success(c, nil)
}

// GetInfo 获取用户信息
func (h *AuthHandler) GetInfo(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "登录状态已过期")
		return
	}

	// 从Redis获取登录用户
	loginUser, err := h.authService.GetLoginUser(claims.UUID)
	if err != nil {
		response.Unauthorized(c, "登录状态已过期")
		return
	}

	// 检查是否需要刷新token
	if loginUser.NeedRefresh() {
		h.authService.RefreshToken(loginUser)
	}

	userInfo, err := h.authService.GetUserInfo(c.Request.Context(), loginUser)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"user":        userInfo.User,
		"roles":       userInfo.Roles,
		"permissions": userInfo.Permissions,
	})
}

// GetRouters 获取路由菜单
func (h *AuthHandler) GetRouters(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		response.Unauthorized(c, "登录状态已过期")
		return
	}

	// 从Redis获取登录用户
	loginUser, err := h.authService.GetLoginUser(claims.UUID)
	if err != nil {
		response.Unauthorized(c, "登录状态已过期")
		return
	}

	routers, err := h.authService.GetRouters(c.Request.Context(), loginUser.UserID)
	if err != nil {
		response.Error(c, err.Error())
		return
	}

	response.Success(c, routers)
}

// CaptchaImage 获取验证码
func (h *AuthHandler) CaptchaImage(c *gin.Context) {
	uuid, imgBase64, err := h.authService.GenerateCaptcha()
	if err != nil {
		response.Error(c, "验证码生成失败")
		return
	}

	response.Success(c, dto.CaptchaImageResponse{
		CaptchaEnabled: true,
		UUID:           uuid,
		Img:            imgBase64,
	})
}

// GetPublicKey 获取RSA公钥
func (h *AuthHandler) GetPublicKey(c *gin.Context) {
	pair := crypto.GetRSAKeyPair()
	response.Success(c, gin.H{
		"publicKey":  pair.PublicKey,
		"privateKey": pair.PrivateKey,
	})
}
