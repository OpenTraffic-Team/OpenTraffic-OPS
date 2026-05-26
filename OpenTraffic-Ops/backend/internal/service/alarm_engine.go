package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/constant"
	"opentraffic-ops-backend/internal/model"
	"opentraffic-ops-backend/internal/repository"
	"opentraffic-ops-backend/internal/utils"
)

// AlarmEngine 告警检查引擎
type AlarmEngine struct {
	db                 *gorm.DB
	alarmRuleRepo      *repository.AlarmRuleRepository
	alarmRecordRepo    *repository.AlarmRecordRepository
	alarmChannelRepo   *repository.AlarmChannelRepository
	hostInfoRepo       *repository.HostInfoRepository
	hostHealthRepo     *repository.HostHealthRepository
	alarmNotifyLogRepo *repository.AlarmNotifyLogRepository

	// 阈值突破开始时间记录：key = "ruleID:hostID"
	breachStartMap sync.Map
}

// NewAlarmEngine 创建告警引擎
func NewAlarmEngine(db *gorm.DB) *AlarmEngine {
	return &AlarmEngine{
		db:                 db,
		alarmRuleRepo:      repository.NewAlarmRuleRepository(db),
		alarmRecordRepo:    repository.NewAlarmRecordRepository(db),
		alarmChannelRepo:   repository.NewAlarmChannelRepository(db),
		hostInfoRepo:       repository.NewHostInfoRepository(db),
		hostHealthRepo:     repository.NewHostHealthRepository(db),
		alarmNotifyLogRepo: repository.NewAlarmNotifyLogRepository(db),
	}
}

// Check 执行一次告警检查
func (e *AlarmEngine) Check() {
	ctx := context.Background()

	// 加载所有启用的告警规则
	rules, err := e.alarmRuleRepo.FindEnabled(ctx)
	if err != nil {
		zap.L().Error("告警引擎加载规则失败", zap.Error(err))
		return
	}

	if len(rules) == 0 {
		return
	}

	// 加载所有主机信息
	hosts, err := e.hostInfoRepo.FindAll(ctx)
	if err != nil {
		zap.L().Error("告警引擎加载主机失败", zap.Error(err))
		return
	}

	// 构建主机映射
	hostMap := make(map[int64]model.HostInfo)
	for _, h := range hosts {
		hostMap[h.ID] = h
	}

	// 获取所有主机的最新健康度数据
	ips := make([]string, 0, len(hosts))
	for _, h := range hosts {
		if h.IP != "" {
			ips = append(ips, h.IP)
		}
	}

	healthMap := make(map[string]model.HostHealth)
	if len(ips) > 0 {
		healths, err := e.hostHealthRepo.FindLatestByIPs(ctx, ips)
		if err != nil {
			zap.L().Error("告警引擎加载健康度失败", zap.Error(err))
		} else {
			for _, h := range healths {
				healthMap[h.IP] = h
			}
		}
	}

	// 遍历每条规则进行检查
	for _, rule := range rules {
		if rule.RuleType == "metric" {
			e.checkMetricRule(ctx, rule, hosts, hostMap, healthMap)
		} else if rule.RuleType == "service" {
			if rule.MetricType == "agent_offline" {
				e.checkAgentOfflineRule(ctx, rule)
			} else {
				e.checkServiceRule(ctx, rule, hosts, hostMap)
			}
		}
	}
}

