package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ComponentType 组件类型
type ComponentType string

const (
	ComponentTypePostgreSQL ComponentType = "postgresql"
	ComponentTypeRedis      ComponentType = "redis"
)

// ComponentStatus 组件状态
type ComponentStatus string

const (
	ComponentStatusInstalling ComponentStatus = "installing"
	ComponentStatusRunning    ComponentStatus = "running"
	ComponentStatusStopped    ComponentStatus = "stopped"
	ComponentStatusError      ComponentStatus = "error"
)

// ComponentConfig 组件配置
type ComponentConfig map[string]interface{}

// Value 实现 driver.Valuer 接口
func (c ComponentConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan 实现 sql.Scanner 接口
func (c *ComponentConfig) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Component 组件实体
type Component struct {
	ID          string          `gorm:"primaryKey" json:"id"`
	Name        string          `gorm:"not null;uniqueIndex" json:"name"`
	Type        ComponentType   `gorm:"not null" json:"type"`
	Image       string          `gorm:"not null" json:"image"`
	Version     string          `json:"version"`
	Status      ComponentStatus `gorm:"not null" json:"status"`
	Config      ComponentConfig `gorm:"type:text" json:"config"`
	ContainerID string          `json:"container_id"`
	ImageSource string          `json:"image_source"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// TableName 指定表名
func (Component) TableName() string {
	return "components"
}
