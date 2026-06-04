package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

func (r *UserRoleRepository) ListRoleIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	var rows []model.SysUserRole
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted = 0", userID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.RoleID)
	}
	return ids, nil
}

func (r *UserRoleRepository) ReplaceUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SysUserRole{}).
			Where("user_id = ? AND deleted = 0", userID).
			Update("deleted", 1).Error; err != nil {
			return err
		}

		for _, rid := range roleIDs {
			ur := model.SysUserRole{
				ID:         uuid.NewString(),
				UserID:     userID,
				RoleID:     rid,
				CreateTime: now,
				UpdateTime: now,
				Deleted:    0,
			}
			if err := tx.Create(&ur).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRoleRepository) CountUsersByRoleID(ctx context.Context, roleID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SysUserRole{}).
		Where("role_id = ? AND deleted = 0", roleID).
		Count(&n).Error
	return n, err
}

func (r *UserRoleRepository) SoftDeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&model.SysUserRole{}).
		Where("user_id = ? AND deleted = 0", userID).
		Update("deleted", 1).Error
}

