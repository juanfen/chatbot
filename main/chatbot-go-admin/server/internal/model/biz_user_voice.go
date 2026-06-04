package model

import "time"

// BizUserVoice maps to biz_user_voice.
type BizUserVoice struct {
	ID              string    `gorm:"column:id"`
	UserID          string    `gorm:"column:user_id"`
	FileName        string    `gorm:"column:file_name"`
	UploadTime      time.Time `gorm:"column:upload_time"`
	DurationSeconds float64   `gorm:"column:duration_seconds"`
	IsActive        int       `gorm:"column:is_active"`
	CreateTime      time.Time `gorm:"column:create_time"`
	UpdateTime      time.Time `gorm:"column:update_time"`
	Deleted         int       `gorm:"column:deleted"`
}

func (BizUserVoice) TableName() string { return "biz_user_voice" }
