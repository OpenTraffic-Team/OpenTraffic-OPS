package service

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"opentraffic-ops-init-backend/internal/model"
	"opentraffic-ops-init-backend/internal/repository"
	"opentraffic-ops-init-backend/pkg/assets"
	"opentraffic-ops-init-backend/pkg/ssh"
	"strings"
	"time"
)

// DeployService 部署服务
type DeployService struct {
	serverService      *ServerService
	deployRecordRepo   *repository.DeployRecordRepository
}

// NewDeployService 创建部署服务
func NewDeployService(serverService *ServerService) *DeployService {
	return &DeployService{
		serverService:    serverService,
		deployRecordRepo: repository.NewDeployRecordRepository(),
	}
}

// DeployRequest 部署请求
type DeployRequest struct {
	ServerID      string  `json:"server_id" binding:"required"`
	BinaryName    string  `json:"binary_name" binding:"required,oneof=opentraffic-ops-proxy opentraffic-ops opentraffic-control opentraffic-perception"`
	Version       string  `json:"version"`        // 可选：部署版本（opentraffic-control/opentraffic-perception 等可重复部署资源使用）
	ConfigContent *string `json:"config_content"` // 可选：自定义配置内容
}

// isRepeatableDeploy 判断该资源是否允许重复部署（每次部署都会产生一条新的成功记录）
func isRepeatableDeploy(binaryName string) bool {
	return binaryName == "opentraffic-control" || binaryName == "opentraffic-perception"
}

