package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// WriteWait 写入超时
	WriteWait = 10 * time.Second
	// PongWait pong等待时间（读超时）
	// 缩短到 20 秒，避免 NAT/防火墙在 30-60 秒内断开空闲连接
	PongWait = 20 * time.Second
	// PingPeriod ping发送周期
	// 明确设为 15 秒，确保每 15 秒至少有一次数据流动
	PingPeriod = 15 * time.Second
	// MaxMessageSize 最大消息大小
	MaxMessageSize = 10 * 1024 * 1024 // 10MB
	// SessionIdleTimeout 会话空闲超时
	SessionIdleTimeout = 30 * time.Minute
	// CallbackTimeout 回调超时（必须小于前端 axios 超时，给后端留出包装响应的时间）
	CallbackTimeout = 25 * time.Second
)

// Session 前端WebSocket会话
type Session struct {
	ID        string
	Conn      *websocket.Conn
	AgentIP   string
	UserID    int64
	Username  string
	LastActive time.Time
	Send      chan []byte
	mu        sync.Mutex
}

// HostProxyConn HostProxy WebSocket连接
type HostProxyConn struct {
	IP       string
	Conn     *websocket.Conn
	LastPing time.Time
	Send     chan []byte
	mu       sync.Mutex
}

// Callback 请求-响应回调
type Callback struct {
	RequestID string
	Channel   chan *Message
	CreatedAt time.Time
}

// Hub WebSocket消息桥接中心
type Hub struct {
	// 前端会话映射: sessionID -> Session
	sessions map[string]*Session
	// HostProxy连接映射: ip -> HostProxyConn
	agents map[string]*HostProxyConn
	// 回调映射: requestID -> Callback
	callbacks map[string]*Callback
	// 操作锁
	sessionMu  sync.RWMutex
	agentMu    sync.RWMutex
	callbackMu sync.RWMutex
	// 停止信号
	stopCh chan struct{}
}

var globalHub *Hub
var once sync.Once

// GetHub 获取全局Hub实例（单例）
func GetHub() *Hub {
	once.Do(func() {
		globalHub = &Hub{
			sessions:  make(map[string]*Session),
			agents:    make(map[string]*HostProxyConn),
			callbacks: make(map[string]*Callback),
			stopCh:    make(chan struct{}),
		}
		// 启动清理goroutine
		go globalHub.cleanupLoop()
	})
	return globalHub
}

// RegisterSession 注册前端会话
func (h *Hub) RegisterSession(session *Session) {
	h.sessionMu.Lock()
	defer h.sessionMu.Unlock()

	// 如果已有相同sessionID，先关闭旧连接
	if old, exists := h.sessions[session.ID]; exists {
		closeSession(old)
	}
	h.sessions[session.ID] = session
	zap.L().Info("前端会话已注册",
		zap.String("sessionId", session.ID),
		zap.String("agentIP", session.AgentIP),
		zap.String("username", session.Username))
}

// UnregisterSession 注销前端会话
func (h *Hub) UnregisterSession(sessionID string) {
	h.sessionMu.Lock()
	defer h.sessionMu.Unlock()

	if session, exists := h.sessions[sessionID]; exists {
		closeSession(session)
		delete(h.sessions, sessionID)
		zap.L().Info("前端会话已注销", zap.String("sessionId", sessionID))
	}
}

// GetSession 获取前端会话
func (h *Hub) GetSession(sessionID string) *Session {
	h.sessionMu.RLock()
	defer h.sessionMu.RUnlock()
	return h.sessions[sessionID]
}

// RegisterHostProxy 注册HostProxy连接
func (h *Hub) RegisterHostProxy(agent *HostProxyConn) {
	h.agentMu.Lock()
	defer h.agentMu.Unlock()

	// 如果已有相同IP，先关闭旧连接
	if old, exists := h.agents[agent.IP]; exists {
		closeHostProxyConn(old)
	}
	h.agents[agent.IP] = agent
	zap.L().Info("HostProxy连接已注册", zap.String("ip", agent.IP))
}

// UnregisterHostProxy 注销HostProxy连接
func (h *Hub) UnregisterHostProxy(ip string) {
	h.agentMu.Lock()
	defer h.agentMu.Unlock()

	if agent, exists := h.agents[ip]; exists {
		closeHostProxyConn(agent)
		delete(h.agents, ip)
		zap.L().Info("HostProxy连接已注销", zap.String("ip", ip))
	}
}

// GetHostProxy 获取HostProxy连接
func (h *Hub) GetHostProxy(ip string) *HostProxyConn {
	h.agentMu.RLock()
	defer h.agentMu.RUnlock()
	return h.agents[ip]
}

// IsHostProxyOnline 检查HostProxy是否在线
func (h *Hub) IsHostProxyOnline(ip string) bool {
	h.agentMu.RLock()
	defer h.agentMu.RUnlock()
	agent, exists := h.agents[ip]
	if !exists {
		return false
	}
	return time.Since(agent.LastPing) < PongWait*2
}

// SendToHostProxy 发送消息到HostProxy（使用写锁，防止发送过程中HostProxy连接被替换导致panic）
func (h *Hub) SendToHostProxy(ip string, msg *Message) error {
	h.agentMu.Lock()
	defer h.agentMu.Unlock()

	agent, exists := h.agents[ip]
	if !exists {
		return ErrHostProxyNotConnected
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case agent.Send <- data:
		return nil
	default:
		return ErrHostProxySendBufferFull
	}
}

