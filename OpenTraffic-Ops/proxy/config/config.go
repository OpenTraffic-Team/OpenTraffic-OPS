package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config Proxy 配置
type Config struct {
	// 平台服务端地址
	PlatformURL string `json:"platformUrl"`

	// 本机 IP（留空则自动检测）
	IP string `json:"ip"`

	// 本机主机名（留空则使用系统主机名）
	HostName string `json:"hostName"`

	// Proxy 版本号
	Version string `json:"version"`

	// 心跳上报间隔（秒），默认 3 秒
	HeartbeatInterval int `json:"heartbeatInterval"`

	// 指令轮询间隔（秒）
	PollInterval int `json:"pollInterval"`

	// 日志级别: debug/info/warn/error
	LogLevel string `json:"logLevel"`

	// 日志文件路径（留空则输出到控制台）
	LogFile string `json:"logFile"`

	// 进程列表配置: 需要监控的进程名列表
	Processes []ProcessConfig `json:"processes"`

	// 远程控制开关（默认启用）
	EnableRemote bool `json:"enableRemote"`

	// WebSocket端点（留空则使用 platformURL 自动推导）
	WSEndpoint string `json:"wsEndpoint"`
}

// ProcessConfig 单个进程监控配置
type ProcessConfig struct {
	Name    string `json:"name"`    // 进程标识名
	Pattern string `json:"pattern"` // 进程匹配模式（用于查找进程）
	ExecCmd string `json:"execCmd"` // 启动命令（用于 startProcess 指令）
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		PlatformURL:       "http://127.0.0.1:8080",
		Version:           "1.0.0",
		HeartbeatInterval: 3,
		PollInterval:      10,
		LogLevel:          "info",
		Processes:         []ProcessConfig{},
		EnableRemote:      true,
	}
}

// Load 从文件加载配置，若文件不存在则创建默认配置
func Load(path string) (*Config, error) {
	if path == "" {
		path = defaultConfigPath()
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := Default()
		if err := cfg.Save(path); err != nil {
			return nil, fmt.Errorf("创建默认配置文件失败: %w", err)
		}
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	return cfg, nil
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	if path == "" {
		path = defaultConfigPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func defaultConfigPath() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "."
	}
	return filepath.Join(home, ".opentraffic-ops-proxy", "config.json")
}
