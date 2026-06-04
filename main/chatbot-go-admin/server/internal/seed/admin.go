package seed

import (
	"context"
	"errors"
	"time"

	"chatbot-go-admin/server/internal/model"
	"chatbot-go-admin/server/internal/util"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EnsureDefaultAdmin ensures there is at least one enabled admin user for first login.
func EnsureDefaultAdmin(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.SysUser{}).Where("deleted = 0 AND is_admin = 1").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := util.HashPassword("admin123")
	if err != nil {
		return err
	}

	now := time.Now()
	u := &model.SysUser{
		ID:           uuid.NewString(),
		Username:     "admin",
		PasswordHash: hash,
		Nickname:     "Administrator",
		Status:       1,
		IsAdmin:      1,
		CreateTime:   now,
		UpdateTime:   now,
		Deleted:      0,
	}

	var existing model.SysUser
	err = db.WithContext(ctx).Where("username = ? AND deleted = 0", "admin").First(&existing).Error
	if err == nil {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if err := db.WithContext(ctx).Create(u).Error; err != nil {
		return err
	}

	logger.Warn("seeded default admin user",
		zap.String("username", "admin"),
		zap.String("password", "admin123"),
		zap.String("hint", "please change password after first login"),
	)

	return nil
}
