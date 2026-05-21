package dto

// AlarmRecordQuery 告警记录查询条件
type AlarmRecordQuery struct {
	PageQuery
	RuleID     int64   `form:"ruleId"`
	HostID     int64   `form:"hostId"`
	HostIP     string  `form:"hostIp"`
	AlarmType  string  `form:"alarmType"`
	MetricType string  `form:"metricType"`
	Status     string  `form:"status"`
	Severity   string  `form:"severity"`
	BeginTime  string  `form:"beginTime"`
	EndTime    string  `form:"endTime"`
}

// AlarmRecordDto 告警记录DTO
type AlarmRecordDto struct {
	ID           int64   `json:"id"`
	RuleID       int64   `json:"ruleId"`
	RuleName     string  `json:"ruleName"`
	HostID       int64   `json:"hostId"`
	HostIP       string  `json:"hostIp"`
	HostName     string  `json:"hostName"`
	AlarmType    string  `json:"alarmType"`
	MetricType   string  `json:"metricType"`
	CurrentValue float64 `json:"currentValue"`
	Threshold    float64 `json:"threshold"`
	Severity     string  `json:"severity"`
	Content      string  `json:"content"`
	Status       string  `json:"status"`
	TriggerTime  string  `json:"triggerTime"`
	ResolveTime  string  `json:"resolveTime"`
	NotifyStatus string  `json:"notifyStatus"`
	CreateTime   string  `json:"createTime"`
}

// AlarmRecordAckRequest 告警确认请求
type AlarmRecordAckRequest struct {
	ID     int64  `json:"id" binding:"required"`
}

// AlarmRecordBatchAckRequest 告警批量确认请求
type AlarmRecordBatchAckRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}
