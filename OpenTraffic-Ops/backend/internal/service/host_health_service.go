package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/model"
	"opentraffic-ops-backend/internal/repository"
	"opentraffic-ops-backend/internal/utils"
)

// HostHealthService 主机健康度服务
type HostHealthService struct {
	healthRepo *repository.HostHealthRepository
	infoRepo   *repository.HostInfoRepository
}

// NewHostHealthService 创建主机健康度服务
func NewHostHealthService(db *gorm.DB) *HostHealthService {
	return &HostHealthService{
		healthRepo: repository.NewHostHealthRepository(db),
		infoRepo:   repository.NewHostInfoRepository(db),
	}
}

// SaveHostHealth 保存主机健康度数据到 PostgreSQL
func (s *HostHealthService) SaveHostHealth(ctx context.Context, hostID int64, ip string, req *dto.HostProxyHeartbeatRequest) error {
	if ip == "" {
		return fmt.Errorf("IP不能为空")
	}

	reportTime := time.Now()
	if req.Timestamp > 0 {
		reportTime = time.Unix(req.Timestamp, 0).In(time.Local)
	}

	health := &model.HostHealth{
		HostID:     hostID,
		IP:         ip,
		CpuUsage:   req.CpuUsage,
		MemUsage:   req.MemUsage,
		MemUsedMb:  int64(req.MemUsedMb),
		DiskUsage:  req.DiskUsage,
		NetInKbps:  req.NetInKbps,
		NetOutKbps: req.NetOutKbps,
		LoadAvg:    req.LoadAvg,
		IsOnline:   boolPtr(true),
		ReportTime: reportTime.Format(utils.TimeFormat),
		CreateTime: utils.NowStr(),
	}

	if err := s.healthRepo.Save(ctx, health); err != nil {
		zap.L().Error("保存主机健康度数据失败",
			zap.String("ip", ip),
			zap.Error(err))
		return err
	}

	return nil
}

// SelectHostInfoVoList 查询主机监控列表（从 PG 查询最新健康度 JOIN host_info）
func (s *HostHealthService) SelectHostInfoVoList(ctx context.Context) ([]dto.HostHealthVo, error) {
	// 查询所有主机
	hosts, err := s.infoRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return []dto.HostHealthVo{}, nil
	}

	// 收集所有IP
	ips := make([]string, 0, len(hosts))
	for _, h := range hosts {
		if h.IP != "" {
			ips = append(ips, h.IP)
		}
	}

	// 查询每个主机的最新健康度数据
	latestHealthMap := s.fetchLatestHealth(ctx, ips)

	// 组装VO
	result := make([]dto.HostHealthVo, 0, len(hosts))
	for _, host := range hosts {
		vo := dto.HostHealthVo{
			IpAddr:       host.IP,
			HostName:     host.Name,
			IsOnline:     host.IsOnline,
			OsType:       host.OsType,
			OsVersion:    host.OsVersion,
			CpuArch:      host.CpuArch,
			CpuCores:     host.CpuCores,
			CpuModel:     host.CpuModel,
			MemTotalMb:   host.MemTotalMb,
			DiskTotalGb:  host.DiskTotalGb,
			ProxyVersion: host.ProxyVersion,
		}
		if host.LastHeartbeat != nil {
			vo.LastHeartbeat = *host.LastHeartbeat
		}

		// 填充最新监控数据
		if health, ok := latestHealthMap[host.IP]; ok {
			vo.CpuUsage = formatFloatStr(health.CpuUsage)
			vo.MemUsage = formatFloatStr(health.MemUsage)
			vo.MemUsageMB = strconv.FormatInt(health.MemUsedMb, 10)
			vo.DiskUsage = formatFloatStr(health.DiskUsage)
			vo.NetIn = formatFloatStr(health.NetInKbps)
			vo.NetOut = formatFloatStr(health.NetOutKbps)
			vo.LoadAvg = health.LoadAvg
			vo.Datetime = health.ReportTime
		} else {
			vo.CpuUsage = "0"
			vo.MemUsage = "0"
			vo.MemUsageMB = "0"
			vo.DiskUsage = "0"
			vo.NetIn = "0"
			vo.NetOut = "0"
			vo.LoadAvg = ""
		}

		result = append(result, vo)
	}

	return result, nil
}

