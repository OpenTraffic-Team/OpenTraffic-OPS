package model

import (
	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/utils"
)

// AlarmNotifyLog 告警通知日志模型 bu_alarm_notify_log
type AlarmNotifyLog struct {
	ID           int64   `json:"id" gorm:"primaryKey;column:id;comment:主键"`
	RecordID     int64   `json:"recordId" gorm:"column:record_id;comment:关联告警记录ID"`
	ChannelID    int64   `json:"channelId" gorm:"column:channel_id;comment:通道ID"`
	ChannelName  string  `json:"channelName" gorm:"column:channel_name;size:100;comment:通道名称"`
	ChannelType  string  `json:"channelType" gorm:"column:channel_type;size:32;comment:通道类型"`
	Status       string  `json:"status" gorm:"column:status;size:16;comment:发送状态"`
	Response     string  `json:"response" gorm:"column:response;type:text;comment:响应内容"`
	SendTime     string  `json:"sendTime" gorm:"column:send_time;type:timestamp;comment:发送时间"`
	CreateTime   string  `json:"createTime" gorm:"column:create_time;type:timestamp;comment:创建时间"`
}

// BeforeCreate 创建前回调
func (l *AlarmNotifyLog) BeforeCreate(tx *gorm.DB) error {
	if l.CreateTime == "" {
		l.CreateTime = utils.NowStr()
	}
	if l.SendTime == "" {
		l.SendTime = utils.NowStr()
	}
	return nil
}

func (AlarmNotifyLog) TableName() string {
	return "bu_alarm_notify_log"
}
