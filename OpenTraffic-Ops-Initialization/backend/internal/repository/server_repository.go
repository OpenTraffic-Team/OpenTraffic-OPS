package repository

import (
	"rtm-initialization-backend/internal/model"
)

// ServerRepository 服务器仓储
type ServerRepository struct{}

// NewServerRepository 创建服务器仓储
func NewServerRepository() *ServerRepository {
	return &ServerRepository{}
}

// Create 创建服务器
func (r *ServerRepository) Create(server *model.Server) error {
	return DB.Create(server).Error
}

// GetByID 根据ID获取服务器
func (r *ServerRepository) GetByID(id string) (*model.Server, error) {
	var server model.Server
	err := DB.First(&server, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// GetByName 根据名称获取服务器
func (r *ServerRepository) GetByName(name string) (*model.Server, error) {
	var server model.Server
	err := DB.First(&server, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// List 获取服务器列表
func (r *ServerRepository) List() ([]model.Server, error) {
	var servers []model.Server
	err := DB.Find(&servers).Error
	return servers, err
}

// Update 更新服务器
func (r *ServerRepository) Update(server *model.Server) error {
	return DB.Save(server).Error
}

// Delete 删除服务器
func (r *ServerRepository) Delete(id string) error {
	return DB.Delete(&model.Server{}, "id = ?", id).Error
}
