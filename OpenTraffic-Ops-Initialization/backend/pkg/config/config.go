package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	EncryptionKey string
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DataDir string
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string
	ExpireHours int
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载.env文件
	_ = godotenv.Load()

	// 获取数据目录
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		// 默认使用当前目录下的data文件夹
		exePath, err := os.Executable()
		if err != nil {
			dataDir = "./data"
		} else {
			dataDir = filepath.Join(filepath.Dir(exePath), "data")
		}
	}

	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			DataDir: dataDir,
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "rtm-init-secret-key-change-in-production"),
			ExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 24),
		},
		EncryptionKey: getEnv("ENCRYPTION_KEY", "rtm-init-encryption-key-32-bytes-long"),
	}

	return cfg, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
