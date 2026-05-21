package service

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"rtm-server/internal/dto"
	"rtm-server/internal/model"
	"rtm-server/internal/repository"
	"rtm-server/internal/utils"
)

// HostOnlineStatus 主机在线状态（存储在sync.Map中）
type HostOnlineStatus struct {
	IP         string `json:"ip"`
	Status     bool   `json:"status"`
	CreateDate string `json:"createDate"`
}

// HostInfoService 主机信息服务
type HostInfoService struct {
	repo *repository.HostInfoRepository
}

var (
	// onlineMap 全局主机在线状态映射（与Java ConcurrentHashMap等效）
	onlineMap sync.Map
	onceLoad  sync.Once
)

// NewHostInfoService 创建主机信息服务
func NewHostInfoService(db *gorm.DB) *HostInfoService {
	svc := &HostInfoService{repo: repository.NewHostInfoRepository(db)}
	// 首次创建时加载在线状态
	onceLoad.Do(func() {
		svc.LoadOnline(context.Background())
	})
	return svc
}

// LoadOnline 加载主机在线状态到内存
func (s *HostInfoService) LoadOnline(ctx context.Context) {
	// 清空现有条目
	onlineMap.Range(func(key, _ interface{}) bool {
		onlineMap.Delete(key)
		return true
	})
	nowStr := utils.NowStr()

	hosts, err := s.repo.FindAll(ctx)
	if err != nil {
		zap.L().Error("加载主机在线状态失败", zap.Error(err))
		return
	}

	for _, host := range hosts {
		status := false
		if host.IsOnline != nil {
			status = *host.IsOnline
		}

		onlineMap.Store(host.IP, &HostOnlineStatus{
			IP:         host.IP,
			Status:     status,
			CreateDate: nowStr,
		})
	}

	zap.L().Info("主机在线状态加载完成", zap.Int("count", len(hosts)))
}

// List 主机信息列表
func (s *HostInfoService) List(ctx context.Context, query *dto.HostInfoQuery) ([]dto.HostInfoDto, int64, error) {
	total, err := s.repo.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	hosts, err := s.repo.FindPage(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 转换为DTO
	result := make([]dto.HostInfoDto, 0, len(hosts))
	for _, host := range hosts {
		dtoItem := dto.HostInfoDto{
			ID:                host.ID,
			IP:                host.IP,
			Name:              host.Name,
			IsOnline:          host.IsOnline,
			OsType:            host.OsType,
			OsVersion:         host.OsVersion,
			CpuArch:           host.CpuArch,
			CpuCores:          host.CpuCores,
			CpuModel:          host.CpuModel,
			MemTotalMb:        host.MemTotalMb,
			DiskTotalGb:       host.DiskTotalGb,
			GpuInfo:           host.GpuInfo,
			MacAddress:        host.MacAddress,
			ProxyVersion:      host.ProxyVersion,
			HeartbeatInterval: host.HeartbeatInterval,
		}
		if host.LastHeartbeat != nil {
			dtoItem.LastHeartbeat = *host.LastHeartbeat
		}
		if host.OfflineTime != nil {
			dtoItem.OfflineTime = *host.OfflineTime
		}
		dtoItem.RegisterTime = host.RegisterTime
		result = append(result, dtoItem)
	}

	return result, total, nil
}

// GetByID 根据ID获取主机信息
func (s *HostInfoService) GetByID(ctx context.Context, id int64) (*model.HostInfo, error) {
	return s.repo.FindByID(ctx, id)
}

// GetByIP 根据IP获取主机信息
func (s *HostInfoService) GetByIP(ctx context.Context, ip string) (*model.HostInfo, error) {
	return s.repo.FindByIP(ctx, ip)
}

// Create 新增主机信息（管理员修改名称等基本信息）
func (s *HostInfoService) Create(ctx context.Context, req *dto.HostInfoCreateRequest) error {
	if _, err := s.repo.FindByID(ctx, req.ID); err != nil {
		return fmt.Errorf("主机不存在")
	}

	return s.repo.Update(ctx, req.ID, map[string]interface{}{
		"name": req.Name,
	})
}

// Update 修改主机信息
func (s *HostInfoService) Update(ctx context.Context, req *dto.HostInfoUpdateRequest) error {
	if _, err := s.repo.FindByID(ctx, req.ID); err != nil {
		return fmt.Errorf("主机不存在")
	}

	return s.repo.Update(ctx, req.ID, map[string]interface{}{
		"name": req.Name,
	})
}

// DeleteByIDs 批量删除主机信息
func (s *HostInfoService) DeleteByIDs(ctx context.Context, ids []int64) error {
	return s.repo.DeleteByIDs(ctx, ids)
}

// DeleteByID 删除主机信息
func (s *HostInfoService) DeleteByID(ctx context.Context, id int64) error {
	return s.repo.DeleteByIDs(ctx, []int64{id})
}

// SelectHostInfoTreeNode 查询主机信息树形节点
func (s *HostInfoService) SelectHostInfoTreeNode(ctx context.Context) ([]dto.HostInfoTreeNode, error) {
	hosts, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.HostInfoTreeNode, 0, len(hosts))
	for _, host := range hosts {
		label := host.Name + "[" + host.IP + "]"
		result = append(result, dto.HostInfoTreeNode{
			Label:    label,
			IP:       host.IP,
			Children: nil,
		})
	}
	return result, nil
}

