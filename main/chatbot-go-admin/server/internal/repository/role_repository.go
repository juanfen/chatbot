package repository

import (
	"context"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) List(ctx context.Context) ([]model.SysRole, error) {
	var roles []model.SysRole
	err := r.db.WithContext(ctx).
		Where("deleted = 0").
		Order("create_time DESC").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) Create(ctx context.Context, role *model.SysRole) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *RoleRepository) UpdateByID(ctx context.Context, id string, updates map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.SysRole{}).
		Where("id = ? AND deleted = 0", id).
		Updates(updates).Error
}

func (r *RoleRepository) FindByIDs(ctx context.Context, ids []string) ([]model.SysRole, error) {
	if len(ids) == 0 {
		return []model.SysRole{}, nil
	}
	var roles []model.SysRole
	err := r.db.WithContext(ctx).
		Where("id IN ? AND deleted = 0", ids).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id string) (*model.SysRole, error) {
	var role model.SysRole
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted = 0", id).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).
		Model(&model.SysRole{}).
		Where("id = ? AND deleted = 0", id).
		Update("deleted", 1)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

