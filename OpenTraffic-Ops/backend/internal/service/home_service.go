package service

import (
	"context"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"rtm-server/internal/dto"
	"rtm-server/internal/model"
	"rtm-server/internal/repository"
)

// HomeService 首页统计服务
type HomeService struct {
	infoRepo         *repository.HostInfoRepository
	alarmChannelRepo *repository.AlarmChannelRepository
	alarmRuleRepo    *repository.AlarmRuleRepository
	alarmRecordRepo  *repository.AlarmRecordRepository
}

// NewHomeService 创建首页统计服务
func NewHomeService(db *gorm.DB) *HomeService {
	return &HomeService{
		infoRepo:         repository.NewHostInfoRepository(db),
		alarmChannelRepo: repository.NewAlarmChannelRepository(db),
		alarmRuleRepo:    repository.NewAlarmRuleRepository(db),
		alarmRecordRepo:  repository.NewAlarmRecordRepository(db),
	}
}

// GetStatisticData 获取首页统计数据
func (s *HomeService) GetStatisticData(ctx context.Context) (*dto.HomeStatisticData, error) {
	var (
		result       dto.HomeStatisticData
		recentRecords []model.AlarmRecord
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		v, err := s.infoRepo.CountTotal(gctx)
		result.HostCount = v
		return err
	})
	g.Go(func() error {
		v, err := s.infoRepo.CountOffline(gctx)
		result.OfflineHostCount = v
		return err
	})
	g.Go(func() error {
		v, err := s.alarmChannelRepo.Count(gctx, &dto.AlarmChannelQuery{})
		result.AlarmChannelCount = v
		return err
	})
	g.Go(func() error {
		v, err := s.alarmRuleRepo.Count(gctx, &dto.AlarmRuleQuery{})
		result.AlarmRuleCount = v
		return err
	})
	g.Go(func() error {
		v, err := s.alarmRecordRepo.CountUnread(gctx)
		result.UnhandledAlarmCount = v
		return err
	})
	g.Go(func() error {
		records, err := s.alarmRecordRepo.FindRecent(gctx, 5)
		if err != nil {
			return err
		}
		recentRecords = records
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	result.RecentAlarms = toAlarmRecordDtoList(recentRecords)
	return &result, nil
}