// checkMetricRule 检查指标类告警规则
func (e *AlarmEngine) checkMetricRule(ctx context.Context, rule model.AlarmRule, hosts []model.HostInfo, hostMap map[int64]model.HostInfo, healthMap map[string]model.HostHealth) {
	// 确定需要检查的主机列表
	targetHosts := e.getTargetHosts(rule, hosts)

	for _, host := range targetHosts {
		health, ok := healthMap[host.IP]
		if !ok {
			continue
		}

		// 获取当前指标值
		currentValue := e.getMetricValue(rule.MetricType, health)
		if currentValue < 0 {
			continue // 不支持的指标类型
		}

		// 判断是否超过阈值
		isBreached := e.compare(currentValue, rule.Threshold, rule.CompareOp)
		key := fmt.Sprintf("%d:%d", rule.ID, host.ID)

		if isBreached {
			// 检查是否已存在未恢复的告警
			existing, err := e.alarmRecordRepo.FindTriggeredByRuleAndHost(ctx, rule.ID, host.ID)
			if err == nil && existing != nil {
				// 已有未恢复告警，更新当前值
				e.alarmRecordRepo.Update(ctx, existing.ID, map[string]interface{}{
					"current_value": currentValue,
				})
				continue
			}

			// 记录突破开始时间
			breachStart, exists := e.breachStartMap.Load(key)
			if !exists {
				e.breachStartMap.Store(key, time.Now())
				continue
			}

			// 检查持续时间
			startTime := breachStart.(time.Time)
			if time.Since(startTime) < time.Duration(rule.Duration)*time.Second {
				continue // 持续时间未满足
			}

			// 触发告警
			content := e.buildAlarmContent(rule.Name, host.IP, host.Name, rule.MetricType, currentValue, rule.Threshold)
			record := &model.AlarmRecord{
				RuleID:       rule.ID,
				RuleName:     rule.Name,
				HostID:       host.ID,
				HostIP:       host.IP,
				HostName:     host.Name,
				AlarmType:    rule.RuleType,
				MetricType:   rule.MetricType,
				CurrentValue: currentValue,
				Threshold:    rule.Threshold,
				Severity:     rule.Severity,
				Content:      content,
				Status:       "triggered",
				TriggerTime:  utils.NowStr(),
				NotifyStatus: "pending",
			}

			if err := e.alarmRecordRepo.Create(ctx, record); err != nil {
				zap.L().Error("创建告警记录失败", zap.Error(err), zap.Int64("ruleId", rule.ID))
			} else {
				zap.L().Info("告警触发",
					zap.String("rule", rule.Name),
					zap.String("host", host.IP),
					zap.Float64("value", currentValue),
					zap.Float64("threshold", rule.Threshold))
				// 发送通知
				e.sendNotification(ctx, record)
			}

			// 清除突破记录
			e.breachStartMap.Delete(key)
		} else {
			// 未超过阈值，清除突破记录
			e.breachStartMap.Delete(key)

			// 检查是否有未恢复的告警，标记为已恢复
			existing, err := e.alarmRecordRepo.FindTriggeredByRuleAndHost(ctx, rule.ID, host.ID)
			if err == nil && existing != nil {
				if err := e.alarmRecordRepo.Update(ctx, existing.ID, map[string]interface{}{
					"status":       "resolved",
					"resolve_time": utils.NowStr(),
				}); err != nil {
					zap.L().Error("更新告警恢复状态失败", zap.Error(err))
				} else {
					zap.L().Info("告警恢复",
						zap.String("rule", rule.Name),
						zap.String("host", host.IP))
				}
			}
		}
	}
}

