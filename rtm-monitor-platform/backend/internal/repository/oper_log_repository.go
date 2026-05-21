package repository

import (
	"context"

	"gorm.io/gorm"
	"rtm-server/internal/dto"
	"rtm-server/internal/model"
)

// OperLogRepository 操作日志数据访问
type OperLogRepository struct {
	db *gorm.DB
}

// NewOperLogRepository 创建操作日志仓库
func NewOperLogRepository(db *gorm.DB) *OperLogRepository {
	return &OperLogRepository{db: db}
}

// Count 统计操作日志数量
func (r *OperLogRepository) Count(ctx context.Context, query *dto.OperLogQuery) (int64, error) {
	db := r.db.WithContext(ctx).Model(&model.SysOperLog{})

	if query.Title != "" {
		db = db.Where("title LIKE ?", "%"+query.Title+"%")
	}
	if query.OperName != "" {
		db = db.Where("oper_name LIKE ?", "%"+query.OperName+"%")
	}
	if query.BusinessType != nil {
		db = db.Where("business_type = ?", *query.BusinessType)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.BeginTime != "" {
		db = db.Where("oper_time >= ?", query.BeginTime)
	}
	if query.EndTime != "" {
		db = db.Where("oper_time <= ?", query.EndTime)
	}

	var total int64
	err := db.Count(&total).Error
	return total, err
}

// FindPage 分页查询操作日志
func (r *OperLogRepository) FindPage(ctx context.Context, query *dto.OperLogQuery) ([]model.SysOperLog, error) {
	db := r.db.WithContext(ctx).Model(&model.SysOperLog{})

	if query.Title != "" {
		db = db.Where("title LIKE ?", "%"+query.Title+"%")
	}
	if query.OperName != "" {
		db = db.Where("oper_name LIKE ?", "%"+query.OperName+"%")
	}
	if query.BusinessType != nil {
		db = db.Where("business_type = ?", *query.BusinessType)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.BeginTime != "" {
		db = db.Where("oper_time >= ?", query.BeginTime)
	}
	if query.EndTime != "" {
		db = db.Where("oper_time <= ?", query.EndTime)
	}

	var logs []model.SysOperLog
	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Order("oper_id DESC").Find(&logs).Error
	return logs, err
}

// GetByID 根据ID获取操作日志
func (r *OperLogRepository) GetByID(ctx context.Context, id int64) (*model.SysOperLog, error) {
	var log model.SysOperLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// Create 创建操作日志
func (r *OperLogRepository) Create(ctx context.Context, operLog *model.SysOperLog) error {
	return r.db.WithContext(ctx).Create(operLog).Error
}

// DeleteByIDs 批量删除操作日志
func (r *OperLogRepository) DeleteByIDs(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Delete(&model.SysOperLog{}, ids).Error
}

// Clean 清空操作日志
func (r *OperLogRepository) Clean(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("TRUNCATE TABLE sys_oper_log").Error
}