// Deploy 执行部署
func (s *DeployService) Deploy(req *DeployRequest, userName string) (*model.DeployRecord, error) {
	// 1. 获取服务器配置（解密凭据）
	sshConfig, server, err := s.serverService.BuildSSHConfig(req.ServerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server config: %w", err)
	}

	// 检查是否已部署过该服务（opentraffic-control 等可重复部署资源除外）
	if !isRepeatableDeploy(req.BinaryName) {
		hasDeployed, err := s.deployRecordRepo.HasSuccessfulDeploy(req.ServerID, req.BinaryName)
		if err != nil {
			return nil, fmt.Errorf("failed to check deploy history: %w", err)
		}
		if hasDeployed {
			return nil, fmt.Errorf("service %s has already been deployed on this server", req.BinaryName)
		}
	}

	// 创建部署记录（路径稍后根据远程架构更新）
	record := &model.DeployRecord{
		ServerID:   server.ID,
		ServerName: server.Name,
		BinaryName: req.BinaryName,
		RemotePath: "",
		Version:    req.Version,
		Status:     string(model.DeployStatusPending),
		CreatedAt:  time.Now(),
	}

	if err := s.deployRecordRepo.Create(record); err != nil {
		return nil, fmt.Errorf("failed to create deploy record: %w", err)
	}

	// 执行部署流程
	var deployLog strings.Builder
	deployLog.WriteString(fmt.Sprintf("[%s] 开始部署 %s 到 %s (%s)\n",
		time.Now().Format("2006-01-02 15:04:05"), req.BinaryName, server.Name, server.Host))

	// 2. 建立SSH连接
	client, err := ssh.NewClient(sshConfig)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] SSH连接失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("ssh connection failed: %w", err)
	}
	defer client.Close()
	deployLog.WriteString(fmt.Sprintf("[%s] SSH连接成功\n", time.Now().Format("2006-01-02 15:04:05")))

	// opentraffic-control / opentraffic-perception 算法包走 tar 包部署分支
	if req.BinaryName == "opentraffic-control" {
		return s.deployTarPackage(client, server, req, record, &deployLog)
	}
	if req.BinaryName == "opentraffic-perception" {
		return s.deployPerceptionPackage(client, server, req, record, &deployLog)
	}

	// 3. 探测远程服务器架构并选择对应二进制
	arch, err := detectRemoteArch(client)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 探测远程架构失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to detect remote architecture: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 远程架构: %s\n", time.Now().Format("2006-01-02 15:04:05"), arch))

	binaryFileName, err := getBinaryFileName(req.BinaryName, arch)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 不支持的架构: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, err
	}
	if !assets.HasBinary(binaryFileName) {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 嵌入式二进制文件不存在: %s\n", binaryFileName))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("binary file not found: %s", binaryFileName)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 使用二进制: %s\n", time.Now().Format("2006-01-02 15:04:05"), binaryFileName))

	// 更新部署记录中的远程路径
	remotePath := filepath.Join(server.DeployPath, binaryFileName)
	record.RemotePath = remotePath
	if err := s.deployRecordRepo.Update(record); err != nil {
		return record, fmt.Errorf("failed to update deploy record: %w", err)
	}

	// 4. 创建远程部署目录
	mkdirCmd := fmt.Sprintf("mkdir -p %s", server.DeployPath)
	if _, err := client.Execute(mkdirCmd); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 创建远程目录失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to create remote directory: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 创建远程目录: %s\n", time.Now().Format("2006-01-02 15:04:05"), server.DeployPath))

	// 5. 读取嵌入的二进制文件
	reader, err := assets.GetBinaryReader(binaryFileName)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 读取二进制文件失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to read binary file: %w", err)
	}
	defer reader.Close()

	// 获取文件大小
	binaryData, err := io.ReadAll(reader)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 读取二进制内容失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to read binary content: %w", err)
	}

	// 5. 通过SFTP上传到远程服务器
	if err := client.UploadFile(bytes.NewReader(binaryData), remotePath, int64(len(binaryData))); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 上传文件失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to upload file: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 上传文件成功: %s (%d bytes)\n",
		time.Now().Format("2006-01-02 15:04:05"), remotePath, len(binaryData)))

	// 6. 设置可执行权限
	chmodCmd := fmt.Sprintf("chmod +x %s", remotePath)
	if _, err := client.Execute(chmodCmd); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 设置可执行权限失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to set executable permission: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 设置可执行权限成功\n", time.Now().Format("2006-01-02 15:04:05")))

	// 6.5 为部署的软件创建/更新配置文件（opentraffic-control 在 tar 包部署分支中单独处理）
	meta, hasMeta := softwareConfigMeta[req.BinaryName]
	if hasMeta && !isControlService(req.BinaryName) {
		configDirCmd := fmt.Sprintf("eval echo %s", meta.ConfigDir)
		configDir, _ := client.Execute(configDirCmd)
		configDir = strings.TrimSpace(configDir)
		if configDir == "" {
			fallbackDir := strings.TrimPrefix(meta.ConfigDir, "~/")
			configDir = fmt.Sprintf("/home/%s/%s", server.Username, fallbackDir)
		}
		configPath := filepath.Join(configDir, meta.ConfigFile)

		if req.ConfigContent != nil && *req.ConfigContent != "" {
			// 用户提供了自定义配置，直接写入
			if err := client.WriteFile([]byte(*req.ConfigContent), configPath); err != nil {
				deployLog.WriteString(fmt.Sprintf("[WARN] 写入配置文件失败: %v\n", err))
			} else {
				deployLog.WriteString(fmt.Sprintf("[%s] 写入配置文件: %s\n", time.Now().Format("2006-01-02 15:04:05"), configPath))
			}
		} else {
			// 未提供配置，检查是否已存在
			_, readErr := client.ReadFile(configPath)
			if readErr != nil {
				// 不存在则创建默认配置
				defaultConfig := getDefaultConfig(req.BinaryName)
				if err := client.WriteFile([]byte(defaultConfig), configPath); err != nil {
					deployLog.WriteString(fmt.Sprintf("[WARN] 创建默认配置文件失败: %v\n", err))
				} else {
					deployLog.WriteString(fmt.Sprintf("[%s] 创建默认配置文件: %s\n", time.Now().Format("2006-01-02 15:04:05"), configPath))
				}
			} else {
				deployLog.WriteString(fmt.Sprintf("[%s] 配置文件已存在，保持不变: %s\n", time.Now().Format("2006-01-02 15:04:05"), configPath))
			}
		}
	}

	// 更新部署记录为成功
	deployLog.WriteString(fmt.Sprintf("[%s] 部署完成\n", time.Now().Format("2006-01-02 15:04:05")))
	record.Status = string(model.DeployStatusSuccess)
	record.Log = deployLog.String()
	record.RemotePath = remotePath
	if err := s.deployRecordRepo.Update(record); err != nil {
		return record, fmt.Errorf("deploy succeeded but failed to update record: %w", err)
	}

	return record, nil
}

