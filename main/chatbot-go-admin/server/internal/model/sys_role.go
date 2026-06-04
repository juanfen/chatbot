package model

import "time"

type SysRole struct {
	ID         string    `gorm:"column:id;primaryKey"`
	Code       string    `gorm:"column:code"`
	Name       string    `gorm:"column:name"`
	Status     int       `gorm:"column:status"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
	Deleted    int       `gorm:"column:deleted"`
}

func (SysRole) TableName() string { return "sys_role" }

