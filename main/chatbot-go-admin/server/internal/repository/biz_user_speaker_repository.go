package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
)

type BizUserSpeakerRepository struct {
	db *gorm.DB
}

func NewBizUserSpeakerRepository(db *gorm.DB) *BizUserSpeakerRepository {
	return &BizUserSpeakerRepository{db: db}
}

func (r *BizUserSpeakerRepository) ListByUserID(ctx context.Context, userID string) ([]model.BizUserSpeaker, error) {
	var rows []model.BizUserSpeaker
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("update_time DESC, speaker_name ASC").
		Find(&rows).Error
	return rows, err
}

func (r *BizUserSpeakerRepository) FindBySpeakerID(ctx context.Context, speakerID string) (*model.BizUserSpeaker, error) {
	var row model.BizUserSpeaker
	err := r.db.WithContext(ctx).Where("speaker_id = ?", speakerID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BizUserSpeakerRepository) FindByIDAndUser(ctx context.Context, speakerID, userID string) (*model.BizUserSpeaker, error) {
	var row model.BizUserSpeaker
	err := r.db.WithContext(ctx).
		Where("speaker_id = ? AND user_id = ?", speakerID, userID).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BizUserSpeakerRepository) FindByNameAndUser(ctx context.Context, userID, name string) (*model.BizUserSpeaker, error) {
	var row model.BizUserSpeaker
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND speaker_name = ?", userID, name).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BizUserSpeakerRepository) Create(ctx context.Context, row *model.BizUserSpeaker) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *BizUserSpeakerRepository) TouchUpdateTime(ctx context.Context, speakerID string) error {
	return r.db.WithContext(ctx).Model(&model.BizUserSpeaker{}).
		Where("speaker_id = ?", speakerID).
		Update("update_time", time.Now()).Error
}

func (r *BizUserSpeakerRepository) UpdatePersonalLines(ctx context.Context, speakerID, personalLinesJSON string) error {
	return r.db.WithContext(ctx).Model(&model.BizUserSpeaker{}).
		Where("speaker_id = ?", speakerID).
		Updates(map[string]any{
			"personal_lines": personalLinesJSON,
			"update_time":    time.Now(),
		}).Error
}
