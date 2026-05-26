package service

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"rtm-server/internal/model"
	"rtm-server/internal/repository"
	"rtm-server/internal/utils"
)

// AlarmNotifier 告警通知服务
type AlarmNotifier struct {
	db                 *gorm.DB
	alarmRecordRepo    *repository.AlarmRecordRepository
	alarmChannelRepo   *repository.AlarmChannelRepository
	alarmNotifyLogRepo *repository.AlarmNotifyLogRepository
}

// NewAlarmNotifier 创建告警通知服务
func NewAlarmNotifier(db *gorm.DB) *AlarmNotifier {
	return &AlarmNotifier{
		db:                 db,
		alarmRecordRepo:    repository.NewAlarmRecordRepository(db),
		alarmChannelRepo:   repository.NewAlarmChannelRepository(db),
		alarmNotifyLogRepo: repository.NewAlarmNotifyLogRepository(db),
	}
}

// SendPlatformNotify 发送平台内部通知
func (n *AlarmNotifier) SendPlatformNotify(ctx context.Context, recordID int64) error {
	if err := n.alarmRecordRepo.Update(ctx, recordID, map[string]interface{}{
		"notify_status": "success",
	}); err != nil {
		zap.L().Error("更新平台通知状态失败", zap.Error(err))
		return err
	}

	log := &model.AlarmNotifyLog{
		RecordID:    recordID,
		ChannelID:   0,
		ChannelName: "平台内部",
		ChannelType: "platform",
		Status:      "success",
		Response:    "平台内部通知",
		SendTime:    utils.NowStr(),
		CreateTime:  utils.NowStr(),
	}
	return n.alarmNotifyLogRepo.Create(ctx, log)
}

// RetryNotify 重试通知（用于平台通知失败时重试）
func (n *AlarmNotifier) RetryNotify(ctx context.Context, recordID int64) error {
	record, err := n.alarmRecordRepo.FindByID(ctx, recordID)
	if err != nil {
		return err
	}

	if record.NotifyStatus == "success" {
		return nil // 已成功，无需重试
	}

	return n.SendPlatformNotify(ctx, recordID)
}
