package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BizUserQuotaRepository struct {
	db *gorm.DB
}

func NewBizUserQuotaRepository(db *gorm.DB) *BizUserQuotaRepository {
	return &BizUserQuotaRepository{db: db}
}

func (r *BizUserQuotaRepository) GetByUserID(ctx context.Context, userID string) (*model.BizUserQuota, error) {
	var row model.BizUserQuota
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *BizUserQuotaRepository) MapByUserIDs(ctx context.Context, userIDs []string) (map[string]model.BizUserQuota, error) {
	out := make(map[string]model.BizUserQuota)
	if len(userIDs) == 0 {
		return out, nil
	}
	var rows []model.BizUserQuota
	if err := r.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, q := range rows {
		out[q.UserID] = q
	}
	return out, nil
}

// Upsert replaces or creates quota row for user.
func (r *BizUserQuotaRepository) Upsert(ctx context.Context, row *model.BizUserQuota) error {
	row.UpdateTime = time.Now()
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"package_id", "package_title", "total_quota", "used_quota", "update_time"}),
	}).Create(row).Error
}

func (r *BizUserQuotaRepository) IncrementUsed(ctx context.Context, userID string, delta int64) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.BizUserQuota{}).
		Where("user_id = ?", userID).
		UpdateColumn("used_quota", gorm.Expr("used_quota + ?", delta))
	if res.Error != nil {
		return 0, res.Error
	}
	var row model.BizUserQuota
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error; err != nil {
		return 0, err
	}
	return row.UsedQuota, nil
}
