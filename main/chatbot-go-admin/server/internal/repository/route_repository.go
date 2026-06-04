package repository

import (
	"context"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
)

type RouteRepository struct {
	db *gorm.DB
}

func NewRouteRepository(db *gorm.DB) *RouteRepository {
	return &RouteRepository{db: db}
}

func (r *RouteRepository) ListAll(ctx context.Context) ([]model.SysRoute, error) {
	var rows []model.SysRoute
	err := r.db.WithContext(ctx).
		Where("deleted = 0").
		Order("parent_id ASC, order_no ASC, id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *RouteRepository) ListEnabled(ctx context.Context) ([]model.SysRoute, error) {
	var rows []model.SysRoute
	err := r.db.WithContext(ctx).
		Where("deleted = 0 AND status = 1").
		Order("parent_id ASC, order_no ASC, id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *RouteRepository) FindByID(ctx context.Context, id string) (*model.SysRoute, error) {
	var row model.SysRoute
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted = 0", id).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RouteRepository) FindByName(ctx context.Context, name string) (*model.SysRoute, error) {
	var row model.SysRoute
	err := r.db.WithContext(ctx).
		Where("name = ? AND deleted = 0", name).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *RouteRepository) Create(ctx context.Context, row *model.SysRoute) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *RouteRepository) UpdateByID(ctx context.Context, id string, updates map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&model.SysRoute{}).
		Where("id = ? AND deleted = 0", id).
		Updates(updates).Error
}

func (r *RouteRepository) CountChildren(ctx context.Context, parentID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.SysRoute{}).
		Where("parent_id = ? AND deleted = 0", parentID).
		Count(&n).Error
	return n, err
}

func (r *RouteRepository) SoftDelete(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).
		Model(&model.SysRoute{}).
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
