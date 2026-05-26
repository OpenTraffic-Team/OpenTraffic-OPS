package model

// ChatSession Agent聊天会话模型 bu_chat_session
type ChatSession struct {
	ID             int64  `json:"id" gorm:"primaryKey;column:id;comment:主键"`
	UserID         int64  `json:"userId" gorm:"column:user_id;comment:所属用户ID"`
	SessionType    string `json:"sessionType" gorm:"column:session_type;size:16;default:'control';comment:会话类型 control/perceive"`
	Title          string `json:"title" gorm:"column:title;size:128;comment:会话标题"`
	AgentSessionID string `json:"agentSessionId" gorm:"column:agent_session_id;size:128;comment:外部Agent服务session_id"`
	LastMessageAt  string `json:"lastMessageAt" gorm:"column:last_message_at;type:timestamp;comment:最近一次消息时间"`
	MessageCount   int    `json:"messageCount" gorm:"column:message_count;default:0;comment:累计消息条数"`
	BaseEntity
}

func (ChatSession) TableName() string {
	return "bu_chat_session"
}
