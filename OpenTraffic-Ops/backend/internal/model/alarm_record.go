package model

import (
	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/utils"
)

// AlarmRecord 告警记录模型 bu_alarm_record
type AlarmRecord struct {
	ID           int64   `json:"id" gorm:"primaryKey;column:id;comment:主键"`
	RuleID       int64   `json:"ruleId" gorm:"column:rule_id;comment:关联规则ID"`
	RuleName     string  `json:"ruleName" gorm:"column:rule_name;size:100;comment:规则名称快照"`
	HostID       int64   `json:"hostId" gorm:"column:host_id;comment:主机ID"`
	HostIP       string  `json:"hostIp" gorm:"column:host_ip;size:64;comment:主机IP"`
	HostName     string  `json:"hostName" gorm:"column:host_name;size:100;comment:主机名称"`
	AlarmType    string  `json:"alarmType" gorm:"column:alarm_type;size:32;comment:告警类型"`
	MetricType   string  `json:"metricType" gorm:"column:metric_type;size:32;comment:指标/服务类型"`
	CurrentValue float64 `json:"currentValue" gorm:"column:current_value;type:numeric(10,2);comment:当前值"`
	Threshold    float64 `json:"threshold" gorm:"column:threshold;type:numeric(10,2);comment:阈值"`
	Severity     string  `json:"severity" gorm:"column:severity;size:16;comment:告警级别"`
	Content      string  `json:"content" gorm:"column:content;type:text;comment:告警内容"`
	Status       string  `json:"status" gorm:"column:status;size:16;comment:状态"`
	TriggerTime  string  `json:"triggerTime" gorm:"column:trigger_time;type:timestamp;comment:触发时间"`
	ResolveTime  *string `json:"resolveTime" gorm:"column:resolve_time;type:timestamp;comment:恢复时间"`
	NotifyStatus string  `json:"notifyStatus" gorm:"column:notify_status;size:16;comment:通知状态"`
	CreateTime   string  `json:"createTime" gorm:"column:create_time;type:timestamp;comment:创建时间"`
}

// BeforeCreate 创建前回调
func (r *AlarmRecord) BeforeCreate(tx *gorm.DB) error {
	now := utils.NowStr()
	if r.CreateTime == "" {
		r.CreateTime = now
	}
	if r.TriggerTime == "" {
		r.TriggerTime = now
	}
	return nil
}

func (AlarmRecord) TableName() string {
	return "bu_alarm_record"
}
