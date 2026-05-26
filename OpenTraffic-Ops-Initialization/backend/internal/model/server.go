package model

import "time"

// AuthType SSH认证方式
type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeKey      AuthType = "key"
)

// Server 目标服务器
type Server struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null;uniqueIndex" json:"name"`        // 显示名称
	Host        string    `gorm:"not null" json:"host"`                    // 主机地址
	Port        int       `gorm:"not null;default:22" json:"port"`         // SSH端口
	Username    string    `gorm:"not null" json:"username"`                // SSH用户名
	AuthType    string    `gorm:"not null" json:"auth_type"`               // password / key
	Password    string    `json:"-"`                                        // 加密存储（密码认证）
	PrivateKey  string    `json:"-"`                                        // 加密存储（密钥认证）
	Passphrase  string    `json:"-"`                                        // 加密存储（密钥密码）
	DeployPath  string    `gorm:"not null;default:/opt/rtm" json:"deploy_path"` // 远程部署路径
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Server) TableName() string {
	return "servers"
}
