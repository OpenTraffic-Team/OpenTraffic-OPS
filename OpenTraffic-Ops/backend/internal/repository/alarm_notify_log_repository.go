package repository

import (
	"context"

	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/model"
)

// AlarmNotifyLogRepository 告警通知日志数据访问
type AlarmNotifyLogRepository struct {
	db *gorm.DB
}

// NewAlarmNotifyLogRepository 创建告警通知日志仓库
func NewAlarmNotifyLogRepository(db *gorm.DB) *AlarmNotifyLogRepository {
	return &AlarmNotifyLogRepository{db: db}
}

// Create 创建日志
func (r *AlarmNotifyLogRepository) Create(ctx context.Context, log *model.AlarmNotifyLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindPage 分页查询
func (r *AlarmNotifyLogRepository) FindPage(ctx context.Context, query *dto.AlarmNotifyLogQuery) ([]model.AlarmNotifyLog, error) {
	db := r.db.WithContext(ctx).Model(&model.AlarmNotifyLog{})

	if query.RecordID > 0 {
		db = db.Where("record_id = ?", query.RecordID)
	}
	if query.ChannelID > 0 {
		db = db.Where("channel_id = ?", query.ChannelID)
	}
	if query.ChannelType != "" {
		db = db.Where("channel_type = ?", query.ChannelType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var logs []model.AlarmNotifyLog
	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&logs).Error
	return logs, err
}

// Count 统计
func (r *AlarmNotifyLogRepository) Count(ctx context.Context, query *dto.AlarmNotifyLogQuery) (int64, error) {
	db := r.db.WithContext(ctx).Model(&model.AlarmNotifyLog{})

	if query.RecordID > 0 {
		db = db.Where("record_id = ?", query.RecordID)
	}
	if query.ChannelID > 0 {
		db = db.Where("channel_id = ?", query.ChannelID)
	}
	if query.ChannelType != "" {
		db = db.Where("channel_type = ?", query.ChannelType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	err := db.Count(&total).Error
	return total, err
}