// fetchLatestHealth 批量获取每个IP的最新健康度记录
func (s *HostHealthService) fetchLatestHealth(ctx context.Context, ips []string) map[string]model.HostHealth {
	result := make(map[string]model.HostHealth)
	if len(ips) == 0 {
		return result
	}

	healths, err := s.healthRepo.FindLatestByIPs(ctx, ips)
	if err != nil {
		zap.L().Warn("查询最新健康度数据失败", zap.Error(err))
		return result
	}

	for _, h := range healths {
		result[h.IP] = h
	}
	return result
}

// GetHostMonHistoryData 获取主机监控历史数据（从 PG 按时间聚合查询）
// queryLevel: 1=日期级别(近15天), 2=小时级别, 3=分钟级别
func (s *HostHealthService) GetHostMonHistoryData(ctx context.Context, ip, queryLevel, queryDate, queryHour string) map[string]interface{} {
	result := make(map[string]interface{})
	var times []string
	var cpuUsages, memUsages, memUsageMBs, diskUsages []float64
	var netIns, netOuts []float64

	switch queryLevel {
	case "1":
		// 日期级别：近15天，每天取平均值
		now := time.Now()
		for i := 14; i >= 0; i-- {
			date := utils.FormatDate(now.AddDate(0, 0, -i))
			times = append(times, date)
			agg := s.aggregateHealthByDate(ctx, ip, date)
			cpuUsages = append(cpuUsages, agg.CpuUsage)
			memUsages = append(memUsages, agg.MemUsage)
			memUsageMBs = append(memUsageMBs, float64(agg.MemUsedMb))
			diskUsages = append(diskUsages, agg.DiskUsage)
			netIns = append(netIns, agg.NetInKbps)
			netOuts = append(netOuts, agg.NetOutKbps)
		}
	case "2":
		// 小时级别：指定日期，每小时取平均值
		if queryDate == "" {
			queryDate = utils.TodayStr()
		}
		for i := 0; i <= 23; i++ {
			times = append(times, fmt.Sprintf("%02d", i))
			agg := s.aggregateHealthByHour(ctx, ip, queryDate, i)
			cpuUsages = append(cpuUsages, agg.CpuUsage)
			memUsages = append(memUsages, agg.MemUsage)
			memUsageMBs = append(memUsageMBs, float64(agg.MemUsedMb))
			diskUsages = append(diskUsages, agg.DiskUsage)
			netIns = append(netIns, agg.NetInKbps)
			netOuts = append(netOuts, agg.NetOutKbps)
		}
	case "3":
		// 分钟级别：指定日期和小时，每分钟取平均值
		if queryDate == "" || queryHour == "" {
			break
		}
		hour, _ := strconv.Atoi(queryHour)
		for i := 0; i <= 59; i++ {
			times = append(times, fmt.Sprintf("%02d", i))
			agg := s.aggregateHealthByMinute(ctx, ip, queryDate, hour, i)
			cpuUsages = append(cpuUsages, agg.CpuUsage)
			memUsages = append(memUsages, agg.MemUsage)
			memUsageMBs = append(memUsageMBs, float64(agg.MemUsedMb))
			diskUsages = append(diskUsages, agg.DiskUsage)
			netIns = append(netIns, agg.NetInKbps)
			netOuts = append(netOuts, agg.NetOutKbps)
		}
	}

	result["times"] = times
	result["cpuUsages"] = cpuUsages
	result["memUsages"] = memUsages
	result["memUsageMBs"] = memUsageMBs
	result["diskUsages"] = diskUsages
	result["netIns"] = netIns
	result["netOuts"] = netOuts
	return result
}

