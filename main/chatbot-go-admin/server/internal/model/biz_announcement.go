package model

import "time"

type BizAnnouncement struct {
	ID          string     `gorm:"column:id;primaryKey"`
	Title       string     `gorm:"column:title"`
	Content     string     `gorm:"column:content"`
	SortNo      int        `gorm:"column:sort_no"`
	Status      int        `gorm:"column:status"`
	PublishTime *time.Time `gorm:"column:publish_time"`
	CreateTime  time.Time  `gorm:"column:create_time"`
	UpdateTime  time.Time  `gorm:"column:update_time"`
	Deleted     int        `gorm:"column:deleted"`
}

func (BizAnnouncement) TableName() string { return "biz_announcement" }
