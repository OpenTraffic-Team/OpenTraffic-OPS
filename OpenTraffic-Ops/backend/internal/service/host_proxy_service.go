package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"opentraffic-ops-backend/internal/config"
	"opentraffic-ops-backend/internal/dto"
)

// HostProxyCommandTTL 待执行指令存活时长
const HostProxyCommandTTL = 5 * time.Minute

// RedisKeyHostCommands Redis 指令队列 key 模板
const RedisKeyHostCommands = "rtm:host:commands:%s"

// HostProxyService HostProxy 业务服务
type HostProxyService struct {
	redis             *redis.Client
	hostService       *HostInfoService
	hostHealthService *HostHealthService
}

// NewHostProxyService 创建 HostProxy 服务
func NewHostProxyService(hostService *HostInfoService, hostHealthService *HostHealthService) *HostProxyService {
	return &HostProxyService{
		redis:             config.RedisPlatform,
		hostService:       hostService,
		hostHealthService: hostHealthService,
	}
}

// Heartbeat 处理 HostProxy 心跳上报（合并健康度数据）
func (s *HostProxyService) Heartbeat(ctx context.Context, req *dto.HostProxyHeartbeatRequest) error {
	if req.IP == "" {
		return fmt.Errorf("IP不能为空")
	}

	// 自动建立记录（若不存在）
	if err := s.hostService.RegisterHandler(ctx, req.IP); err != nil {
		zap.L().Warn("HostProxy 心跳自动建立记录失败",
			zap.String("ip", req.IP),
			zap.Error(err))
	}

	// 获取主机信息（用于 host_id）
	host, err := s.hostService.GetByIP(ctx, req.IP)
	if err != nil {
		zap.L().Warn("HostProxy 心跳查询主机失败",
			zap.String("ip", req.IP),
			zap.Error(err))
		// 继续执行，host_id 为 0
	}

	// 更新心跳时间和在线状态（同时更新 heartbeat_interval）
	if err := s.hostService.UpdateHeartbeat(ctx, req.IP, req.HeartbeatInterval); err != nil {
		zap.L().Warn("HostProxy 心跳更新时间失败",
			zap.String("ip", req.IP),
			zap.Int("heartbeatInterval", req.HeartbeatInterval),
			zap.Error(err))
	} else {
		zap.L().Debug("HostProxy 心跳更新成功",
			zap.String("ip", req.IP),
			zap.Int("heartbeatInterval", req.HeartbeatInterval))
	}

	var hostID int64
	if host != nil {
		hostID = host.ID
	}

	// 插入健康度数据到 PG
	if err := s.hostHealthService.SaveHostHealth(ctx, hostID, req.IP, req); err != nil {
		zap.L().Warn("HostProxy 保存健康度数据失败",
			zap.String("ip", req.IP),
			zap.Error(err))
	}

	return nil
}

// Register 处理 HostProxy 注册（首次启动）
func (s *HostProxyService) Register(ctx context.Context, req *dto.HostProxyRegisterRequest) (*dto.HostProxyRegisterResponse, error) {
	host, err := s.hostService.UpsertHostProxyInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	resp := &dto.HostProxyRegisterResponse{
		Registered: true,
		Message:    "主机已注册，可正常上报数据",
	}

	// 如果主机名为空，提示管理员配置
	if host.Name == "" {
		resp.Message = "主机自动入库成功，请在管理后台配置主机名称"
	}

	return resp, nil
}

// Poll HostProxy 拉取待执行指令
func (s *HostProxyService) Poll(ctx context.Context, ip string) ([]dto.HostProxyCommand, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis client is nil")
	}

	key := fmt.Sprintf(RedisKeyHostCommands, ip)
	commands := make([]dto.HostProxyCommand, 0)

	// 全部弹出后下发，避免重复执行
	for {
		val, err := s.redis.LPop(ctx, key).Result()
		if err == redis.Nil {
			break
		}
		if err != nil {
			zap.L().Error("Redis LPop 指令失败",
				zap.String("ip", ip),
				zap.Error(err))
			return nil, err
		}

		var cmd dto.HostProxyCommand
		if err := json.Unmarshal([]byte(val), &cmd); err != nil {
			zap.L().Warn("HostProxy 指令反序列化失败",
				zap.String("ip", ip),
				zap.String("raw", val),
				zap.Error(err))
			continue
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

// AckCommand HostProxy 上报指令执行结果
func (s *HostProxyService) AckCommand(ctx context.Context, req *dto.HostProxyCommandAckRequest) error {
	zap.L().Info("HostProxy 指令执行结果",
		zap.String("ip", req.IP),
		zap.String("commandId", req.CommandID),
		zap.Bool("success", req.Success),
		zap.String("message", req.Message))
	// 当前实现仅记录日志，可扩展为持久化执行结果
	return nil
}
