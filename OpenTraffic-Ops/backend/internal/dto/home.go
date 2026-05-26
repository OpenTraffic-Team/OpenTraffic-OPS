package dto

// HomeStatisticData 首页统计数据
type HomeStatisticData struct {
	HostCount           int64            `json:"hostCount"`
	OfflineHostCount    int64            `json:"offlineHostCount"`
	AlarmChannelCount   int64            `json:"alarmChannelCount"`
	AlarmRuleCount      int64            `json:"alarmRuleCount"`
	UnhandledAlarmCount int64            `json:"unhandledAlarmCount"`
	RecentAlarms        []AlarmRecordDto `json:"recentAlarms"`
}
