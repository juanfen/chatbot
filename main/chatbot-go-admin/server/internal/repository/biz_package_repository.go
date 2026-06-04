package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
)

type BizPackageRepository struct {
	db *gorm.DB
}

func NewBizPackageRepository(db *gorm.DB) *BizPackageRepository {
	return &BizPackageRepository{db: db}
}

func (r *BizPackageRepository) ListEnabled(ctx context.Context) ([]model.BizPackage, error) {
	var rows []model.BizPackage
	err := r.db.WithContext(ctx).
		Where("deleted = 0 AND status = 1").
		Order("sort_no ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *BizPackageRepository) ListAll(ctx context.Context) ([]model.BizPackage, error) {
	var rows []model.BizPackage
	err := r.db.WithContext(ctx).
		Where("deleted = 0").
		Order("sort_no ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *BizPackageRepository) FindByID(ctx context.Context, id string) (*model.BizPackage, error) {
	var row model.BizPackage
	err := r.db.WithContext(ctx).Where("id = ? AND deleted = 0", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BizPackageRepository) Create(ctx context.Context, row *model.BizPackage) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *BizPackageRepository) UpdateByID(ctx context.Context, id string, updates map[string]any) error {
	updates["update_time"] = time.Now()
	return r.db.WithContext(ctx).Model(&model.BizPackage{}).Where("id = ? AND deleted = 0", id).Updates(updates).Error
}

func (r *BizPackageRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.BizPackage{}).
		Where("id = ? AND deleted = 0", id).
		Updates(map[string]any{"deleted": 1, "update_time": now}).Error
}
