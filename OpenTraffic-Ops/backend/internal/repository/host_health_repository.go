package repository

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"rtm-server/internal/model"
)

// HostHealthRepository 主机健康度数据访问
type HostHealthRepository struct {
	db *gorm.DB
}

// NewHostHealthRepository 创建主机健康度仓库
func NewHostHealthRepository(db *gorm.DB) *HostHealthRepository {
	return &HostHealthRepository{db: db}
}

// Save 保存主机健康度记录
func (r *HostHealthRepository) Save(ctx context.Context, health *model.HostHealth) error {
	return r.db.WithContext(ctx).Create(health).Error
}

// FindLatestByIPs 批量查询每个IP的最新健康度记录
func (r *HostHealthRepository) FindLatestByIPs(ctx context.Context, ips []string) ([]model.HostHealth, error) {
	if len(ips) == 0 {
		return []model.HostHealth{}, nil
	}

	var healths []model.HostHealth
	err := r.db.WithContext(ctx).Raw(`
		SELECT DISTINCT ON (ip) *
		FROM bu_host_health
		WHERE ip IN ?
		ORDER BY ip, report_time DESC
	`, ips).Scan(&healths).Error
	return healths, err
}

// healthAggregateResult 聚合查询结果结构
type healthAggregateResult struct {
	CpuUsage   float64
	MemUsage   float64
	MemUsedMb  float64
	DiskUsage  float64
	NetInKbps  float64
	NetOutKbps float64
	Count      int64
}

// AggregateByDate 按日期聚合健康度数据（使用 DATE() 避免时区问题）
func (r *HostHealthRepository) AggregateByDate(ctx context.Context, ip string, date string) (cpu, mem, memMB, disk, netIn, netOut float64, count int64) {
	var result healthAggregateResult

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(AVG(cpu_usage), 0) as cpu_usage,
			COALESCE(AVG(mem_usage), 0) as mem_usage,
			COALESCE(AVG(mem_used_mb), 0) as mem_used_mb,
			COALESCE(AVG(disk_usage), 0) as disk_usage,
			COALESCE(AVG(net_in_kbps), 0) as net_in_kbps,
			COALESCE(AVG(net_out_kbps), 0) as net_out_kbps,
			COUNT(*) as count
		FROM bu_host_health
		WHERE ip = ? AND DATE(report_time) = ?
	`, ip, date).Scan(&result).Error

	if err != nil {
		zap.L().Warn("按日期聚合健康度数据失败", zap.String("ip", ip), zap.String("date", date), zap.Error(err))
		return 0, 0, 0, 0, 0, 0, 0
	}

	return result.CpuUsage, result.MemUsage, result.MemUsedMb, result.DiskUsage, result.NetInKbps, result.NetOutKbps, result.Count
}

// AggregateByHour 按小时聚合健康度数据（使用 DATE() + EXTRACT() 避免时区问题）
func (r *HostHealthRepository) AggregateByHour(ctx context.Context, ip string, date string, hour int) (cpu, mem, memMB, disk, netIn, netOut float64, count int64) {
	var result healthAggregateResult

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(AVG(cpu_usage), 0) as cpu_usage,
			COALESCE(AVG(mem_usage), 0) as mem_usage,
			COALESCE(AVG(mem_used_mb), 0) as mem_used_mb,
			COALESCE(AVG(disk_usage), 0) as disk_usage,
			COALESCE(AVG(net_in_kbps), 0) as net_in_kbps,
			COALESCE(AVG(net_out_kbps), 0) as net_out_kbps,
			COUNT(*) as count
		FROM bu_host_health
		WHERE ip = ? AND DATE(report_time) = ? AND EXTRACT(HOUR FROM report_time) = ?
	`, ip, date, hour).Scan(&result).Error

	if err != nil {
		zap.L().Warn("按小时聚合健康度数据失败", zap.String("ip", ip), zap.String("date", date), zap.Int("hour", hour), zap.Error(err))
		return 0, 0, 0, 0, 0, 0, 0
	}

	return result.CpuUsage, result.MemUsage, result.MemUsedMb, result.DiskUsage, result.NetInKbps, result.NetOutKbps, result.Count
}

// AggregateByMinute 按分钟聚合健康度数据（使用 DATE() + EXTRACT() 避免时区问题）
func (r *HostHealthRepository) AggregateByMinute(ctx context.Context, ip string, date string, hour, minute int) (cpu, mem, memMB, disk, netIn, netOut float64, count int64) {
	var result healthAggregateResult

	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(AVG(cpu_usage), 0) as cpu_usage,
			COALESCE(AVG(mem_usage), 0) as mem_usage,
			COALESCE(AVG(mem_used_mb), 0) as mem_used_mb,
			COALESCE(AVG(disk_usage), 0) as disk_usage,
			COALESCE(AVG(net_in_kbps), 0) as net_in_kbps,
			COALESCE(AVG(net_out_kbps), 0) as net_out_kbps,
			COUNT(*) as count
		FROM bu_host_health
		WHERE ip = ? AND DATE(report_time) = ? AND EXTRACT(HOUR FROM report_time) = ? AND EXTRACT(MINUTE FROM report_time) = ?
	`, ip, date, hour, minute).Scan(&result).Error

	if err != nil {
		zap.L().Warn("按分钟聚合健康度数据失败", zap.String("ip", ip), zap.String("date", date), zap.Int("hour", hour), zap.Int("minute", minute), zap.Error(err))
		return 0, 0, 0, 0, 0, 0, 0
	}

	return result.CpuUsage, result.MemUsage, result.MemUsedMb, result.DiskUsage, result.NetInKbps, result.NetOutKbps, result.Count
}

// CleanOldData 清理指定天数之前的健康度数据
func (r *HostHealthRepository) CleanOldData(ctx context.Context, days int) (int64, error) {
	// 使用 NOW() - INTERVAL 在数据库侧计算，避免时区转换问题
	result := r.db.WithContext(ctx).Where("report_time < NOW() - INTERVAL '? days'", days).Delete(&model.HostHealth{})
	return result.RowsAffected, result.Error
}
