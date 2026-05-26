package service

import (
	"context"

	"gorm.io/gorm"
	"rtm-server/internal/dto"
	"rtm-server/internal/model"
	"rtm-server/internal/repository"
	"rtm-server/internal/utils"
)

// OperLogService 操作日志服务
type OperLogService struct {
	repo *repository.OperLogRepository
}

// NewOperLogService 创建操作日志服务
func NewOperLogService(db *gorm.DB) *OperLogService {
	return &OperLogService{
		repo: repository.NewOperLogRepository(db),
	}
}

// List 操作日志列表
func (s *OperLogService) List(ctx context.Context, query *dto.OperLogQuery) ([]model.SysOperLog, int64, error) {
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

// GetByID 根据ID获取操作日志
func (s *OperLogService) GetByID(ctx context.Context, id int64) (*model.SysOperLog, error) {
	return s.repo.GetByID(ctx, id)
}

// Create 创建操作日志
func (s *OperLogService) Create(ctx context.Context, req *dto.OperLogCreateRequest) error {
	operLog := &model.SysOperLog{
		Title:         req.Title,
		BusinessType:  req.BusinessType,
		Method:        req.Method,
		RequestMethod: req.RequestMethod,
		OperatorType:  req.OperatorType,
		OperName:      req.OperName,
		DeptName:      req.DeptName,
		OperURL:       req.OperURL,
		OperIP:        req.OperIP,
		OperLocation:  req.OperLocation,
		OperParam:     req.OperParam,
		JsonResult:    req.JsonResult,
		Status:        req.Status,
		ErrorMsg:      req.ErrorMsg,
		OperTime:      utils.NowStr(),
		CostTime:      req.CostTime,
	}
	return s.repo.Create(ctx, operLog)
}

// DeleteByIDs 批量删除操作日志
func (s *OperLogService) DeleteByIDs(ctx context.Context, ids []int64) error {
	return s.repo.DeleteByIDs(ctx, ids)
}

// Clean 清空操作日志
func (s *OperLogService) Clean(ctx context.Context) error {
	return s.repo.Clean(ctx)
}