// checkServiceRule 检查服务类告警规则
func (e *AlarmEngine) checkServiceRule(ctx context.Context, rule model.AlarmRule, hosts []model.HostInfo, hostMap map[int64]model.HostInfo) {
	targetHosts := e.getTargetHosts(rule, hosts)

	for _, host := range targetHosts {
		var isBreached bool
		var currentValue float64

		switch rule.MetricType {
		case "host_offline":
			isBreached = host.IsOnline == nil || !*host.IsOnline
			if isBreached {
				currentValue = 0
			} else {
				currentValue = 1
			}
		default:
			continue
		}

		key := fmt.Sprintf("%d:%d", rule.ID, host.ID)

		if isBreached {
			// 检查是否已存在未恢复的告警
			existing, err := e.alarmRecordRepo.FindTriggeredByRuleAndHost(ctx, rule.ID, host.ID)
			if err == nil && existing != nil {
				continue
			}

			// 记录突破开始时间
			breachStart, exists := e.breachStartMap.Load(key)
			if !exists {
				e.breachStartMap.Store(key, time.Now())
				continue
			}

			// 检查持续时间
			startTime := breachStart.(time.Time)
			if time.Since(startTime) < time.Duration(rule.Duration)*time.Second {
				continue
			}

			// 触发告警
			content := e.buildAlarmContent(rule.Name, host.IP, host.Name, rule.MetricType, currentValue, rule.Threshold)
			record := &model.AlarmRecord{
				RuleID:       rule.ID,
				RuleName:     rule.Name,
				HostID:       host.ID,
				HostIP:       host.IP,
				HostName:     host.Name,
				AlarmType:    rule.RuleType,
				MetricType:   rule.MetricType,
				CurrentValue: currentValue,
				Threshold:    rule.Threshold,
				Severity:     rule.Severity,
				Content:      content,
				Status:       "triggered",
				TriggerTime:  utils.NowStr(),
				NotifyStatus: "pending",
			}

			if err := e.alarmRecordRepo.Create(ctx, record); err != nil {
				zap.L().Error("创建服务告警记录失败", zap.Error(err), zap.Int64("ruleId", rule.ID))
			} else {
				zap.L().Info("服务告警触发",
					zap.String("rule", rule.Name),
					zap.String("host", host.IP),
					zap.String("type", rule.MetricType))
				e.sendNotification(ctx, record)
			}

			e.breachStartMap.Delete(key)
		} else {
			e.breachStartMap.Delete(key)

			// 检查是否有未恢复的告警
			existing, err := e.alarmRecordRepo.FindTriggeredByRuleAndHost(ctx, rule.ID, host.ID)
			if err == nil && existing != nil {
				if err := e.alarmRecordRepo.Update(ctx, existing.ID, map[string]interface{}{
					"status":       "resolved",
					"resolve_time": utils.NowStr(),
				}); err != nil {
					zap.L().Error("更新服务告警恢复状态失败", zap.Error(err))
				} else {
					zap.L().Info("服务告警恢复",
						zap.String("rule", rule.Name),
						zap.String("host", host.IP))
				}
			}
		}
	}
}

// checkAgentOfflineRule 检查控制Agent离线规则（通过HTTP探测）
func (e *AlarmEngine) checkAgentOfflineRule(ctx context.Context, rule model.AlarmRule) {
	// HTTP探测控制Agent接口（host_id=0 表示全局检测）
	isOnline := e.probeAgentOnline()
	isBreached := !isOnline
	var currentValue float64
	if isBreached {
		currentValue = 0
	} else {
		currentValue = 1
	}

	key := fmt.Sprintf("%d:agent", rule.ID)

	if isBreached {
		// 检查是否已存在未恢复的告警
		existing, err := e.alarmRecordRepo.FindTriggeredByRuleAndHost(ctx, rule.ID, 0)
		if err == nil && existing != nil {
			return
		}

		// 记录突破开始时间
		breachStart, exists := e.breachStartMap.Load(key)
		if !exists {
			e.breachStartMap.Store(key, time.Now())
			return
		}

		// 检查持续时间
		startTime := breachStart.(time.Time)
		if time.Since(startTime) < time.Duration(rule.Duration)*time.Second {
			return
		}

		// 触发告警（全局告警，不关联特定主机）
		content := e.buildAlarmContent(rule.Name, "控制Agent", "控制Agent", rule.MetricType, currentValue, rule.Threshold)
		record := &model.AlarmRecord{
			RuleID:       rule.ID,
			RuleName:     rule.Name,
			HostID:       0,
			HostIP:       "控制Agent",
			HostName:     "控制Agent",
			AlarmType:    rule.RuleType,
			MetricType:   rule.MetricType,
			CurrentValue: currentValue,
			Threshold:    rule.Threshold,
			Severity:     rule.Severity,
			Content:      content,
			Status:       "triggered",
			TriggerTime:  utils.NowStr(),
			NotifyStatus: "pending",
		}

		if err := e.alarmRecordRepo.Create(ctx, record); err != nil {
			zap.L().Error("创建控制Agent离线告警记录失败", zap.Error(err), zap.Int64("ruleId", rule.ID))
		} else {
			zap.L().Info("控制Agent离线告警触发", zap.String("rule", rule.Name))
			e.sendNotification(ctx, record)
		}

		e.breachStartMap.Delete(key)
	} else {
		e.breachStartMap.Delete(key)

		// 检查是否有未恢复的告警，标记为已恢复
		existing, err := e.alarmRecordRepo.FindTriggeredByRuleAndHost(ctx, rule.ID, 0)
		if err == nil && existing != nil {
			if err := e.alarmRecordRepo.Update(ctx, existing.ID, map[string]interface{}{
				"status":       "resolved",
				"resolve_time": utils.NowStr(),
			}); err != nil {
				zap.L().Error("更新控制Agent告警恢复状态失败", zap.Error(err))
			} else {
				zap.L().Info("控制Agent离线告警恢复", zap.String("rule", rule.Name))
			}
		}
	}
}

