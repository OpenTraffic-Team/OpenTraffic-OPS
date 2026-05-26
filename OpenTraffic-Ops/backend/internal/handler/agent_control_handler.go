package handler

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AgentControlHandler Agent API 代理处理器
type AgentControlHandler struct {
	proxy  *httputil.ReverseProxy
	target string
}

// NewAgentControlHandler 创建 Agent 代理处理器
func NewAgentControlHandler(target string) *AgentControlHandler {
	targetURL, err := url.Parse(target)
	if err != nil {
		zap.L().Fatal("Failed to parse agent control target URL", zap.Error(err))
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义 Director，修改请求路径和头部
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 将 /api/agent-control/* 映射到目标服务的 /*
		// 例如：/api/agent-control/api/agent/algorithm/analyze-params -> /api/agent/algorithm/analyze-params
		path := req.URL.Path
		if strings.HasPrefix(path, "/api/agent-control/") {
			req.URL.Path = strings.TrimPrefix(path, "/api/agent-control")
		}
		if req.URL.RawPath != "" && strings.HasPrefix(req.URL.RawPath, "/api/agent-control/") {
			req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, "/api/agent-control")
		}

		// 确保 Host 头部正确
		req.Host = targetURL.Host

		// 移除不必要的头部，避免问题
		req.Header.Del("Origin")
		req.Header.Del("Referer")
	}

	// 自定义 Transport，增加超时控制
	proxy.Transport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// 自定义 ModifyResponse，处理响应
	proxy.ModifyResponse = func(resp *http.Response) error {
		// SSE 响应保持原样，不修改
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/event-stream") {
			return nil
		}
		return nil
	}

	// 自定义 ErrorHandler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		zap.L().Error("Agent control error", zap.Error(err), zap.String("path", r.URL.Path))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"code":502,"msg":"Agent服务暂不可用","data":null}`))
	}

	return &AgentControlHandler{proxy: proxy, target: target}
}

// RegisterRoutes 注册代理路由
func (h *AgentControlHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 所有 /api/agent-control/* 请求都走代理
	r.Any("/api/agent-control/*path", h.Proxy)
}

// Proxy 代理请求到外部 Agent API
func (h *AgentControlHandler) Proxy(c *gin.Context) {
	// 记录代理请求日志
	zap.L().Debug("Agent control request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("query", c.Request.URL.RawQuery),
	)

	// 对于 SSE 请求，确保头部正确设置
	if c.Request.Header.Get("Accept") == "text/event-stream" {
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
	}

	// 使用 ReverseProxy 转发请求
	h.proxy.ServeHTTP(c.Writer, c.Request)
}

// ProxyWithBody 手动代理（用于需要自定义处理的场景）
func (h *AgentControlHandler) ProxyWithBody(c *gin.Context) {
	targetURL, err := url.Parse(h.target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "代理目标URL解析失败"})
		return
	}

	// 构建目标URL
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api/agent-control/") {
		path = strings.TrimPrefix(path, "/api/agent-control")
	}

	target := targetURL.String() + path
	if c.Request.URL.RawQuery != "" {
		target += "?" + c.Request.URL.RawQuery
	}

	// 读取请求体
	var bodyReader io.Reader
	if c.Request.Body != nil {
		bodyReader = c.Request.Body
	}

	// 创建新请求
	req, err := http.NewRequest(c.Request.Method, target, bodyReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "创建代理请求失败"})
		return
	}

	// 复制请求头
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Host = targetURL.Host

	// 发送请求
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error("Agent control request failed", zap.Error(err), zap.String("target", target))
		c.JSON(http.StatusBadGateway, gin.H{"code": 502, "msg": "Agent服务请求失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
