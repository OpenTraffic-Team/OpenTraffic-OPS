package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"rtm-server/internal/dto"
	"rtm-server/internal/service"
	"rtm-server/internal/utils"
)

// 业务类型常量（与Java BusinessType对应）
const (
	BusinessTypeOther    = 0 // 其它
	BusinessTypeInsert   = 1 // 新增
	BusinessTypeUpdate   = 2 // 修改
	BusinessTypeDelete   = 3 // 删除
	BusinessTypeGrant    = 4 // 授权
	BusinessTypeExport   = 5 // 导出
	BusinessTypeImport   = 6 // 导入
	BusinessTypeForce    = 7 // 强退
	BusinessTypeGencode  = 8 // 生成代码
	BusinessTypeClean    = 9 // 清空数据
)

// 操作人类型常量（与Java OperatorType对应）
const (
	OperatorTypeOther  = 0 // 其它
	OperatorTypeManage = 1 // 后台用户
	OperatorTypeMobile = 2 // 手机端用户
)

// 敏感字段，记录日志时需要排除
var sensitiveFields = []string{"password", "oldPassword", "newPassword", "confirmPassword"}

// OperLogConfig 操作日志配置
type OperLogConfig struct {
	Title         string // 模块标题
	BusinessType  int    // 业务类型
	OperatorType  int    // 操作人类型
	IsSaveRequestData  bool // 是否保存请求参数
	IsSaveResponseData bool // 是否保存响应参数
}

// DefaultOperLogConfig 返回默认配置
func DefaultOperLogConfig(title string, businessType int) OperLogConfig {
	return OperLogConfig{
		Title:              title,
		BusinessType:       businessType,
		OperatorType:       OperatorTypeManage,
		IsSaveRequestData:  true,
		IsSaveResponseData: true,
	}
}

// bodyLogWriter 用于捕获响应体的Writer
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// OperLog 操作日志中间件
// 使用方式: router.POST("/path", middleware.OperLog(middleware.DefaultOperLogConfig("标题", 1), logService), handler)
func OperLog(cfg OperLogConfig, logService *service.OperLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 捕获响应体
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 保存请求体（因为只能读一次）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 继续处理请求
		c.Next()

		// 请求处理完成后记录日志
		duration := time.Since(start).Milliseconds()

		// 只记录成功的操作或发生异常的操作
		// 如果handler中有错误，c.Errors会被设置
		status := 0 // 0=正常
		var errorMsg string
		if len(c.Errors) > 0 {
			status = 1 // 1=异常
			errorMsg = c.Errors.String()
			if len(errorMsg) > 2000 {
				errorMsg = errorMsg[:2000]
			}
		}

		// 获取操作人
		operName := GetUsername(c)

		// 获取请求参数
		var operParam string
		if cfg.IsSaveRequestData {
			operParam = getRequestParam(c, requestBody)
			if len(operParam) > 2000 {
				operParam = operParam[:2000]
			}
		}

		// 获取响应结果
		var jsonResult string
		if cfg.IsSaveResponseData {
			jsonResult = blw.body.String()
			if len(jsonResult) > 2000 {
				jsonResult = jsonResult[:2000]
			}
		}

		// 构建方法名
		handlerName := c.HandlerName()
		if handlerName == "" {
			handlerName = c.Request.URL.Path
		}

		// 创建日志记录
		logReq := &dto.OperLogCreateRequest{
			Title:         cfg.Title,
			BusinessType:  cfg.BusinessType,
			Method:        handlerName,
			RequestMethod: c.Request.Method,
			OperatorType:  cfg.OperatorType,
			OperName:      operName,
			OperURL:       c.Request.URL.Path,
			OperIP:        utils.GetClientIP(c),
			OperParam:     operParam,
			JsonResult:    jsonResult,
			Status:        status,
			ErrorMsg:      errorMsg,
			CostTime:      duration,
		}

		// 异步记录日志（不阻塞响应）
		go func() {
			ctx := c.Request.Context()
			_ = logService.Create(ctx, logReq)
		}()
	}
}

// getRequestParam 获取请求参数
func getRequestParam(c *gin.Context, body []byte) string {
	requestMethod := c.Request.Method
	if requestMethod == "PUT" || requestMethod == "POST" {
		// 从请求体获取参数，并过滤敏感字段
		if len(body) > 0 {
			return filterSensitiveFields(string(body))
		}
		return ""
	}
	// GET/DELETE 从URL查询参数获取
	params := make(map[string]interface{})
	for k, v := range c.Request.URL.Query() {
		if len(v) == 1 {
			params[k] = v[0]
		} else {
			params[k] = v
		}
	}
	if len(params) == 0 {
		return ""
	}
	jsonBytes, _ := json.Marshal(params)
	return filterSensitiveFields(string(jsonBytes))
}

// filterSensitiveFields 过滤敏感字段
func filterSensitiveFields(jsonStr string) string {
	if jsonStr == "" {
		return ""
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// 如果不是JSON，尝试作为数组处理
		var arr []map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &arr); err != nil {
			return jsonStr
		}
		for _, item := range arr {
			for _, field := range sensitiveFields {
				if _, exists := item[field]; exists {
					item[field] = "******"
				}
			}
		}
		filtered, _ := json.Marshal(arr)
		return string(filtered)
	}
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			data[field] = "******"
		}
	}
	filtered, _ := json.Marshal(data)
	return string(filtered)
}

// IsOperLogSkipPath 判断路径是否跳过操作日志记录
func IsOperLogSkipPath(path string) bool {
	skipPaths := []string{
		"/ping",
		"/login",
		"/logout",
		"/captchaImage",
		"/refresh",
		"/getInfo",
		"/getRouters",
	}
	for _, p := range skipPaths {
		if strings.HasSuffix(path, p) {
			return true
		}
	}
	return false
}
