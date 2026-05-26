package wsclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"opentraffic-ops-proxy/config"
	"opentraffic-ops-proxy/executor"
	"opentraffic-ops-proxy/filemanager"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 20 * time.Second // 读超时 20 秒，与后端一致
	pingPeriod     = 15 * time.Second // 明确 15 秒 ping 周期，防止 NAT/防火墙断开
	maxMessageSize = 10 * 1024 * 1024 // 10MB
	reconnectMin   = 5 * time.Second
	reconnectMax   = 5 * time.Minute
)

// Client WebSocket客户端
type Client struct {
	cfg      *config.Config
	ip       string
	conn     *websocket.Conn
	send     chan []byte
	stopCh   chan struct{}
	doneCh   chan struct{} // 通知连接断开，触发重连
	wg       sync.WaitGroup // 跟踪读写 goroutine，确保旧 goroutine 完全退出后再重连
	mu       sync.RWMutex
	// 终端会话映射: sessionId -> *executor.ShellExecutor
	sessions map[string]*executor.ShellExecutor
	sessMu   sync.RWMutex
	// 文件管理器
	fileMgr *filemanager.Manager
}

// New 创建WebSocket客户端
func New(cfg *config.Config, ip string) *Client {
	return &Client{
		cfg:      cfg,
		ip:       ip,
		send:     make(chan []byte, 256),
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}, 1), // 缓冲 1，避免 readPump 与 connectLoop 的竞争
		sessions: make(map[string]*executor.ShellExecutor),
		fileMgr:  filemanager.New(),
	}
}

// Start 启动WebSocket客户端（包含自动重连）
func (c *Client) Start() {
	if !c.cfg.EnableRemote {
		log.Println("[WS] 远程控制已禁用，不启动WebSocket客户端")
		return
	}

	go c.connectLoop()
}

// Stop 停止WebSocket客户端
func (c *Client) Stop() {
	select {
	case <-c.stopCh:
		// 已经关闭
	default:
		close(c.stopCh)
	}
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()
	// 等待读写 goroutine 退出
	c.wg.Wait()
}

// IsConnected 是否已连接
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}

// connectLoop 连接循环（自动重连）
// 核心设计：使用 sync.WaitGroup 确保旧读写 goroutine 完全退出后再建立新连接，
// 杜绝新旧 goroutine 竞争关闭同一连接的问题。
func (c *Client) connectLoop() {
	backoff := reconnectMin

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		if err := c.connect(); err != nil {
			log.Printf("[WS] 连接失败: %v, %v后重试", err, backoff)
			select {
			case <-time.After(backoff):
				// 指数退避
				backoff *= 2
				if backoff > reconnectMax {
					backoff = reconnectMax
				}
				continue
			case <-c.stopCh:
				return
			}
		}

		// 连接成功，重置退避时间
		backoff = reconnectMin

		// 等待连接断开或停止信号
		select {
		case <-c.doneCh:
			log.Println("[WS] 收到断开信号，等待旧 goroutine 退出...")
			c.wg.Wait() // 关键：确保读写 goroutine 完全退出
			log.Println("[WS] 旧 goroutine 已退出，准备重连")

			// 创建新的 doneCh
			c.mu.Lock()
			select {
			case <-c.doneCh:
			default:
			}
			c.doneCh = make(chan struct{}, 1)
			c.mu.Unlock()

			// 短暂等待后重连
			select {
			case <-time.After(1 * time.Second):
			case <-c.stopCh:
				return
			}
		case <-c.stopCh:
			return
		}
	}
}

// connect 建立WebSocket连接
func (c *Client) connect() error {
	wsEndpoint := c.cfg.WSEndpoint
	if wsEndpoint == "" {
		// 默认使用platformURL的ws://协议
		u, err := url.Parse(c.cfg.PlatformURL)
		if err != nil {
			return fmt.Errorf("解析平台URL失败: %w", err)
		}
		scheme := "ws"
		if u.Scheme == "https" {
			scheme = "wss"
		}
		wsEndpoint = fmt.Sprintf("%s://%s/api/v1/proxy/ws", scheme, u.Host)
	}

	u, err := url.Parse(wsEndpoint)
	if err != nil {
		return fmt.Errorf("解析WS端点失败: %w", err)
	}

	q := u.Query()
	q.Set("ip", c.ip)
	u.RawQuery = q.Encode()

	log.Printf("[WS] 正在连接: %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial失败: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	log.Printf("[WS] 连接成功")

	// 启动读写goroutine
	c.wg.Add(2)
	go func() {
		defer c.wg.Done()
		c.writePump(conn)
	}()
	go func() {
		defer c.wg.Done()
		c.readPump(conn)
	}()

	return nil
}

// readPump 读泵
// 职责：读取消息、处理 pong、检测连接断开。
// 只有 readPump 负责关闭连接和通知 connectLoop 重连。
func (c *Client) readPump(conn *websocket.Conn) {
	defer func() {
		// 只由 readPump 关闭连接，避免与 writePump 竞争
		conn.Close()
		c.mu.Lock()
		if c.conn == conn {
			c.conn = nil
		}
		c.mu.Unlock()
		// 通知 connectLoop 需要重连
		c.mu.Lock()
		select {
		case c.doneCh <- struct{}{}:
		default:
		}
		c.mu.Unlock()
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] 连接异常关闭: %v", err)
			}
			break
		}

		// 处理消息
		go c.handleMessage(data)
	}
}

