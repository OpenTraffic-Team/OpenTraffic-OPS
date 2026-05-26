package model

// HostHealth 主机健康度模型 bu_host_health
// 注意：ReportTime 使用 string 类型，避免 pgx 驱动对 time.Time 做 UTC 转换
type HostHealth struct {
	ID         int64     `json:"id" gorm:"primaryKey;column:id;comment:主键"`
	HostID     int64     `json:"hostId" gorm:"column:host_id;comment:关联主机ID"`
	IP         string    `json:"ip" gorm:"column:ip;size:64;comment:主机IP"`
	CpuUsage   float64   `json:"cpuUsage" gorm:"column:cpu_usage;type:numeric(5,2);comment:CPU使用率%"`
	MemUsage   float64   `json:"memUsage" gorm:"column:mem_usage;type:numeric(5,2);comment:内存使用率%"`
	MemUsedMb  int64     `json:"memUsedMb" gorm:"column:mem_used_mb;comment:内存使用MB"`
	DiskUsage  float64   `json:"diskUsage" gorm:"column:disk_usage;type:numeric(5,2);comment:磁盘使用率%"`
	NetInKbps  float64   `json:"netInKbps" gorm:"column:net_in_kbps;type:numeric(10,2);comment:网络入流量KB/s"`
	NetOutKbps float64   `json:"netOutKbps" gorm:"column:net_out_kbps;type:numeric(10,2);comment:网络出流量KB/s"`
	LoadAvg    string    `json:"loadAvg" gorm:"column:load_avg;size:50;comment:系统负载"`
	IsOnline   *bool     `json:"isOnline" gorm:"column:is_online;comment:上报时在线状态"`
	ReportTime string `json:"reportTime" gorm:"column:report_time;type:timestamp;comment:Agent上报时间"`
	CreateTime string `json:"createTime" gorm:"column:create_time;type:timestamp;comment:记录入库时间"`
}

func (HostHealth) TableName() string {
	return "bu_host_health"
}
