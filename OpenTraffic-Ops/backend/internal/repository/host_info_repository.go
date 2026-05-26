package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/model"
)

// HostInfoRepository 主机信息数据访问
type HostInfoRepository struct {
	db *gorm.DB
}

// NewHostInfoRepository 创建主机信息仓库
func NewHostInfoRepository(db *gorm.DB) *HostInfoRepository {
	return &HostInfoRepository{db: db}
}

// FindAll 查询所有主机
func (r *HostInfoRepository) FindAll(ctx context.Context) ([]model.HostInfo, error) {
	var hosts []model.HostInfo
	err := r.db.WithContext(ctx).Find(&hosts).Error
	return hosts, err
}

// FindByID 根据ID查询主机
func (r *HostInfoRepository) FindByID(ctx context.Context, id int64) (*model.HostInfo, error) {
	var host model.HostInfo
	err := r.db.WithContext(ctx).First(&host, id).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

// FindByIP 根据IP查询主机
func (r *HostInfoRepository) FindByIP(ctx context.Context, ip string) (*model.HostInfo, error) {
	var host model.HostInfo
	err := r.db.WithContext(ctx).Where("ip = ?", ip).First(&host).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

// Count 带条件统计主机数量
func (r *HostInfoRepository) Count(ctx context.Context, query *dto.HostInfoQuery) (int64, error) {
	db := r.db.WithContext(ctx).Model(&model.HostInfo{})

	if query.IP != "" {
		db = db.Where("ip LIKE ?", "%"+query.IP+"%")
	}
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.IsOnline != nil {
		db = db.Where("is_online = ?", *query.IsOnline)
	}
	if query.OsType != "" {
		db = db.Where("os_type = ?", query.OsType)
	}

	var total int64
	err := db.Count(&total).Error
	return total, err
}

// FindPage 分页查询主机列表
func (r *HostInfoRepository) FindPage(ctx context.Context, query *dto.HostInfoQuery) ([]model.HostInfo, error) {
	db := r.db.WithContext(ctx).Model(&model.HostInfo{})

	if query.IP != "" {
		db = db.Where("ip LIKE ?", "%"+query.IP+"%")
	}
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.IsOnline != nil {
		db = db.Where("is_online = ?", *query.IsOnline)
	}
	if query.OsType != "" {
		db = db.Where("os_type = ?", query.OsType)
	}

	var hosts []model.HostInfo
	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&hosts).Error
	return hosts, err
}

// Create 插入主机记录
func (r *HostInfoRepository) Create(ctx context.Context, host *model.HostInfo) error {
	return r.db.WithContext(ctx).Create(host).Error
}

// Update 根据ID更新主机信息
func (r *HostInfoRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.HostInfo{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateByIP 根据IP更新主机信息
func (r *HostInfoRepository) UpdateByIP(ctx context.Context, ip string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.HostInfo{}).Where("ip = ?", ip).Updates(updates).Error
}

// UpdateBatchByIP 批量根据IP更新主机信息
func (r *HostInfoRepository) UpdateBatchByIP(ctx context.Context, ips []string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.HostInfo{}).Where("ip IN ?", ips).Updates(updates).Error
}

// DeleteByIDs 批量删除主机
func (r *HostInfoRepository) DeleteByIDs(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Delete(&model.HostInfo{}, ids).Error
}

// CountTotal 统计主机总数
func (r *HostInfoRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.HostInfo{}).Count(&count).Error
	return count, err
}

// CountOffline 统计离线主机数量
func (r *HostInfoRepository) CountOffline(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.HostInfo{}).Where("is_online = ?", false).Count(&count).Error
	return count, err
}

// MarkOfflineByTimeout 根据超时阈值标记离线主机（单条SQL，避免读取-更新竞态）
// 超时阈值 = heartbeat_interval * 5 秒，按各主机自身的心跳间隔计算
// 兜底策略：heartbeat_interval 最小值为 30 秒，避免旧数据或异常值导致误判
func (r *HostInfoRepository) MarkOfflineByTimeout(ctx context.Context) (int64, error) {
	// 使用 GREATEST(COALESCE(heartbeat_interval, 30), 30) 确保最小心跳间隔为 30 秒
	// COALESCE 处理 NULL，GREATEST 处理 0 或过小值
	// 最小超时阈值 = 30 * 5 = 150 秒
	result := r.db.WithContext(ctx).Model(&model.HostInfo{}).
		Where("is_online = ?", true).
		Where("last_heartbeat < NOW() - (GREATEST(COALESCE(heartbeat_interval, 30), 30) * 5 * interval '1 second')").
		Updates(map[string]interface{}{
			"is_online":    false,
			"offline_time": time.Now().Format("2006-01-02 15:04:05"),
		})
	return result.RowsAffected, result.Error
}

// UpsertByIP 根据IP插入或更新（Agent注册使用）
func (r *HostInfoRepository) UpsertByIP(ctx context.Context, host *model.HostInfo) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "ip"}}, DoNothing: true}).
		Create(host).Error
}
