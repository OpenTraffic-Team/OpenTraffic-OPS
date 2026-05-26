package dto

// AlarmRuleQuery 告警规则查询条件
type AlarmRuleQuery struct {
	PageQuery
	Name       string `form:"name"`
	RuleType   string `form:"ruleType"`
	MetricType string `form:"metricType"`
	HostID     int64  `form:"hostId"`
	Status     string `form:"status"`
}

// AlarmRuleCreateRequest 告警规则创建请求
type AlarmRuleCreateRequest struct {
	Name       string  `json:"name" binding:"required"`
	RuleType   string  `json:"ruleType" binding:"required"`
	MetricType string  `json:"metricType" binding:"required"`
	HostID     int64   `json:"hostId"`
	Threshold  float64 `json:"threshold"`
	CompareOp  string  `json:"compareOp"`
	Duration   int     `json:"duration"`
	Severity   string  `json:"severity"`
	ChannelIDs string  `json:"channelIds"`
	Status     string  `json:"status"`
	Remark     string  `json:"remark"`
}

// AlarmRuleUpdateRequest 告警规则更新请求
type AlarmRuleUpdateRequest struct {
	ID         int64   `json:"id" binding:"required"`
	Name       string  `json:"name" binding:"required"`
	RuleType   string  `json:"ruleType" binding:"required"`
	MetricType string  `json:"metricType" binding:"required"`
	HostID     int64   `json:"hostId"`
	Threshold  float64 `json:"threshold"`
	CompareOp  string  `json:"compareOp"`
	Duration   int     `json:"duration"`
	Severity   string  `json:"severity"`
	ChannelIDs string  `json:"channelIds"`
	Status     string  `json:"status"`
	Remark     string  `json:"remark"`
}

// AlarmRuleDto 告警规则DTO
type AlarmRuleDto struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	RuleType   string  `json:"ruleType"`
	MetricType string  `json:"metricType"`
	HostID     int64   `json:"hostId"`
	Threshold  float64 `json:"threshold"`
	CompareOp  string  `json:"compareOp"`
	Duration   int     `json:"duration"`
	Severity   string  `json:"severity"`
	ChannelIDs string  `json:"channelIds"`
	Status     string  `json:"status"`
	Remark     string  `json:"remark"`
	CreateBy   string  `json:"createBy"`
	CreateTime string  `json:"createTime"`
	UpdateBy   string  `json:"updateBy"`
	UpdateTime string  `json:"updateTime"`
}

// AlarmRuleStatusRequest 告警规则状态更新请求
type AlarmRuleStatusRequest struct {
	ID     int64  `json:"id" binding:"required"`
	Status string `json:"status" binding:"required"`
}
