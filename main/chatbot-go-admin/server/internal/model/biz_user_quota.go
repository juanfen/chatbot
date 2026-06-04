package model

import "time"

type BizUserQuota struct {
	UserID        string    `gorm:"column:user_id;primaryKey"`
	PackageID     string    `gorm:"column:package_id"`
	PackageTitle  string    `gorm:"column:package_title"`
	TotalQuota    int64     `gorm:"column:total_quota"`
	UsedQuota     int64     `gorm:"column:used_quota"`
	UpdateTime    time.Time `gorm:"column:update_time"`
}

func (BizUserQuota) TableName() string { return "biz_user_quota" }
