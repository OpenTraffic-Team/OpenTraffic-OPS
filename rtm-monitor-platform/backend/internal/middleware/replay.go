package middleware

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"rtm-server/pkg/cache"
	"rtm-server/pkg/response"
)

// Replay 防重放攻击中间件
// 基于时间戳和请求签名的简单防重放机制
func Replay(redisCache *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对非GET请求进行防重放检查
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		timestamp := c.GetHeader("X-Timestamp")
		signature := c.GetHeader("X-Signature")

		// 如果没有签名头，跳过检查（兼容旧客户端）
		if timestamp == "" || signature == "" {
			c.Next()
			return
		}

		// 检查时间戳是否过期（5分钟有效期）
		ts, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			response.Error(c, "Invalid timestamp format")
			c.Abort()
			return
		}

		if time.Since(ts).Abs() > 5*time.Minute {
			response.Error(c, "Request expired")
			c.Abort()
			return
		}

		// 检查签名是否已使用
		key := fmt.Sprintf("replay:%x", md5.Sum([]byte(signature)))
		exists, _ := redisCache.Exists(key)
		if exists > 0 {
			response.Error(c, "Duplicate request detected")
			c.Abort()
			return
		}

		// 标记签名已使用（10分钟过期）
		redisCache.Set(key, "1", 10*time.Minute)

		c.Next()
	}
}

// BuildSignature 构建请求签名（客户端使用）
func BuildSignature(method, path, body, secret string, timestamp int64) string {
	data := fmt.Sprintf("%s|%s|%s|%d|%s", strings.ToUpper(method), path, body, timestamp, secret)
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

// readBody 读取请求体
func readBody(c *gin.Context) string {
	body, _ := io.ReadAll(c.Request.Body)
	return string(body)
}
