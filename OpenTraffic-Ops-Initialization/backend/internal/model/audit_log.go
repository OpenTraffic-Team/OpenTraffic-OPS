package model

import "time"

// AuditLog 操作日志
type AuditLog struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Action       string    `gorm:"not null" json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Details      string    `gorm:"type:text" json:"details"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}
