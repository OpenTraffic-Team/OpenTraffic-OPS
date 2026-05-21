package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"rtm-server/internal/dto"
	"rtm-server/internal/model"
	"rtm-server/internal/repository"
	"rtm-server/internal/utils"
)

// AlarmChannelService 告警通道服务
type AlarmChannelService struct {
	repo *repository.AlarmChannelRepository
}

// NewAlarmChannelService 创建告警通道服务
func NewAlarmChannelService(db *gorm.DB) *AlarmChannelService {
	return &AlarmChannelService{repo: repository.NewAlarmChannelRepository(db)}
}

// List 告警通道列表
func (s *AlarmChannelService) List(ctx context.Context, query *dto.AlarmChannelQuery) ([]dto.AlarmChannelDto, int64, error) {
	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	channels, err := s.repo.FindPage(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.AlarmChannelDto, 0, len(channels))
	for _, ch := range channels {
		result = append(result, dto.AlarmChannelDto{
			ID:          ch.ID,
			Name:        ch.Name,
			ChannelType: ch.ChannelType,
			Config:      ch.Config,
			Status:      ch.Status,
			IsDefault:   ch.IsDefault,
			Remark:      ch.Remark,
			CreateBy:    ch.CreateBy,
			CreateTime:  ch.CreateTime,
			UpdateBy:    ch.UpdateBy,
			UpdateTime:  ch.UpdateTime,
		})
	}
	return result, total, nil
}

// GetByID 根据ID获取
func (s *AlarmChannelService) GetByID(ctx context.Context, id int64) (*dto.AlarmChannelDto, error) {
	ch, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &dto.AlarmChannelDto{
		ID:          ch.ID,
		Name:        ch.Name,
		ChannelType: ch.ChannelType,
		Config:      ch.Config,
		Status:      ch.Status,
		IsDefault:   ch.IsDefault,
		Remark:      ch.Remark,
		CreateBy:    ch.CreateBy,
		CreateTime:  ch.CreateTime,
		UpdateBy:    ch.UpdateBy,
		UpdateTime:  ch.UpdateTime,
	}, nil
}

// Create 创建告警通道
func (s *AlarmChannelService) Create(ctx context.Context, req *dto.AlarmChannelCreateRequest, createBy string) error {
	channel := &model.AlarmChannel{
		Name:        req.Name,
		ChannelType: req.ChannelType,
		Config:      req.Config,
		Status:      req.Status,
		IsDefault:   req.IsDefault,
		BaseEntity: model.BaseEntity{
			Remark:   req.Remark,
			CreateBy: createBy,
		},
	}
	if channel.Status == "" {
		channel.Status = "0"
	}
	if channel.IsDefault == "" {
		channel.IsDefault = "0"
	}
	return s.repo.Create(ctx, channel)
}

// Update 更新告警通道
func (s *AlarmChannelService) Update(ctx context.Context, req *dto.AlarmChannelUpdateRequest, updateBy string) error {
	_, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("告警通道不存在")
	}

	return s.repo.Update(ctx, req.ID, map[string]interface{}{
		"name":         req.Name,
		"channel_type": req.ChannelType,
		"config":       req.Config,
		"status":       req.Status,
		"is_default":   req.IsDefault,
		"remark":       req.Remark,
		"update_by":    updateBy,
		"update_time":  utils.NowStr(),
	})
}

// DeleteByIDs 批量删除
func (s *AlarmChannelService) DeleteByIDs(ctx context.Context, ids []int64) error {
	return s.repo.DeleteByIDs(ctx, ids)
}

// UpdateStatus 更新状态
func (s *AlarmChannelService) UpdateStatus(ctx context.Context, req *dto.AlarmChannelStatusRequest, updateBy string) error {
	return s.repo.Update(ctx, req.ID, map[string]interface{}{
		"status":      req.Status,
		"update_by":   updateBy,
		"update_time": utils.NowStr(),
	})
}

// SetDefault 设置默认通道
func (s *AlarmChannelService) SetDefault(ctx context.Context, req *dto.AlarmChannelSetDefaultRequest, updateBy string) error {
	ch, err := s.repo.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("告警通道不存在")
	}

	// 如果设为默认，先清除同类型的其他默认通道
	if req.IsDefault == "1" {
		if err := s.repo.ClearDefault(ctx, ch.ChannelType); err != nil {
			return err
		}
	}

	return s.repo.Update(ctx, req.ID, map[string]interface{}{
		"is_default":  req.IsDefault,
		"update_by":   updateBy,
		"update_time": utils.NowStr(),
	})
}

// FindAll 查询所有通道（用于下拉选择）
func (s *AlarmChannelService) FindAll(ctx context.Context) ([]dto.AlarmChannelDto, error) {
	channels, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.AlarmChannelDto, 0, len(channels))
	for _, ch := range channels {
		result = append(result, dto.AlarmChannelDto{
			ID:          ch.ID,
			Name:        ch.Name,
			ChannelType: ch.ChannelType,
			Config:      ch.Config,
			Status:      ch.Status,
			IsDefault:   ch.IsDefault,
			Remark:      ch.Remark,
		})
	}
	return result, nil
}
