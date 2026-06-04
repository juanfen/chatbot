package model

import "time"

// BizUserPreference maps to biz_user_preference.
type BizUserPreference struct {
	UserID       string    `gorm:"column:user_id"`
	StandardText string    `gorm:"column:standard_text"`
	CreateTime   time.Time `gorm:"column:create_time"`
	UpdateTime   time.Time `gorm:"column:update_time"`
}

func (BizUserPreference) TableName() string { return "biz_user_preference" }
