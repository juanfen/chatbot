package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
)

type BizUserVoiceRepository struct {
	db *gorm.DB
}

func NewBizUserVoiceRepository(db *gorm.DB) *BizUserVoiceRepository {
	return &BizUserVoiceRepository{db: db}
}

func (r *BizUserVoiceRepository) Create(ctx context.Context, row *model.BizUserVoice) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *BizUserVoiceRepository) ListByUserID(ctx context.Context, userID string) ([]model.BizUserVoice, error) {
	var rows []model.BizUserVoice
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted = 0", userID).
		Order("upload_time DESC, id DESC").
		Find(&rows).Error
	return rows, err
}

func (r *BizUserVoiceRepository) FindByIDAndUser(ctx context.Context, id, userID string) (*model.BizUserVoice, error) {
	var row model.BizUserVoice
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ? AND deleted = 0", id, userID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BizUserVoiceRepository) ActivateOne(ctx context.Context, userID, voiceID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		if err := tx.Model(&model.BizUserVoice{}).
			Where("user_id = ? AND deleted = 0", userID).
			Updates(map[string]any{"is_active": 0, "update_time": now}).Error; err != nil {
			return err
		}
		res := tx.Model(&model.BizUserVoice{}).
			Where("id = ? AND user_id = ? AND deleted = 0", voiceID, userID).
			Updates(map[string]any{"is_active": 1, "update_time": now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
