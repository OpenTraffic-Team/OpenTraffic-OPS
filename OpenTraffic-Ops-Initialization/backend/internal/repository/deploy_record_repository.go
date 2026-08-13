package repository

import (
	"opentraffic-ops-init-backend/internal/model"

	"gorm.io/gorm"
)

// DeployRecordRepository 部署记录仓储
type DeployRecordRepository struct{}

// NewDeployRecordRepository 创建部署记录仓储
func NewDeployRecordRepository() *DeployRecordRepository {
	return &DeployRecordRepository{}
}

// Create 创建部署记录
func (r *DeployRecordRepository) Create(record *model.DeployRecord) error {
	return DB.Create(record).Error
}

// GetByID 根据ID获取部署记录
func (r *DeployRecordRepository) GetByID(id int) (*model.DeployRecord, error) {
	var record model.DeployRecord
	err := DB.First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// List 获取部署记录列表
func (r *DeployRecordRepository) List() ([]model.DeployRecord, error) {
	var records []model.DeployRecord
	err := DB.Order("created_at DESC").Find(&records).Error
	return records, err
}

// ListByServerID 根据服务器ID获取部署记录
func (r *DeployRecordRepository) ListByServerID(serverID string) ([]model.DeployRecord, error) {
	var records []model.DeployRecord
	err := DB.Where("server_id = ?", serverID).Order("created_at DESC").Find(&records).Error
	return records, err
}

// Update 更新部署记录
func (r *DeployRecordRepository) Update(record *model.DeployRecord) error {
	return DB.Save(record).Error
}

// UpdateStatus 更新部署状态
func (r *DeployRecordRepository) UpdateStatus(id int, status model.DeployStatus, log string) error {
	return DB.Model(&model.DeployRecord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status,
			"log":    log,
		}).Error
}

// UpdateProgress 增量更新部署进度与日志，供部署后台任务上报，不触碰状态
func (r *DeployRecordRepository) UpdateProgress(id int, progress int, log string) error {
	return DB.Model(&model.DeployRecord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"progress": progress,
			"log":      log,
		}).Error
}

// UpdateRemotePath 只更新部署记录的远程路径，避免整行 Save 覆盖后台任务写入的进度
func (r *DeployRecordRepository) UpdateRemotePath(id int, remotePath string) error {
	return DB.Model(&model.DeployRecord{}).
		Where("id = ?", id).
		Update("remote_path", remotePath).Error
}

// FailStalePending 将残留 pending 状态的记录标记为失败（服务重启导致部署中断）
func (r *DeployRecordRepository) FailStalePending() error {
	return DB.Model(&model.DeployRecord{}).
		Where("status = ?", string(model.DeployStatusPending)).
		Updates(map[string]interface{}{
			"status": string(model.DeployStatusFailed),
			"log":    gorm.Expr("log || ?", "\n[ERROR] 服务重启，部署中断"),
		}).Error
}

// HasSuccessfulDeploy 检查指定服务器是否已成功部署过指定二进制文件
func (r *DeployRecordRepository) HasSuccessfulDeploy(serverID string, binaryName string) (bool, error) {
	var count int64
	err := DB.Model(&model.DeployRecord{}).
		Where("server_id = ? AND binary_name = ? AND status = ?", serverID, binaryName, string(model.DeployStatusSuccess)).
		Count(&count).Error
	return count > 0, err
}

// GetLatestSuccessfulDeploy 获取指定服务器指定二进制文件的最新成功部署记录
func (r *DeployRecordRepository) GetLatestSuccessfulDeploy(serverID string, binaryName string) (*model.DeployRecord, error) {
	var record model.DeployRecord
	err := DB.Where("server_id = ? AND binary_name = ? AND status = ?", serverID, binaryName, string(model.DeployStatusSuccess)).
		Order("created_at DESC").First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// DeleteByServerAndBinary 删除指定服务器指定二进制文件的所有部署记录
func (r *DeployRecordRepository) DeleteByServerAndBinary(serverID string, binaryName string) error {
	return DB.Where("server_id = ? AND binary_name = ?", serverID, binaryName).Delete(&model.DeployRecord{}).Error
}