// GetHostOnlineStatusByMap 从内存获取主机在线状态 (-1:不存在 0:离线 1:在线)
func (s *HostInfoService) GetHostOnlineStatusByMap(ip string) int {
	val, ok := onlineMap.Load(ip)
	if !ok {
		return -1
	}
	status := val.(*HostOnlineStatus)
	if status.Status {
		return 1
	}
	return 0
}

// SetOnlineStatus 设置主机在线状态
func (s *HostInfoService) SetOnlineStatus(ctx context.Context, ip string, status bool) error {
	nowStr := utils.NowStr()

	update := map[string]interface{}{
		"is_online":      status,
		"last_heartbeat": nowStr,
	}
	if !status {
		update["offline_time"] = nowStr
	}
	if err := s.repo.UpdateByIP(ctx, ip, update); err != nil {
		return err
	}

	updateOnlineMap(ip, status, utils.NowStr())
	return nil
}

// SetOnlineStatusBatch 批量设置主机在线状态（一次 UPDATE 完成，避免 N 次 SQL）
func (s *HostInfoService) SetOnlineStatusBatch(ctx context.Context, ips []string, status bool) error {
	if len(ips) == 0 {
		return nil
	}
	nowStr := utils.NowStr()
	update := map[string]interface{}{
		"is_online":      status,
		"last_heartbeat": nowStr,
	}
	if !status {
		update["offline_time"] = nowStr
	}
	if err := s.repo.UpdateBatchByIP(ctx, ips, update); err != nil {
		return err
	}
	for _, ip := range ips {
		updateOnlineMap(ip, status, utils.NowStr())
	}
	return nil
}

// updateOnlineMap 更新内存中的在线状态条目
func updateOnlineMap(ip string, status bool, nowStr string) {
	val, ok := onlineMap.Load(ip)
	if ok {
		hostStatus := val.(*HostOnlineStatus)
		hostStatus.Status = status
		hostStatus.CreateDate = nowStr
		onlineMap.Store(ip, hostStatus)
		return
	}
	onlineMap.Store(ip, &HostOnlineStatus{
		IP:         ip,
		Status:     status,
		CreateDate: nowStr,
	})
}

// UpdateHeartbeat 更新主机心跳时间（HostProxy 心跳上报时调用）
func (s *HostInfoService) UpdateHeartbeat(ctx context.Context, ip string, heartbeatInterval int) error {
	nowStr := utils.NowStr()

	// 兜底：心跳间隔必须 >= 30 秒，避免异常值导致离线检测误判
	if heartbeatInterval <= 0 {
		heartbeatInterval = 30
	}

	updates := map[string]interface{}{
		"is_online":          true,
		"last_heartbeat":     nowStr,
		"heartbeat_interval": heartbeatInterval,
	}
	if err := s.repo.UpdateByIP(ctx, ip, updates); err != nil {
		return err
	}

	updateOnlineMap(ip, true, utils.NowStr())
	return nil
}