// healthAggregate 健康度聚合结果
type healthAggregate struct {
	CpuUsage   float64
	MemUsage   float64
	MemUsedMb  int64
	DiskUsage  float64
	NetInKbps  float64
	NetOutKbps float64
}

// aggregateHealthByDate 按日期聚合健康度数据
func (s *HostHealthService) aggregateHealthByDate(ctx context.Context, ip, date string) healthAggregate {
	var agg healthAggregate
	if ip == "" || date == "" {
		return agg
	}

	cpu, mem, memMB, disk, netIn, netOut, count := s.healthRepo.AggregateByDate(ctx, ip, date)
	if count > 0 {
		agg.CpuUsage = cpu
		agg.MemUsage = mem
		agg.MemUsedMb = int64(memMB)
		agg.DiskUsage = disk
		agg.NetInKbps = netIn
		agg.NetOutKbps = netOut
	}
	return agg
}

// aggregateHealthByHour 按小时聚合健康度数据
func (s *HostHealthService) aggregateHealthByHour(ctx context.Context, ip, date string, hour int) healthAggregate {
	var agg healthAggregate
	if ip == "" || date == "" {
		return agg
	}

	cpu, mem, memMB, disk, netIn, netOut, count := s.healthRepo.AggregateByHour(ctx, ip, date, hour)
	if count > 0 {
		agg.CpuUsage = cpu
		agg.MemUsage = mem
		agg.MemUsedMb = int64(memMB)
		agg.DiskUsage = disk
		agg.NetInKbps = netIn
		agg.NetOutKbps = netOut
	}
	return agg
}

// aggregateHealthByMinute 按分钟聚合健康度数据
func (s *HostHealthService) aggregateHealthByMinute(ctx context.Context, ip, date string, hour, minute int) healthAggregate {
	var agg healthAggregate
	if ip == "" || date == "" {
		return agg
	}

	cpu, mem, memMB, disk, netIn, netOut, count := s.healthRepo.AggregateByMinute(ctx, ip, date, hour, minute)
	if count > 0 {
		agg.CpuUsage = cpu
		agg.MemUsage = mem
		agg.MemUsedMb = int64(memMB)
		agg.DiskUsage = disk
		agg.NetInKbps = netIn
		agg.NetOutKbps = netOut
	}
	return agg
}

// GetProMonHistoryData 获取进程监控历史数据
// 进程历史数据当前仅记录 status 状态，CPU/内存等明细暂不持久化
func (s *HostHealthService) GetProMonHistoryData(ctx context.Context, ip, process, queryLevel, queryDate, queryHour string) map[string]interface{} {
	result := make(map[string]interface{})
	var times []string
	var cpuUsages, memUsageMBs []float64

	switch queryLevel {
	case "1":
		now := time.Now()
		for i := 14; i >= 0; i-- {
			times = append(times, utils.FormatDate(now.AddDate(0, 0, -i)))
			cpuUsages = append(cpuUsages, 0)
			memUsageMBs = append(memUsageMBs, 0)
		}
	case "2":
		if queryDate != "" {
			for i := 0; i <= 23; i++ {
				times = append(times, fmt.Sprintf("%02d", i))
				cpuUsages = append(cpuUsages, 0)
				memUsageMBs = append(memUsageMBs, 0)
			}
		}
	case "3":
		if queryDate != "" && queryHour != "" {
			for i := 0; i <= 59; i++ {
				times = append(times, fmt.Sprintf("%02d", i))
				cpuUsages = append(cpuUsages, 0)
				memUsageMBs = append(memUsageMBs, 0)
			}
		}
	}

	result["times"] = times
	result["cpuUsages"] = cpuUsages
	result["memUsageMBs"] = memUsageMBs
	return result
}

// CleanOldHealthData 清理指定天数之前的健康度数据
func (s *HostHealthService) CleanOldHealthData(ctx context.Context, days int) (int64, error) {
	return s.healthRepo.CleanOldData(ctx, days)
}

// formatFloatStr 格式化浮点数为字符串（保留2位小数）
func formatFloatStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
