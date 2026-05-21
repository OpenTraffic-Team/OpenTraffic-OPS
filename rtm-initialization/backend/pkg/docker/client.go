package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Client Docker客户端封装
type Client struct {
	client *client.Client
}

// NewClient 创建Docker客户端
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}
	return &Client{client: cli}, nil
}

// Close 关闭客户端
func (c *Client) Close() error {
	return c.client.Close()
}

// Ping 检查Docker连接
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx)
	return err
}

// ContainerConfig 容器配置
type ContainerConfig struct {
	Name       string
	Image      string
	Command    []string
	Env        []string
	Ports      map[string]string // 主机端口:容器端口
	Volumes    map[string]string // 主机路径:容器路径
	Networks   []string
	User       string
	AutoRemove bool
	SkipPull   bool
}

// CreateContainer 创建容器
func (c *Client) CreateContainer(ctx context.Context, config *ContainerConfig) (string, error) {
	// 配置容器
	containerConfig := &container.Config{
		Image: config.Image,
		Cmd:   config.Command,
		Env:   config.Env,
		User:  config.User,
	}

	// 端口映射
	portBindings := make(nat.PortMap)
	exposedPorts := make(nat.PortSet)
	for hostPort, containerPort := range config.Ports {
		portKey := nat.Port(fmt.Sprintf("%s/tcp", containerPort))
		portBindings[portKey] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
		exposedPorts[portKey] = struct{}{}
	}
	containerConfig.ExposedPorts = exposedPorts

	// 主机配置
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		AutoRemove:   config.AutoRemove,
	}

	// 卷挂载
	binds := make([]string, 0, len(config.Volumes))
	for hostPath, containerPath := range config.Volumes {
		binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}
	hostConfig.Binds = binds

	// 网络配置
	networkConfig := &network.NetworkingConfig{}
	if len(config.Networks) > 0 {
		endpointConfig := make(map[string]*network.EndpointSettings)
		for _, netName := range config.Networks {
			endpointConfig[netName] = &network.EndpointSettings{}
		}
		networkConfig.EndpointsConfig = endpointConfig
	}

	// 创建容器
	resp, err := c.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		config.Name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return resp.ID, nil
}

// StartContainer 启动容器
func (c *Client) StartContainer(ctx context.Context, containerID string) error {
	return c.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

// StopContainer 停止容器
func (c *Client) StopContainer(ctx context.Context, containerID string, timeout *time.Duration) error {
	stopTimeout := 0
	if timeout != nil {
		stopTimeout = int(timeout.Seconds())
	}
	return c.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &stopTimeout})
}

// RestartContainer 重启容器
func (c *Client) RestartContainer(ctx context.Context, containerID string, timeout *time.Duration) error {
	restartTimeout := 0
	if timeout != nil {
		restartTimeout = int(timeout.Seconds())
	}
	return c.client.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &restartTimeout})
}

// RemoveContainer 删除容器
func (c *Client) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return c.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: force,
	})
}

// GetContainer 获取容器信息
func (c *Client) GetContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return c.client.ContainerInspect(ctx, containerID)
}

// ListContainers 列出所有容器
func (c *Client) ListContainers(ctx context.Context, all bool) ([]types.Container, error) {
	return c.client.ContainerList(ctx, types.ContainerListOptions{All: all})
}

// GetContainerStats 获取容器统计信息
func (c *Client) GetContainerStats(ctx context.Context, containerID string) (types.StatsJSON, error) {
	stats, err := c.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return types.StatsJSON{}, err
	}
	defer stats.Body.Close()

	var statsJSON types.StatsJSON
	err = json.NewDecoder(stats.Body).Decode(&statsJSON)
	if err != nil {
		return types.StatsJSON{}, err
	}

	return statsJSON, nil
}

// GetContainerLogs 获取容器日志
func (c *Client) GetContainerLogs(ctx context.Context, containerID string, tail string) (io.ReadCloser, error) {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	}
	return c.client.ContainerLogs(ctx, containerID, options)
}

// PullImage 拉取镜像
func (c *Client) PullImage(ctx context.Context, image string) (io.ReadCloser, error) {
	return c.client.ImagePull(ctx, image, types.ImagePullOptions{})
}

// ListImages 列出镜像
func (c *Client) ListImages(ctx context.Context) ([]types.ImageSummary, error) {
	return c.client.ImageList(ctx, types.ImageListOptions{})
}

// LoadImage 从 tar 包加载镜像
func (c *Client) LoadImage(ctx context.Context, reader io.Reader) (string, error) {
	resp, err := c.client.ImageLoad(ctx, reader, false)
	if err != nil {
		return "", fmt.Errorf("failed to load image: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应 JSON stream，解析 Loaded image 字段
	var loadedImage string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		var result struct {
			Stream string `json:"stream"`
		}
		if err := json.Unmarshal([]byte(line), &result); err == nil && result.Stream != "" {
			// Docker 返回格式通常为 "Loaded image: xxx\n"
			if strings.HasPrefix(result.Stream, "Loaded image:") {
				loadedImage = strings.TrimSpace(strings.TrimPrefix(result.Stream, "Loaded image:"))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read load response: %w", err)
	}

	if loadedImage == "" {
		return "", fmt.Errorf("could not determine loaded image name from response")
	}

	return loadedImage, nil
}

// ExecCreate 在容器中执行命令
func (c *Client) ExecCreate(ctx context.Context, containerID string, cmd []string) (string, error) {
	config := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}
	resp, err := c.client.ContainerExecCreate(ctx, containerID, config)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

// ExecStart 执行命令
func (c *Client) ExecStart(ctx context.Context, execID string) (io.Reader, error) {
	config := types.ExecStartCheck{}
	hijackedResp, err := c.client.ContainerExecAttach(ctx, execID, config)
	if err != nil {
		return nil, err
	}
	return hijackedResp.Reader, nil
}
