package model

import "time"

// BizDeviceAudio maps to biz_device_audio.
type BizDeviceAudio struct {
	ID         string    `gorm:"column:id"`
	MacID      string    `gorm:"column:mac_id"`
	AudioBlob  []byte    `gorm:"column:audio_blob"`
	AudioLen   string    `gorm:"column:audio_len"`
	Context    string    `gorm:"column:context"`
	CreateTime time.Time `gorm:"column:create_time"`
	ExpireTime  time.Time `gorm:"column:expire_time"`
	DeleteTime  time.Time `gorm:"column:delete_time"`
	SpeakerID   *string   `gorm:"column:speaker_id"`
	BoundUserID *string   `gorm:"column:bound_user_id"`
}

func (BizDeviceAudio) TableName() string { return "biz_device_audio" }
