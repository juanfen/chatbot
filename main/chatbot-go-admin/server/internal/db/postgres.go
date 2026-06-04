package db

import (
	"fmt"
	"time"

	"chatbot-go-admin/server/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenPostgres opens a PostgreSQL connection using GORM.
func OpenPostgres(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	gormCfg := &gorm.Config{
		PrepareStmt: false,
		Logger:      logger.Default.LogMode(logger.Silent),
	}
	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// 部分 PG 实例（尤其远程自建库）默认 search_path 为空，会导致
	// CREATE TABLE 报 SQLSTATE 3F000: no schema has been selected to create in
	if err := db.Exec(`SET search_path TO public`).Error; err != nil {
		return nil, fmt.Errorf("set search_path to public failed: %w", err)
	}

	return db, nil
}

