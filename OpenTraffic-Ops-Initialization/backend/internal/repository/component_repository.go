package repository

import (
	"opentraffic-ops-init-backend/internal/model"
)

// ComponentRepository 组件仓储
type ComponentRepository struct{}

// NewComponentRepository 创建组件仓储
func NewComponentRepository() *ComponentRepository {
	return &ComponentRepository{}
}

// Create 创建组件
func (r *ComponentRepository) Create(component *model.Component) error {
	return DB.Create(component).Error
}

// GetByID 根据ID获取组件
func (r *ComponentRepository) GetByID(id string) (*model.Component, error) {
	var component model.Component
	err := DB.First(&component, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// GetByName 根据名称获取组件
func (r *ComponentRepository) GetByName(name string) (*model.Component, error) {
	var component model.Component
	err := DB.First(&component, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &component, nil
}

// List 获取组件列表
func (r *ComponentRepository) List() ([]model.Component, error) {
	var components []model.Component
	err := DB.Find(&components).Error
	return components, err
}

// Update 更新组件
func (r *ComponentRepository) Update(component *model.Component) error {
	return DB.Save(component).Error
}

// Delete 删除组件
func (r *ComponentRepository) Delete(id string) error {
	return DB.Delete(&model.Component{}, "id = ?", id).Error
}

// UpdateStatus 更新组件状态
func (r *ComponentRepository) UpdateStatus(id string, status model.ComponentStatus) error {
	return DB.Model(&model.Component{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// UpdateContainerID 更新容器ID
func (r *ComponentRepository) UpdateContainerID(id, containerID string) error {
	return DB.Model(&model.Component{}).
		Where("id = ?", id).
		Update("container_id", containerID).Error
}
