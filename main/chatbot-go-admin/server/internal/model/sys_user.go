package model

import "time"

type SysUser struct {
	ID           string    `gorm:"column:id;primaryKey"`
	Username     string    `gorm:"column:username"`
	PasswordHash string    `gorm:"column:password_hash"`
	Nickname     string    `gorm:"column:nickname"`
	Status       int       `gorm:"column:status"`
	IsAdmin      int       `gorm:"column:is_admin"`
	CreateTime   time.Time `gorm:"column:create_time"`
	UpdateTime   time.Time `gorm:"column:update_time"`
	Deleted      int       `gorm:"column:deleted"`
}

func (SysUser) TableName() string { return "sys_user" }

