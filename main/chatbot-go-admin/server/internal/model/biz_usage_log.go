package model

import "time"

type BizUsageLog struct {
	ID         string    `gorm:"column:id;primaryKey"`
	UserID     string    `gorm:"column:user_id"`
	Channel    string    `gorm:"column:channel"`
	RefText    string    `gorm:"column:ref_text"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (BizUsageLog) TableName() string { return "biz_usage_log" }
