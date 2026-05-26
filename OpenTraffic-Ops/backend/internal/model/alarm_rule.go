package model

// AlarmRule 告警规则模型 bu_alarm_rule
type AlarmRule struct {
	ID         int64   `json:"id" gorm:"primaryKey;column:id;comment:主键"`
	Name       string  `json:"name" gorm:"column:name;size:100;comment:规则名称"`
	RuleType   string  `json:"ruleType" gorm:"column:rule_type;size:32;comment:规则类型"`
	MetricType string  `json:"metricType" gorm:"column:metric_type;size:32;comment:指标/服务类型"`
	HostID     int64   `json:"hostId" gorm:"column:host_id;comment:关联主机ID，0表示全部主机"`
	Threshold  float64 `json:"threshold" gorm:"column:threshold;type:numeric(10,2);comment:阈值"`
	CompareOp  string  `json:"compareOp" gorm:"column:compare_op;size:16;comment:比较运算符"`
	Duration   int     `json:"duration" gorm:"column:duration;comment:持续时间（秒）"`
	Severity   string  `json:"severity" gorm:"column:severity;size:16;comment:告警级别"`
	ChannelIDs string  `json:"channelIds" gorm:"column:channel_ids;type:text;comment:告警通道ID数组JSON"`
	Status     string  `json:"status" gorm:"column:status;size:1;default:0;comment:状态（0启用 1禁用）"`
	BaseEntity
}

func (AlarmRule) TableName() string {
	return "bu_alarm_rule"
}
