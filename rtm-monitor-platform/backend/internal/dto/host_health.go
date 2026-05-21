package dto

// HostHealthVo 主机健康度VO
type HostHealthVo struct {
	IpAddr          string         `json:"ip"`
	HostName        string         `json:"name"`
	IsOnline        *bool          `json:"isOnline"`
	OsType          string         `json:"osType"`
	OsVersion       string         `json:"osVersion"`
	CpuArch         string         `json:"cpuArch"`
	CpuCores        int            `json:"cpuCores"`
	CpuModel        string         `json:"cpuModel"`
	MemTotalMb      int64          `json:"memTotalMb"`
	DiskTotalGb     int64          `json:"diskTotalGb"`
	ProxyVersion    string         `json:"agentVersion"`
	LastHeartbeat   string         `json:"lastHeartbeat"`
	CpuUsage        string         `json:"cpuUsage"`
	MemUsage        string         `json:"memUsage"`
	MemUsageMB      string         `json:"memUsageMB"`
	DiskUsage       string         `json:"diskUsage"`
	NetIn           string         `json:"netIn"`
	NetOut          string         `json:"netOut"`
	LoadAvg         string         `json:"loadAvg"`
	Datetime        string         `json:"datetime"`
	MonProcessCount int            `json:"monProcessCount"`
	ProcessVos      []MonProcessVo `json:"processVos"`
}

// MonProcessVo 进程监控VO
type MonProcessVo struct {
	Process    string `json:"process"`
	Desc       string `json:"desc"`
	CpuUsage   string `json:"cpuUsage"`
	MemUsageMB string `json:"memUsageMB"`
	Status     string `json:"status"`
	Datetime   string `json:"datetime"`
}
