package model

// ChatMessage Agent聊天消息模型 bu_chat_message
// 追加写，无 BaseEntity；删除会话时随会话一并硬删除
type ChatMessage struct {
	ID         int64  `json:"id" gorm:"primaryKey;column:id;comment:主键"`
	SessionID  int64  `json:"sessionId" gorm:"column:session_id;comment:所属会话ID"`
	Role       string `json:"role" gorm:"column:role;size:16;comment:消息角色 user/assistant"`
	Content    string `json:"content" gorm:"column:content;type:text;comment:消息内容（markdown）"`
	Seq        int    `json:"seq" gorm:"column:seq;default:0;comment:会话内顺序号"`
	CreateTime string `json:"createTime" gorm:"column:create_time;type:timestamp;comment:入库时间"`
}

func (ChatMessage) TableName() string {
	return "bu_chat_message"
}
