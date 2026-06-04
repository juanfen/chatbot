package model

import "time"

type SysUserRole struct {
	ID         string    `gorm:"column:id;primaryKey"`
	UserID     string    `gorm:"column:user_id"`
	RoleID     string    `gorm:"column:role_id"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
	Deleted    int       `gorm:"column:deleted"`
}

func (SysUserRole) TableName() string { return "sys_user_role" }

