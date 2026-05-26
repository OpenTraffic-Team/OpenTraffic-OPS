package model

// HostInfo 主机信息模型 bu_host_info
// 注意：时间字段使用 string 类型，避免 pgx 驱动对 time.Time 做 UTC 转换
// 可为空的字段使用 *string，确保 GORM 插入 NULL 而非空字符串
type HostInfo struct {
	ID                int64   `json:"id" gorm:"primaryKey;column:id;comment:主键"`
	IP                string  `json:"ip" gorm:"column:ip;size:64;uniqueIndex:idx_bu_host_info_ip;comment:主机IP地址"`
	Name              string  `json:"name" gorm:"column:name;size:128;comment:主机名称"`
	OsType            string  `json:"osType" gorm:"column:os_type;size:50;comment:操作系统类型"`
	OsVersion         string  `json:"osVersion" gorm:"column:os_version;size:100;comment:操作系统版本"`
	CpuArch           string  `json:"cpuArch" gorm:"column:cpu_arch;size:20;comment:CPU架构"`
	CpuCores          int     `json:"cpuCores" gorm:"column:cpu_cores;comment:CPU逻辑核数"`
	CpuModel          string  `json:"cpuModel" gorm:"column:cpu_model;size:200;comment:CPU型号"`
	MemTotalMb        int64   `json:"memTotalMb" gorm:"column:mem_total_mb;comment:内存总量MB"`
	DiskTotalGb       int64   `json:"diskTotalGb" gorm:"column:disk_total_gb;comment:磁盘总量GB"`
	GpuInfo           string  `json:"gpuInfo" gorm:"column:gpu_info;type:text;comment:显卡信息JSON"`
	MacAddress        string  `json:"macAddress" gorm:"column:mac_address;size:100;comment:主网卡MAC地址"`
	ProxyVersion      string  `json:"agentVersion" gorm:"column:proxy_version;size:50;comment:Proxy版本号"`
	HeartbeatInterval int     `json:"heartbeatInterval" gorm:"column:heartbeat_interval;default:3;comment:心跳上报间隔（秒）"`
	IsOnline          *bool   `json:"isOnline" gorm:"column:is_online;comment:是否在线"`
	RegisterTime      string  `json:"registerTime" gorm:"column:register_time;type:timestamp;comment:Agent首次上报时间"`
	OfflineTime       *string `json:"offlineTime" gorm:"column:offline_time;type:timestamp;comment:离线时间"`
	LastHeartbeat     *string `json:"lastHeartbeat" gorm:"column:last_heartbeat;type:timestamp;comment:最后心跳时间"`
}

func (HostInfo) TableName() string {
	return "bu_host_info"
}
