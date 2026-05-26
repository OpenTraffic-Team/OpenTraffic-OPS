package service

import (
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
	BinaryName    string  `json:"binary_name" binding:"required,oneof=opentraffic-ops-proxy opentraffic-ops"`
	ConfigContent *string `json:"config_content"` // 可选：自定义配置内容
}

// binaryFileMap 二进制文件名映射
var binaryFileMap = map[string]string{
	"opentraffic-ops-proxy": "opentraffic-ops-proxy-linux-amd64",
	"opentraffic-ops":       "opentraffic-ops-linux-amd64",
}

// Deploy 执行部署
func (s *DeployService) Deploy(req *DeployRequest, userName string) (*model.DeployRecord, error) {
	// 1. 获取服务器配置（解密凭据）
	sshConfig, server, err := s.serverService.BuildSSHConfig(req.ServerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server config: %w", err)
	}

	// 检查是否已部署过该服务
	hasDeployed, err := s.deployRecordRepo.HasSuccessfulDeploy(req.ServerID, req.BinaryName)
	if err != nil {
		return nil, fmt.Errorf("failed to check deploy history: %w", err)
	}
	if hasDeployed {
		return nil, fmt.Errorf("service %s has already been deployed on this server", req.BinaryName)
	}

	// 检查二进制文件是否存在
	binaryFileName, ok := binaryFileMap[req.BinaryName]
	if !ok {
		return nil, fmt.Errorf("unknown binary name: %s", req.BinaryName)
	}
	if !assets.HasBinary(binaryFileName) {
		return nil, fmt.Errorf("binary file not found: %s", binaryFileName)
	}

	// 创建部署记录
	remotePath := filepath.Join(server.DeployPath, binaryFileName)
	record := &model.DeployRecord{
		ServerID:   server.ID,
		ServerName: server.Name,
		BinaryName: req.BinaryName,
		RemotePath: remotePath,
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

	// 3. 创建远程部署目录
	mkdirCmd := fmt.Sprintf("mkdir -p %s", server.DeployPath)
	if _, err := client.Execute(mkdirCmd); err != nil {
		deployLog.WriteString(fmt.Sprintf("[ERROR] 创建远程目录失败: %v\n", err))
		s.updateRecordFailed(record.ID, deployLog.String())
		return record, fmt.Errorf("failed to create remote directory: %w", err)
	}
	deployLog.WriteString(fmt.Sprintf("[%s] 创建远程目录: %s\n", time.Now().Format("2006-01-02 15:04:05"), server.DeployPath))

	// 4. 读取嵌入的二进制文件
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

	// 6.5 为部署的软件创建/更新配置文件
	meta, hasMeta := softwareConfigMeta[req.BinaryName]
	if hasMeta {
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

// updateRecordFailed 更新部署记录为失败
func (s *DeployService) updateRecordFailed(id int, log string) {
	_ = s.deployRecordRepo.UpdateStatus(id, model.DeployStatusFailed, log)
}

// UndeployRequest 卸载请求
 type UndeployRequest struct {
	ServerID   string `json:"server_id" binding:"required"`
	BinaryName string `json:"binary_name" binding:"required,oneof=opentraffic-ops-proxy opentraffic-ops"`
}

// Undeploy 执行卸载
func (s *DeployService) Undeploy(req *UndeployRequest) error {
	// 1. 获取最新的成功部署记录
	record, err := s.deployRecordRepo.GetLatestSuccessfulDeploy(req.ServerID, req.BinaryName)
	if err != nil {
		return fmt.Errorf("no successful deploy record found for %s on this server", req.BinaryName)
	}

	// 2. 获取服务器配置（解密凭据）
	sshConfig, _, err := s.serverService.BuildSSHConfig(req.ServerID)
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}

	// 3. 建立SSH连接
	client, err := ssh.NewClient(sshConfig)
	if err != nil {
		return fmt.Errorf("ssh connection failed: %w", err)
	}
	defer client.Close()

	binaryFileName := binaryFileMap[req.BinaryName]
	remotePath := filepath.Join(record.RemotePath)
	if remotePath == "" {
		// 如果记录中没有远程路径，从服务器配置中构造
		_, server, _ := s.serverService.BuildSSHConfig(req.ServerID)
		if server != nil {
			remotePath = filepath.Join(server.DeployPath, binaryFileName)
		}
	}

	// 4. 停止进程并删除pid文件
	_, server, _ := s.serverService.BuildSSHConfig(req.ServerID)
	if server != nil {
		pidFile := pidFilePath(server.DeployPath, req.BinaryName)
		_, _ = client.Execute(fmt.Sprintf("if [ -f %s ]; then kill $(cat %s) 2>/dev/null; rm -f %s; fi", pidFile, pidFile, pidFile))
	}

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
