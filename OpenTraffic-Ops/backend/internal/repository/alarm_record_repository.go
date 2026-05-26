package repository

import (
	"context"

	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/model"
)

// AlarmRecordRepository 告警记录数据访问
type AlarmRecordRepository struct {
	db *gorm.DB
}

// NewAlarmRecordRepository 创建告警记录仓库
func NewAlarmRecordRepository(db *gorm.DB) *AlarmRecordRepository {
	return &AlarmRecordRepository{db: db}
}

// FindByID 根据ID查询
func (r *AlarmRecordRepository) FindByID(ctx context.Context, id int64) (*model.AlarmRecord, error) {
	var record model.AlarmRecord
	err := r.db.WithContext(ctx).First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Count 带条件统计
func (r *AlarmRecordRepository) Count(ctx context.Context, query *dto.AlarmRecordQuery) (int64, error) {
	db := r.db.WithContext(ctx).Model(&model.AlarmRecord{})

	if query.RuleID > 0 {
		db = db.Where("rule_id = ?", query.RuleID)
	}
	if query.HostID > 0 {
		db = db.Where("host_id = ?", query.HostID)
	}
	if query.HostIP != "" {
		db = db.Where("host_ip LIKE ?", "%"+query.HostIP+"%")
	}
	if query.AlarmType != "" {
		db = db.Where("alarm_type = ?", query.AlarmType)
	}
	if query.MetricType != "" {
		db = db.Where("metric_type = ?", query.MetricType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Severity != "" {
		db = db.Where("severity = ?", query.Severity)
	}
	if query.BeginTime != "" {
		db = db.Where("trigger_time >= ?", query.BeginTime)
	}
	if query.EndTime != "" {
		db = db.Where("trigger_time <= ?", query.EndTime)
	}

	var total int64
	err := db.Count(&total).Error
	return total, err
}

// FindPage 分页查询
func (r *AlarmRecordRepository) FindPage(ctx context.Context, query *dto.AlarmRecordQuery) ([]model.AlarmRecord, error) {
	db := r.db.WithContext(ctx).Model(&model.AlarmRecord{})

	if query.RuleID > 0 {
		db = db.Where("rule_id = ?", query.RuleID)
	}
	if query.HostID > 0 {
		db = db.Where("host_id = ?", query.HostID)
	}
	if query.HostIP != "" {
		db = db.Where("host_ip LIKE ?", "%"+query.HostIP+"%")
	}
	if query.AlarmType != "" {
		db = db.Where("alarm_type = ?", query.AlarmType)
	}
	if query.MetricType != "" {
		db = db.Where("metric_type = ?", query.MetricType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Severity != "" {
		db = db.Where("severity = ?", query.Severity)
	}
	if query.BeginTime != "" {
		db = db.Where("trigger_time >= ?", query.BeginTime)
	}
	if query.EndTime != "" {
		db = db.Where("trigger_time <= ?", query.EndTime)
	}

	var records []model.AlarmRecord
	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Order("trigger_time DESC").Find(&records).Error
	return records, err
}

// Create 创建
func (r *AlarmRecordRepository) Create(ctx context.Context, record *model.AlarmRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// Update 更新
func (r *AlarmRecordRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.AlarmRecord{}).Where("id = ?", id).Updates(updates).Error
}

// BatchAck 批量确认
func (r *AlarmRecordRepository) BatchAck(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.AlarmRecord{}).Where("id IN ?", ids).Update("status", "acknowledged").Error
}

// FindTriggeredByRuleAndHost 查询指定规则和主机的未恢复告警
func (r *AlarmRecordRepository) FindTriggeredByRuleAndHost(ctx context.Context, ruleID, hostID int64) (*model.AlarmRecord, error) {
	var record model.AlarmRecord
	err := r.db.WithContext(ctx).
		Where("rule_id = ? AND host_id = ? AND status = ?", ruleID, hostID, "triggered").
		Order("trigger_time DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// CountUnread 统计未读（未恢复且未确认）告警数量
func (r *AlarmRecordRepository) CountUnread(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AlarmRecord{}).
		Where("status = ?", "triggered").
		Count(&count).Error
	return count, err
}

// FindRecent 查询最近 N 条告警记录（按 trigger_time 倒序）
func (r *AlarmRecordRepository) FindRecent(ctx context.Context, limit int) ([]model.AlarmRecord, error) {
	var records []model.AlarmRecord
	err := r.db.WithContext(ctx).
		Order("trigger_time DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}