// probeAgentOnline 探测控制Agent HTTP接口是否可达
func (e *AlarmEngine) probeAgentOnline() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://47.110.91.222:8010")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode < 500
}

// getTargetHosts 获取规则需要检查的主机列表
func (e *AlarmEngine) getTargetHosts(rule model.AlarmRule, allHosts []model.HostInfo) []model.HostInfo {
	if rule.HostID > 0 {
		for _, h := range allHosts {
			if h.ID == rule.HostID {
				return []model.HostInfo{h}
			}
		}
		return []model.HostInfo{}
	}
	return allHosts
}

// getMetricValue 获取指标值
func (e *AlarmEngine) getMetricValue(metricType string, health model.HostHealth) float64 {
	switch metricType {
	case "cpu":
		return health.CpuUsage
	case "mem":
		return health.MemUsage
	case "disk":
		return health.DiskUsage
	case "network":
		// 使用网络入流量作为指标
		return health.NetInKbps
	case "load":
		// 解析负载字符串的第一个值
		return e.parseLoadAvg(health.LoadAvg)
	default:
		return -1
	}
}

// parseLoadAvg 解析负载平均值
func (e *AlarmEngine) parseLoadAvg(loadAvg string) float64 {
	if loadAvg == "" {
		return 0
	}
	parts := strings.Split(loadAvg, ",")
	if len(parts) == 0 {
		return 0
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0
	}
	return val
}

// compare 比较值
func (e *AlarmEngine) compare(current, threshold float64, op string) bool {
	switch op {
	case "gt":
		return current > threshold
	case "lt":
		return current < threshold
	case "ge":
		return current >= threshold
	case "le":
		return current <= threshold
	case "eq":
		return current == threshold
	default:
		return current > threshold
	}
}

// buildAlarmContent 构建告警内容
func (e *AlarmEngine) buildAlarmContent(ruleName, hostIP, hostName, metricType string, currentValue, threshold float64) string {
	metricName := map[string]string{
		"cpu":           "CPU使用率",
		"mem":           "内存使用率",
		"disk":          "磁盘使用率",
		"network":       "网络入流量",
		"load":          "系统负载",
		"host_offline":  "主机离线",
		"agent_offline": "Agent离线",
	}
	name, ok := metricName[metricType]
	if !ok {
		name = metricType
	}

	return fmt.Sprintf("告警规则：%s\n主机：%s(%s)\n指标：%s\n当前值：%.2f\n阈值：%.2f",
		ruleName, hostIP, hostName, name, currentValue, threshold)
}

