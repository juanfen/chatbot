package model

import "time"

type SysRoleRoute struct {
	ID         string    `gorm:"column:id;primaryKey"`
	RoleID     string    `gorm:"column:role_id"`
	RouteID    string    `gorm:"column:route_id"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
	Deleted    int       `gorm:"column:deleted"`
}

func (SysRoleRoute) TableName() string { return "sys_role_route" }
