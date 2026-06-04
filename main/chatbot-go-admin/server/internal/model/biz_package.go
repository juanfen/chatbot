package model

import "time"

type BizPackage struct {
	ID          string    `gorm:"column:id;primaryKey"`
	Code        string    `gorm:"column:code"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	QuotaCalls  int64     `gorm:"column:quota_calls"`
	SortNo      int       `gorm:"column:sort_no"`
	Status      int       `gorm:"column:status"`
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
	Deleted     int       `gorm:"column:deleted"`
}

func (BizPackage) TableName() string { return "biz_package" }
