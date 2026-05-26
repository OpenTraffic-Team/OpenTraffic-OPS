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

// AlarmRecordService 告警记录服务
type AlarmRecordService struct {
	repo *repository.AlarmRecordRepository
}

// NewAlarmRecordService 创建告警记录服务
func NewAlarmRecordService(db *gorm.DB) *AlarmRecordService {
	return &AlarmRecordService{repo: repository.NewAlarmRecordRepository(db)}
}

// toAlarmRecordDto 将告警记录模型转换为 DTO
func toAlarmRecordDto(r *model.AlarmRecord) dto.AlarmRecordDto {
	item := dto.AlarmRecordDto{
		ID:           r.ID,
		RuleID:       r.RuleID,
		RuleName:     r.RuleName,
		HostID:       r.HostID,
		HostIP:       r.HostIP,
		HostName:     r.HostName,
		AlarmType:    r.AlarmType,
		MetricType:   r.MetricType,
		CurrentValue: r.CurrentValue,
		Threshold:    r.Threshold,
		Severity:     r.Severity,
		Content:      r.Content,
		Status:       r.Status,
		TriggerTime:  r.TriggerTime,
		NotifyStatus: r.NotifyStatus,
		CreateTime:   r.CreateTime,
	}
	if r.ResolveTime != nil {
		item.ResolveTime = *r.ResolveTime
	}
	return item
}

// toAlarmRecordDtoList 批量转换告警记录模型为 DTO 切片
func toAlarmRecordDtoList(records []model.AlarmRecord) []dto.AlarmRecordDto {
	result := make([]dto.AlarmRecordDto, 0, len(records))
	for i := range records {
		result = append(result, toAlarmRecordDto(&records[i]))
	}
	return result
}

// List 告警记录列表
func (s *AlarmRecordService) List(ctx context.Context, query *dto.AlarmRecordQuery) ([]dto.AlarmRecordDto, int64, error) {
	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	records, err := s.repo.FindPage(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return toAlarmRecordDtoList(records), total, nil
}

// GetByID 根据ID获取
func (s *AlarmRecordService) GetByID(ctx context.Context, id int64) (*dto.AlarmRecordDto, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	item := toAlarmRecordDto(r)
	return &item, nil
}

// Acknowledge 确认告警
func (s *AlarmRecordService) Acknowledge(ctx context.Context, id int64) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("告警记录不存在")
	}

	return s.repo.Update(ctx, id, map[string]interface{}{
		"status": "acknowledged",
	})
}

// BatchAcknowledge 批量确认
func (s *AlarmRecordService) BatchAcknowledge(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("告警ID不能为空")
	}
	return s.repo.BatchAck(ctx, ids)
}

// Create 创建告警记录（告警引擎使用）
func (s *AlarmRecordService) Create(ctx context.Context, record *model.AlarmRecord) error {
	return s.repo.Create(ctx, record)
}

// UpdateResolved 更新为已恢复状态
func (s *AlarmRecordService) UpdateResolved(ctx context.Context, id int64) error {
	return s.repo.Update(ctx, id, map[string]interface{}{
		"status":       "resolved",
		"resolve_time": utils.NowStr(),
	})
}

// FindTriggeredByRuleAndHost 查询未恢复告警
func (s *AlarmRecordService) FindTriggeredByRuleAndHost(ctx context.Context, ruleID, hostID int64) (*model.AlarmRecord, error) {
	return s.repo.FindTriggeredByRuleAndHost(ctx, ruleID, hostID)
}

// CountUnread 统计未读告警
func (s *AlarmRecordService) CountUnread(ctx context.Context) (int64, error) {
	return s.repo.CountUnread(ctx)
}
