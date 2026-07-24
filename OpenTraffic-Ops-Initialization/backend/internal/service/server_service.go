package service

import (
	"encoding/json"
	"fmt"
	"log"
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
		ConfigFile:   "opentraffic-ops-proxy-config.json",
		EmbeddedName: "opentraffic-ops-proxy-config.json",
	},
	"opentraffic-ops": {
		ConfigDir:    "~/.opentraffic-ops",
		ConfigFile:   "opentraffic-ops-config.yaml",
		EmbeddedName: "opentraffic-ops-config.yaml",
	},
	"opentraffic-control": {
		ConfigDir:    "",
		ConfigFile:   "mq_config.json",
		EmbeddedName: "opentraffic-control-config.json",
	},
	"opentraffic-perception": {
		ConfigDir:    "",
		ConfigFile:   "config.json",
		EmbeddedName: "opentraffic-perception-config.json",
	},
}

// controlServiceConfig opentraffic-control 服务配置
var controlServiceConfig = struct {
	DirName        string
	StartScript    string
	ProcessPattern string
	PidFileName    string
}{
	DirName:        "opentraffic-control",
	StartScript:    "start_algo.sh",
	ProcessPattern: "run_algorithms.py",
	PidFileName:    "opentraffic-control.pid",
}

// perceptionServiceConfig opentraffic-perception 服务配置
var perceptionServiceConfig = struct {
	DirName         string
	StartScript     string
	StopScript      string
	StatusScript    string
	InstallScript   string
	ConfigureScript string
	PidFileName     string
}{
	DirName:         "opentraffic-perception",
	StartScript:     "deploy/start.sh",
	StopScript:      "deploy/stop.sh",
	StatusScript:    "deploy/status.sh",
	InstallScript:   "deploy/install.sh",
	ConfigureScript: "deploy/configure.sh",
	PidFileName:     "run_perception.pid",
}

// resolvePerceptionWorkDir 根据远程实际目录判断 perception 工作目录（支持扁平解压或带架构子目录）
func resolvePerceptionWorkDir(client *ssh.Client, deployPath string) (string, error) {
	baseDir := filepath.ToSlash(filepath.Join(deployPath, perceptionServiceConfig.DirName))
	candidates := []string{
		filepath.ToSlash(filepath.Join(baseDir, "opentraffic-perception-linux-loong64")),
		filepath.ToSlash(filepath.Join(baseDir, "opentraffic-perception-linux-arm64")),
		filepath.ToSlash(filepath.Join(baseDir, "opentraffic-perception-linux-amd64")),
		baseDir,
	}
	for _, dir := range candidates {
		scriptPath := filepath.ToSlash(filepath.Join(dir, "deploy", "install.sh"))
		out, _ := client.Execute(fmt.Sprintf("test -f %s && echo exists || echo missing", scriptPath))
		if strings.TrimSpace(out) == "exists" {
			return dir, nil
		}
	}
	return "", fmt.Errorf("未找到 opentraffic-perception 工作目录，请确认已部署")
}

// resolvePerceptionProcessPattern 根据工作目录内容判断感知服务实际进程入口
func resolvePerceptionProcessPattern(client *ssh.Client, workDir string) string {
	loongsonEntry := filepath.ToSlash(filepath.Join(workDir, "loongson_deploy", "main_loongson.py"))
	out, _ := client.Execute(fmt.Sprintf("test -f %s && echo exists || echo missing", loongsonEntry))
	if strings.TrimSpace(out) == "exists" {
		return "main_loongson.py"
	}
	rknnEntry := filepath.ToSlash(filepath.Join(workDir, "run_local_rknn.sh"))
	out, _ = client.Execute(fmt.Sprintf("test -f %s && echo exists || echo missing", rknnEntry))
	if strings.TrimSpace(out) == "exists" {
		return "run_local_rknn.sh"
	}
	return "run_main.py"
}

// pkillPattern 生成 pkill/pgrep 使用的防自匹配正则
func pkillPattern(processPattern string) string {
	if processPattern == "" {
		return ""
	}
	return fmt.Sprintf("[%c]%s", processPattern[0], processPattern[1:])
}

// isControlService 判断是否为 opentraffic-control 服务
func isControlService(name string) bool {
	return name == controlServiceConfig.DirName || name == "opentraffic-control-linux-amd64"
}

// isPerceptionService 判断是否为 opentraffic-perception 服务
func isPerceptionService(name string) bool {
	return name == perceptionServiceConfig.DirName || name == "opentraffic-perception-linux-amd64"
}

