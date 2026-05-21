package service

import (
	"context"
	"fmt"
	"os"
	"rtm-initialization-backend/internal/model"
	"rtm-initialization-backend/internal/repository"
	"rtm-initialization-backend/pkg/assets"
	"rtm-initialization-backend/pkg/crypto"
	"rtm-initialization-backend/pkg/docker"
	"strings"
	"time"
)

// ComponentService 组件服务
type ComponentService struct {
	componentRepo    *repository.ComponentRepository
	auditLogRepo     *repository.AuditLogRepository
	dockerManager    *docker.Manager
	encryptor        *crypto.Encryptor
}

// NewComponentService 创建组件服务
func NewComponentService(encryptor *crypto.Encryptor) (*ComponentService, error) {
	dockerMgr, err := docker.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker manager: %w", err)
	}

	return &ComponentService{
		componentRepo:       repository.NewComponentRepository(),
		auditLogRepo:        repository.NewAuditLogRepository(),
		dockerManager:       dockerMgr,
		encryptor:           encryptor,
	}, nil
}

// Close 关闭服务
func (s *ComponentService) Close() error {
	return s.dockerManager.Close()
}

// builtInComponentCatalog 内置组件目录定义
var builtInComponentCatalog = []model.ComponentCatalogItem{
	{
		Type:           model.ComponentTypePostgreSQL,
		Name:           "PostgreSQL",
		Description:    "关系型数据库，用于持久化数据存储",
		DefaultImage:   "postgres:16-alpine",
		DefaultVersion: "16-alpine",
		DefaultPort:    "5432",
		EmbeddedImage:  "postgres-16-alpine.tar",
		DefaultConfig: model.ComponentConfig{
			"port": "5432",
			"env": map[string]interface{}{
				"POSTGRES_USER":     "admin",
				"POSTGRES_PASSWORD": "admin123",
				"POSTGRES_DB":       "myappdb",
				"TZ":                "Asia/Shanghai",
			},
			"volumes": []string{
				"postgres-data:/var/lib/postgresql/data",
			},
			"command": []string{
				"postgres",
				"-c", "max_connections=200",
				"-c", "shared_buffers=512MB",
				"-c", "effective_cache_size=1536MB",
				"-c", "work_mem=16MB",
				"-c", "maintenance_work_mem=128MB",
				"-c", "log_statement=all",
			},
		},
	},
	{
		Type:           model.ComponentTypeRedis,
		Name:           "Redis",
		Description:    "高性能键值缓存数据库",
		DefaultImage:   "redis:7-alpine",
		DefaultVersion: "7-alpine",
		DefaultPort:    "6379",
		EmbeddedImage:  "redis-7-alpine.tar",
		DefaultConfig: model.ComponentConfig{
			"port": "6379",
			"volumes": []string{
				"redis-data:/data",
			},
			"command": []string{
				"redis-server",
				"--appendonly", "yes",
				"--requirepass", "admin123",
				"--maxmemory", "1536mb",
				"--maxmemory-policy", "allkeys-lru",
				"--protected-mode", "yes",
				"--loglevel", "notice",
				"--save", "900 1",
				"--save", "300 10",
				"--save", "60 10000",
			},
		},
	},
}

