package model

import "time"

// BizUserSpeaker maps to biz_user_speaker.
type BizUserSpeaker struct {
	SpeakerID     string    `gorm:"column:speaker_id;primaryKey"`
	UserID        string    `gorm:"column:user_id"`
	SpeakerName   string    `gorm:"column:speaker_name"`
	PersonalLines string    `gorm:"column:personal_lines"`
	CreateTime    time.Time `gorm:"column:create_time"`
	UpdateTime    time.Time `gorm:"column:update_time"`
}

func (BizUserSpeaker) TableName() string { return "biz_user_speaker" }