// deployTarPackage 部署 tar 包资源（opentraffic-control）
func (s *DeployService) deployTarPackage(client *ssh.Client, server *model.Server, req *DeployRequest, record *model.DeployRecord, deployLog *strings.Builder) (*model.DeployRecord, error) {
	const packageDir = "opentraffic-control"

	// 探测远程服务器架构并选择对应 tar 包
	arch, err := detectRemoteArch(client)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 探测远程架构失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to detect remote architecture: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 远程架构: %s\n", time.Now().Format("2006-01-02 15:04:05"), arch))

	tarFileName, err := getControlTarFileName(arch)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 不支持的架构: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, err
	}
	if !assets.HasBinary(tarFileName) {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 嵌入式 tar 包不存在: %s\n", tarFileName))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("tar package not found: %s", tarFileName)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 使用 tar 包: %s\n", time.Now().Format("2006-01-02 15:04:05"), tarFileName))

	remoteDir := filepath.Join(server.DeployPath, packageDir)
	remoteTarPath := filepath.Join(remoteDir, tarFileName)

	// 创建远程部署目录
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if _, err := client.Execute(mkdirCmd); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 创建远程目录失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to create remote directory: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 创建远程目录: %s\n", time.Now().Format("2006-01-02 15:04:05"), remoteDir))

	// 读取嵌入的 tar 包
	reader, err := assets.GetBinaryReader(tarFileName)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 读取 tar 包失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to read tar package: %w", err)
	}
	defer reader.Close()

	tarData, err := io.ReadAll(reader)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 读取 tar 包内容失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to read tar content: %w", err)
	}

	// 上传 tar 包
	if err := client.UploadFile(bytes.NewReader(tarData), remoteTarPath, int64(len(tarData))); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 上传 tar 包失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to upload tar package: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 上传 tar 包成功: %s (%d bytes)\n",
		time.Now().Format("2006-01-02 15:04:05"), remoteTarPath, len(tarData)))

	// 检测 tar 包是否存在单一顶层目录，若存在则剥离，确保内容直接落到 remoteDir 下
	stripOpt := ""
	if root, ok := tarHasSingleRootDir(tarData); ok {
		stripOpt = "--strip-components=1"
		deployLog.WriteString(fmt.Sprintf("[%s] tar 包存在单一顶层目录 %s，解压时剥离\n",
			time.Now().Format("2006-01-02 15:04:05"), root))
	}

	// 解压 tar 包
	extractParts := []string{fmt.Sprintf("cd %s", remoteDir), "&&", "tar", "-xf", tarFileName}
	if stripOpt != "" {
		extractParts = append(extractParts, stripOpt)
	}
	extractParts = append(extractParts, "&&", "rm", "-f", tarFileName)
	extractCmd := strings.Join(extractParts, " ")
	if _, err := client.ExecuteWithTimeout(extractCmd, 120*time.Second); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 解压 tar 包失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to extract tar package: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 解压 tar 包成功\n", time.Now().Format("2006-01-02 15:04:05")))

	// 龙芯/ARM 架构需确保 trafficlight_env 虚拟环境已部署到部署目录
	if envPackage := controlEnvPackage(arch); envPackage != "" {
		if err := s.ensureControlPythonEnv(client, remoteDir, envPackage, deployLog, record); err != nil {
			return record, err
		}
	}

	// 写入用户自定义配置（mq_config.json）
	if req.ConfigContent != nil && *req.ConfigContent != "" {
		configDir := filepath.Join(remoteDir, "config")
		configPath := filepath.Join(configDir, "mq_config.json")
		mkdirConfigCmd := fmt.Sprintf("mkdir -p %s", configDir)
		if _, err := client.Execute(mkdirConfigCmd); err != nil {
			deployLog.WriteString(fmt.Sprintf("[WARN] 创建配置目录失败: %v\n", err))
		} else if err := client.WriteFile([]byte(*req.ConfigContent), configPath); err != nil {
			deployLog.WriteString(fmt.Sprintf("[WARN] 写入 mq_config.json 失败: %v\n", err))
		} else {
			deployLog.WriteString(fmt.Sprintf("[%s] 写入配置文件: %s\n", time.Now().Format("2006-01-02 15:04:05"), configPath))
		}
	}

	// 为启动脚本赋予可执行权限
	startScriptPath := filepath.Join(remoteDir, controlServiceConfig.StartScript)
	chmodCmd := fmt.Sprintf("chmod +x %s", startScriptPath)
	if _, err := client.Execute(chmodCmd); err != nil {
		deployLog.WriteString(fmt.Sprintf("[WARN] 设置启动脚本可执行权限失败: %v\n", err))
	} else {
		deployLog.WriteString(fmt.Sprintf("[%s] 设置启动脚本可执行权限: %s\n", time.Now().Format("2006-01-02 15:04:05"), startScriptPath))
	}

	// 更新部署记录为成功
	deployLog.WriteString(fmt.Sprintf("[%s] 部署完成\n", time.Now().Format("2006-01-02 15:04:05")))
	record.Status = string(model.DeployStatusSuccess)
	record.Log = deployLog.String()
	record.RemotePath = remoteDir
	if err := s.deployRecordRepo.Update(record); err != nil {
		return record, fmt.Errorf("deploy succeeded but failed to update record: %w", err)
	}

	return record, nil
}

