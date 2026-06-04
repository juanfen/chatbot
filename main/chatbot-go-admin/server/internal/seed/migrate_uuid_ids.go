package seed

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MigratePrefixedIDsToUUID migrates legacy prefixed seed IDs (user_admin/role_admin/route_*)
// to pure UUID IDs, and updates link tables accordingly.
//
// This is intentionally idempotent: if the legacy rows are not found, it will do nothing.
func MigratePrefixedIDsToUUID(ctx context.Context, db *gorm.DB, logger *zap.Logger) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Admin user
		adminUserID := "00000000-0000-0000-0000-000000000001"
		if err := migrateID(tx, "sys_user", "id", "username", "admin", "user_admin", adminUserID); err != nil {
			return err
		}

		// Admin role
		adminRoleID := "00000000-0000-0000-0000-000000000002"
		if err := migrateID(tx, "sys_role", "id", "code", "admin", "role_admin", adminRoleID); err != nil {
			return err
		}

		// Routes
		routeHomeID := "00000000-0000-0000-0000-000000000010"
		routeSysUserID := "00000000-0000-0000-0000-000000000011"
		routeSysRoleID := "00000000-0000-0000-0000-000000000012"
		routeSysRouteID := "00000000-0000-0000-0000-000000000013"

		if err := migrateID(tx, "sys_route", "id", "name", "home", "route_home", routeHomeID); err != nil {
			return err
		}
		if err := migrateID(tx, "sys_route", "id", "name", "sys-user", "route_sys_user", routeSysUserID); err != nil {
			return err
		}
		if err := migrateID(tx, "sys_route", "id", "name", "sys-role", "route_sys_role", routeSysRoleID); err != nil {
			return err
		}
		if err := migrateID(tx, "sys_route", "id", "name", "sys-route", "route_sys_route", routeSysRouteID); err != nil {
			return err
		}

		// Link tables: rebuild seed relations in an idempotent way.
		// sys_user_role (admin -> admin role)
		if err := tx.Exec(
			`DELETE FROM sys_user_role WHERE deleted = 0 AND (
				(user_id IN (?, ?) AND role_id IN (?, ?))
			 )`,
			"user_admin", adminUserID, "role_admin", adminRoleID,
		).Error; err != nil {
			return err
		}
		if err := tx.Exec(
			`INSERT INTO sys_user_role (id, user_id, role_id, deleted)
			 VALUES (?, ?, ?, 0)
			 ON CONFLICT DO NOTHING`,
			uuid.NewString(), adminUserID, adminRoleID,
		).Error; err != nil {
			return err
		}

		// sys_role_route (admin role -> seed routes)
		if err := tx.Exec(
			`DELETE FROM sys_role_route WHERE deleted = 0 AND role_id IN (?, ?)`,
			"role_admin", adminRoleID,
		).Error; err != nil {
			return err
		}

		seedPairs := [][2]string{
			{adminRoleID, routeHomeID},
			{adminRoleID, routeSysUserID},
			{adminRoleID, routeSysRoleID},
			{adminRoleID, routeSysRouteID},
		}
		for _, p := range seedPairs {
			if err := tx.Exec(
				`INSERT INTO sys_role_route (id, role_id, route_id, deleted)
				 VALUES (?, ?, ?, 0)
				 ON CONFLICT DO NOTHING`,
				uuid.NewString(), p[0], p[1],
			).Error; err != nil {
				return err
			}
		}

		logger.Info("migrated legacy prefixed ids to uuid when needed")
		return nil
	})
}

func migrateID(tx *gorm.DB, table, idCol, keyCol, keyVal, legacyID, newID string) error {
	// If there is a row with legacyID, move it to newID.
	// If newID already exists, just leave legacy row as-is (it will be filtered by unique constraints elsewhere).
	type row struct{ ID string }
	var r row
	if err := tx.Raw(
		`SELECT `+idCol+` AS id FROM `+table+` WHERE deleted = 0 AND `+idCol+` = ? LIMIT 1`,
		legacyID,
	).Scan(&r).Error; err != nil {
		return err
	}
	if r.ID == "" {
		return nil
	}
	return tx.Exec(
		`UPDATE `+table+` SET `+idCol+` = ? WHERE `+idCol+` = ? AND deleted = 0`,
		newID, legacyID,
	).Error
}

