package dto

// ChatSessionQuery 聊天会话查询条件（分页）
type ChatSessionQuery struct {
	PageQuery
	SessionType string `form:"sessionType" json:"sessionType"`
}

// ChatSessionDto 聊天会话列表项
type ChatSessionDto struct {
	ID             int64  `json:"id"`
	SessionType    string `json:"sessionType"`
	Title          string `json:"title"`
	AgentSessionID string `json:"agentSessionId"`
	LastMessageAt  string `json:"lastMessageAt"`
	MessageCount   int    `json:"messageCount"`
	CreateTime     string `json:"createTime"`
	UpdateTime     string `json:"updateTime"`
}

// ChatMessageDto 聊天消息项
type ChatMessageDto struct {
	ID         int64  `json:"id"`
	Role       string `json:"role"`
	Content    string `json:"content"`
	Seq        int    `json:"seq"`
	CreateTime string `json:"createTime"`
}

// ChatSessionDetailDto 聊天会话详情（含消息）
type ChatSessionDetailDto struct {
	ChatSessionDto
	Messages []ChatMessageDto `json:"messages"`
}

// ChatTurnRequest 一轮对话持久化请求
// sessionId 为 0 时表示新建会话；否则向已有会话追加
// sessionType 仅在新建会话时生效，取值 control/perceive，缺省为 control
type ChatTurnRequest struct {
	SessionID        int64  `json:"sessionId"`
	SessionType      string `json:"sessionType"`
	AgentSessionID   string `json:"agentSessionId"`
	UserMessage      string `json:"userMessage" binding:"required"`
	AssistantMessage string `json:"assistantMessage" binding:"required"`
}

// ChatTurnResponse 一轮对话持久化响应
type ChatTurnResponse struct {
	SessionID     int64  `json:"sessionId"`
	Title         string `json:"title"`
	LastMessageAt string `json:"lastMessageAt"`
	MessageCount  int    `json:"messageCount"`
}

// ChatSessionRenameRequest 重命名会话请求
type ChatSessionRenameRequest struct {
	ID    int64  `json:"id" binding:"required"`
	Title string `json:"title" binding:"required"`
}