// GetComponentCatalog 获取组件目录（含实时状态）
func (s *ComponentService) GetComponentCatalog(ctx context.Context) ([]model.ComponentCatalogItemWithStatus, error) {
	// 获取所有已安装组件
	installedComponents, err := s.componentRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list components: %w", err)
	}

	// 尝试连接Docker获取容器列表
	dockerAvailable := true
	var dockerErr string
	containers, err := s.dockerManager.ListContainers(ctx, true)
	if err != nil {
		dockerAvailable = false
		dockerErr = err.Error()
		containers = nil
	}

	result := make([]model.ComponentCatalogItemWithStatus, 0, len(builtInComponentCatalog))
	for _, item := range builtInComponentCatalog {
		statusItem := model.ComponentCatalogItemWithStatus{
			ComponentCatalogItem: item,
			Installed:            false,
			DockerAvailable:      dockerAvailable,
			DockerError:          dockerErr,
		}

		// 查找数据库中是否有该类型的组件记录
		var matched *model.Component
		for _, c := range installedComponents {
			if c.Type == item.Type {
				matched = &c
				break
			}
		}

		if matched != nil {
			statusItem.Installed = true
			statusItem.ComponentID = matched.ID
			statusItem.Status = matched.Status
			statusItem.ContainerID = matched.ContainerID

			if dockerAvailable && matched.ContainerID != "" {
				// 通过容器ID查找容器实时状态
				for _, cnt := range containers {
					if cnt.ID == matched.ContainerID || strings.HasPrefix(cnt.ID, matched.ContainerID) {
						statusItem.ContainerState = cnt.State
						if cnt.State == "running" {
							statusItem.Status = model.ComponentStatusRunning
						} else if cnt.State == "exited" {
							statusItem.Status = model.ComponentStatusStopped
						} else if cnt.State == "dead" {
							statusItem.Status = model.ComponentStatusError
						}
						break
					}
				}
			}
		} else if dockerAvailable {
			// 未在数据库中记录，尝试按镜像名匹配容器
			for _, cnt := range containers {
				for _, imgName := range cnt.Names {
					// 容器名格式为 /rtm-xxx
					nameLower := strings.ToLower(imgName)
					if strings.Contains(nameLower, fmt.Sprintf("rtm-%s", strings.ToLower(string(item.Type)))) ||
						strings.Contains(nameLower, strings.ToLower(item.Name)) {
						statusItem.ContainerState = cnt.State
						if cnt.State == "running" {
							statusItem.Status = model.ComponentStatusRunning
						} else {
							statusItem.Status = model.ComponentStatusStopped
						}
						break
					}
				}
				// 也按镜像匹配
				if statusItem.ContainerState == "" {
					imgLower := strings.ToLower(cnt.Image)
					if strings.Contains(imgLower, strings.ToLower(item.DefaultImage)) ||
						strings.Contains(imgLower, strings.ToLower(strings.Split(item.DefaultImage, ":")[0])) {
						statusItem.ContainerState = cnt.State
						if cnt.State == "running" {
							statusItem.Status = model.ComponentStatusRunning
						} else {
							statusItem.Status = model.ComponentStatusStopped
						}
					}
				}
			}
		}

		result = append(result, statusItem)
	}

	return result, nil
}

// InstallComponentRequest 安装组件请求
type InstallComponentRequest struct {
	Name            string                      `json:"name"`
	Type            model.ComponentType         `json:"type"`
	Image           string                      `json:"image"`
	Version         string                      `json:"version"`
	Config          model.ComponentConfig       `json:"config"`
	ImageSource     string                      `json:"image_source"`
	ImageFilePath   string                      `json:"-"`
	UserName        string                      `json:"-"`
}

