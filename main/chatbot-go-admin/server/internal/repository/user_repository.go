package repository

import (
	"context"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.SysUser, error) {
	var u model.SysUser
	err := r.db.WithContext(ctx).
		Where("username = ? AND deleted = 0", username).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.SysUser, error) {
	var u model.SysUser
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted = 0", id).
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) List(ctx context.Context, keyword string, offset int, limit int) ([]model.SysUser, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.SysUser{}).Where("deleted = 0")
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("username ILIKE ? OR nickname ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.SysUser
	if err := q.Order("create_time DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepository) Create(ctx context.Context, u *model.SysUser) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepository) UpdateByID(ctx context.Context, id string, updates map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.SysUser{}).
		Where("id = ? AND deleted = 0", id).
		Updates(updates).Error
}

func (r *UserRepository) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).
		Model(&model.SysUser{}).
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

func (r *UserRepository) CountAdmins(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SysUser{}).
		Where("deleted = 0 AND is_admin = 1 AND status = 1").
		Count(&n).Error
	return n, err
}

