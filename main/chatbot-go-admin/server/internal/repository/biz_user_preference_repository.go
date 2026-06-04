package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
)

type BizUserPreferenceRepository struct {
	db *gorm.DB
}

func NewBizUserPreferenceRepository(db *gorm.DB) *BizUserPreferenceRepository {
	return &BizUserPreferenceRepository{db: db}
}

func (r *BizUserPreferenceRepository) FindByUserID(ctx context.Context, userID string) (*model.BizUserPreference, error) {
	var row model.BizUserPreference
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BizUserPreferenceRepository) UpsertStandardText(ctx context.Context, userID, text string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO biz_user_preference (user_id, standard_text, create_time, update_time)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			standard_text = EXCLUDED.standard_text,
			update_time = EXCLUDED.update_time
	`, userID, text, now, now).Error
}