// InstallComponent 安装组件
func (s *ComponentService) InstallComponent(ctx context.Context, req *InstallComponentRequest) (*model.Component, error) {
	// 检查名称是否已存在
	_, err := s.componentRepo.GetByName(req.Name)
	if err == nil {
		return nil, fmt.Errorf("component name already exists: %s", req.Name)
	}

	imageToUse := req.Image

	// 上传模式：加载本地镜像包
	if req.ImageSource == "upload" && req.ImageFilePath != "" {
		file, err := os.Open(req.ImageFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open image file: %w", err)
		}
		defer file.Close()
		defer os.Remove(req.ImageFilePath)

		loadedImage, err := s.dockerManager.LoadImage(ctx, file)
		if err != nil {
			return nil, fmt.Errorf("failed to load image: %w", err)
		}
		imageToUse = loadedImage
	}

	// 内置镜像模式：从嵌入资源加载镜像
	if req.ImageSource == "embedded" {
		catalogItem, err := s.getCatalogItem(req.Type)
		if err != nil {
			return nil, err
		}
		if catalogItem.EmbeddedImage == "" {
			return nil, fmt.Errorf("no embedded image defined for component type: %s", req.Type)
		}

		reader, err := assets.GetImageReader(catalogItem.EmbeddedImage)
		if err != nil {
			return nil, fmt.Errorf("embedded image not available: %w", err)
		}
		defer reader.Close()

		loadedImage, err := s.dockerManager.LoadImage(ctx, reader)
		if err != nil {
			return nil, fmt.Errorf("failed to load embedded image: %w", err)
		}
		imageToUse = loadedImage
		if imageToUse == "" {
			imageToUse = catalogItem.DefaultImage
		}
	}

	// 如果仍未确定镜像，对内置类型使用默认镜像
	if imageToUse == "" {
		if catalogItem, err := s.getCatalogItem(req.Type); err == nil {
			imageToUse = catalogItem.DefaultImage
		}
	}

	// 合并默认配置（确保内置组件带有必需的环境变量等默认值）
	if catalogItem, err := s.getCatalogItem(req.Type); err == nil {
		mergedConfig := make(model.ComponentConfig)
		for k, v := range catalogItem.DefaultConfig {
			mergedConfig[k] = v
		}
		for k, v := range req.Config {
			mergedConfig[k] = v
		}
		req.Config = mergedConfig
	}

	// 创建组件记录
	component := &model.Component{
		ID:          generateUUID(),
		Name:        req.Name,
		Type:        req.Type,
		Image:       imageToUse,
		Version:     req.Version,
		Status:      model.ComponentStatusInstalling,
		Config:      req.Config,
		ImageSource: req.ImageSource,
	}

	err = s.componentRepo.Create(component)
	if err != nil {
		return nil, fmt.Errorf("failed to create component: %w", err)
	}

	// 构建Docker配置
	dockerConfig, err := s.buildDockerConfig(req, imageToUse)
	if err != nil {
		s.componentRepo.UpdateStatus(component.ID, model.ComponentStatusError)
		return nil, fmt.Errorf("failed to build docker config: %w", err)
	}

	// 安装容器
	containerID, err := s.dockerManager.InstallComponent(ctx, dockerConfig)
	if err != nil {
		s.componentRepo.UpdateStatus(component.ID, model.ComponentStatusError)
		return nil, fmt.Errorf("failed to install container: %w", err)
	}

	// 更新组件信息
	component.ContainerID = containerID
	component.Status = model.ComponentStatusRunning
	err = s.componentRepo.Update(component)
	if err != nil {
		return nil, fmt.Errorf("failed to update component: %w", err)
	}

	// 记录审计日志
	s.createAuditLog(req.UserName, "install_component", "component", component.ID, fmt.Sprintf("Installed component: %s", req.Name))

	return component, nil
}

// UninstallComponent 卸载组件
func (s *ComponentService) UninstallComponent(ctx context.Context, id, userName string) error {
	component, err := s.componentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("component not found: %w", err)
	}

	// 卸载容器
	if component.ContainerID != "" {
		err = s.dockerManager.UninstallComponent(ctx, component.ContainerID)
		if err != nil {
			return fmt.Errorf("failed to uninstall container: %w", err)
		}
	}

	// 删除组件记录
	err = s.componentRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete component: %w", err)
	}

	// 记录审计日志
	s.createAuditLog(userName, "uninstall_component", "component", id, fmt.Sprintf("Uninstalled component: %s", component.Name))

	return nil
}

// StartComponent 启动组件
func (s *ComponentService) StartComponent(ctx context.Context, id, userName string) error {
	component, err := s.componentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("component not found: %w", err)
	}

	if component.ContainerID == "" {
		return fmt.Errorf("component has no container")
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	err = dockerClient.StartContainer(ctx, component.ContainerID)
	if err != nil {
		s.componentRepo.UpdateStatus(id, model.ComponentStatusError)
		return fmt.Errorf("failed to start container: %w", err)
	}

	s.componentRepo.UpdateStatus(id, model.ComponentStatusRunning)
	s.createAuditLog(userName, "start_component", "component", id, fmt.Sprintf("Started component: %s", component.Name))

	return nil
}

// StopComponent 停止组件
func (s *ComponentService) StopComponent(ctx context.Context, id, userName string) error {
	component, err := s.componentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("component not found: %w", err)
	}

	if component.ContainerID == "" {
		return fmt.Errorf("component has no container")
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	timeout := 30 * time.Second
	err = dockerClient.StopContainer(ctx, component.ContainerID, &timeout)
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	s.componentRepo.UpdateStatus(id, model.ComponentStatusStopped)
	s.createAuditLog(userName, "stop_component", "component", id, fmt.Sprintf("Stopped component: %s", component.Name))

	return nil
}

// RestartComponent 重启组件
func (s *ComponentService) RestartComponent(ctx context.Context, id, userName string) error {
	component, err := s.componentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("component not found: %w", err)
	}

	if component.ContainerID == "" {
		return fmt.Errorf("component has no container")
	}

	dockerClient, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	timeout := 30 * time.Second
	err = dockerClient.RestartContainer(ctx, component.ContainerID, &timeout)
	if err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}

	s.componentRepo.UpdateStatus(id, model.ComponentStatusRunning)
	s.createAuditLog(userName, "restart_component", "component", id, fmt.Sprintf("Restarted component: %s", component.Name))

	return nil
}

