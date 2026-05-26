package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"opentraffic-ops-backend/internal/config"
	"opentraffic-ops-backend/pkg/jwt"
)

// WSAuth WebSocket JWT认证中间件
// 从 query param "token" 提取并校验 JWT
func WSAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "令牌不能为空"})
			return
		}

		// 提取token（Bearer前缀）
		parts := strings.SplitN(tokenString, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			tokenString = parts[1]
		}

		// 解析token
		claims, err := jwt.ParseToken(tokenString, config.GlobalConfig.JWT.Secret)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "登录状态已过期"})
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "令牌验证失败"})
			}
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