// deployPerceptionPackage 部署 opentraffic-perception tar 包资源（支持 amd64 / arm64）
func (s *DeployService) deployPerceptionPackage(client *ssh.Client, server *model.Server, req *DeployRequest, record *model.DeployRecord, deployLog *strings.Builder) (*model.DeployRecord, error) {
	const packageDir = "opentraffic-perception"

	arch, err := detectRemoteArch(client)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 探测远程架构失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to detect remote architecture: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 远程架构: %s\n", time.Now().Format("2006-01-02 15:04:05"), arch))

	tarFileName, err := getPerceptionTarFileName(arch)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, err
	}

	if !assets.HasBinary(tarFileName) {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 嵌入式 tar 包不存在: %s\n", tarFileName))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("tar package not found: %s", tarFileName)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 使用 tar 包: %s\n", time.Now().Format("2006-01-02 15:04:05"), tarFileName))

	remoteDir := filepath.Join(server.DeployPath, packageDir)
	remoteTarPath := filepath.Join(remoteDir, tarFileName)

	// 创建远程部署目录
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if _, err := client.Execute(mkdirCmd); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 创建远程目录失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to create remote directory: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 创建远程目录: %s\n", time.Now().Format("2006-01-02 15:04:05"), remoteDir))

	// 读取嵌入的 tar 包
	reader, err := assets.GetBinaryReader(tarFileName)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 读取 tar 包失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to read tar package: %w", err)
	}
	defer reader.Close()

	tarData, err := io.ReadAll(reader)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 读取 tar 包内容失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to read tar content: %w", err)
	}

	// 上传 tar 包
	if err := client.UploadFile(bytes.NewReader(tarData), remoteTarPath, int64(len(tarData))); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 上传 tar 包失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to upload tar package: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 上传 tar 包成功: %s (%d bytes)\n",
		time.Now().Format("2006-01-02 15:04:05"), remoteTarPath, len(tarData)))

	// 解压 tar 包到远程目录
	extractCmd := fmt.Sprintf("cd %s && tar -xf %s && rm -f %s", remoteDir, tarFileName, tarFileName)
	if _, err := client.ExecuteWithTimeout(extractCmd, 120*time.Second); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 解压 tar 包失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to extract tar package: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 解压 tar 包成功\n", time.Now().Format("2006-01-02 15:04:05")))

	// 根据实际解压结构定位工作目录（扁平或带架构子目录）
	workDir, err := resolvePerceptionWorkDir(client, server.DeployPath)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 定位 perception 工作目录失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to resolve perception work directory: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] perception 工作目录: %s\n", time.Now().Format("2006-01-02 15:04:05"), workDir))

	// 运行 install.sh 准备运行环境（无参数，脚本会按 conda / venv 依次回退）
	installCmd := fmt.Sprintf("cd %s && bash deploy/install.sh", workDir)
	if _, err := client.ExecuteWithTimeout(installCmd, 600*time.Second); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 运行 install.sh 失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to run deploy/install.sh: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 运行 install.sh 成功\n", time.Now().Format("2006-01-02 15:04:05")))

	// 运行 configure.sh 生成默认 drivers/config.json
	configureCmd := fmt.Sprintf("cd %s && bash deploy/configure.sh", workDir)
	if _, err := client.ExecuteWithTimeout(configureCmd, 60*time.Second); err != nil {
		deployLog.WriteString(fmt.Sprintf("[WARN] 运行 configure.sh 失败: %v\n", err))
	} else {
		deployLog.WriteString(fmt.Sprintf("[%s] 运行 configure.sh 成功\n", time.Now().Format("2006-01-02 15:04:05")))
	}

	// 写入用户自定义配置（drivers/config.json）
	if req.ConfigContent != nil && *req.ConfigContent != "" {
		configPath := filepath.Join(workDir, "drivers", "config.json")
		if err := client.WriteFile([]byte(*req.ConfigContent), configPath); err != nil {
			deployLog.WriteString(fmt.Sprintf("[WARN] 写入 drivers/config.json 失败: %v\n", err))
		} else {
			deployLog.WriteString(fmt.Sprintf("[%s] 写入配置文件: %s\n", time.Now().Format("2006-01-02 15:04:05"), configPath))
		}
	}

	// 为 deploy 目录下的脚本赋予可执行权限
	chmodCmd := fmt.Sprintf("chmod +x %s/deploy/*.sh", workDir)
	if _, err := client.Execute(chmodCmd); err != nil {
		deployLog.WriteString(fmt.Sprintf("[WARN] 设置 deploy 脚本可执行权限失败: %v\n", err))
	} else {
		deployLog.WriteString(fmt.Sprintf("[%s] 设置 deploy 脚本可执行权限\n", time.Now().Format("2006-01-02 15:04:05")))
	}

	// 更新部署记录为成功
	deployLog.WriteString(fmt.Sprintf("[%s] 部署完成\n", time.Now().Format("2006-01-02 15:04:05")))
	record.Status = string(model.DeployStatusSuccess)
	record.Log = deployLog.String()
	record.RemotePath = workDir
	if err := s.deployRecordRepo.Update(record); err != nil {
		return record, fmt.Errorf("deploy succeeded but failed to update record: %w", err)
	}

	return record, nil
}

