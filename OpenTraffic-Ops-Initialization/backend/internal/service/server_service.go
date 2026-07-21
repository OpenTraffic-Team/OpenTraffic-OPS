package service

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"opentraffic-ops-init-backend/internal/model"
	"opentraffic-ops-init-backend/internal/repository"
	"opentraffic-ops-init-backend/pkg/assets"
	"opentraffic-ops-init-backend/pkg/crypto"
	"opentraffic-ops-init-backend/pkg/ssh"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ServerService 服务器服务
type ServerService struct {
	serverRepo *repository.ServerRepository
	encryptor  *crypto.Encryptor
}

// NewServerService 创建服务器服务
func NewServerService(encryptor *crypto.Encryptor) *ServerService {
	return &ServerService{
		serverRepo: repository.NewServerRepository(),
		encryptor:  encryptor,
	}
}

// CreateServerRequest 创建服务器请求
type CreateServerRequest struct {
	Name        string `json:"name" binding:"required"`
	Host        string `json:"host" binding:"required"`
	Port        int    `json:"port" binding:"required,min=1,max=65535"`
	Username    string `json:"username" binding:"required"`
	AuthType    string `json:"auth_type" binding:"required,oneof=password key"`
	Password    string `json:"password"`
	PrivateKey  string `json:"private_key"`
	Passphrase  string `json:"passphrase"`
	DeployPath  string `json:"deploy_path" binding:"required"`
	Description string `json:"description"`
}

// CreateServer 创建服务器
func (s *ServerService) CreateServer(req *CreateServerRequest) (*model.Server, error) {
	// 检查名称是否已存在
	_, err := s.serverRepo.GetByName(req.Name)
	if err == nil {
		return nil, fmt.Errorf("server name already exists: %s", req.Name)
	}

	// 加密敏感信息
	encryptedPassword, _ := s.encryptor.Encrypt(req.Password)
	encryptedPrivateKey, _ := s.encryptor.Encrypt(req.PrivateKey)
	encryptedPassphrase, _ := s.encryptor.Encrypt(req.Passphrase)

	server := &model.Server{
		ID:          generateUUID(),
		Name:        req.Name,
		Host:        req.Host,
		Port:        req.Port,
		Username:    req.Username,
		AuthType:    req.AuthType,
		Password:    encryptedPassword,
		PrivateKey:  encryptedPrivateKey,
		Passphrase:  encryptedPassphrase,
		DeployPath:  req.DeployPath,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.serverRepo.Create(server); err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	return server, nil
}

// UpdateServerRequest 更新服务器请求
type UpdateServerRequest struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	AuthType    string `json:"auth_type" binding:"omitempty,oneof=password key"`
	Password    string `json:"password"`
	PrivateKey  string `json:"private_key"`
	Passphrase  string `json:"passphrase"`
	DeployPath  string `json:"deploy_path"`
	Description string `json:"description"`
}

// UpdateServer 更新服务器
func (s *ServerService) UpdateServer(id string, req *UpdateServerRequest) (*model.Server, error) {
	server, err := s.serverRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	// 如果名称变更，检查新名称是否已存在
	if req.Name != "" && req.Name != server.Name {
		_, err := s.serverRepo.GetByName(req.Name)
		if err == nil {
			return nil, fmt.Errorf("server name already exists: %s", req.Name)
		}
		server.Name = req.Name
	}

	if req.Host != "" {
		server.Host = req.Host
	}
	if req.Port > 0 {
		server.Port = req.Port
	}
	if req.Username != "" {
		server.Username = req.Username
	}
	if req.AuthType != "" {
		server.AuthType = req.AuthType
	}
	if req.Password != "" {
		encryptedPassword, _ := s.encryptor.Encrypt(req.Password)
		server.Password = encryptedPassword
	}
	if req.PrivateKey != "" {
		encryptedPrivateKey, _ := s.encryptor.Encrypt(req.PrivateKey)
		server.PrivateKey = encryptedPrivateKey
	}
	if req.Passphrase != "" {
		encryptedPassphrase, _ := s.encryptor.Encrypt(req.Passphrase)
		server.Passphrase = encryptedPassphrase
	}
	if req.DeployPath != "" {
		server.DeployPath = req.DeployPath
	}
	server.Description = req.Description
	server.UpdatedAt = time.Now()

	if err := s.serverRepo.Update(server); err != nil {
		return nil, fmt.Errorf("failed to update server: %w", err)
	}

	return server, nil
}

// DeleteServer 删除服务器
func (s *ServerService) DeleteServer(id string) error {
	if err := s.serverRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}
	return nil
}