// writePump 写泵
// 职责：从 send channel 读取消息并写入连接、定时发送原生 ping。
// writePump 不拥有连接，不调用 Close，只负责写入。
// 当写入失败时，直接退出，由 readPump 检测到连接关闭后统一处理。
func (c *Client) writePump(conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		// 注意：writePump 不关闭连接！关闭由 readPump 统一负责
	}()

	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				// send channel 被关闭，优雅退出
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[WS] 写入消息失败: %v", err)
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[WS] 发送 ping 失败: %v", err)
				return
			}
		}
	}
}

// handleMessage 处理收到的消息
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[WS] 消息解析失败: %v", err)
		return
	}

	// 检查远程控制是否启用
	if !c.cfg.EnableRemote {
		c.sendError(msg.SessionID, msg.RequestID, "远程控制已禁用")
		return
	}

	switch msg.Type {
	case TypePing:
		// 收到心跳探测，回复 pong
		c.sendMessage(&Message{Type: TypePong})
	case TypeInput:
		c.handleInput(&msg)
	case TypeResize:
		c.handleResize(&msg)
	case TypeFileList:
		c.handleFileList(&msg)
	case TypeFileRead:
		c.handleFileRead(&msg)
	case TypeFileWrite:
		c.handleFileWrite(&msg)
	case TypeFileDelete:
		c.handleFileDelete(&msg)
	case TypeFileUpload:
		c.handleFileUpload(&msg)
	case TypeFileDownload:
		c.handleFileDownload(&msg)
	case TypeFileMkdir:
		c.handleFileMkdir(&msg)
	default:
		log.Printf("[WS] 未知消息类型: %s", msg.Type)
	}
}

// handleInput 处理终端输入
func (c *Client) handleInput(msg *Message) {
	session := c.getOrCreateSession(msg.SessionID)
	if session == nil {
		c.sendError(msg.SessionID, msg.RequestID, "无法创建终端会话")
		return
	}

	if err := session.Write([]byte(msg.Data)); err != nil {
		log.Printf("[WS] 写入shell失败: %v", err)
		c.sendError(msg.SessionID, msg.RequestID, "写入shell失败: "+err.Error())
	}
}

// handleResize 处理终端大小调整
func (c *Client) handleResize(msg *Message) {
	// 使用 getOrCreateSession，确保 shell 已创建（首次 resize 时自动创建）
	session := c.getOrCreateSession(msg.SessionID)
	if session == nil {
		return
	}

	if err := session.Resize(msg.Cols, msg.Rows); err != nil {
		log.Printf("[WS] resize失败: %v", err)
	}
}

// getOrCreateSession 获取或创建终端会话
func (c *Client) getOrCreateSession(sessionID string) *executor.ShellExecutor {
	c.sessMu.Lock()
	defer c.sessMu.Unlock()

	if session, exists := c.sessions[sessionID]; exists {
		if session.IsRunning() {
			return session
		}
		// 会话已结束，删除旧会话
		delete(c.sessions, sessionID)
	}

	// 创建新会话
	session, err := executor.NewShellExecutor()
	if err != nil {
		log.Printf("[WS] 创建shell失败: %v", err)
		return nil
	}

	c.sessions[sessionID] = session

	// 启动输出转发goroutine
	go c.forwardOutput(sessionID, session)

	return session
}

// getSession 获取终端会话
func (c *Client) getSession(sessionID string) *executor.ShellExecutor {
	c.sessMu.RLock()
	defer c.sessMu.RUnlock()
	return c.sessions[sessionID]
}

// forwardOutput 转发shell输出到WebSocket（PTY模式：单一输出流）
func (c *Client) forwardOutput(sessionID string, session *executor.ShellExecutor) {
	buf := make([]byte, 4096)

	// PTY已合并stdout和stderr，只需一个读取循环
	for {
		n, err := session.Read(buf)
		if err != nil {
			// EOF 和 input/output error 都是 PTY 正常关闭的表现，不需要打印错误日志
			if err.Error() != "EOF" && !strings.Contains(err.Error(), "input/output error") {
				log.Printf("[WS] PTY读取错误: %v", err)
			}
			return
		}
		if n > 0 {
			msg := &Message{
				Type:      TypeOutput,
				Data:      string(buf[:n]),
				SessionID: sessionID,
			}
			c.sendMessage(msg)
		}
	}
}

