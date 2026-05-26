package dto

// AlarmNotifyLogQuery 告警通知日志查询条件
type AlarmNotifyLogQuery struct {
	PageQuery
	RecordID    int64  `form:"recordId"`
	ChannelID   int64  `form:"channelId"`
	ChannelType string `form:"channelType"`
	Status      string `form:"status"`
}

// AlarmNotifyLogDto 告警通知日志DTO
type AlarmNotifyLogDto struct {
	ID          int64  `json:"id"`
	RecordID    int64  `json:"recordId"`
	ChannelID   int64  `json:"channelId"`
	ChannelName string `json:"channelName"`
	ChannelType string `json:"channelType"`
	Status      string `json:"status"`
	Response    string `json:"response"`
	SendTime    string `json:"sendTime"`
	CreateTime  string `json:"createTime"`
}
