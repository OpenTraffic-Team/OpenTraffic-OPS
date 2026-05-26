package service

import (
	"context"
	"fmt"
	"opentraffic-ops-init-backend/internal/model"
	"opentraffic-ops-init-backend/internal/repository"
	"opentraffic-ops-init-backend/pkg/docker"
)

// MonitorService 监控服务
type MonitorService struct {
	componentRepo *repository.ComponentRepository
	dockerManager *docker.Manager
}

// NewMonitorService 创建监控服务
func NewMonitorService() (*MonitorService, error) {
	dockerMgr, err := docker.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker manager: %w", err)
	}

	return &MonitorService{
		componentRepo: repository.NewComponentRepository(),
		dockerManager: dockerMgr,
	}, nil
}

// Close 关闭服务
func (s *MonitorService) Close() error {
	return s.dockerManager.Close()
}

// Overview 总览信息
type Overview struct {
	TotalComponents    int                       `json:"total_components"`
	RunningComponents  int                       `json:"running_components"`
	StoppedComponents  int                       `json:"stopped_components"`
	ErrorComponents    int                       `json:"error_components"`
	ComponentsByType   map[string]int            `json:"components_by_type"`
}

// GetOverview 获取总览信息
func (s *MonitorService) GetOverview(ctx context.Context) (*Overview, error) {
	components, err := s.componentRepo.List()
	if err != nil {
		return nil, err
	}

	overview := &Overview{
		TotalComponents:  len(components),
		ComponentsByType: make(map[string]int),
	}

	for i := range components {
		s.syncComponentStatus(ctx, &components[i])
		switch components[i].Status {
		case model.ComponentStatusRunning:
			overview.RunningComponents++
		case model.ComponentStatusStopped:
			overview.StoppedComponents++
		case model.ComponentStatusError:
			overview.ErrorComponents++
		}

		overview.ComponentsByType[string(components[i].Type)]++
	}

	return overview, nil
}

// ComponentDetail 组件详情
type ComponentDetail struct {
	Component *model.Component     `json:"component"`
	Info      *docker.ContainerInfo `json:"info"`
	Stats     *docker.ContainerStats `json:"stats"`
}

// GetComponentDetail 获取组件详情
func (s *MonitorService) GetComponentDetail(ctx context.Context, id string) (*ComponentDetail, error) {
	component, err := s.componentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("component not found: %w", err)
	}

	s.syncComponentStatus(ctx, component)

	detail := &ComponentDetail{
		Component: component,
	}

	// 获取容器信息
	if component.ContainerID != "" {
		info, err := s.dockerManager.GetContainerInfo(ctx, component.ContainerID)
		if err == nil {
			detail.Info = info
		}

		// 获取统计信息
		stats, err := s.dockerManager.GetContainerStats(ctx, component.ContainerID)
		if err == nil {
			detail.Stats = stats
		}
	}

	return detail, nil
}

// GetAllComponentDetails 获取所有组件详情
func (s *MonitorService) GetAllComponentDetails(ctx context.Context) ([]ComponentDetail, error) {
	components, err := s.componentRepo.List()
	if err != nil {
		return nil, err
	}

	details := make([]ComponentDetail, 0, len(components))
	for _, comp := range components {
		detail, err := s.GetComponentDetail(ctx, comp.ID)
		if err != nil {
			// 如果获取详情失败，至少包含组件基本信息
			detail = &ComponentDetail{Component: &comp}
		}
		details = append(details, *detail)
	}

	return details, nil
}

// syncComponentStatus 同步组件的实时 Docker 状态到数据库
func (s *MonitorService) syncComponentStatus(ctx context.Context, component *model.Component) {
	if component.ContainerID == "" {
		return
	}
	info, err := s.dockerManager.GetContainerInfo(ctx, component.ContainerID)
	if err != nil {
		// 容器不存在或 Docker 不可用
		if component.Status != model.ComponentStatusError {
			component.Status = model.ComponentStatusError
			_ = s.componentRepo.UpdateStatus(component.ID, model.ComponentStatusError)
		}
		return
	}

	var newStatus model.ComponentStatus
	switch info.State {
	case "running", "restarting":
		newStatus = model.ComponentStatusRunning
	case "stopped", "paused":
		newStatus = model.ComponentStatusStopped
	case "dead", "oom_killed":
		newStatus = model.ComponentStatusError
	default:
		newStatus = model.ComponentStatusError
	}

	if component.Status != newStatus {
		component.Status = newStatus
		_ = s.componentRepo.UpdateStatus(component.ID, newStatus)
	}
}
