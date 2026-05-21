package wsclient

import "encoding/json"

// MessageType WebSocket消息类型
type MessageType string

const (
	// 心跳相关
	TypePing MessageType = "ping" // 心跳探测
	TypePong MessageType = "pong" // 心跳响应

	// 终端相关
	TypeInput  MessageType = "input"  // 前端 -> HostProxy: 终端输入
	TypeOutput MessageType = "output" // HostProxy -> 前端: 终端输出
	TypeResize MessageType = "resize" // 前端 -> HostProxy: 调整终端大小

	// 文件操作相关
	TypeFileList     MessageType = "file_list"     // 列出目录
	TypeFileRead     MessageType = "file_read"     // 读取文件
	TypeFileWrite    MessageType = "file_write"    // 写入文件
	TypeFileDelete   MessageType = "file_delete"   // 删除文件
	TypeFileUpload   MessageType = "file_upload"   // 上传文件
	TypeFileDownload MessageType = "file_download" // 下载文件
	TypeFileMkdir    MessageType = "file_mkdir"    // 创建目录

	// 响应类型
	TypeResponse MessageType = "response" // 通用响应（带 requestId）
	TypeError    MessageType = "error"    // 错误消息
)

// Message WebSocket通用消息结构（与后端兼容）
type Message struct {
	Type      MessageType     `json:"type"`                // 消息类型
	Data      string          `json:"data,omitempty"`      // 文本数据（如终端输入/输出）
	SessionID string          `json:"sessionId,omitempty"` // 终端会话ID
	RequestID string          `json:"requestId,omitempty"` // 请求ID（用于请求-响应模式）
	Cols      int             `json:"cols,omitempty"`      // 终端列数
	Rows      int             `json:"rows,omitempty"`      // 终端行数
	Payload   json.RawMessage `json:"payload,omitempty"`   // JSON结构化数据（文件操作等）
	Success   bool            `json:"success,omitempty"`   // 操作是否成功
	Error     string          `json:"error,omitempty"`     // 错误信息
}

// FileListPayload 文件列表请求/响应载荷
type FileListPayload struct {
	Path  string     `json:"path"`            // 目录路径
	Files []FileInfo `json:"files,omitempty"` // 文件列表
}

// FileReadPayload 文件读取请求/响应载荷
type FileReadPayload struct {
	Path    string `json:"path"`              // 文件路径
	Content string `json:"content,omitempty"` // Base64编码的文件内容
	Size    int64  `json:"size,omitempty"`    // 文件大小
}

// FileWritePayload 文件写入请求载荷
type FileWritePayload struct {
	Path    string `json:"path"`    // 文件路径
	Content string `json:"content"` // Base64编码的内容
}

// FileDeletePayload 文件删除请求载荷
type FileDeletePayload struct {
	Path string `json:"path"` // 文件/目录路径
}

// FileUploadPayload 文件上传请求载荷
type FileUploadPayload struct {
	Path     string `json:"path"`     // 目标目录
	FileName string `json:"fileName"` // 文件名
	Content  string `json:"content"`  // Base64编码的文件内容
}

// FileDownloadPayload 文件下载请求载荷
type FileDownloadPayload struct {
	Path string `json:"path"` // 文件路径
}

// FileMkdirPayload 创建目录请求载荷
type FileMkdirPayload struct {
	Path string `json:"path"` // 目录路径
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string `json:"name"`              // 文件名
	Path    string `json:"path"`              // 完整路径
	Size    int64  `json:"size"`              // 文件大小
	IsDir   bool   `json:"isDir"`             // 是否为目录
	ModTime int64  `json:"modTime"`           // 修改时间（Unix秒）
	Mode    string `json:"mode,omitempty"`    // 权限模式
}
