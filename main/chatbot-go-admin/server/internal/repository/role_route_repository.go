package repository

import (
	"context"
	"time"

	"chatbot-go-admin/server/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRouteRepository struct {
	db *gorm.DB
}

func NewRoleRouteRepository(db *gorm.DB) *RoleRouteRepository {
	return &RoleRouteRepository{db: db}
}

func (r *RoleRouteRepository) ListRouteIDsByRoleID(ctx context.Context, roleID string) ([]string, error) {
	var rows []model.SysRoleRoute
	err := r.db.WithContext(ctx).
		Where("role_id = ? AND deleted = 0", roleID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.RouteID)
	}
	return out, nil
}

func (r *RoleRouteRepository) ListRouteIDsByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	if len(roleIDs) == 0 {
		return out, nil
	}
	var rows []model.SysRoleRoute
	err := r.db.WithContext(ctx).
		Where("role_id IN ? AND deleted = 0", roleIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.RoleID] = append(out[row.RoleID], row.RouteID)
	}
	return out, nil
}

func (r *RoleRouteRepository) ReplaceRoleRoutes(ctx context.Context, roleID string, routeIDs []string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("role_id = ?", roleID).Delete(&model.SysRoleRoute{}).Error; err != nil {
			return err
		}
		for _, rid := range routeIDs {
			row := &model.SysRoleRoute{
				ID:         uuid.NewString(),
				RoleID:     roleID,
				RouteID:    rid,
				CreateTime: now,
				UpdateTime: now,
				Deleted:    0,
			}
			if err := tx.Create(row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *RoleRouteRepository) SoftDeleteByRouteID(ctx context.Context, routeID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.SysRoleRoute{}).
		Where("route_id = ? AND deleted = 0", routeID).
		Updates(map[string]any{"deleted": 1, "update_time": now}).Error
}

func (r *RoleRouteRepository) SoftDeleteByRoleID(ctx context.Context, roleID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.SysRoleRoute{}).
		Where("role_id = ? AND deleted = 0", roleID).
		Updates(map[string]any{"deleted": 1, "update_time": now}).Error
}
