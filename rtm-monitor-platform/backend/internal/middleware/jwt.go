package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"rtm-server/internal/config"
	"rtm-server/pkg/jwt"
	"rtm-server/pkg/response"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取token
		authHeader := c.GetHeader(config.GlobalConfig.JWT.Header)
		if authHeader == "" {
			response.Unauthorized(c, "令牌不能为空")
			c.Abort()
			return
		}

		// 提取token（Bearer前缀）
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "令牌格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析token
		claims, err := jwt.ParseToken(tokenString, config.GlobalConfig.JWT.Secret)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				response.Unauthorized(c, "登录状态已过期")
			} else {
				response.Unauthorized(c, "令牌验证失败")
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("uuid", claims.UUID)
		c.Set("claims", claims)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get("userId")
	if !exists {
		return 0
	}
	if id, ok := userID.(int64); ok {
		return id
	}
	return 0
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	if name, ok := username.(string); ok {
		return name
	}
	return ""
}

// GetUUID 从上下文获取UUID
func GetUUID(c *gin.Context) string {
	uuid, exists := c.Get("uuid")
	if !exists {
		return ""
	}
	if id, ok := uuid.(string); ok {
		return id
	}
	return ""
}

// GetClaims 从上下文获取JWT声明
func GetClaims(c *gin.Context) *jwt.Claims {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}
	if c, ok := claims.(*jwt.Claims); ok {
		return c
	}
	return nil
}
