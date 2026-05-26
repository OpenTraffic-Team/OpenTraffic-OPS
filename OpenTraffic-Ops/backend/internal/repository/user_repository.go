package repository

import (
	"context"

	"gorm.io/gorm"
	"rtm-server/internal/dto"
	"rtm-server/internal/model"
)

// UserRepository 用户数据访问
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Count 带条件统计用户数量
func (r *UserRepository) Count(ctx context.Context, query *dto.UserQuery) (int64, error) {
	db := r.db.WithContext(ctx).Model(&model.SysUser{}).Where("del_flag = '0'")

	if query.UserName != "" {
		db = db.Where("user_name LIKE ?", "%"+query.UserName+"%")
	}
	if query.NickName != "" {
		db = db.Where("nick_name LIKE ?", "%"+query.NickName+"%")
	}
	if query.Phonenumber != "" {
		db = db.Where("phonenumber LIKE ?", "%"+query.Phonenumber+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	err := db.Count(&total).Error
	return total, err
}

// FindPage 分页查询用户列表
func (r *UserRepository) FindPage(ctx context.Context, query *dto.UserQuery) ([]model.SysUser, error) {
	db := r.db.WithContext(ctx).Model(&model.SysUser{}).Where("del_flag = '0'")

	if query.UserName != "" {
		db = db.Where("user_name LIKE ?", "%"+query.UserName+"%")
	}
	if query.NickName != "" {
		db = db.Where("nick_name LIKE ?", "%"+query.NickName+"%")
	}
	if query.Phonenumber != "" {
		db = db.Where("phonenumber LIKE ?", "%"+query.Phonenumber+"%")
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var users []model.SysUser
	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Order("user_id DESC").Find(&users).Error
	return users, err
}

// FindByUsername 根据用户名查找用户
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).
		Where("user_name = ? AND del_flag = '0'", username).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据ID查找用户
func (r *UserRepository) FindByID(ctx context.Context, userID int64) (*model.SysUser, error) {
	var user model.SysUser
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND del_flag = '0'", userID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.SysUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.SysUser) error {
	return r.db.WithContext(ctx).Model(user).Updates(user).Error
}

// DeleteSoft 软删除用户
func (r *UserRepository) DeleteSoft(ctx context.Context, userIDs []int64) error {
	return r.db.WithContext(ctx).Model(&model.SysUser{}).
		Where("user_id IN ?", userIDs).
		Update("del_flag", "2").Error
}

// ResetPassword 重置密码
func (r *UserRepository) ResetPassword(ctx context.Context, userID int64, hashedPwd string) error {
	return r.db.WithContext(ctx).Model(&model.SysUser{}).
		Where("user_id = ?", userID).
		Update("password", hashedPwd).Error
}

// ChangeStatus 修改状态
func (r *UserRepository) ChangeStatus(ctx context.Context, userID int64, status string) error {
	return r.db.WithContext(ctx).Model(&model.SysUser{}).
		Where("user_id = ?", userID).
		Update("status", status).Error
}

// UpdateLoginInfo 更新登录信息
func (r *UserRepository) UpdateLoginInfo(ctx context.Context, userID int64, ip, loginTime string) error {
	return r.db.WithContext(ctx).
		Model(&model.SysUser{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"login_ip":   ip,
			"login_date": loginTime,
		}).Error
}
