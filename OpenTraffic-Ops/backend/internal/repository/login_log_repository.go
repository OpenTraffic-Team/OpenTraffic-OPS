package repository

import (
	"context"

	"gorm.io/gorm"
	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/model"
)

// LoginLogRepository 登录日志数据访问
type LoginLogRepository struct {
	db *gorm.DB
}

// NewLoginLogRepository 创建登录日志仓库
func NewLoginLogRepository(db *gorm.DB) *LoginLogRepository {
	return &LoginLogRepository{db: db}
}

// Count 统计登录日志数量
func (r *LoginLogRepository) Count(ctx context.Context, query *dto.LoginLogQuery) (int64, error) {
	db := r.db.WithContext(ctx).Model(&model.SysLoginLog{})

	if query.UserName != "" {
		db = db.Where("user_name LIKE ?", "%"+query.UserName+"%")
	}
	if query.IPAddr != "" {
		db = db.Where("ipaddr LIKE ?", "%"+query.IPAddr+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.BeginTime != "" {
		db = db.Where("login_time >= ?", query.BeginTime)
	}
	if query.EndTime != "" {
		db = db.Where("login_time <= ?", query.EndTime)
	}

	var total int64
	err := db.Count(&total).Error
	return total, err
}

// FindPage 分页查询登录日志
func (r *LoginLogRepository) FindPage(ctx context.Context, query *dto.LoginLogQuery) ([]model.SysLoginLog, error) {
	db := r.db.WithContext(ctx).Model(&model.SysLoginLog{})

	if query.UserName != "" {
		db = db.Where("user_name LIKE ?", "%"+query.UserName+"%")
	}
	if query.IPAddr != "" {
		db = db.Where("ipaddr LIKE ?", "%"+query.IPAddr+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.BeginTime != "" {
		db = db.Where("login_time >= ?", query.BeginTime)
	}
	if query.EndTime != "" {
		db = db.Where("login_time <= ?", query.EndTime)
	}

	var logs []model.SysLoginLog
	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Order("info_id DESC").Find(&logs).Error
	return logs, err
}

// GetByID 根据ID获取登录日志
func (r *LoginLogRepository) GetByID(ctx context.Context, id int64) (*model.SysLoginLog, error) {
	var log model.SysLoginLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// Create 创建登录日志
func (r *LoginLogRepository) Create(ctx context.Context, loginLog *model.SysLoginLog) error {
	return r.db.WithContext(ctx).Create(loginLog).Error
}

// DeleteByIDs 批量删除登录日志
func (r *LoginLogRepository) DeleteByIDs(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Delete(&model.SysLoginLog{}, ids).Error
}

// Clean 清空登录日志
func (r *LoginLogRepository) Clean(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("TRUNCATE TABLE sys_login_log").Error
}
