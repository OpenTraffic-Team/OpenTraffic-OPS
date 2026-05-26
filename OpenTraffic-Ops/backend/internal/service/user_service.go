package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"rtm-server/internal/dto"
	"rtm-server/internal/model"
	"rtm-server/internal/repository"
	"rtm-server/pkg/crypto"
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(db),
	}
}

// List 用户列表
func (s *UserService) List(ctx context.Context, query *dto.UserQuery) ([]model.SysUser, int64, error) {
	total, err := s.userRepo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	users, err := s.userRepo.FindPage(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(ctx context.Context, userID int64) (*model.SysUser, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// Create 创建用户
func (s *UserService) Create(ctx context.Context, req *dto.UserCreateRequest, createBy string) error {
	// 检查用户名唯一
	if _, err := s.userRepo.FindByUsername(ctx, req.UserName); err == nil {
		return fmt.Errorf("新增用户'%s'失败，登录账号已存在", req.UserName)
	}

	// 加密密码
	hashedPwd, err := crypto.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &model.SysUser{
		UserName:    req.UserName,
		NickName:    req.NickName,
		Password:    hashedPwd,
		Email:       req.Email,
		Phonenumber: req.Phonenumber,
		Sex:         req.Sex,
		Status:      req.Status,
		BaseEntity: model.BaseEntity{
			Remark: req.Remark,
		},
	}

	return s.userRepo.Create(ctx, user)
}

// Update 更新用户
func (s *UserService) Update(ctx context.Context, req *dto.UserUpdateRequest, updateBy string) error {
	user := &model.SysUser{
		UserID:      req.UserID,
		UserName:    req.UserName,
		NickName:    req.NickName,
		Email:       req.Email,
		Phonenumber: req.Phonenumber,
		Sex:         req.Sex,
		Status:      req.Status,
		BaseEntity: model.BaseEntity{
			Remark: req.Remark,
		},
	}

	return s.userRepo.Update(ctx, user)
}

// Delete 删除用户
func (s *UserService) Delete(ctx context.Context, userIDs []int64) error {
	return s.userRepo.DeleteSoft(ctx, userIDs)
}

// ResetPwd 重置密码
func (s *UserService) ResetPwd(ctx context.Context, userID int64, password string) error {
	hashedPwd, err := crypto.HashPassword(password)
	if err != nil {
		return err
	}
	return s.userRepo.ResetPassword(ctx, userID, hashedPwd)
}

// ChangeStatus 修改状态
func (s *UserService) ChangeStatus(ctx context.Context, userID int64, status string) error {
	return s.userRepo.ChangeStatus(ctx, userID, status)
}