// sendMessage 发送消息（带5秒超时阻塞，避免静默丢弃响应导致后端永远等待）
func (c *Client) sendMessage(msg *Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WS] 消息编码失败: %v", err)
		return
	}

	select {
	case c.send <- data:
	case <-time.After(5 * time.Second):
		log.Printf("[WS] 发送缓冲区已满，消息丢弃: type=%s", msg.Type)
	}
}

// sendResponse 发送响应
func (c *Client) sendResponse(requestID string, payload interface{}, success bool) {
	payloadData, _ := json.Marshal(payload)
	msg := &Message{
		Type:      TypeResponse,
		RequestID: requestID,
		Success:   success,
		Payload:   payloadData,
	}
	c.sendMessage(msg)
}

// sendError 发送错误响应
func (c *Client) sendError(sessionID, requestID, errMsg string) {
	msg := &Message{
		Type:      TypeError,
		SessionID: sessionID,
		RequestID: requestID,
		Error:     errMsg,
	}
	c.sendMessage(msg)
}

// --- 文件操作处理器 ---

func (c *Client) handleFileList(msg *Message) {
	var req FileListPayload
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.sendError("", msg.RequestID, "解析请求失败: "+err.Error())
		return
	}

	files, err := c.fileMgr.List(req.Path)
	if err != nil {
		c.sendError("", msg.RequestID, err.Error())
		return
	}

	// 类型转换
	wsFiles := make([]FileInfo, len(files))
	for i, f := range files {
		wsFiles[i] = FileInfo{
			Name:    f.Name,
			Path:    f.Path,
			Size:    f.Size,
			IsDir:   f.IsDir,
			ModTime: f.ModTime,
			Mode:    f.Mode,
		}
	}

	resp := FileListPayload{
		Path:  req.Path,
		Files: wsFiles,
	}
	c.sendResponse(msg.RequestID, resp, true)
}

func (c *Client) handleFileRead(msg *Message) {
	var req FileReadPayload
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.sendError("", msg.RequestID, "解析请求失败: "+err.Error())
		return
	}

	content, size, err := c.fileMgr.Read(req.Path)
	if err != nil {
		c.sendError("", msg.RequestID, err.Error())
		return
	}

	resp := FileReadPayload{
		Path:    req.Path,
		Content: content,
		Size:    size,
	}
	c.sendResponse(msg.RequestID, resp, true)
}

func (c *Client) handleFileWrite(msg *Message) {
	var req FileWritePayload
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.sendError("", msg.RequestID, "解析请求失败: "+err.Error())
		return
	}

	if err := c.fileMgr.Write(req.Path, req.Content); err != nil {
		c.sendError("", msg.RequestID, err.Error())
		return
	}

	c.sendResponse(msg.RequestID, map[string]string{"path": req.Path}, true)
}

func (c *Client) handleFileDelete(msg *Message) {
	var req FileDeletePayload
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.sendError("", msg.RequestID, "解析请求失败: "+err.Error())
		return
	}

	if err := c.fileMgr.Delete(req.Path); err != nil {
		c.sendError("", msg.RequestID, err.Error())
		return
	}

	c.sendResponse(msg.RequestID, map[string]string{"path": req.Path}, true)
}

func (c *Client) handleFileUpload(msg *Message) {
	var req FileUploadPayload
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.sendError("", msg.RequestID, "解析请求失败: "+err.Error())
		return
	}

	if err := c.fileMgr.Upload(req.Path, req.FileName, req.Content); err != nil {
		c.sendError("", msg.RequestID, err.Error())
		return
	}

	c.sendResponse(msg.RequestID, map[string]string{"path": req.Path, "fileName": req.FileName}, true)
}

func (c *Client) handleFileDownload(msg *Message) {
	var req FileDownloadPayload
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.sendError("", msg.RequestID, "解析请求失败: "+err.Error())
		return
	}

	content, size, err := c.fileMgr.Download(req.Path)
	if err != nil {
		c.sendError("", msg.RequestID, err.Error())
		return
	}

	resp := FileReadPayload{
		Path:    req.Path,
		Content: content,
		Size:    size,
	}
	c.sendResponse(msg.RequestID, resp, true)
}

func (c *Client) handleFileMkdir(msg *Message) {
	var req FileMkdirPayload
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.sendError("", msg.RequestID, "解析请求失败: "+err.Error())
		return
	}

	if err := c.fileMgr.Mkdir(req.Path); err != nil {
		c.sendError("", msg.RequestID, err.Error())
		return
	}

	c.sendResponse(msg.RequestID, map[string]string{"path": req.Path}, true)
}

// getHomeDir 获取home目录（用于默认路径）
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return home
}
