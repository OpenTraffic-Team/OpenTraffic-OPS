package dto

// HostProxy 指令类型
const (
	HostProxyCmdStartProcess   = "startProcess"
	HostProxyCmdStopProcess    = "stopProcess"
	HostProxyCmdRestartProcess = "restartProcess"
)

// HostProxyHeartbeatRequest HostProxy心跳上报请求（合并健康度数据）
type HostProxyHeartbeatRequest struct {
	IP                string  `json:"ip" binding:"required"`
	HostName          string  `json:"hostName"`
	ProxyVersion      string  `json:"agentVersion"`
	HeartbeatInterval int     `json:"heartbeatInterval"`
	// 健康度数据
	CpuUsage   float64 `json:"cpuUsage"`
	MemUsage   float64 `json:"memUsage"`
	MemUsedMb  float64 `json:"memUsedMb"`
	DiskUsage  float64 `json:"diskUsage"`
	NetInKbps  float64 `json:"netInKbps"`
	NetOutKbps float64 `json:"netOutKbps"`
	LoadAvg    string  `json:"loadAvg"`
	Timestamp  int64   `json:"timestamp"`
	// 进程指标（可选）
	Processes []HostProxyProcessMetric `json:"processes"`
}

// HostProxyProcessMetric HostProxy上报的进程指标
type HostProxyProcessMetric struct {
	Process    string  `json:"process"`
	Status     string  `json:"status"` // 0:停止 1:运行
	CpuUsage   float64 `json:"cpuUsage"`
	MemUsageMb float64 `json:"memUsageMb"`
}

// HostProxyRegisterRequest HostProxy注册请求（精简版）
type HostProxyRegisterRequest struct {
	IP           string `json:"ip" binding:"required"`
	HostName     string `json:"hostName"`
	OsType       string `json:"osType"`
	OsVersion    string `json:"osVersion"`
	CpuArch      string `json:"cpuArch"`
	CpuCores     int    `json:"cpuCores"`
	CpuModel     string `json:"cpuModel"`
	MemTotalMb   int64  `json:"memTotalMb"`
	DiskTotalGb  int64  `json:"diskTotalGb"`
	GpuInfo      string `json:"gpuInfo"`      // JSON
	MacAddress   string `json:"macAddress"`
	ProxyVersion string `json:"agentVersion"`
}

// HostProxyRegisterResponse HostProxy注册响应
type HostProxyRegisterResponse struct {
	Registered bool   `json:"registered"`
	Message    string `json:"message"`
}

// HostProxyPollRequest HostProxy轮询请求
type HostProxyPollRequest struct {
	IP string `json:"ip" form:"ip" binding:"required"`
}

// HostProxyPollResponse HostProxy轮询响应
type HostProxyPollResponse struct {
	Commands []HostProxyCommand `json:"commands"`
}

// HostProxyCommand HostProxy待执行指令
type HostProxyCommand struct {
	CommandID  string `json:"commandId"`
	Type       string `json:"type"` // 见 HostProxyCmd* 常量
	Process    string `json:"process"`
	Params     string `json:"params"`
	CreateTime string `json:"createTime"`
}

// HostProxyCommandAckRequest HostProxy指令执行结果上报
type HostProxyCommandAckRequest struct {
	IP        string `json:"ip" binding:"required"`
	CommandID string `json:"commandId" binding:"required"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
}
