package repository

import (
	"opentraffic-ops-init-backend/internal/model"
)

// AuditLogRepository 审计日志仓储
type AuditLogRepository struct{}

// NewAuditLogRepository 创建审计日志仓储
func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{}
}

// Create 创建审计日志
func (r *AuditLogRepository) Create(log *model.AuditLog) error {
	return DB.Create(log).Error
}

// GetByID 根据ID获取日志
func (r *AuditLogRepository) GetByID(id int) (*model.AuditLog, error) {
	var log model.AuditLog
	err := DB.First(&log, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// List 获取日志列表
func (r *AuditLogRepository) List(limit int, offset int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	query := DB.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	err := query.Find(&logs).Error
	return logs, err
}

// ListByUser 根据用户ID获取日志列表
func (r *AuditLogRepository) ListByUser(userID string, limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	query := DB.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&logs).Error
	return logs, err
}