// GetComponent 获取组件（含实时 Docker 状态同步）
func (s *ComponentService) GetComponent(ctx context.Context, id string) (*model.Component, error) {
	component, err := s.componentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	s.syncComponentStatus(ctx, component)
	return component, nil
}

// ListComponents 获取组件列表（含实时 Docker 状态同步）
func (s *ComponentService) ListComponents(ctx context.Context) ([]model.Component, error) {
	components, err := s.componentRepo.List()
	if err != nil {
		return nil, err
	}
	for i := range components {
		s.syncComponentStatus(ctx, &components[i])
	}
	return components, nil
}

// GetComponentLogs 获取组件日志
func (s *ComponentService) GetComponentLogs(ctx context.Context, id, tail string) (string, error) {
	component, err := s.componentRepo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("component not found: %w", err)
	}

	if component.ContainerID == "" {
		return "", fmt.Errorf("component has no container")
	}

	return s.dockerManager.GetContainerLogs(ctx, component.ContainerID, tail)
}

// GetComponentStats 获取组件统计信息
func (s *ComponentService) GetComponentStats(ctx context.Context, id string) (*docker.ContainerStats, error) {
	component, err := s.componentRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("component not found: %w", err)
	}

	if component.ContainerID == "" {
		return nil, fmt.Errorf("component has no container")
	}

	return s.dockerManager.GetContainerStats(ctx, component.ContainerID)
}

// UpdateComponentConfig 更新组件配置
func (s *ComponentService) UpdateComponentConfig(id string, config model.ComponentConfig, userName string) error {
	component, err := s.componentRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("component not found: %w", err)
	}

	component.Config = config
	err = s.componentRepo.Update(component)
	if err != nil {
		return fmt.Errorf("failed to update component: %w", err)
	}

	s.createAuditLog(userName, "update_config", "component", id, fmt.Sprintf("Updated config for component: %s", component.Name))

	return nil
}

// syncComponentStatus 同步组件的实时 Docker 状态到数据库
func (s *ComponentService) syncComponentStatus(ctx context.Context, component *model.Component) {
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

// getCatalogItem 根据类型获取内置组件目录项
func (s *ComponentService) getCatalogItem(componentType model.ComponentType) (model.ComponentCatalogItem, error) {
	for _, item := range builtInComponentCatalog {
		if item.Type == componentType {
			return item, nil
		}
	}
	return model.ComponentCatalogItem{}, fmt.Errorf("unknown component type: %s", componentType)
}

// buildDockerConfig 构建Docker配置
func (s *ComponentService) buildDockerConfig(req *InstallComponentRequest, imageToUse string) (*docker.ContainerConfig, error) {
	config := &docker.ContainerConfig{
		Name:       fmt.Sprintf("rtm-%s", req.Name),
		Image:      imageToUse,
		Env:        []string{},
		Ports:      make(map[string]string),
		Volumes:    make(map[string]string),
		AutoRemove: false,
		SkipPull:   req.ImageSource == "upload" || req.ImageSource == "embedded",
	}

	// 根据配置设置环境变量、端口和卷
	if port, ok := req.Config["port"].(string); ok {
		config.Ports[port] = port
	}

	if envMap, ok := req.Config["env"].(map[string]interface{}); ok {
		for k, v := range envMap {
			config.Env = append(config.Env, fmt.Sprintf("%s=%v", k, v))
		}
	}

	if volumes, ok := req.Config["volumes"].([]interface{}); ok {
		for _, v := range volumes {
			if volStr, ok := v.(string); ok {
				parts := strings.SplitN(volStr, ":", 2)
				if len(parts) == 2 {
					config.Volumes[parts[0]] = parts[1]
				}
			}
		}
	}

	if cmdList, ok := req.Config["command"].([]interface{}); ok {
		for _, v := range cmdList {
			if cmdStr, ok := v.(string); ok {
				config.Command = append(config.Command, cmdStr)
			}
		}
	}

	return config, nil
}

// createAuditLog 创建审计日志
func (s *ComponentService) createAuditLog(username, action, resourceType, resourceID, details string) {
	log := &model.AuditLog{
		UserID:       username,
		Username:     username,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	}
	_ = s.auditLogRepo.Create(log)
}

func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
