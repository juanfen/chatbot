package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"chatbot-go-admin/server/internal/util"
	"gorm.io/gorm"
)

type BizDeviceAudioRepository struct {
	db *gorm.DB
}

func NewBizDeviceAudioRepository(db *gorm.DB) *BizDeviceAudioRepository {
	return &BizDeviceAudioRepository{db: db}
}

func (r *BizDeviceAudioRepository) Create(ctx context.Context, row *model.BizDeviceAudio) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// ListSpeakerIDsByMacID returns distinct speaker_id values already bound for the device.
func (r *BizDeviceAudioRepository) ListSpeakerIDsByMacID(ctx context.Context, macID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&model.BizDeviceAudio{}).
		Where("mac_id = ? AND speaker_id IS NOT NULL AND speaker_id <> ''", macID).
		Distinct("speaker_id").
		Order("speaker_id ASC").
		Pluck("speaker_id", &ids).Error
	return ids, err
}

// ListVisibleByMacID returns rows not yet expired (now < expire_time).
func (r *BizDeviceAudioRepository) ListVisibleByMacID(ctx context.Context, macID string) ([]model.BizDeviceAudio, error) {
	var rows []model.BizDeviceAudio
	err := r.db.WithContext(ctx).
		Where("mac_id = ? AND expire_time > ?", macID, time.Now()).
		Order("create_time DESC").
		Find(&rows).Error
	return rows, err
}

// ListManageable returns rows visible in admin UI: not expired or permanently bound.
func (r *BizDeviceAudioRepository) ListManageable(ctx context.Context, page, pageSize int) ([]model.BizDeviceAudio, int64, error) {
	now := time.Now()
	q := r.db.WithContext(ctx).Model(&model.BizDeviceAudio{}).
		Where("expire_time > ? OR speaker_id IS NOT NULL", now)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	var rows []model.BizDeviceAudio
	err := q.Order("create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}

func (r *BizDeviceAudioRepository) FindByID(ctx context.Context, id string) (*model.BizDeviceAudio, error) {
	var row model.BizDeviceAudio
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BizDeviceAudioRepository) FindAudioBlobByID(ctx context.Context, id string) ([]byte, error) {
	var blob []byte
	err := r.db.WithContext(ctx).Model(&model.BizDeviceAudio{}).
		Select("audio_blob").
		Where("id = ?", id).
		Scan(&blob).Error
	return blob, err
}

func (r *BizDeviceAudioRepository) BindSpeaker(ctx context.Context, audioID, speakerID, userID string) error {
	perpetual := util.PerpetualTime
	return r.db.WithContext(ctx).Model(&model.BizDeviceAudio{}).
		Where("id = ?", audioID).
		Updates(map[string]any{
			"speaker_id":    speakerID,
			"bound_user_id": userID,
			"expire_time":   perpetual,
			"delete_time":   perpetual,
		}).Error
}

// ListMacIDsBySpeakerID returns distinct device MAC addresses bound to a speaker.
func (r *BizDeviceAudioRepository) ListMacIDsBySpeakerID(ctx context.Context, speakerID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&model.BizDeviceAudio{}).
		Where("speaker_id = ? AND mac_id <> ''", speakerID).
		Distinct("mac_id").
		Order("mac_id ASC").
		Pluck("mac_id", &ids).Error
	return ids, err
}

// DeleteDue physically removes rows whose delete_time has passed (skips perpetual rows).
func (r *BizDeviceAudioRepository) DeleteDue(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("delete_time <= ? AND delete_time < ?", time.Now(), util.PerpetualTime).
		Delete(&model.BizDeviceAudio{})
	return res.RowsAffected, res.Error
}
