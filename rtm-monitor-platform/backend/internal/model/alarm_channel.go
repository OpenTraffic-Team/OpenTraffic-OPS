package model

// AlarmChannel 告警通道模型 bu_alarm_channel
type AlarmChannel struct {
	ID          int64  `json:"id" gorm:"primaryKey;column:id;comment:主键"`
	Name        string `json:"name" gorm:"column:name;size:100;comment:通道名称"`
	ChannelType string `json:"channelType" gorm:"column:channel_type;size:32;comment:通道类型"`
	Config      string `json:"config" gorm:"column:config;type:text;comment:通道配置JSON"`
	Status      string `json:"status" gorm:"column:status;size:1;default:0;comment:状态（0启用 1禁用）"`
	IsDefault   string `json:"isDefault" gorm:"column:is_default;size:1;default:0;comment:是否默认（0否 1是）"`
	BaseEntity
}

func (AlarmChannel) TableName() string {
	return "bu_alarm_channel"
}
