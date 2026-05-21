package service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	db                *gorm.DB
	hostInfoService   *HostInfoService
	hostHealthService *HostHealthService
	alarmEngine       *AlarmEngine
	tasks             map[string]*scheduledTask
	stopCh            chan struct{}
}

// scheduledTask 定时任务
type scheduledTask struct {
	name     string
	interval time.Duration
	fn       func()
	ticker   *time.Ticker
}

// NewScheduler 创建定时任务调度器
func NewScheduler(db *gorm.DB, hostInfoService *HostInfoService, hostHealthService *HostHealthService) *Scheduler {
	return &Scheduler{
		db:                db,
		hostInfoService:   hostInfoService,
		hostHealthService: hostHealthService,
		alarmEngine:       NewAlarmEngine(db),
		tasks:             make(map[string]*scheduledTask),
		stopCh:            make(chan struct{}),
	}
}

// Start 启动所有定时任务
func (s *Scheduler) Start() {
	zap.L().Info("定时任务调度器启动")

	// 主机离线检测任务 - 每60秒执行一次（基于 3 秒心跳，5个周期=15秒超时）
	// 使用单条SQL原子操作，避免读取-更新竞态
	s.registerTask("dealOffline", 60*time.Second, s.dealOfflineTask)

	// 告警检查任务 - 每30秒执行一次
	s.registerTask("alarmCheck", 30*time.Second, s.alarmCheckTask)

	// 主机健康度数据清理任务 - 每天凌晨3点30分执行（保留7天）
	s.registerCronTask("cleanHostHealth", "30 3 * * *", s.cleanHostHealthTask)
}

// registerTask 注册固定间隔任务
func (s *Scheduler) registerTask(name string, interval time.Duration, fn func()) {
	task := &scheduledTask{
		name:     name,
		interval: interval,
		fn:       fn,
		ticker:   time.NewTicker(interval),
	}
	s.tasks[name] = task

	go func() {
		// 立即执行一次
		fn()
		for {
			select {
			case <-task.ticker.C:
				zap.L().Debug("定时任务执行", zap.String("task", name))
				fn()
			case <-s.stopCh:
				return
			}
		}
	}()

	zap.L().Info("定时任务已注册",
		zap.String("task", name),
		zap.Duration("interval", interval))
}

// registerCronTask 注册cron风格任务（简化实现，基于每日检查）
func (s *Scheduler) registerCronTask(name string, cronExpr string, fn func()) {
	// 解析cron表达式 (简化: 只支持 "分 时 * * *" 格式)
	var minute, hour int
	fmt.Sscanf(cronExpr, "%d %d * * *", &minute, &hour)

	// 计算下次执行时间
	nextRun := s.nextRunTime(hour, minute)
	initialDelay := time.Until(nextRun)

	go func() {
		// 等待到首次执行时间
		select {
		case <-time.After(initialDelay):
			fn()
		case <-s.stopCh:
			return
		}

		// 之后每24小时执行一次
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				zap.L().Debug("定时任务执行", zap.String("task", name))
				fn()
			case <-s.stopCh:
				return
			}
		}
	}()

	zap.L().Info("定时任务已注册",
		zap.String("task", name),
		zap.String("cron", cronExpr),
		zap.Time("nextRun", nextRun))
}

// nextRunTime 计算下次执行时间
func (s *Scheduler) nextRunTime(hour, minute int) time.Time {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if next.Before(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// Stop 停止所有定时任务
func (s *Scheduler) Stop() {
	zap.L().Info("定时任务调度器停止")
	close(s.stopCh)
	for _, task := range s.tasks {
		if task.ticker != nil {
			task.ticker.Stop()
		}
	}
}

// ========== 定时任务实现 ==========

// dealOfflineTask 主机离线检测任务
// 核心策略：使用单条SQL原子操作，WHERE 条件同时检查 is_online=true 和 last_heartbeat < 阈值
// 超时阈值 = heartbeat_interval * 5 秒，按各主机自身的心跳间隔在 SQL 中计算
// 兜底：heartbeat_interval 最小值为 30 秒，确保最小超时阈值为 150 秒
func (s *Scheduler) dealOfflineTask() {
	ctx := context.Background()

	rowsAffected, err := s.hostInfoService.MarkOfflineByTimeout(ctx)
	if err != nil {
		zap.L().Error("主机离线检测失败", zap.Error(err))
		return
	}

	if rowsAffected > 0 {
		zap.L().Warn("主机离线检测完成，有主机被标记为离线",
			zap.Int64("offlineCount", rowsAffected))
	} else {
		zap.L().Debug("主机离线检测完成，无主机离线")
	}
}

// cleanHostHealthTask 主机健康度数据清理任务（保留7天）
func (s *Scheduler) cleanHostHealthTask() {
	zap.L().Info("执行主机健康度数据清理任务")

	ctx := context.Background()
	deleted, err := s.hostHealthService.CleanOldHealthData(ctx, 7)
	if err != nil {
		zap.L().Error("清理主机健康度数据失败", zap.Error(err))
		return
	}

	zap.L().Info("主机健康度数据清理完成", zap.Int64("deleted", deleted))
}

// alarmCheckTask 告警检查任务
func (s *Scheduler) alarmCheckTask() {
	if s.alarmEngine == nil {
		return
	}
	s.alarmEngine.Check()
}