// isTarPackageService 判断是否为 tar 包部署的算法服务（control / perception）
func isTarPackageService(name string) bool {
	return isControlService(name) || isPerceptionService(name)
}

// tarPackageWorkDir 返回 tar 包服务的实际工作目录（仅用于 control）
func tarPackageWorkDir(deployPath string, softwareName string) string {
	return filepath.ToSlash(filepath.Join(deployPath, controlServiceConfig.DirName))
}

// tarPackageConfigDir 返回 tar 包服务的配置文件所在目录（仅用于 control）
func tarPackageConfigDir(deployPath string, softwareName string) string {
	return filepath.ToSlash(filepath.Join(tarPackageWorkDir(deployPath, softwareName), "config"))
}

// tarPackageConfigPathForService 返回 tar 包服务的完整配置文件路径（control 固定，perception 动态检测架构子目录）
func tarPackageConfigPathForService(client *ssh.Client, deployPath string, softwareName string, configFile string) (string, error) {
	if isPerceptionService(softwareName) {
		workDir, err := resolvePerceptionWorkDir(client, deployPath)
		if err != nil {
			return "", err
		}
		return filepath.ToSlash(filepath.Join(workDir, "drivers", configFile)), nil
	}
	return filepath.ToSlash(filepath.Join(tarPackageConfigDir(deployPath, softwareName), configFile)), nil
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

	if isTarPackageService(softwareName) {
		configPath, err := tarPackageConfigPathForService(client, creds.Server.DeployPath, softwareName, meta.ConfigFile)
		if err != nil {
			return getDefaultConfig(softwareName), nil
		}
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
	configPath := filepath.ToSlash(filepath.Join(configDir, meta.ConfigFile))

	data, err := client.ReadFile(configPath)
	if err != nil {
		return getDefaultConfig(softwareName), nil
	}
	return string(data), nil
}

// UpdateSoftwareConfig 更新远程服务器上指定软件的配置
func (s *ServerService) UpdateSoftwareConfig(id string, softwareName string, content string) error {
	log.Printf("[DEBUG] UpdateSoftwareConfig called: serverID=%s software=%s contentLen=%d", id, softwareName, len(content))
	meta, ok := softwareConfigMeta[softwareName]
	if !ok {
		log.Printf("[DEBUG] UpdateSoftwareConfig unknown software: %s", softwareName)
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

	if isTarPackageService(softwareName) {
		log.Printf("[DEBUG] UpdateSoftwareConfig isTarPackageService=true")
		// perception 可能存在扁平目录和架构子目录两种结构，写入所有有效工作目录
		if isPerceptionService(softwareName) {
			baseDir := filepath.ToSlash(filepath.Join(creds.Server.DeployPath, perceptionServiceConfig.DirName))
			candidates := []string{
				filepath.ToSlash(filepath.Join(baseDir, "opentraffic-perception-linux-loong64")),
				filepath.ToSlash(filepath.Join(baseDir, "opentraffic-perception-linux-arm64")),
				filepath.ToSlash(filepath.Join(baseDir, "opentraffic-perception-linux-amd64")),
				baseDir,
			}
			var lastErr error
			written := false
			for _, dir := range candidates {
				installPath := filepath.ToSlash(filepath.Join(dir, "deploy", "install.sh"))
				out, _ := client.Execute(fmt.Sprintf("test -f %s && echo exists || echo missing", installPath))
				exists := strings.TrimSpace(out) == "exists"
				log.Printf("[DEBUG] checking candidate dir=%s installExists=%v", dir, exists)
				if !exists {
					continue
				}
				configPath := filepath.ToSlash(filepath.Join(dir, "drivers", meta.ConfigFile))
				configDir := filepath.ToSlash(filepath.Dir(configPath))
				mkdirCmd := fmt.Sprintf("mkdir -p %s", configDir)
				if _, err := client.Execute(mkdirCmd); err != nil {
					lastErr = fmt.Errorf("failed to create config directory %s: %w", configDir, err)
					log.Printf("[DEBUG] mkdir failed %s: %v", configDir, err)
					continue
				}
				if err := client.WriteFile([]byte(content), configPath); err != nil {
					lastErr = fmt.Errorf("failed to write config file %s: %w", configPath, err)
					log.Printf("[DEBUG] WriteFile failed %s: %v", configPath, err)
					continue
				}
				log.Printf("[DEBUG] WriteFile succeeded %s", configPath)
				written = true
			}
			if !written {
				if lastErr != nil {
					return lastErr
				}
				return fmt.Errorf("failed to locate perception config path")
			}
			return nil
		}

		configPath, err := tarPackageConfigPathForService(client, creds.Server.DeployPath, softwareName, meta.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to locate config path: %w", err)
		}
		configDir := filepath.ToSlash(filepath.Dir(configPath))
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
	configPath := filepath.ToSlash(filepath.Join(configDir, meta.ConfigFile))

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
	return ok || isTarPackageService(name)
}

// pidFilePath 生成pid文件远程路径
func pidFilePath(deployPath string, softwareName string) string {
	return filepath.ToSlash(filepath.Join(deployPath, softwareName+".pid"))
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

	if isTarPackageService(softwareName) {
		return s.getTarPackageServiceStatus(client, softwareName, creds.Server.DeployPath)
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

// getTarPackageServiceStatus 获取 tar 包算法服务（control / perception）运行状态
func (s *ServerService) getTarPackageServiceStatus(client *ssh.Client, softwareName string, deployPath string) (string, error) {
	var pidFile, statusScript, processPattern, workDir string
	if isPerceptionService(softwareName) {
		var err error
		workDir, err = resolvePerceptionWorkDir(client, deployPath)
		if err != nil {
			return "stopped", nil
		}
		pidFile = filepath.ToSlash(filepath.Join(workDir, perceptionServiceConfig.PidFileName))
		statusScript = filepath.ToSlash(filepath.Join(workDir, perceptionServiceConfig.StatusScript))
		processPattern = resolvePerceptionProcessPattern(client, workDir)
	} else {
		workDir = tarPackageWorkDir(deployPath, softwareName)
		pidFile = filepath.ToSlash(filepath.Join(workDir, controlServiceConfig.PidFileName))
		processPattern = controlServiceConfig.ProcessPattern
	}

	// 优先使用服务自带的 status.sh，不存在时回退到 pid 文件 + pgrep
	var checkCmd string
	if statusScript != "" {
		checkCmd = fmt.Sprintf(
			"if [ -x %s ]; then cd %s && bash %s; elif [ -f %s ] && kill -0 $(cat %s) 2>/dev/null; then echo running; elif pgrep -f '%s' >/dev/null 2>&1; then echo running; else echo stopped; fi",
			statusScript, workDir, statusScript, pidFile, pidFile, pkillPattern(processPattern),
		)
	} else {
		checkCmd = fmt.Sprintf(
			"if [ -f %s ] && kill -0 $(cat %s) 2>/dev/null; then echo running; elif pgrep -f '%s' >/dev/null 2>&1; then echo running; else echo stopped; fi",
			pidFile, pidFile, pkillPattern(processPattern),
		)
	}
	output, err := client.Execute(checkCmd)
	if err != nil {
		return "unknown", nil
	}
	// status.sh 可能输出多行（如 "status: running\npid: ..."），取第一行非空内容判断
	firstLine := ""
	for _, line := range strings.Split(output, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			firstLine = strings.ToLower(trimmed)
			break
		}
	}
	if firstLine == "running" || strings.HasPrefix(firstLine, "status: running") {
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

	if isTarPackageService(softwareName) {
		return s.startTarPackageService(client, softwareName, creds.Server.DeployPath)
	}

	arch, err := detectRemoteArch(client)
	if err != nil {
		return err
	}
	binaryName, err := getBinaryFileName(softwareName, arch)
	if err != nil {
		return err
	}

	remotePath := filepath.ToSlash(filepath.Join(creds.Server.DeployPath, binaryName))
	pidFile := pidFilePath(creds.Server.DeployPath, softwareName)
	// setsid 创建新 session，使后台进程脱离 SSH session 追踪；
	// setsid 启动 sh，sh 直接后台运行 binary 并输出真实 PID，确保 pidfile 记录的是 binary 的 PID
	startCmd := fmt.Sprintf("cd %s && setsid sh -c '%s > /dev/null 2>&1 </dev/null & echo \"$!\"' > %s", creds.Server.DeployPath, remotePath, pidFile)
	if _, err := client.ExecuteWithTimeout(startCmd, 60*time.Second); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return nil
}

// startTarPackageService 启动 tar 包算法服务（control / perception）
func (s *ServerService) startTarPackageService(client *ssh.Client, softwareName string, deployPath string) error {
	if isPerceptionService(softwareName) {
		return s.startPerceptionService(client, deployPath)
	}
	return s.startControlService(client, deployPath)
}

// startControlService 启动 opentraffic-control
func (s *ServerService) startControlService(client *ssh.Client, deployPath string) error {
	deployDir := filepath.ToSlash(filepath.Join(deployPath, controlServiceConfig.DirName))
	pidFile := filepath.ToSlash(filepath.Join(deployDir, controlServiceConfig.PidFileName))

	// 启动前校验 mq_config.json，缺失或无效时阻止启动并给出引导
	configPath := filepath.ToSlash(filepath.Join(deployDir, "config", "mq_config.json"))
	configData, err := client.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("配置文件 %s 不存在，请先在配置管理中填写并保存 mq_config.json 后再启动", configPath)
	}
	var mqConfig map[string]interface{}
	if err := json.Unmarshal(configData, &mqConfig); err != nil {
		return fmt.Errorf("配置文件 mq_config.json 不是合法的 JSON: %v，请在配置管理中修正后再启动", err)
	}
	if addr, ok := mqConfig["redis_addr"].(string); !ok || strings.TrimSpace(addr) == "" {
		return fmt.Errorf("配置文件 mq_config.json 缺少 redis_addr，请在配置管理中填写 Redis 地址后再启动")
	}

	// 先停止可能已存在的进程，避免重复启动；使用 [r] 模式避免 pkill 匹配自身
	_, _ = client.Execute("pkill -f '[r]un_algorithms.py' 2>/dev/null || true")

	arch, _ := detectRemoteArch(client)

	// 使用 bash 显式执行启动脚本，避免脚本含 CRLF 时 shebang 解析失败
	startCmd := fmt.Sprintf("cd %s && bash %s && sleep 2 && pgrep -f '[r]un_algorithms.py' > %s",
		deployDir, controlServiceConfig.StartScript, pidFile)
	if _, err := client.ExecuteWithTimeout(startCmd, 120*time.Second); err != nil {
		// 附带 run.log 尾部，便于定位启动失败原因
		if controlEnvPackage(arch) != "" {
			venvPython := filepath.ToSlash(filepath.Join(deployDir, "trafficlight_env", "bin", "python3"))
			checkOut, _ := client.Execute(fmt.Sprintf("test -f %s && echo exists || echo missing", venvPython))
			if strings.TrimSpace(checkOut) != "exists" {
				return fmt.Errorf("failed to start control service: Python 环境缺失（%s 不存在），请重新部署 opentraffic-control 以自动安装 trafficlight_env", venvPython)
			}
		}
		logTail, _ := client.Execute(fmt.Sprintf("tail -20 %s 2>/dev/null", filepath.ToSlash(filepath.Join(deployDir, "run.log"))))
		if strings.TrimSpace(logTail) != "" {
			return fmt.Errorf("failed to start control service: %w, run.log:\n%s", err, strings.TrimSpace(logTail))
		}
		return fmt.Errorf("failed to start control service: %w", err)
	}
	return nil
}

// getPerceptionVideoPath 从感知配置中安全读取第一个摄像头的 rtsp_url
func getPerceptionVideoPath(config map[string]interface{}) (string, bool) {
	intersection, ok := config["intersection"].(map[string]interface{})
	if !ok {
		return "", false
	}
	cameras, ok := intersection["cameras"].([]interface{})
	if !ok || len(cameras) == 0 {
		return "", false
	}
	camera0, ok := cameras[0].(map[string]interface{})
	if !ok {
		return "", false
	}
	rtspURL, ok := camera0["rtsp_url"].(string)
	return rtspURL, ok
}

// startPerceptionService 启动 opentraffic-perception
func (s *ServerService) startPerceptionService(client *ssh.Client, deployPath string) error {
	workDir, err := resolvePerceptionWorkDir(client, deployPath)
	if err != nil {
		return err
	}
	processPattern := resolvePerceptionProcessPattern(client, workDir)

	// 启动前校验 drivers/config.json，缺失或无效时阻止启动并给出引导
	configPath := filepath.ToSlash(filepath.Join(workDir, "drivers", "config.json"))
	log.Printf("[DEBUG] startPerceptionService reading config from %s", configPath)
	configData, err := client.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("配置文件 %s 不存在，请先在配置管理中填写并保存 drivers/config.json 后再启动", configPath)
	}
	var perceptionConfig map[string]interface{}
	if err := json.Unmarshal(configData, &perceptionConfig); err != nil {
		return fmt.Errorf("配置文件 drivers/config.json 不是合法的 JSON: %v，请在配置管理中修正后再启动", err)
	}

	requiredFields := []struct {
		name   string
		getter func(map[string]interface{}) (string, bool)
	}{
		{"intersection.cameras[0].rtsp_url", getPerceptionVideoPath},
		{"radarReferenceJsonl", func(c map[string]interface{}) (string, bool) { v, ok := c["radarReferenceJsonl"].(string); return v, ok }},
		{"jsonlOutputDir", func(c map[string]interface{}) (string, bool) { v, ok := c["jsonlOutputDir"].(string); return v, ok }},
	}
	for _, field := range requiredFields {
		val, ok := field.getter(perceptionConfig)
		if !ok || strings.TrimSpace(val) == "" {
			return fmt.Errorf("配置文件 drivers/config.json 缺少 %s，请在配置管理中填写有效路径后再启动", field.name)
		}
	}

	// 先停止可能已存在的进程
	_, _ = client.Execute(fmt.Sprintf("pkill -f '%s' 2>/dev/null || true", pkillPattern(processPattern)))

	// start.sh 内部已用 nohup 后台启动 run_local_rknn.sh 并写 pid 文件，同步执行即可
	startCmd := fmt.Sprintf("cd %s && bash %s", workDir, perceptionServiceConfig.StartScript)
	if _, err := client.ExecuteWithTimeout(startCmd, 60*time.Second); err != nil {
		logTail, _ := client.Execute(fmt.Sprintf("tail -30 %s 2>/dev/null", filepath.ToSlash(filepath.Join(workDir, "run.log"))))
		if strings.TrimSpace(logTail) != "" {
			return fmt.Errorf("failed to start perception service: %w, run.log:\n%s", err, strings.TrimSpace(logTail))
		}
		return fmt.Errorf("failed to start perception service: %w", err)
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

	if isTarPackageService(softwareName) {
		return s.stopTarPackageService(client, softwareName, creds.Server.DeployPath)
	}

	pidFile := pidFilePath(creds.Server.DeployPath, softwareName)
	stopCmd := fmt.Sprintf("if [ -f %s ]; then kill $(cat %s) 2>/dev/null; rm -f %s; fi", pidFile, pidFile, pidFile)
	if _, err := client.ExecuteWithTimeout(stopCmd, 60*time.Second); err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	return nil
}

// stopTarPackageService 停止 tar 包算法服务（control / perception）
func (s *ServerService) stopTarPackageService(client *ssh.Client, softwareName string, deployPath string) error {
	if isPerceptionService(softwareName) {
		return s.stopPerceptionService(client, deployPath)
	}
	return s.stopControlService(client, deployPath)
}

// stopControlService 停止 opentraffic-control
func (s *ServerService) stopControlService(client *ssh.Client, deployPath string) error {
	deployDir := filepath.ToSlash(filepath.Join(deployPath, controlServiceConfig.DirName))
	pidFile := filepath.ToSlash(filepath.Join(deployDir, controlServiceConfig.PidFileName))
	stopCmd := fmt.Sprintf(
		"if [ -f %s ]; then kill $(cat %s) 2>/dev/null; rm -f %s; fi; pkill -f '[r]un_algorithms.py' 2>/dev/null || true",
		pidFile, pidFile, pidFile,
	)
	if _, err := client.ExecuteWithTimeout(stopCmd, 60*time.Second); err != nil {
		return fmt.Errorf("failed to stop control service: %w", err)
	}
	return nil
}

// stopPerceptionService 停止 opentraffic-perception
func (s *ServerService) stopPerceptionService(client *ssh.Client, deployPath string) error {
	workDir, err := resolvePerceptionWorkDir(client, deployPath)
	if err != nil {
		return err
	}
	pidFile := filepath.ToSlash(filepath.Join(workDir, perceptionServiceConfig.PidFileName))
	processPattern := resolvePerceptionProcessPattern(client, workDir)

	// 优先使用 deploy/stop.sh，同时清理 pid 文件，并回退到 pkill
	stopScript := filepath.ToSlash(filepath.Join(workDir, perceptionServiceConfig.StopScript))
	stopCmd := fmt.Sprintf(
		"if [ -x %s ]; then cd %s && bash %s; fi; "+
			"if [ -f %s ]; then kill $(cat %s) 2>/dev/null; rm -f %s; fi; "+
			"pkill -f '%s' 2>/dev/null || true",
		stopScript, workDir, stopScript, pidFile, pidFile, pidFile, pkillPattern(processPattern),
	)
	if _, err := client.ExecuteWithTimeout(stopCmd, 60*time.Second); err != nil {
		return fmt.Errorf("failed to stop perception service: %w", err)
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
