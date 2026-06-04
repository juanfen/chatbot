package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
)

type BizAnnouncementRepository struct {
	db *gorm.DB
}

func NewBizAnnouncementRepository(db *gorm.DB) *BizAnnouncementRepository {
	return &BizAnnouncementRepository{db: db}
}

func (r *BizAnnouncementRepository) ListPublished(ctx context.Context, limit int) ([]model.BizAnnouncement, error) {
	var rows []model.BizAnnouncement
	q := r.db.WithContext(ctx).
		Where("deleted = 0 AND status = 1").
		Order("sort_no ASC, publish_time DESC NULLS LAST, id ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (r *BizAnnouncementRepository) ListAll(ctx context.Context) ([]model.BizAnnouncement, error) {
	var rows []model.BizAnnouncement
	err := r.db.WithContext(ctx).
		Where("deleted = 0").
		Order("sort_no ASC, id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *BizAnnouncementRepository) FindByID(ctx context.Context, id string) (*model.BizAnnouncement, error) {
	var row model.BizAnnouncement
	err := r.db.WithContext(ctx).Where("id = ? AND deleted = 0", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BizAnnouncementRepository) Create(ctx context.Context, row *model.BizAnnouncement) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *BizAnnouncementRepository) UpdateByID(ctx context.Context, id string, updates map[string]any) error {
	updates["update_time"] = time.Now()
	return r.db.WithContext(ctx).Model(&model.BizAnnouncement{}).Where("id = ? AND deleted = 0", id).Updates(updates).Error
}

func (r *BizAnnouncementRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.BizAnnouncement{}).
		Where("id = ? AND deleted = 0", id).
		Updates(map[string]any{"deleted": 1, "update_time": now}).Error
}
