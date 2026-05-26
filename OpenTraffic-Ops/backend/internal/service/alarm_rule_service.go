package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/model"
	"opentraffic-ops-backend/internal/repository"
	"opentraffic-ops-backend/internal/utils"
)

// AlarmRuleService 告警规则服务
type AlarmRuleService struct {
	repo *repository.AlarmRuleRepository
}

// NewAlarmRuleService 创建告警规则服务
func NewAlarmRuleService(db *gorm.DB) *AlarmRuleService {
	return &AlarmRuleService{repo: repository.NewAlarmRuleRepository(db)}
}

// List 告警规则列表
func (s *AlarmRuleService) List(ctx context.Context, query *dto.AlarmRuleQuery) ([]dto.AlarmRuleDto, int64, error) {
	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	rules, err := s.repo.FindPage(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.AlarmRuleDto, 0, len(rules))
	for _, rule := range rules {
		result = append(result, dto.AlarmRuleDto{
			ID:         rule.ID,
			Name:       rule.Name,
			RuleType:   rule.RuleType,
			MetricType: rule.MetricType,
			HostID:     rule.HostID,
			Threshold:  rule.Threshold,
			CompareOp:  rule.CompareOp,
			Duration:   rule.Duration,
			Severity:   rule.Severity,
			ChannelIDs: rule.ChannelIDs,
			Status:     rule.Status,
			Remark:     rule.Remark,
			CreateBy:   rule.CreateBy,
			CreateTime: rule.CreateTime,
			UpdateBy:   rule.UpdateBy,
			UpdateTime: rule.UpdateTime,
		})
	}
	return result, total, nil
}

// GetByID 根据ID获取
func (s *AlarmRuleService) GetByID(ctx context.Context, id int64) (*dto.AlarmRuleDto, error) {
	rule, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.AlarmRuleDto{
		ID:         rule.ID,
		Name:       rule.Name,
		RuleType:   rule.RuleType,
		MetricType: rule.MetricType,
		HostID:     rule.HostID,
		Threshold:  rule.Threshold,
		CompareOp:  rule.CompareOp,
		Duration:   rule.Duration,
		Severity:   rule.Severity,
		ChannelIDs: rule.ChannelIDs,
		Status:     rule.Status,
		Remark:     rule.Remark,
		CreateBy:   rule.CreateBy,
		CreateTime: rule.CreateTime,
		UpdateBy:   rule.UpdateBy,
		UpdateTime: rule.UpdateTime,
	}, nil
}

// Create 创建告警规则
func (s *AlarmRuleService) Create(ctx context.Context, req *dto.AlarmRuleCreateRequest, createBy string) error {
	rule := &model.AlarmRule{
		Name:       req.Name,
		RuleType:   req.RuleType,
		MetricType: req.MetricType,
		HostID:     req.HostID,
		Threshold:  req.Threshold,
		CompareOp:  req.CompareOp,
		Duration:   req.Duration,
		Severity:   req.Severity,
		ChannelIDs: req.ChannelIDs,
		Status:     req.Status,
		BaseEntity: model.BaseEntity{
			Remark:   req.Remark,
			CreateBy: createBy,
		},
	}
	if rule.Status == "" {
		rule.Status = "0"
	}
	if rule.Severity == "" {
		rule.Severity = "warning"
	}
	if rule.CompareOp == "" {
		rule.CompareOp = "gt"
	}
	return s.repo.Create(ctx, rule)
}

// Update 更新告警规则
func (s *AlarmRuleService) Update(ctx context.Context, req *dto.AlarmRuleUpdateRequest, updateBy string) error {
	_, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("告警规则不存在")
	}

	return s.repo.Update(ctx, req.ID, map[string]interface{}{
		"name":        req.Name,
		"rule_type":   req.RuleType,
		"metric_type": req.MetricType,
		"host_id":     req.HostID,
		"threshold":   req.Threshold,
		"compare_op":  req.CompareOp,
		"duration":    req.Duration,
		"severity":    req.Severity,
		"channel_ids": req.ChannelIDs,
		"status":      req.Status,
		"remark":      req.Remark,
		"update_by":   updateBy,
		"update_time": utils.NowStr(),
	})
}

// DeleteByIDs 批量删除
func (s *AlarmRuleService) DeleteByIDs(ctx context.Context, ids []int64) error {
	return s.repo.DeleteByIDs(ctx, ids)
}

// UpdateStatus 更新状态
func (s *AlarmRuleService) UpdateStatus(ctx context.Context, req *dto.AlarmRuleStatusRequest, updateBy string) error {
	return s.repo.Update(ctx, req.ID, map[string]interface{}{
		"status":      req.Status,
		"update_by":   updateBy,
		"update_time": utils.NowStr(),
	})
}

// FindEnabled 查询所有启用的规则（告警引擎使用）
func (s *AlarmRuleService) FindEnabled(ctx context.Context) ([]model.AlarmRule, error) {
	return s.repo.FindEnabled(ctx)
}
