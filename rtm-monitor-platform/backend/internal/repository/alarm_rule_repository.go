package repository

import (
	"context"

	"gorm.io/gorm"

	"rtm-server/internal/dto"
	"rtm-server/internal/model"
)

// AlarmRuleRepository 告警规则数据访问
type AlarmRuleRepository struct {
	db *gorm.DB
}

// NewAlarmRuleRepository 创建告警规则仓库
func NewAlarmRuleRepository(db *gorm.DB) *AlarmRuleRepository {
	return &AlarmRuleRepository{db: db}
}

// FindAll 查询所有规则
func (r *AlarmRuleRepository) FindAll(ctx context.Context) ([]model.AlarmRule, error) {
	var rules []model.AlarmRule
	err := r.db.WithContext(ctx).Where("del_flag = ?", "0").Order("id DESC").Find(&rules).Error
	return rules, err
}

// FindByID 根据ID查询
func (r *AlarmRuleRepository) FindByID(ctx context.Context, id int64) (*model.AlarmRule, error) {
	var rule model.AlarmRule
	err := r.db.WithContext(ctx).Where("del_flag = ?", "0").First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// FindEnabled 查询所有启用的规则
func (r *AlarmRuleRepository) FindEnabled(ctx context.Context) ([]model.AlarmRule, error) {
	var rules []model.AlarmRule
	err := r.db.WithContext(ctx).Where("del_flag = ? AND status = ?", "0", "0").Find(&rules).Error
	return rules, err
}

// Count 带条件统计
func (r *AlarmRuleRepository) Count(ctx context.Context, query *dto.AlarmRuleQuery) (int64, error) {
	db := r.db.WithContext(ctx).Model(&model.AlarmRule{}).Where("del_flag = ?", "0")

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.RuleType != "" {
		db = db.Where("rule_type = ?", query.RuleType)
	}
	if query.MetricType != "" {
		db = db.Where("metric_type = ?", query.MetricType)
	}
	if query.HostID > 0 {
		db = db.Where("host_id = ? OR host_id = 0", query.HostID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	err := db.Count(&total).Error
	return total, err
}

// FindPage 分页查询
func (r *AlarmRuleRepository) FindPage(ctx context.Context, query *dto.AlarmRuleQuery) ([]model.AlarmRule, error) {
	db := r.db.WithContext(ctx).Model(&model.AlarmRule{}).Where("del_flag = ?", "0")

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.RuleType != "" {
		db = db.Where("rule_type = ?", query.RuleType)
	}
	if query.MetricType != "" {
		db = db.Where("metric_type = ?", query.MetricType)
	}
	if query.HostID > 0 {
		db = db.Where("host_id = ? OR host_id = 0", query.HostID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var rules []model.AlarmRule
	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&rules).Error
	return rules, err
}

// Create 创建
func (r *AlarmRuleRepository) Create(ctx context.Context, rule *model.AlarmRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// Update 更新
func (r *AlarmRuleRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.AlarmRule{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteByIDs 批量删除（逻辑删除）
func (r *AlarmRuleRepository) DeleteByIDs(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.AlarmRule{}).Where("id IN ?", ids).Update("del_flag", "2").Error
}
