package service

import (
	"context"

	"gorm.io/gorm"
	"rtm-server/internal/dto"
	"rtm-server/internal/model"
	"rtm-server/internal/repository"
	"rtm-server/internal/utils"
)

// LoginLogService 登录日志服务
type LoginLogService struct {
	repo *repository.LoginLogRepository
}

// NewLoginLogService 创建登录日志服务
func NewLoginLogService(db *gorm.DB) *LoginLogService {
	return &LoginLogService{
		repo: repository.NewLoginLogRepository(db),
	}
}

// List 登录日志列表
func (s *LoginLogService) List(ctx context.Context, query *dto.LoginLogQuery) ([]model.SysLoginLog, int64, error) {
	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	logs, err := s.repo.FindPage(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID 根据ID获取登录日志
func (s *LoginLogService) GetByID(ctx context.Context, id int64) (*model.SysLoginLog, error) {
	return s.repo.GetByID(ctx, id)
}

// Create 创建登录日志
func (s *LoginLogService) Create(ctx context.Context, req *dto.LoginLogCreateRequest) error {
	loginLog := &model.SysLoginLog{
		UserName:      req.UserName,
		IPAddr:        req.IPAddr,
		LoginLocation: req.LoginLocation,
		Browser:       req.Browser,
		OS:            req.OS,
		Status:        req.Status,
		Msg:           req.Msg,
		LoginTime:     utils.NowStr(),
	}
	return s.repo.Create(ctx, loginLog)
}

// DeleteByIDs 批量删除登录日志
func (s *LoginLogService) DeleteByIDs(ctx context.Context, ids []int64) error {
	return s.repo.DeleteByIDs(ctx, ids)
}

// Clean 清空登录日志
func (s *LoginLogService) Clean(ctx context.Context) error {
	return s.repo.Clean(ctx)
}
