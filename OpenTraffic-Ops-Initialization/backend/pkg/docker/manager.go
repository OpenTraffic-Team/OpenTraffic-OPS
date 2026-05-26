package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
)

// ContainerInfo 容器信息
type ContainerInfo struct {
	ID         string
	Name       string
	Image      string
	Status     string
	State      string
	Running    bool
	Paused     bool
	Restarting bool
	OOMKilled  bool
	Dead       bool
	Pid        int
	Created    time.Time
}

// ContainerStats 容器统计信息
type ContainerStats struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	MemoryLimit float64 `json:"memory_limit"`
	NetworkRx   uint64  `json:"network_rx"`
	NetworkTx   uint64  `json:"network_tx"`
	BlockRead   uint64  `json:"block_read"`
	BlockWrite  uint64  `json:"block_write"`
}

// Manager 容器管理器
type Manager struct {
	client *Client
}

// NewManager 创建容器管理器
func NewManager() (*Manager, error) {
	cli, err := NewClient()
	if err != nil {
		return nil, err
	}
	return &Manager{client: cli}, nil
}

// Close 关闭管理器
func (m *Manager) Close() error {
	return m.client.Close()
}

// Ping 检查Docker连接
func (m *Manager) Ping(ctx context.Context) error {
	return m.client.Ping(ctx)
}

// ListContainers 列出所有容器
func (m *Manager) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	return m.client.ListContainers(ctx, all)
}

// LoadImage 加载镜像
func (m *Manager) LoadImage(ctx context.Context, reader io.Reader) (string, error) {
	return m.client.LoadImage(ctx, reader)
}

// InstallComponent 安装组件
func (m *Manager) InstallComponent(ctx context.Context, config *ContainerConfig) (string, error) {
	if !config.SkipPull {
		// 检查镜像是否存在，不存在则拉取
		err := m.ensureImage(ctx, config.Image)
		if err != nil {
			return "", fmt.Errorf("failed to ensure image: %w", err)
		}
	}

	// 创建容器
	containerID, err := m.client.CreateContainer(ctx, config)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// 启动容器
	err = m.client.StartContainer(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return containerID, nil
}

// UninstallComponent 卸载组件
func (m *Manager) UninstallComponent(ctx context.Context, containerID string) error {
	// 先停止容器
	timeout := 30 * time.Second
	err := m.client.StopContainer(ctx, containerID, &timeout)
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	// 删除容器
	err = m.client.RemoveContainer(ctx, containerID, true)
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

// GetContainerInfo 获取容器信息
func (m *Manager) GetContainerInfo(ctx context.Context, containerID string) (*ContainerInfo, error) {
	container, err := m.client.GetContainer(ctx, containerID)
	if err != nil {
		return nil, err
	}

	// 解析创建时间
	createdTime, err := time.Parse(time.RFC3339Nano, container.Created)
	if err != nil {
		createdTime = time.Now()
	}

	info := &ContainerInfo{
		ID:         container.ID,
		Name:       container.Name,
		Image:      container.Config.Image,
		Status:     container.State.Status,
		State:      getStateString(container.State),
		Running:    container.State.Running,
		Paused:     container.State.Paused,
		Restarting: container.State.Restarting,
		OOMKilled:  container.State.OOMKilled,
		Dead:       container.State.Dead,
		Pid:        container.State.Pid,
		Created:    createdTime,
	}

	return info, nil
}

// GetContainerStats 获取容器统计信息
func (m *Manager) GetContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	stats, err := m.client.GetContainerStats(ctx, containerID)
	if err != nil {
		return nil, err
	}

	// 计算CPU使用率
	var cpuUsage float64
	cpuDelta := stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage
	systemDelta := stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage
	cpuCount := len(stats.CPUStats.CPUUsage.PercpuUsage)
	if cpuCount == 0 {
		// fallback: use OnlineCPUs if available
		cpuCount = int(stats.CPUStats.OnlineCPUs)
	}
	if cpuCount == 0 {
		cpuCount = 1
	}
	if systemDelta > 0 {
		cpuUsage = float64(cpuDelta) / float64(systemDelta) * float64(cpuCount) * 100.0
	}

	// 计算内存使用
	var memoryUsage, memoryLimit float64
	memoryLimit = float64(stats.MemoryStats.Limit)
	if stats.MemoryStats.Usage > 0 {
		cache := stats.MemoryStats.Stats["cache"]
		if stats.MemoryStats.Usage >= cache {
			memoryUsage = float64(stats.MemoryStats.Usage - cache)
		} else {
			memoryUsage = float64(stats.MemoryStats.Usage)
		}
	}

	// 网络IO
	var networkRx, networkTx uint64
	for _, network := range stats.Networks {
		networkRx += network.RxBytes
		networkTx += network.TxBytes
	}

	// 磁盘IO
	var blockRead, blockWrite uint64
	for _, blkio := range stats.BlkioStats.IoServiceBytesRecursive {
		if blkio.Op == "Read" {
			blockRead += blkio.Value
		} else if blkio.Op == "Write" {
			blockWrite += blkio.Value
		}
	}

	return &ContainerStats{
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		MemoryLimit: memoryLimit,
		NetworkRx:   networkRx,
		NetworkTx:   networkTx,
		BlockRead:   blockRead,
		BlockWrite:  blockWrite,
	}, nil
}

// GetContainerLogs 获取容器日志
func (m *Manager) GetContainerLogs(ctx context.Context, containerID string, tail string) (string, error) {
	reader, err := m.client.GetContainerLogs(ctx, containerID, tail)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

// ensureImage 确保镜像存在
func (m *Manager) ensureImage(ctx context.Context, image string) error {
	// 检查镜像是否已存在
	images, err := m.client.ListImages(ctx)
	if err != nil {
		return err
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == image || tag == image+":latest" {
				return nil
			}
		}
	}

	// 拉取镜像
	reader, err := m.client.PullImage(ctx, image)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 等待拉取完成
	_, err = io.Copy(io.Discard, reader)
	return err
}

// getStateString 获取状态字符串
func getStateString(state *types.ContainerState) string {
	if state.Running {
		return "running"
	}
	if state.Paused {
		return "paused"
	}
	if state.Restarting {
		return "restarting"
	}
	if state.OOMKilled {
		return "oom_killed"
	}
	if state.Dead {
		return "dead"
	}
	return "stopped"
}
