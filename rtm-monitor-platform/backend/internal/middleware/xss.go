package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"rtm-server/internal/config"
	"rtm-server/internal/utils"
)

// XSS XSS过滤中间件
func XSS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.GlobalConfig.XSS.Enabled {
			c.Next()
			return
		}

		// 检查是否在排除列表中
		path := c.Request.URL.Path
		for _, exclude := range config.GlobalConfig.XSS.Excludes {
			if strings.HasPrefix(path, exclude) {
				c.Next()
				return
			}
		}

		// 检查是否匹配URL模式
		matched := false
		for _, pattern := range config.GlobalConfig.XSS.URLPatterns {
			if pattern == path {
				matched = true
				break
			}
			if strings.HasSuffix(pattern, "/*") {
				prefix := strings.TrimSuffix(pattern, "/*")
				if strings.HasPrefix(path, prefix) {
					matched = true
					break
				}
			}
		}

		if !matched {
			c.Next()
			return
		}

		// 对POST/PUT/PATCH请求的请求体进行XSS过滤
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodPatch {
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") ||
				strings.Contains(contentType, "application/x-www-form-urlencoded") {

				body, err := io.ReadAll(c.Request.Body)
				if err == nil && len(body) > 0 {
					filtered := utils.EscapeHtml(string(body))
					c.Request.Body = io.NopCloser(bytes.NewBufferString(filtered))
					c.Request.ContentLength = int64(len(filtered))
				}
			}
		}

		c.Next()
	}
}