// getPerceptionTarFileName 根据远程架构生成 opentraffic-perception tar 包文件名
func getPerceptionTarFileName(arch string) (tarFileName string, err error) {
	suffix, ok := archToBinarySuffix[strings.ToLower(strings.TrimSpace(arch))]
	if !ok {
		return "", fmt.Errorf("unsupported architecture for perception: %s", arch)
	}
	if suffix != "linux-amd64" && suffix != "linux-arm64" {
		return "", fmt.Errorf("opentraffic-perception does not support architecture: %s", suffix)
	}
	return fmt.Sprintf("opentraffic-perception-%s.tar", suffix), nil
}

// getControlTarFileName 根据远程架构生成 opentraffic-control tar 包文件名
func getControlTarFileName(arch string) (string, error) {
	suffix, ok := archToBinarySuffix[strings.ToLower(strings.TrimSpace(arch))]
	if !ok {
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}
	return fmt.Sprintf("opentraffic-control-%s.tar", suffix), nil
}

// tarHasSingleRootDir 检查 tar 包是否只有一个顶层目录；若是，返回该目录名
func tarHasSingleRootDir(data []byte) (string, bool) {
	r := tar.NewReader(bytes.NewReader(data))
	root := ""
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", false
		}
		name := strings.TrimPrefix(h.Name, "./")
		if root == "" {
			if h.Typeflag == tar.TypeDir {
				if !strings.HasSuffix(name, "/") {
					name += "/"
				}
				root = name
				continue
			}
			idx := strings.Index(name, "/")
			if idx == -1 {
				return "", false
			}
			root = name[:idx+1]
			continue
		}
		if !strings.HasPrefix(name, root) {
			return "", false
		}
	}
	return root, root != ""
}

