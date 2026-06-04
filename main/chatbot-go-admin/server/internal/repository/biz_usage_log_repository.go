package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"gorm.io/gorm"
)

type BizUsageLogRepository struct {
	db *gorm.DB
}

func NewBizUsageLogRepository(db *gorm.DB) *BizUsageLogRepository {
	return &BizUsageLogRepository{db: db}
}

func (r *BizUsageLogRepository) Create(ctx context.Context, row *model.BizUsageLog) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *BizUsageLogRepository) ListByUser(ctx context.Context, userID string, offset, limit int) ([]model.BizUsageLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.BizUsageLog{}).Where("user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.BizUsageLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("create_time DESC").
		Offset(offset).Limit(limit).
		Find(&rows).Error
	return rows, total, err
}

type DailyCount struct {
	Day   string `gorm:"column:day"`
	Count int64  `gorm:"column:cnt"`
}

func (r *BizUsageLogRepository) DailyCountsSince(ctx context.Context, userID string, since time.Time) ([]DailyCount, error) {
	var out []DailyCount
	err := r.db.WithContext(ctx).Raw(`
		SELECT to_char(create_time::date, 'YYYY-MM-DD') AS day, COUNT(*)::bigint AS cnt
		FROM biz_usage_log
		WHERE user_id = ? AND create_time >= ?
		GROUP BY 1
		ORDER BY 1
	`, userID, since).Scan(&out).Error
	return out, err
}
