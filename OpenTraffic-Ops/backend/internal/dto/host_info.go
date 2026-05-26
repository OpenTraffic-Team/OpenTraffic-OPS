package dto

// HostInfoQuery 主机信息查询条件
type HostInfoQuery struct {
	PageQuery
	IP       string `form:"ip"`
	Name     string `form:"name"`
	IsOnline *bool  `form:"isOnline"`
	OsType   string `form:"osType"`
}

// HostInfoCreateRequest 主机信息创建/更新请求（管理员修改名称等基本信息）
type HostInfoCreateRequest struct {
	ID   int64  `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// HostInfoUpdateRequest 主机信息更新请求
type HostInfoUpdateRequest struct {
	ID   int64  `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// HostInfoTreeNode 主机信息树形节点
type HostInfoTreeNode struct {
	Label    string             `json:"label"`
	IP       string             `json:"ip"`
	Children []HostInfoTreeNode `json:"children,omitempty"`
}

// HostInfoDto 主机信息DTO
type HostInfoDto struct {
	ID                int64  `json:"id"`
	IP                string `json:"ip"`
	Name              string `json:"name"`
	IsOnline          *bool  `json:"isOnline"`
	OsType            string `json:"osType"`
	OsVersion         string `json:"osVersion"`
	CpuArch           string `json:"cpuArch"`
	CpuCores          int    `json:"cpuCores"`
	CpuModel          string `json:"cpuModel"`
	MemTotalMb        int64  `json:"memTotalMb"`
	DiskTotalGb       int64  `json:"diskTotalGb"`
	GpuInfo           string `json:"gpuInfo"`
	MacAddress        string `json:"macAddress"`
	ProxyVersion      string `json:"agentVersion"`
	HeartbeatInterval int    `json:"heartbeatInterval"`
	LastHeartbeat     string `json:"lastHeartbeat"`
	OfflineTime       string `json:"offlineTime"`
	RegisterTime      string `json:"registerTime"`
}