// RegisterHandler HostProxy 首次上报时自动创建记录（若不存在）
// 使用 OnConflict(DoNothing) 避免竞争
func (s *HostInfoService) RegisterHandler(ctx context.Context, ip string) error {
	if ip == "" {
		return nil
	}

	host := &model.HostInfo{
		IP:                ip,
		Name:              "",
		IsOnline:          boolPtr(false),
		HeartbeatInterval: 30, // 兜底默认值，避免首次心跳前被误判离线
		RegisterTime:      utils.NowStr(),
	}
	return s.repo.UpsertByIP(ctx, host)
}

// UpsertHostProxyInfo HostProxy 注册时新增或更新主机硬件信息
func (s *HostInfoService) UpsertHostProxyInfo(ctx context.Context, req *dto.HostProxyRegisterRequest) (*model.HostInfo, error) {
	if req.IP == "" {
		return nil, fmt.Errorf("IP不能为空")
	}

	host, err := s.repo.FindByIP(ctx, req.IP)
	nowStr := utils.NowStr()

	if err == gorm.ErrRecordNotFound {
		// 不存在则插入新记录
		// heartbeat_interval 默认 30 秒，避免新注册主机在首次心跳前被误判离线
		newHost := &model.HostInfo{
			IP:                req.IP,
			Name:              req.HostName,
			OsType:            req.OsType,
			OsVersion:         req.OsVersion,
			CpuArch:           req.CpuArch,
			CpuCores:          req.CpuCores,
			CpuModel:          req.CpuModel,
			MemTotalMb:        req.MemTotalMb,
			DiskTotalGb:       req.DiskTotalGb,
			GpuInfo:           req.GpuInfo,
			MacAddress:        req.MacAddress,
			ProxyVersion:      req.ProxyVersion,
			HeartbeatInterval: 30,
			IsOnline:          boolPtr(false),
			RegisterTime:      nowStr,
		}
		if err := s.repo.Create(ctx, newHost); err != nil {
			return nil, err
		}
		return newHost, nil
	}

	if err != nil {
		return nil, err
	}

	// 已存在，更新硬件信息
	// 同时设置 heartbeat_interval 为 30（兜底），等首次心跳后再更新为实际值
	update := map[string]interface{}{
		"os_type":            req.OsType,
		"os_version":         req.OsVersion,
		"cpu_arch":           req.CpuArch,
		"cpu_cores":          req.CpuCores,
		"cpu_model":          req.CpuModel,
		"mem_total_mb":       req.MemTotalMb,
		"disk_total_gb":      req.DiskTotalGb,
		"gpu_info":           req.GpuInfo,
		"mac_address":        req.MacAddress,
		"proxy_version":      req.ProxyVersion,
		"last_heartbeat":     nowStr,
		"heartbeat_interval": 30,
	}
	if req.HostName != "" && host.Name == "" {
		update["name"] = req.HostName
	}
	if err := s.repo.Update(ctx, host.ID, update); err != nil {
		return nil, err
	}

	return host, nil
}

// GetOnlineMap 获取在线状态映射（用于其他服务查询）
func (s *HostInfoService) GetOnlineMap() *sync.Map {
	return &onlineMap
}

// GetAllHosts 获取所有主机
func (s *HostInfoService) GetAllHosts(ctx context.Context) ([]model.HostInfo, error) {
	return s.repo.FindAll(ctx)
}

// MarkOfflineByTimeout 根据超时阈值标记离线主机（单条SQL原子操作，避免读取-更新竞态）
// 超时阈值 = heartbeat_interval * 5 秒，按各主机自身的心跳间隔计算
func (s *HostInfoService) MarkOfflineByTimeout(ctx context.Context) (int64, error) {
	return s.repo.MarkOfflineByTimeout(ctx)
}

// boolPtr bool转指针
func boolPtr(b bool) *bool {
	return &b
}
