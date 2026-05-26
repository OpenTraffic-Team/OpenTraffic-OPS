package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AgentPerceiveHandler 感知Agent API 代理处理器
type AgentPerceiveHandler struct {
	proxy  *httputil.ReverseProxy
	target string
}

// NewAgentPerceiveHandler 创建感知Agent代理处理器
func NewAgentPerceiveHandler(target string) *AgentPerceiveHandler {
	targetURL, err := url.Parse(target)
	if err != nil {
		zap.L().Fatal("Failed to parse agent perceive target URL", zap.Error(err))
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义 Director，修改请求路径和头部
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 将 /api/agent-perceive/* 映射到目标服务的 /*
		// 例如：/api/agent-perceive/api/health -> /api/health
		path := req.URL.Path
		if strings.HasPrefix(path, "/api/agent-perceive/") {
			req.URL.Path = strings.TrimPrefix(path, "/api/agent-perceive")
		}
		if req.URL.RawPath != "" && strings.HasPrefix(req.URL.RawPath, "/api/agent-perceive/") {
			req.URL.RawPath = strings.TrimPrefix(req.URL.RawPath, "/api/agent-perceive")
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
		// 禁用响应缓冲，配合 SSE 流式响应
		DisableCompression: true,
	}

	// SSE 需要逐 chunk 刷新到下游，避免 ReverseProxy 在默认 1s 间隔下聚合
	// FlushInterval = -1 表示每次写入立即 flush
	proxy.FlushInterval = -1

	// 自定义 ModifyResponse，处理响应（SSE 保持原样）
	proxy.ModifyResponse = func(resp *http.Response) error {
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/event-stream") {
			// 显式告知上游中间件/客户端不要缓冲
			resp.Header.Set("Cache-Control", "no-cache")
			resp.Header.Set("Connection", "keep-alive")
			resp.Header.Set("X-Accel-Buffering", "no")
		}
		return nil
	}

	// 自定义 ErrorHandler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		zap.L().Error("Agent perceive error", zap.Error(err), zap.String("path", r.URL.Path))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"code":502,"msg":"感知Agent服务暂不可用","data":null}`))
	}

	return &AgentPerceiveHandler{proxy: proxy, target: target}
}

// RegisterRoutes 注册代理路由
func (h *AgentPerceiveHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 所有 /api/agent-perceive/* 请求都走代理
	r.Any("/api/agent-perceive/*path", h.Proxy)
}

// Proxy 代理请求到外部感知Agent API
func (h *AgentPerceiveHandler) Proxy(c *gin.Context) {
	// 记录代理请求日志
	zap.L().Debug("Agent perceive request",
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("query", c.Request.URL.RawQuery),
	)

	// 对于 SSE 请求，确保头部正确设置
	if c.Request.Header.Get("Accept") == "text/event-stream" {
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("X-Accel-Buffering", "no")
	}

	// 使用 ReverseProxy 转发请求
	h.proxy.ServeHTTP(c.Writer, c.Request)
}
