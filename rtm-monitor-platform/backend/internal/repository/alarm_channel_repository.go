package repository

import (
	"context"

	"gorm.io/gorm"

	"rtm-server/internal/dto"
	"rtm-server/internal/model"
)

// AlarmChannelRepository 告警通道数据访问
type AlarmChannelRepository struct {
	db *gorm.DB
}

// NewAlarmChannelRepository 创建告警通道仓库
func NewAlarmChannelRepository(db *gorm.DB) *AlarmChannelRepository {
	return &AlarmChannelRepository{db: db}
}

// FindAll 查询所有通道
func (r *AlarmChannelRepository) FindAll(ctx context.Context) ([]model.AlarmChannel, error) {
	var channels []model.AlarmChannel
	err := r.db.WithContext(ctx).Where("del_flag = ?", "0").Order("id DESC").Find(&channels).Error
	return channels, err
}

// FindByID 根据ID查询
func (r *AlarmChannelRepository) FindByID(ctx context.Context, id int64) (*model.AlarmChannel, error) {
	var ch model.AlarmChannel
	err := r.db.WithContext(ctx).Where("del_flag = ?", "0").First(&ch, id).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// Count 带条件统计
func (r *AlarmChannelRepository) Count(ctx context.Context, query *dto.AlarmChannelQuery) (int64, error) {
	db := r.db.WithContext(ctx).Model(&model.AlarmChannel{}).Where("del_flag = ?", "0")

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.ChannelType != "" {
		db = db.Where("channel_type = ?", query.ChannelType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	err := db.Count(&total).Error
	return total, err
}

// FindPage 分页查询
func (r *AlarmChannelRepository) FindPage(ctx context.Context, query *dto.AlarmChannelQuery) ([]model.AlarmChannel, error) {
	db := r.db.WithContext(ctx).Model(&model.AlarmChannel{}).Where("del_flag = ?", "0")

	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}
	if query.ChannelType != "" {
		db = db.Where("channel_type = ?", query.ChannelType)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var channels []model.AlarmChannel
	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Order("id DESC").Find(&channels).Error
	return channels, err
}

// Create 创建
func (r *AlarmChannelRepository) Create(ctx context.Context, channel *model.AlarmChannel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

// Update 更新
func (r *AlarmChannelRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.AlarmChannel{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteByIDs 批量删除（逻辑删除）
func (r *AlarmChannelRepository) DeleteByIDs(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Model(&model.AlarmChannel{}).Where("id IN ?", ids).Update("del_flag", "2").Error
}

// ClearDefault 清除所有默认通道
func (r *AlarmChannelRepository) ClearDefault(ctx context.Context, channelType string) error {
	return r.db.WithContext(ctx).Model(&model.AlarmChannel{}).Where("channel_type = ?", channelType).Update("is_default", "0").Error
}