// GetServer 获取服务器
func (s *ServerService) GetServer(id string) (*model.Server, error) {
	return s.serverRepo.GetByID(id)
}

// ListServers 获取服务器列表
func (s *ServerService) ListServers() ([]model.Server, error) {
	return s.serverRepo.List()
}

// DecryptedCredentials 解密后的凭据
type DecryptedCredentials struct {
	Server     *model.Server
	Password   string
	PrivateKey string
	Passphrase string
}

// GetDecryptedCredentials 获取解密后的凭据
func (s *ServerService) GetDecryptedCredentials(id string) (*DecryptedCredentials, error) {
	server, err := s.serverRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("server not found: %w", err)
	}

	password, _ := s.encryptor.Decrypt(server.Password)
	privateKey, _ := s.encryptor.Decrypt(server.PrivateKey)
	passphrase, _ := s.encryptor.Decrypt(server.Passphrase)

	return &DecryptedCredentials{
		Server:     server,
		Password:   password,
		PrivateKey: privateKey,
		Passphrase: passphrase,
	}, nil
}

// TestConnection 测试SSH连接
func (s *ServerService) TestConnection(id string) error {
	creds, err := s.GetDecryptedCredentials(id)
	if err != nil {
		return err
	}

	client, err := ssh.NewClient(&ssh.Config{
		Host:       creds.Server.Host,
		Port:       creds.Server.Port,
		Username:   creds.Server.Username,
		AuthType:   creds.Server.AuthType,
		Password:   creds.Password,
		PrivateKey: creds.PrivateKey,
		Passphrase: creds.Passphrase,
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	return client.TestConnection()
}

// BuildSSHConfig 根据服务器ID构建SSH配置
func (s *ServerService) BuildSSHConfig(id string) (*ssh.Config, *model.Server, error) {
	creds, err := s.GetDecryptedCredentials(id)
	if err != nil {
		return nil, nil, err
	}

	config := &ssh.Config{
		Host:       creds.Server.Host,
		Port:       creds.Server.Port,
		Username:   creds.Server.Username,
		AuthType:   creds.Server.AuthType,
		Password:   creds.Password,
		PrivateKey: creds.PrivateKey,
		Passphrase: creds.Passphrase,
	}

	return config, creds.Server, nil
}

// softwareConfigMeta 软件配置元数据
var softwareConfigMeta = map[string]struct {
	ConfigDir    string
	ConfigFile   string
	EmbeddedName string
}{
	"opentraffic-ops-proxy": {
		ConfigDir:    "~/.opentraffic-ops-proxy",
		ConfigFile:   "config.json",
		EmbeddedName: "config.json",
	},
	"opentraffic-ops": {
		ConfigDir:    "~/.opentraffic-ops",
		ConfigFile:   "config.yaml",
		EmbeddedName: "config.yaml",
	},
	"opentraffic-control": {
		ConfigDir:    "",
		ConfigFile:   "mq_config.json",
		EmbeddedName: "mq_config.json",
	},
}

// controlServiceConfig opentraffic-control 服务配置
var controlServiceConfig = struct {
	DirName     string
	StartScript string
	ProcessPattern string
	PidFileName string
}{
	DirName:        "opentraffic-control",
	StartScript:    "start_algo.sh",
	ProcessPattern: "run_algorithms.py",
	PidFileName:    "opentraffic-control.pid",
}

// isControlService 判断是否为 opentraffic-control 服务
func isControlService(name string) bool {
	return name == controlServiceConfig.DirName || name == "opentraffic-control-linux-amd64"
}

// getDefaultConfig 从嵌入资源读取指定软件的默认配置，如不存在则返回空JSON
func getDefaultConfig(softwareName string) string {
	meta, ok := softwareConfigMeta[softwareName]
	if !ok {
		return "{}"
	}
	if assets.HasConfig(meta.EmbeddedName) {
		data, err := assets.ReadEmbeddedFile(meta.EmbeddedName)
		if err == nil {
			return string(data)
		}
	}
	return "{}"
}

// getDefaultProxyConfig 从嵌入资源读取默认 proxy 配置
func getDefaultProxyConfig() string {
	return getDefaultConfig("opentraffic-ops-proxy")
}

// GetDefaultSoftwareConfig 获取指定软件的默认配置（嵌入资源）
func (s *ServerService) GetDefaultSoftwareConfig(softwareName string) (string, error) {
	_, ok := softwareConfigMeta[softwareName]
	if !ok {
		return "", fmt.Errorf("unknown software: %s", softwareName)
	}
	return getDefaultConfig(softwareName), nil
}

// GetSoftwareConfig 获取远程服务器上指定软件的配置
func (s *ServerService) GetSoftwareConfig(id string, softwareName string) (string, error) {
	meta, ok := softwareConfigMeta[softwareName]
	if !ok {
		return "", fmt.Errorf("unknown software: %s", softwareName)
	}

	creds, err := s.GetDecryptedCredentials(id)
	if err != nil {
		return "", err
	}

	client, err := ssh.NewClient(&ssh.Config{
		Host:       creds.Server.Host,
		Port:       creds.Server.Port,
		Username:   creds.Server.Username,
		AuthType:   creds.Server.AuthType,
		Password:   creds.Password,
		PrivateKey: creds.PrivateKey,
		Passphrase: creds.Passphrase,
	})
	if err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	if isControlService(softwareName) {
		configPath := filepath.Join(creds.Server.DeployPath, controlServiceConfig.DirName, "config", meta.ConfigFile)
		data, err := client.ReadFile(configPath)
		if err != nil {
			return getDefaultConfig(softwareName), nil
		}
		return string(data), nil
	}

	configDirCmd := fmt.Sprintf("eval echo %s", meta.ConfigDir)
	configDir, _ := client.Execute(configDirCmd)
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		fallbackDir := strings.TrimPrefix(meta.ConfigDir, "~/")
		configDir = fmt.Sprintf("/home/%s/%s", creds.Server.Username, fallbackDir)
	}
	configPath := filepath.Join(configDir, meta.ConfigFile)

	data, err := client.ReadFile(configPath)
	if err != nil {
		return getDefaultConfig(softwareName), nil
	}
	return string(data), nil
}

// UpdateSoftwareConfig 更新远程服务器上指定软件的配置
func (s *ServerService) UpdateSoftwareConfig(id string, softwareName string, content string) error {
	meta, ok := softwareConfigMeta[softwareName]
	if !ok {
		return fmt.Errorf("unknown software: %s", softwareName)
	}

	// 根据配置文件扩展名校验格式
	ext := strings.ToLower(filepath.Ext(meta.ConfigFile))
	if ext == ".yaml" || ext == ".yml" {
		var yml interface{}
		if err := yaml.Unmarshal([]byte(content), &yml); err != nil {
			return fmt.Errorf("invalid yaml format: %w", err)
		}
	} else {
		var js map[string]interface{}
		if err := json.Unmarshal([]byte(content), &js); err != nil {
			return fmt.Errorf("invalid json format: %w", err)
		}
	}

	creds, err := s.GetDecryptedCredentials(id)
	if err != nil {
		return err
	}

	client, err := ssh.NewClient(&ssh.Config{
		Host:       creds.Server.Host,
		Port:       creds.Server.Port,
		Username:   creds.Server.Username,
		AuthType:   creds.Server.AuthType,
		Password:   creds.Password,
		PrivateKey: creds.PrivateKey,
		Passphrase: creds.Passphrase,
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	if isControlService(softwareName) {
		configDir := filepath.Join(creds.Server.DeployPath, controlServiceConfig.DirName, "config")
		configPath := filepath.Join(configDir, meta.ConfigFile)
		mkdirCmd := fmt.Sprintf("mkdir -p %s", configDir)
		if _, err := client.Execute(mkdirCmd); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if err := client.WriteFile([]byte(content), configPath); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
		return nil
	}

	configDirCmd := fmt.Sprintf("eval echo %s", meta.ConfigDir)
	configDir, _ := client.Execute(configDirCmd)
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		fallbackDir := strings.TrimPrefix(meta.ConfigDir, "~/")
		configDir = fmt.Sprintf("/home/%s/%s", creds.Server.Username, fallbackDir)
	}
	configPath := filepath.Join(configDir, meta.ConfigFile)

	if err := client.WriteFile([]byte(content), configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// UpdateProxyConfigRequest 更新 proxy 配置请求
type UpdateProxyConfigRequest struct {
	Content string `json:"content" binding:"required"`
}

// GetProxyConfig 获取远程服务器上的 proxy 配置
func (s *ServerService) GetProxyConfig(id string) (string, error) {
	return s.GetSoftwareConfig(id, "opentraffic-ops-proxy")
}

// UpdateProxyConfig 更新远程服务器上的 proxy 配置
func (s *ServerService) UpdateProxyConfig(id string, content string) error {
	return s.UpdateSoftwareConfig(id, "opentraffic-ops-proxy", content)
}

// archToBinarySuffix 将 uname -m 输出映射到二进制文件名后缀
var archToBinarySuffix = map[string]string{
	"x86_64":      "linux-amd64",
	"amd64":       "linux-amd64",
	"aarch64":     "linux-arm64",
	"arm64":       "linux-arm64",
	"loongarch64": "linux-loong64",
}

// getBinaryFileName 根据软件名和远程架构生成二进制文件名
func getBinaryFileName(binaryName string, arch string) (string, error) {
	suffix, ok := archToBinarySuffix[strings.ToLower(strings.TrimSpace(arch))]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}
	return fmt.Sprintf("%s-%s", binaryName, suffix), nil
}

// detectRemoteArch 通过 SSH 执行 uname -m 获取远程服务器架构
func detectRemoteArch(client *ssh.Client) (string, error) {
	output, err := client.Execute("uname -m")
	if err != nil {
		return "", fmt.Errorf("failed to detect remote architecture: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// isValidSoftwareName 校验软件名是否受支持（服务状态管理）
func isValidSoftwareName(name string) bool {
	_, ok := softwareConfigMeta[name]
	return ok || isControlService(name)
}

// pidFilePath 生成pid文件远程路径
func pidFilePath(deployPath string, softwareName string) string {
	return filepath.Join(deployPath, softwareName+".pid")
}

// GetServiceStatus 获取指定软件的运行状态（通过pid文件检测）
func (s *ServerService) GetServiceStatus(id string, softwareName string) (string, error) {
	if !isValidSoftwareName(softwareName) {
		return "unknown", fmt.Errorf("unknown software: %s", softwareName)
	}

	creds, err := s.GetDecryptedCredentials(id)
	if err != nil {
		return "unknown", err
	}

	client, err := ssh.NewClient(&ssh.Config{
		Host:       creds.Server.Host,
		Port:       creds.Server.Port,
		Username:   creds.Server.Username,
		AuthType:   creds.Server.AuthType,
		Password:   creds.Password,
		PrivateKey: creds.PrivateKey,
		Passphrase: creds.Passphrase,
	})
	if err != nil {
		return "unknown", fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	if isControlService(softwareName) {
		return s.getControlServiceStatus(client, creds.Server.DeployPath)
	}

	pidFile := pidFilePath(creds.Server.DeployPath, softwareName)
	checkCmd := fmt.Sprintf("if [ -f %s ] && kill -0 $(cat %s) 2>/dev/null; then echo running; else echo stopped; fi", pidFile, pidFile)
	output, err := client.Execute(checkCmd)
	if err != nil {
		return "unknown", nil
	}
	status := strings.TrimSpace(output)
	if status == "running" {
		return "running", nil
	}
	return "stopped", nil
}

// getControlServiceStatus 获取 opentraffic-control 运行状态
func (s *ServerService) getControlServiceStatus(client *ssh.Client, deployPath string) (string, error) {
	pidFile := filepath.Join(deployPath, controlServiceConfig.DirName, controlServiceConfig.PidFileName)
	checkCmd := fmt.Sprintf(
		"if [ -f %s ] && kill -0 $(cat %s) 2>/dev/null; then echo running; elif pgrep -f '[r]un_algorithms.py' >/dev/null 2>&1; then echo running; else echo stopped; fi",
		pidFile, pidFile,
	)
	output, err := client.Execute(checkCmd)
	if err != nil {
		return "unknown", nil
	}
	status := strings.TrimSpace(output)
	if status == "running" {
		return "running", nil
	}
	return "stopped", nil
}

// StartService 启动指定软件（使用pid文件管理）
func (s *ServerService) StartService(id string, softwareName string) error {
	if !isValidSoftwareName(softwareName) {
		return fmt.Errorf("unknown software: %s", softwareName)
	}

	creds, err := s.GetDecryptedCredentials(id)
	if err != nil {
		return err
	}

	client, err := ssh.NewClient(&ssh.Config{
		Host:       creds.Server.Host,
		Port:       creds.Server.Port,
		Username:   creds.Server.Username,
		AuthType:   creds.Server.AuthType,
		Password:   creds.Password,
		PrivateKey: creds.PrivateKey,
		Passphrase: creds.Passphrase,
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	if isControlService(softwareName) {
		return s.startControlService(client, creds.Server.DeployPath)
	}

	arch, err := detectRemoteArch(client)
	if err != nil {
		return err
	}
	binaryName, err := getBinaryFileName(softwareName, arch)
	if err != nil {
		return err
	}

	remotePath := filepath.Join(creds.Server.DeployPath, binaryName)
	pidFile := pidFilePath(creds.Server.DeployPath, softwareName)
	// setsid 创建新 session，使后台进程脱离 SSH session 追踪；
	// setsid 启动 sh，sh 直接后台运行 binary 并输出真实 PID，确保 pidfile 记录的是 binary 的 PID
	startCmd := fmt.Sprintf("cd %s && setsid sh -c '%s > /dev/null 2>&1 </dev/null & echo \"$!\"' > %s", creds.Server.DeployPath, remotePath, pidFile)
	if _, err := client.ExecuteWithTimeout(startCmd, 60*time.Second); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}

// startControlService 启动 opentraffic-control
func (s *ServerService) startControlService(client *ssh.Client, deployPath string) error {
	deployDir := filepath.Join(deployPath, controlServiceConfig.DirName)
	pidFile := filepath.Join(deployDir, controlServiceConfig.PidFileName)

	// 先停止可能已存在的进程，避免重复启动；使用 [r] 模式避免 pkill 匹配自身
	_, _ = client.Execute("pkill -f '[r]un_algorithms.py' 2>/dev/null || true")

	arch, _ := detectRemoteArch(client)

	// 启动脚本内部已后台运行；执行后等待进程拉起并写入 pid 文件
	startCmd := fmt.Sprintf("cd %s && ./%s && sleep 2 && pgrep -f '[r]un_algorithms.py' > %s",
		deployDir, controlServiceConfig.StartScript, pidFile)
	if _, err := client.ExecuteWithTimeout(startCmd, 120*time.Second); err != nil {
		// 附带 run.log 尾部，便于定位启动失败原因
		if controlEnvPackage(arch) != "" {
			venvPython := filepath.Join(deployDir, "trafficlight_env", "bin", "python3")
			checkOut, _ := client.Execute(fmt.Sprintf("test -f %s && echo exists || echo missing", venvPython))
			if strings.TrimSpace(checkOut) != "exists" {
				return fmt.Errorf("failed to start control service: Python 环境缺失（%s 不存在），请重新部署 opentraffic-control 以自动安装 trafficlight_env", venvPython)
			}
		}
		logTail, _ := client.Execute(fmt.Sprintf("tail -20 %s 2>/dev/null", filepath.Join(deployDir, "run.log")))
		if strings.TrimSpace(logTail) != "" {
			return fmt.Errorf("failed to start control service: %w, run.log:\n%s", err, strings.TrimSpace(logTail))
		}
		return fmt.Errorf("failed to start control service: %w", err)
	}
	return nil
}

// StopService 停止指定软件（通过pid文件停止）
func (s *ServerService) StopService(id string, softwareName string) error {
	if !isValidSoftwareName(softwareName) {
		return fmt.Errorf("unknown software: %s", softwareName)
	}

	creds, err := s.GetDecryptedCredentials(id)
	if err != nil {
		return err
	}

	client, err := ssh.NewClient(&ssh.Config{
		Host:       creds.Server.Host,
		Port:       creds.Server.Port,
		Username:   creds.Server.Username,
		AuthType:   creds.Server.AuthType,
		Password:   creds.Password,
		PrivateKey: creds.PrivateKey,
		Passphrase: creds.Passphrase,
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	if isControlService(softwareName) {
		return s.stopControlService(client, creds.Server.DeployPath)
	}

	pidFile := pidFilePath(creds.Server.DeployPath, softwareName)
	stopCmd := fmt.Sprintf("if [ -f %s ]; then kill $(cat %s) 2>/dev/null; rm -f %s; fi", pidFile, pidFile, pidFile)
	if _, err := client.ExecuteWithTimeout(stopCmd, 60*time.Second); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	return nil
}

// stopControlService 停止 opentraffic-control
func (s *ServerService) stopControlService(client *ssh.Client, deployPath string) error {
	deployDir := filepath.Join(deployPath, controlServiceConfig.DirName)
	pidFile := filepath.Join(deployDir, controlServiceConfig.PidFileName)
	stopCmd := fmt.Sprintf(
		"if [ -f %s ]; then kill $(cat %s) 2>/dev/null; rm -f %s; fi; pkill -f '[r]un_algorithms.py' 2>/dev/null || true",
		pidFile, pidFile, pidFile,
	)
	if _, err := client.ExecuteWithTimeout(stopCmd, 60*time.Second); err != nil {
		return fmt.Errorf("failed to stop control service: %w", err)
	}
	return nil
}

// RestartService 重启指定软件
func (s *ServerService) RestartService(id string, softwareName string) error {
	if err := s.StopService(id, softwareName); err != nil {
		return fmt.Errorf("failed to stop service during restart: %w", err)
	}
	return s.StartService(id, softwareName)
}
