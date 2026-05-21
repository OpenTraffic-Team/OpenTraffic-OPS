package collector

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemMetrics 系统监控指标
type SystemMetrics struct {
	CpuUsage   float64 `json:"cpuUsage"`
	MemUsage   float64 `json:"memUsage"`
	MemUsageMb float64 `json:"memUsageMb"`
	DiskUsage  float64 `json:"diskUsage"`
	NetIn      float64 `json:"netIn"`      // KB/s
	NetOut     float64 `json:"netOut"`     // KB/s
	LoadAvg    string  `json:"loadAvg"`
	GpuUsage   string  `json:"gpuUsage"`   // JSON 字符串，暂不支持
	Timestamp  int64   `json:"timestamp"`
}

// SystemInfo 系统静态信息（注册时用）
type SystemInfo struct {
	OsType      string `json:"osType"`
	OsVersion   string `json:"osVersion"`
	CpuArch     string `json:"cpuArch"`
	CpuCores    int    `json:"cpuCores"`
	CpuModel    string `json:"cpuModel"`
	MemTotalMb  uint64 `json:"memTotalMb"`  // MB
	DiskTotalGb uint64 `json:"diskTotalGb"` // GB
	GpuInfo     string `json:"gpuInfo"`
	MacAddress  string `json:"macAddress"`
}

var lastNetIO *net.IOCountersStat
var lastNetIOTime time.Time

// CollectSystemMetrics 采集系统动态指标
func CollectSystemMetrics() (*SystemMetrics, error) {
	m := &SystemMetrics{Timestamp: time.Now().Unix()}

	// CPU
	percents, err := cpu.Percent(100*time.Millisecond, false)
	if err == nil && len(percents) > 0 {
		m.CpuUsage = percents[0]
	}

	// Memory
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		m.MemUsage = vmStat.UsedPercent
		m.MemUsageMb = float64(vmStat.Used) / 1024 / 1024
	}

	// Disk (根分区)
	diskStat, err := disk.Usage("/")
	if err != nil && runtime.GOOS == "windows" {
		diskStat, err = disk.Usage("C:")
	}
	if err == nil {
		m.DiskUsage = diskStat.UsedPercent
	}

	// Load
	avg, err := load.Avg()
	if err == nil {
		m.LoadAvg = fmt.Sprintf("%.2f,%.2f,%.2f", avg.Load1, avg.Load5, avg.Load15)
	}

	// Network (计算差值)
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		m.NetIn, m.NetOut = calcNetSpeed(&netIO[0])
	}

	// GPU 暂不支持
	m.GpuUsage = ""

	return m, nil
}

// CollectSystemInfo 采集系统静态信息
func CollectSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// OS
	info.OsType = runtime.GOOS
	hostInfo, err := host.Info()
	if err == nil {
		info.OsVersion = hostInfo.PlatformVersion
	}
	info.CpuArch = runtime.GOARCH

	// CPU
	cpuInfos, err := cpu.Info()
	if err == nil && len(cpuInfos) > 0 {
		info.CpuModel = cpuInfos[0].ModelName
	}
	cpuCounts, err := cpu.Counts(true)
	if err == nil {
		info.CpuCores = cpuCounts
	}

	// Memory
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		info.MemTotalMb = vmStat.Total / 1024 / 1024
	}

	// Disk
	diskStat, err := disk.Usage("/")
	if err != nil && runtime.GOOS == "windows" {
		diskStat, err = disk.Usage("C:")
	}
	if err == nil {
		info.DiskTotalGb = diskStat.Total / 1024 / 1024 / 1024
	}

	// Network interfaces - 取第一个 MAC 地址
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.HardwareAddr != "" {
				info.MacAddress = iface.HardwareAddr
				break
			}
		}
	}

	return info, nil
}

func calcNetSpeed(current *net.IOCountersStat) (inKBps, outKBps float64) {
	if lastNetIO == nil {
		lastNetIO = current
		lastNetIOTime = time.Now()
		return 0, 0
	}

	elapsed := time.Since(lastNetIOTime).Seconds()
	if elapsed <= 0 {
		return 0, 0
	}

	inBytes := float64(current.BytesRecv - lastNetIO.BytesRecv)
	outBytes := float64(current.BytesSent - lastNetIO.BytesSent)

	inKBps = inBytes / 1024 / elapsed
	outKBps = outBytes / 1024 / elapsed

	lastNetIO = current
	lastNetIOTime = time.Now()
	return
}

func toJSON(v interface{}) string {
	// 简化的 JSON 序列化
	b, _ := json.Marshal(v)
	return string(b)
}