// updateRecordFailed 更新部署记录为失败
func (s *DeployService) updateRecordFailed(id int, log string) {
	_ = s.deployRecordRepo.UpdateStatus(id, model.DeployStatusFailed, log)
}

// controlEnvPackage 返回该架构 control 服务所需的 Python 环境包名；空串表示无需环境包
func controlEnvPackage(arch string) string {
	switch strings.ToLower(strings.TrimSpace(arch)) {
	case "x86_64", "amd64":
		return "trafficlight-amd64.tar.gz"
	case "loongarch64":
		return "trafficlight-loong64.tar.gz"
	case "aarch64", "arm64":
		return "trafficlight-arm64.tar.gz"
	}
	return ""
}

// ensureControlPythonEnv 确保 trafficlight_env 虚拟环境已解压到部署目录
func (s *DeployService) ensureControlPythonEnv(client *ssh.Client, remoteDir string, packageName string, deployLog *strings.Builder, record *model.DeployRecord) error {
	pythonPath := filepath.Join(remoteDir, "trafficlight_env", "bin", "python3")

	checkCmd := fmt.Sprintf("test -f %s && echo exists || echo missing", pythonPath)
	output, err := client.Execute(checkCmd)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 检查 Python 环境失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return fmt.Errorf("failed to check python env: %w", err)
	}
	if strings.TrimSpace(output) == "exists" {
		deployLog.WriteString(fmt.Sprintf("[%s] Python 环境已存在: %s\n",
			time.Now().Format("2006-01-02 15:04:05"), pythonPath))
		return nil
	}

	if !assets.HasBinary(packageName) {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 嵌入式 Python 环境包不存在: %s\n", packageName))
		s.updateRecordFailed(record.ID, deployLog.String())
		return fmt.Errorf("embedded python env package not found: %s", packageName)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] Python 环境不存在，开始部署 %s\n",
		time.Now().Format("2006-01-02 15:04:05"), packageName))

	reader, err := assets.GetBinaryReader(packageName)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 读取 Python 环境包失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return fmt.Errorf("failed to read python env package: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 读取 Python 环境包内容失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return fmt.Errorf("failed to read python env content: %w", err)
	}

	remoteTarPath := filepath.Join(remoteDir, packageName)
	if err := client.UploadFile(bytes.NewReader(data), remoteTarPath, int64(len(data))); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 上传 Python 环境包失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return fmt.Errorf("failed to upload python env package: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 上传 Python 环境包成功: %s (%d bytes)\n",
		time.Now().Format("2006-01-02 15:04:05"), remoteTarPath, len(data)))

	extractCmd := fmt.Sprintf("cd %s && tar -xzf %s && rm -f %s",
		remoteDir, packageName, packageName)
	if _, err := client.ExecuteWithTimeout(extractCmd, 300*time.Second); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 解压 Python 环境包失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return fmt.Errorf("failed to extract python env package: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] Python 环境部署完成: %s\n",
		time.Now().Format("2006-01-02 15:04:05"), pythonPath))
	return nil
}

// UndeployRequest 卸载请求
type UndeployRequest struct {
	ServerID   string `json:"server_id" binding:"required"`
	BinaryName string `json:"binary_name" binding:"required,oneof=opentraffic-ops-proxy opentraffic-ops opentraffic-control opentraffic-perception"`
}

// isLegacyControlName 兼容旧记录中使用的 opentraffic-control-linux-amd64 名称
func isLegacyControlName(binaryName string) bool {
	return binaryName == "opentraffic-control-linux-amd64"
}

// isLegacyPerceptionName 兼容旧记录中使用的 opentraffic-perception-linux-amd64 名称
func isLegacyPerceptionName(binaryName string) bool {
	return binaryName == "opentraffic-perception-linux-amd64"
}

