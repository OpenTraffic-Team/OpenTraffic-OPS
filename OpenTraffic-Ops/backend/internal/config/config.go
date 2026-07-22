package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var (
	Version   = "dev"
	BuildTime = ""
	GoVersion = ""
)

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// Config 应用配置结构体
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	App        AppConfig        `mapstructure:"app"`
	Datasource DatasourceConfig `mapstructure:"datasource"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	User       UserConfig       `mapstructure:"user"`
	XSS        XSSConfig        `mapstructure:"xss"`
	Log        LogConfig        `mapstructure:"log"`
	Agent      AgentConfig      `mapstructure:"agent"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type AppConfig struct {
	Name           string            `mapstructure:"name"`
	Version        string            `mapstructure:"version"`
	CopyrightYear  string            `mapstructure:"copyrightYear"`
	DemoEnabled    bool              `mapstructure:"demoEnabled"`
	CaptchaType    string            `mapstructure:"captchaType"`
	AddressEnabled bool              `mapstructure:"addressEnabled"`
	Profile        string            `mapstructure:"profile"`
}

type DatasourceConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Database        string `mapstructure:"database"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"maxIdleConns"`
	MaxOpenConns    int    `mapstructure:"maxOpenConns"`
	ConnMaxLifetime int    `mapstructure:"connMaxLifetime"`
	LogLevel        string `mapstructure:"logLevel"`
}

func (d *DatasourceConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		d.Host, d.Port, d.Username, d.Password, d.Database)
}

type RedisInstanceConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"poolSize"`
	MinIdleConns int    `mapstructure:"minIdleConns"`
}

type RedisConfig struct {
	Platform RedisInstanceConfig `mapstructure:"platform"`
	Edge     RedisInstanceConfig `mapstructure:"edge"`
}

func (r *RedisInstanceConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Header     string `mapstructure:"header"`
	Secret     string `mapstructure:"secret"`
	ExpireTime int    `mapstructure:"expireTime"`
}

type UserConfig struct {
	Password PasswordConfig `mapstructure:"password"`
}

type PasswordConfig struct {
	MaxRetryCount int `mapstructure:"maxRetryCount"`
	LockTime      int `mapstructure:"lockTime"`
}

type XSSConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	Excludes    []string `mapstructure:"excludes"`
	URLPatterns []string `mapstructure:"urlPatterns"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"maxSize"`
	MaxBackups int    `mapstructure:"maxBackups"`
	MaxAge     int    `mapstructure:"maxAge"`
	Compress   bool   `mapstructure:"compress"`
}

type AgentConfig struct {
	Control  string `mapstructure:"control"`
	Perceive string `mapstructure:"perceive"`
}

// LoadConfig 加载配置文件
// 固定从 ~/.opentraffic-ops/opentraffic-ops-config.yaml 加载
func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	configFile := filepath.Join(home, ".opentraffic-ops", "opentraffic-ops-config.yaml")
	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	v.AutomaticEnv()
	v.SetEnvPrefix("RTM")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	GlobalConfig = &cfg
	return &cfg, nil
}
