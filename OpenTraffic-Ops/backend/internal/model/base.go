package model

import (
	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/utils"
)

// BaseEntity 基础实体（嵌入到各模型中）
// 注意：时间字段使用 string 类型，避免 pgx 驱动对 time.Time 做 UTC 转换
type BaseEntity struct {
	CreateBy   string `json:"createBy" gorm:"column:create_by;size:64;comment:创建者"`
	CreateTime string `json:"createTime" gorm:"column:create_time;type:timestamp;comment:创建时间"`
	UpdateBy   string `json:"updateBy" gorm:"column:update_by;size:64;comment:更新者"`
	UpdateTime string `json:"updateTime" gorm:"column:update_time;type:timestamp;comment:更新时间"`
	Remark     string `json:"remark" gorm:"column:remark;size:500;comment:备注"`
	DelFlag    string `json:"delFlag" gorm:"column:del_flag;size:1;default:0;comment:删除标志（0代表存在 2代表删除）"`
}

// BeforeCreate 创建前回调
func (b *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	now := utils.NowStr()
	b.CreateTime = now
	b.UpdateTime = now
	b.DelFlag = "0"
	return nil
}

// BeforeUpdate 更新前回调
func (b *BaseEntity) BeforeUpdate(tx *gorm.DB) error {
	b.UpdateTime = utils.NowStr()
	return nil
}
