package model

import "time"

// DeployStatus 部署状态
type DeployStatus string

const (
	DeployStatusPending DeployStatus = "pending"
	DeployStatusSuccess DeployStatus = "success"
	DeployStatusFailed  DeployStatus = "failed"
)

// DeployRecord 部署记录
type DeployRecord struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ServerID       string    `gorm:"not null" json:"server_id"`
	ServerName     string    `gorm:"not null" json:"server_name"`
	BinaryName     string    `gorm:"not null" json:"binary_name"`        // opentraffic-ops-proxy / opentraffic-ops
	RemotePath     string    `json:"remote_path"`                         // 远程文件路径
	Version        string    `json:"version"`                             // 部署版本（algo_md 等可重复部署资源使用）
	Status         string    `gorm:"not null" json:"status"`              // pending / success / failed
	Log            string    `gorm:"type:text" json:"log"`                // 部署日志
	CreatedAt      time.Time `json:"created_at"`
}

// TableName 指定表名
func (DeployRecord) TableName() string {
	return "deploy_records"
}