// Undeploy 执行卸载
func (s *DeployService) Undeploy(req *UndeployRequest) error {
	// 1. 获取服务器配置（解密凭据）
	sshConfig, server, err := s.serverService.BuildSSHConfig(req.ServerID)
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	// 2. 建立SSH连接
	client, err := ssh.NewClient(sshConfig)
	if err != nil {
		return fmt.Errorf("ssh connection failed: %w", err)
	}
	defer client.Close()

	// opentraffic-control 算法包卸载：先停止服务，再删除整个目录
	if req.BinaryName == "opentraffic-control" || isLegacyControlName(req.BinaryName) {
		_ = s.serverService.StopService(req.ServerID, "opentraffic-control")
		remoteDir := filepath.Join(server.DeployPath, "opentraffic-control")
		_, _ = client.Execute(fmt.Sprintf("rm -rf %s", remoteDir))
		// 同时兼容旧路径
		oldRemoteDir := filepath.Join(server.DeployPath, "ops/opentraffic-control")
		_, _ = client.Execute(fmt.Sprintf("rm -rf %s", oldRemoteDir))
		// 同时删除新名称与旧名称（旧版本使用过 opentraffic-control-linux-amd64）的部署记录
		_ = s.deployRecordRepo.DeleteByServerAndBinary(req.ServerID, "opentraffic-control")
		_ = s.deployRecordRepo.DeleteByServerAndBinary(req.ServerID, "opentraffic-control-linux-amd64")
		return nil
	}

	// opentraffic-perception 算法包卸载：先停止服务，再删除整个目录
	if req.BinaryName == "opentraffic-perception" || isLegacyPerceptionName(req.BinaryName) {
		_ = s.serverService.StopService(req.ServerID, "opentraffic-perception")
		remoteDir := filepath.Join(server.DeployPath, "opentraffic-perception")
		_, _ = client.Execute(fmt.Sprintf("rm -rf %s", remoteDir))
		_ = s.deployRecordRepo.DeleteByServerAndBinary(req.ServerID, "opentraffic-perception")
		_ = s.deployRecordRepo.DeleteByServerAndBinary(req.ServerID, "opentraffic-perception-linux-amd64")
		return nil
	}

	// 3. 获取最新的成功部署记录
	record, err := s.deployRecordRepo.GetLatestSuccessfulDeploy(req.ServerID, req.BinaryName)
	if err != nil {
		return fmt.Errorf("no successful deploy record found for %s on this server", req.BinaryName)
	}

	binaryFileName := ""
	remotePath := record.RemotePath
	if remotePath == "" {
		// 如果记录中没有远程路径，探测远程架构后构造
		arch, err := detectRemoteArch(client)
		if err == nil {
			binaryFileName, _ = getBinaryFileName(req.BinaryName, arch)
		}
		if binaryFileName == "" {
			// 兼容旧记录：未探测到架构时默认 amd64
			binaryFileName = fmt.Sprintf("%s-linux-amd64", req.BinaryName)
		}
		remotePath = filepath.Join(server.DeployPath, binaryFileName)
	}

	// 4. 停止进程并删除pid文件
	pidFile := pidFilePath(server.DeployPath, req.BinaryName)
	_, _ = client.Execute(fmt.Sprintf("if [ -f %s ]; then kill $(cat %s) 2>/dev/null; rm -f %s; fi", pidFile, pidFile, pidFile))

	// 5. 删除远程二进制文件
	if remotePath != "" {
		_, _ = client.Execute(fmt.Sprintf("rm -f %s", remotePath))
	}

	// 7. 删除部署记录
	if err := s.deployRecordRepo.DeleteByServerAndBinary(req.ServerID, req.BinaryName); err != nil {
		return fmt.Errorf("undeploy succeeded but failed to delete record: %w", err)
	}

	return nil
}

// ListRecords 获取部署记录列表
func (s *DeployService) ListRecords(serverID string) ([]model.DeployRecord, error) {
	if serverID != "" {
		return s.deployRecordRepo.ListByServerID(serverID)
	}
	return s.deployRecordRepo.List()
}

// GetRecord 获取部署记录详情
func (s *DeployService) GetRecord(id int) (*model.DeployRecord, error) {
	return s.deployRecordRepo.GetByID(id)
}