// sendNotification 发送告警通知
func (e *AlarmEngine) sendNotification(ctx context.Context, record *model.AlarmRecord) {
	// 获取规则对应的通道
	rule, err := e.alarmRuleRepo.FindByID(ctx, record.RuleID)
	if err != nil || rule.ChannelIDs == "" {
		// 没有配置通道，仅记录平台内部通知
		e.sendPlatformNotification(ctx, record)
		if err := e.alarmRecordRepo.Update(ctx, record.ID, map[string]interface{}{
			"notify_status": constant.NotifyStatusSuccess,
		}); err != nil {
			zap.L().Error("更新告警通知状态失败", zap.Error(err))
		}
		return
	}

	// 解析通道ID列表
	channelIDs := e.parseChannelIDs(rule.ChannelIDs)
	if len(channelIDs) == 0 {
		e.sendPlatformNotification(ctx, record)
		if err := e.alarmRecordRepo.Update(ctx, record.ID, map[string]interface{}{
			"notify_status": constant.NotifyStatusSuccess,
		}); err != nil {
			zap.L().Error("更新告警通知状态失败", zap.Error(err))
		}
		return
	}

	// 获取所有通道信息
	channels, err := e.alarmChannelRepo.FindAll(ctx)
	if err != nil {
		zap.L().Error("获取告警通道失败", zap.Error(err))
		e.sendPlatformNotification(ctx, record)
		if err := e.alarmRecordRepo.Update(ctx, record.ID, map[string]interface{}{
			"notify_status": constant.NotifyStatusSuccess,
		}); err != nil {
			zap.L().Error("更新告警通知状态失败", zap.Error(err))
		}
		return
	}

	// 构建通道映射
	channelMap := make(map[int64]model.AlarmChannel)
	for _, ch := range channels {
		channelMap[ch.ID] = ch
	}

	successCount, failCount := 0, 0
	// 遍历通道发送通知
	for _, cid := range channelIDs {
		ch, ok := channelMap[cid]
		if !ok || ch.Status != "0" {
			continue
		}

		var sendOk bool
		switch ch.ChannelType {
		case constant.ChannelTypePlatform:
			sendOk, _ = e.sendPlatformNotification(ctx, record)
		case constant.ChannelTypeEmail:
			sendOk, _ = e.sendEmailNotification(ctx, record, ch)
		case constant.ChannelTypeDingTalk:
			sendOk, _ = e.sendDingTalkNotification(ctx, record, ch)
		case constant.ChannelTypeWechat:
			sendOk, _ = e.sendWechatNotification(ctx, record, ch)
		}
		if sendOk {
			successCount++
		} else {
			failCount++
		}
	}

	if successCount > 0 {
		if err := e.alarmRecordRepo.Update(ctx, record.ID, map[string]interface{}{
			"notify_status": constant.NotifyStatusSuccess,
		}); err != nil {
			zap.L().Error("更新告警通知状态失败", zap.Error(err))
		}
	} else if failCount > 0 {
		if err := e.alarmRecordRepo.Update(ctx, record.ID, map[string]interface{}{
			"notify_status": constant.NotifyStatusFailed,
		}); err != nil {
			zap.L().Error("更新告警通知状态失败", zap.Error(err))
		}
	}
}

// parseChannelIDs 解析通道ID列表
func (e *AlarmEngine) parseChannelIDs(channelIDsStr string) []int64 {
	var ids []int64
	channelIDsStr = strings.Trim(channelIDsStr, "[]")
	if channelIDsStr == "" {
		return ids
	}

	parts := strings.Split(channelIDsStr, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		id, err := strconv.ParseInt(p, 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// sendPlatformNotification 平台内部通知
func (e *AlarmEngine) sendPlatformNotification(ctx context.Context, record *model.AlarmRecord) (bool, string) {
	// 记录通知日志
	log := &model.AlarmNotifyLog{
		RecordID:    record.ID,
		ChannelID:   0,
		ChannelName: "平台内部",
		ChannelType: constant.ChannelTypePlatform,
		Status:      constant.NotifyStatusSuccess,
		Response:    "平台内部通知",
		SendTime:    utils.NowStr(),
		CreateTime:  utils.NowStr(),
	}
	if err := e.alarmNotifyLogRepo.Create(ctx, log); err != nil {
		zap.L().Error("创建平台通知日志失败", zap.Error(err))
	}
	return true, "平台内部通知"
}

// writeNotifyLog 统一写入通知日志
func (e *AlarmEngine) writeNotifyLog(ctx context.Context, record *model.AlarmRecord, ch model.AlarmChannel, ok bool, resp string) (bool, string) {
	status := constant.NotifyStatusFailed
	if ok {
		status = constant.NotifyStatusSuccess
	}
	log := &model.AlarmNotifyLog{
		RecordID:    record.ID,
		ChannelID:   ch.ID,
		ChannelName: ch.Name,
		ChannelType: ch.ChannelType,
		Status:      status,
		Response:    resp,
		SendTime:    utils.NowStr(),
		CreateTime:  utils.NowStr(),
	}
	if err := e.alarmNotifyLogRepo.Create(ctx, log); err != nil {
		zap.L().Error("创建通知日志失败", zap.Error(err))
	}
	return ok, resp
}
