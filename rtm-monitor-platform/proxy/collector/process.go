package collector

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessMetric 单个进程指标
type ProcessMetric struct {
	Process    string  `json:"process"`
	Status     string  `json:"status"`     // "0"=停止 "1"=运行
	CpuUsage   float64 `json:"cpuUsage"`
	MemUsageMb float64 `json:"memUsageMb"`
}

// CollectProcessMetrics 采集配置的进程指标
func CollectProcessMetrics(names []string) []ProcessMetric {
	results := make([]ProcessMetric, 0, len(names))
	if len(names) == 0 {
		return results
	}

	// 获取所有进程
	procs, err := process.Processes()
	if err != nil {
		return results
	}

	for _, name := range names {
		pm := ProcessMetric{Process: name, Status: "0"}
		for _, p := range procs {
			pname, _ := p.Name()
			if !strings.Contains(pname, name) {
				continue
			}
			pm.Status = "1"
			cpuPercent, _ := p.CPUPercent()
			pm.CpuUsage += cpuPercent
			memInfo, _ := p.MemoryInfo()
			if memInfo != nil {
				pm.MemUsageMb += float64(memInfo.RSS) / 1024 / 1024
			}
		}
		results = append(results, pm)
	}

	return results
}

// FindProcessByName 查找进程是否存在并返回 PID
func FindProcessByName(name string) (int32, error) {
	procs, err := process.Processes()
	if err != nil {
		return 0, err
	}
	for _, p := range procs {
		pname, _ := p.Name()
		if strings.Contains(pname, name) {
			return p.Pid, nil
		}
	}
	return 0, fmt.Errorf("进程 %s 未找到", name)
}
