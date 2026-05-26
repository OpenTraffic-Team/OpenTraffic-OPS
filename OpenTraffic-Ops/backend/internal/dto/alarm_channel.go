package dto

// AlarmChannelQuery 告警通道查询条件
type AlarmChannelQuery struct {
	PageQuery
	Name        string `form:"name"`
	ChannelType string `form:"channelType"`
	Status      string `form:"status"`
}

// AlarmChannelCreateRequest 告警通道创建请求
type AlarmChannelCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	ChannelType string `json:"channelType" binding:"required"`
	Config      string `json:"config"`
	Status      string `json:"status"`
	IsDefault   string `json:"isDefault"`
	Remark      string `json:"remark"`
}

// AlarmChannelUpdateRequest 告警通道更新请求
type AlarmChannelUpdateRequest struct {
	ID          int64  `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	ChannelType string `json:"channelType" binding:"required"`
	Config      string `json:"config"`
	Status      string `json:"status"`
	IsDefault   string `json:"isDefault"`
	Remark      string `json:"remark"`
}

// AlarmChannelDto 告警通道DTO
type AlarmChannelDto struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ChannelType string `json:"channelType"`
	Config      string `json:"config"`
	Status      string `json:"status"`
	IsDefault   string `json:"isDefault"`
	Remark      string `json:"remark"`
	CreateBy    string `json:"createBy"`
	CreateTime  string `json:"createTime"`
	UpdateBy    string `json:"updateBy"`
	UpdateTime  string `json:"updateTime"`
}

// AlarmChannelStatusRequest 告警通道状态更新请求
type AlarmChannelStatusRequest struct {
	ID     int64  `json:"id" binding:"required"`
	Status string `json:"status" binding:"required"`
}

// AlarmChannelSetDefaultRequest 设置默认通道请求
type AlarmChannelSetDefaultRequest struct {
	ID        int64  `json:"id" binding:"required"`
	IsDefault string `json:"isDefault" binding:"required"`
}