// SendToSession 发送消息到前端会话
func (h *Hub) SendToSession(sessionID string, msg *Message) error {
	session := h.GetSession(sessionID)
	if session == nil {
		return ErrSessionNotFound
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case session.Send <- data:
		return nil
	default:
		return ErrSessionSendBufferFull
	}
}

// Bridge 桥接消息: 从session转发到HostProxy（终端输入）
func (h *Hub) Bridge(sessionID string, msg *Message) error {
	session := h.GetSession(sessionID)
	if session == nil {
		return ErrSessionNotFound
	}

	// 更新最后活跃时间
	session.LastActive = time.Now()

	return h.SendToHostProxy(session.AgentIP, msg)
}

// BridgeOutput 桥接输出: 从HostProxy转发到session（终端输出）
func (h *Hub) BridgeOutput(sessionID string, msg *Message) error {
	return h.SendToSession(sessionID, msg)
}

// RegisterCallback 注册请求-响应回调
func (h *Hub) RegisterCallback(requestID string) chan *Message {
	ch := make(chan *Message, 1)

	h.callbackMu.Lock()
	defer h.callbackMu.Unlock()

	// 清理过期的回调
	h.cleanupExpiredCallbacksLocked()

	h.callbacks[requestID] = &Callback{
		RequestID: requestID,
		Channel:   ch,
		CreatedAt: time.Now(),
	}
	return ch
}

// UnregisterCallback 注销回调
func (h *Hub) UnregisterCallback(requestID string) {
	h.callbackMu.Lock()
	defer h.callbackMu.Unlock()

	if cb, exists := h.callbacks[requestID]; exists {
		close(cb.Channel)
		delete(h.callbacks, requestID)
	}
}

// DispatchResponse 分发响应到对应回调
func (h *Hub) DispatchResponse(msg *Message) {
	if msg.RequestID == "" {
		return
	}

	h.callbackMu.Lock()
	defer h.callbackMu.Unlock()

	if cb, exists := h.callbacks[msg.RequestID]; exists {
		// 使用非阻塞发送，避免 channel 满或已关闭时阻塞 DispatchResponse
		select {
		case cb.Channel <- msg:
		default:
			zap.L().Warn("回调channel已满或已关闭，响应丢弃",
				zap.String("requestID", msg.RequestID))
		}
		delete(h.callbacks, msg.RequestID)
	}
}

// Request 通过WebSocket发送请求并等待响应（支持ctx取消，当前端断开时立即释放）
func (h *Hub) Request(ctx context.Context, agentIP string, msg *Message) (*Message, error) {
	if msg.RequestID == "" {
		msg.RequestID = generateRequestID()
	}

	ch := h.RegisterCallback(msg.RequestID)
	defer h.UnregisterCallback(msg.RequestID)

	if err := h.SendToHostProxy(agentIP, msg); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(CallbackTimeout):
		zap.L().Warn("HostProxy请求超时",
			zap.String("agentIP", agentIP),
			zap.String("requestID", msg.RequestID),
			zap.String("msgType", string(msg.Type)))
		return nil, ErrCallbackTimeout
	}
}

// cleanupLoop 定期清理过期会话和回调（30秒一次，避免超时回调堆积）
func (h *Hub) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.cleanupSessions()
			h.cleanupExpiredCallbacks()
		case <-h.stopCh:
			return
		}
	}
}

// cleanupSessions 清理空闲会话
func (h *Hub) cleanupSessions() {
	h.sessionMu.Lock()
	defer h.sessionMu.Unlock()

	now := time.Now()
	for id, session := range h.sessions {
		if now.Sub(session.LastActive) > SessionIdleTimeout {
			closeSession(session)
			delete(h.sessions, id)
			zap.L().Info("空闲会话已清理", zap.String("sessionId", id))
		}
	}
}

// cleanupExpiredCallbacks 清理过期回调
func (h *Hub) cleanupExpiredCallbacks() {
	h.callbackMu.Lock()
	defer h.callbackMu.Unlock()
	h.cleanupExpiredCallbacksLocked()
}

func (h *Hub) cleanupExpiredCallbacksLocked() {
	now := time.Now()
	for id, cb := range h.callbacks {
		if now.Sub(cb.CreatedAt) > CallbackTimeout {
			close(cb.Channel)
			delete(h.callbacks, id)
		}
	}
}

// closeSession 安全关闭会话
func closeSession(s *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 使用 recover 防止重复关闭 channel 导致 panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// channel 已被关闭，忽略
			}
		}()
		close(s.Send)
	}()
	s.Conn.Close()
}

// closeHostProxyConn 安全关闭HostProxy连接
func closeHostProxyConn(a *HostProxyConn) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// 使用 recover 防止重复关闭 channel 导致 panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// channel 已被关闭，忽略
			}
		}()
		close(a.Send)
	}()
	a.Conn.Close()
}

// requestIDCounter 原子递增计数器，确保并发时 RequestID 唯一
var requestIDCounter uint64

// generateRequestID 生成全局唯一请求ID（时间戳+原子计数器+随机后缀）
func generateRequestID() string {
	return fmt.Sprintf("%s-%d-%s",
		time.Now().Format("20060102150405.000"),
		atomic.AddUint64(&requestIDCounter, 1),
		randomString(4))
}

// randomString 生成随机字符串（仅作为可读后缀，不保证唯一性）
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
