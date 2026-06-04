package model

import "time"

type SysRoute struct {
	ID          string    `gorm:"column:id;primaryKey"`
	ParentID    string    `gorm:"column:parent_id"`
	Name        string    `gorm:"column:name"`
	Path        string    `gorm:"column:path"`
	Component   string    `gorm:"column:component"`
	Title       string    `gorm:"column:title"`
	I18nKey     string    `gorm:"column:i18n_key"`
	Icon        string    `gorm:"column:icon"`
	OrderNo     int       `gorm:"column:order_no"`
	HideInMenu  int       `gorm:"column:hide_in_menu"`
	KeepAlive   int       `gorm:"column:keep_alive"`
	Status      int       `gorm:"column:status"`
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
	Deleted     int       `gorm:"column:deleted"`
}

func (SysRoute) TableName() string { return "sys_route" }
