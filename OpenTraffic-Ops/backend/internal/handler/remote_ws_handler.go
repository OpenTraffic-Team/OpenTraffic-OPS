package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"opentraffic-ops-backend/internal/middleware"
	"opentraffic-ops-backend/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域
	},
}

// HandleTerminalWS 前端终端WebSocket连接
func HandleTerminalWS(c *gin.Context) {
	// JWT已在中间件中校验，从上下文获取用户信息
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	hostIP := c.Query("hostIp")
	if hostIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "hostIp不能为空"})
		return
	}

	// 检查目标HostProxy是否在线
	hub := ws.GetHub()
	if !hub.IsHostProxyOnline(hostIP) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "目标主机不在线"})
		return
	}

	// 升级WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("WebSocket升级失败", zap.Error(err))
		return
	}

	// 生成sessionID
	sessionID := generateSessionID()

	// 创建会话
	session := &ws.Session{
		ID:         sessionID,
		Conn:       conn,
		AgentIP:    hostIP,
		UserID:     userID,
		Username:   username,
		LastActive: time.Now(),
		Send:       make(chan []byte, 256),
	}

	// 注册到Hub
	hub.RegisterSession(session)
	defer hub.UnregisterSession(sessionID)

	// 启动读写goroutine
	go sessionWritePump(session)
	sessionReadPump(session, hub)
}

// HandleHostProxyWS HostProxy WebSocket连接
func HandleHostProxyWS(c *gin.Context) {
	hostProxyIP := c.Query("ip")
	if hostProxyIP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "ip不能为空"})
		return
	}

	// 升级WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("HostProxy WebSocket升级失败", zap.Error(err))
		return
	}

	zap.L().Info("HostProxy WebSocket连接已建立", zap.String("ip", hostProxyIP))

	// 创建HostProxy连接
	hostProxy := &ws.HostProxyConn{
		IP:       hostProxyIP,
		Conn:     conn,
		LastPing: time.Now(),
		Send:     make(chan []byte, 256),
	}

	// 注册到Hub
	hub := ws.GetHub()
	hub.RegisterHostProxy(hostProxy)
	defer func() {
		hub.UnregisterHostProxy(hostProxyIP)
		zap.L().Info("HostProxy WebSocket连接已关闭", zap.String("ip", hostProxyIP))
	}()

	// 启动读写goroutine
	go hostProxyWritePump(hostProxy)
	hostProxyReadPump(hostProxy, hub)
}

// sessionReadPump 前端会话读泵
func sessionReadPump(session *ws.Session, hub *ws.Hub) {
	defer func() {
		hub.UnregisterSession(session.ID)
	}()

	session.Conn.SetReadLimit(ws.MaxMessageSize)
	session.Conn.SetReadDeadline(time.Now().Add(ws.PongWait))
	session.Conn.SetPongHandler(func(string) error {
		session.Conn.SetReadDeadline(time.Now().Add(ws.PongWait))
		return nil
	})

	for {
		_, data, err := session.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.L().Warn("前端连接异常关闭",
					zap.String("sessionId", session.ID),
					zap.Error(err))
			}
			break
		}

		// 解析消息
		var msg ws.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			zap.L().Warn("前端消息解析失败",
				zap.String("sessionId", session.ID),
				zap.Error(err))
			continue
		}

		// 确保sessionID正确
		msg.SessionID = session.ID

		// 桥接到HostProxy
		if err := hub.Bridge(session.ID, &msg); err != nil {
			zap.L().Warn("消息桥接失败",
				zap.String("sessionId", session.ID),
				zap.Error(err))

			// 发送错误消息回前端
			errMsg := &ws.Message{
				Type:      ws.TypeError,
				Error:     err.Error(),
				SessionID: session.ID,
			}
			hub.SendToSession(session.ID, errMsg)
		}
	}
}

// sessionWritePump 前端会话写泵
func sessionWritePump(session *ws.Session) {
	ticker := time.NewTicker(ws.PingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case data, ok := <-session.Send:
			if !ok {
				session.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			session.Conn.SetWriteDeadline(time.Now().Add(ws.WriteWait))
			if err := session.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			session.Conn.SetWriteDeadline(time.Now().Add(ws.WriteWait))
			if err := session.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// hostProxyReadPump HostProxy读泵
func hostProxyReadPump(hostProxy *ws.HostProxyConn, hub *ws.Hub) {
	defer func() {
		hub.UnregisterHostProxy(hostProxy.IP)
	}()

	hostProxy.Conn.SetReadLimit(ws.MaxMessageSize)
	hostProxy.Conn.SetReadDeadline(time.Now().Add(ws.PongWait))
	hostProxy.Conn.SetPongHandler(func(string) error {
		hostProxy.LastPing = time.Now()
		hostProxy.Conn.SetReadDeadline(time.Now().Add(ws.PongWait))
		return nil
	})

	for {
		_, data, err := hostProxy.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				zap.L().Warn("HostProxy连接异常关闭",
					zap.String("ip", hostProxy.IP),
					zap.Error(err))
			}
			break
		}

		// 更新最后ping时间
		hostProxy.LastPing = time.Now()

		// 解析消息
		var msg ws.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			zap.L().Warn("HostProxy消息解析失败",
				zap.String("ip", hostProxy.IP),
				zap.Error(err))
			continue
		}

		// 处理心跳探测（JSON ping/pong）
		if msg.Type == ws.TypePing {
			pong := &ws.Message{Type: ws.TypePong}
			pongData, _ := json.Marshal(pong)
			hostProxy.Conn.SetWriteDeadline(time.Now().Add(ws.WriteWait))
			if err := hostProxy.Conn.WriteMessage(websocket.TextMessage, pongData); err != nil {
				zap.L().Warn("HostProxy pong发送失败", zap.String("ip", hostProxy.IP), zap.Error(err))
			}
			zap.L().Debug("HostProxy JSON ping/pong", zap.String("ip", hostProxy.IP))
			continue
		}
		if msg.Type == ws.TypePong {
			// JSON pong 只更新 LastPing，原生 pong 已由 SetPongHandler 处理
			continue
		}

		// 处理带requestId的响应消息（文件操作等）
		if msg.RequestID != "" {
			hub.DispatchResponse(&msg)
			continue
		}

		// 终端输出消息：桥接到对应session
		if msg.Type == ws.TypeOutput && msg.SessionID != "" {
			if err := hub.BridgeOutput(msg.SessionID, &msg); err != nil {
				zap.L().Warn("输出桥接失败",
					zap.String("sessionId", msg.SessionID),
					zap.Error(err))
			}
			continue
		}

		// 其他类型消息：直接转发到对应session
		if msg.SessionID != "" {
			if err := hub.SendToSession(msg.SessionID, &msg); err != nil {
				zap.L().Warn("消息转发到session失败",
					zap.String("sessionId", msg.SessionID),
					zap.Error(err))
			}
		}
	}
}

// hostProxyWritePump HostProxy写泵
// 不拥有连接，不调用 Close。写入失败时直接退出，由 hostProxyReadPump 统一关闭连接。
func hostProxyWritePump(hostProxy *ws.HostProxyConn) {
	ticker := time.NewTicker(ws.PingPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case data, ok := <-hostProxy.Send:
			if !ok {
				hostProxy.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			hostProxy.Conn.SetWriteDeadline(time.Now().Add(ws.WriteWait))
			if err := hostProxy.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			hostProxy.Conn.SetWriteDeadline(time.Now().Add(ws.WriteWait))
			if err := hostProxy.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
