package repository

import (
	"rtm-initialization-backend/internal/model"
)

// UserRepository 用户仓储
type UserRepository struct{}

// NewUserRepository 创建用户仓储
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
	return DB.Create(user).Error
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(id string) (*model.User, error) {
	var user model.User
	err := DB.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := DB.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表
func (r *UserRepository) List() ([]model.User, error) {
	var users []model.User
	err := DB.Find(&users).Error
	return users, err
}

// Update 更新用户
func (r *UserRepository) Update(user *model.User) error {
	return DB.Save(user).Error
}

// Delete 删除用户
func (r *UserRepository) Delete(id string) error {
	return DB.Delete(&model.User{}, "id = ?", id).Error
}
