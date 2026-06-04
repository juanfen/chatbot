package schema

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CheckAndInit scans schemaDir/*.sql and ensures tables exist.
// If a table does not exist, it executes its schema SQL file.
func CheckAndInit(ctx context.Context, db *gorm.DB, logger *zap.Logger, schemaDir string) error {
	if err := db.WithContext(ctx).Exec(`SET search_path TO public`).Error; err != nil {
		return fmt.Errorf("set search_path to public failed: %w", err)
	}

	entries, err := os.ReadDir(schemaDir)
	if err != nil {
		return fmt.Errorf("read schema dir %q failed: %w", schemaDir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, filepath.Join(schemaDir, name))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		tableName := strings.TrimSuffix(filepath.Base(f), ".sql")
		if tableName == "" {
			continue
		}

		exists, err := tableExists(ctx, db, tableName)
		if err != nil {
			return err
		}

		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read schema file %q failed: %w", f, err)
		}
		sqlText := strings.TrimSpace(string(sqlBytes))
		if sqlText == "" {
			continue
		}

		if exists {
			if exp, ok := expectedSchemas[tableName]; ok {
				if err := checkTableColumns(ctx, db, tableName, exp); err != nil {
					return err
				}
			} else {
				logger.Info("no expected schema found, skip diff check", zap.String("table", tableName))
			}
		} else {
			logger.Info("table missing, will apply schema", zap.String("table", tableName))
		}

		logger.Info("apply schema file (idempotent)", zap.String("table", tableName), zap.String("file", f))
		if err := db.WithContext(ctx).Exec(sqlText).Error; err != nil {
			return fmt.Errorf("exec schema file %q failed: %w", f, err)
		}
	}
	return nil
}

type dbColumn struct {
	ColumnName string `gorm:"column:column_name"`
	DataType   string `gorm:"column:data_type"`
	IsNullable string `gorm:"column:is_nullable"`
}

func tableExists(ctx context.Context, db *gorm.DB, table string) (bool, error) {
	var exists bool
	err := db.WithContext(ctx).Raw(
		`SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = ?
		)`,
		table,
	).Scan(&exists).Error
	if err != nil {
		return false, fmt.Errorf("check table exists %q failed: %w", table, err)
	}
	return exists, nil
}

func checkTableColumns(ctx context.Context, db *gorm.DB, table string, exp expectedTable) error {
	var cols []dbColumn
	if err := db.WithContext(ctx).Raw(
		`SELECT column_name, data_type, is_nullable
		 FROM information_schema.columns
		 WHERE table_schema='public' AND table_name=?
		 ORDER BY ordinal_position`,
		table,
	).Scan(&cols).Error; err != nil {
		return fmt.Errorf("load table columns %q failed: %w", table, err)
	}

	actual := map[string]dbColumn{}
	for _, c := range cols {
		actual[c.ColumnName] = c
	}

	var diffs []string
	for _, ec := range exp.Columns {
		ac, ok := actual[ec.Name]
		if !ok {
			diffs = append(diffs, fmt.Sprintf("missing column: %s.%s", table, ec.Name))
			continue
		}
		if !matchesDataType(ac.DataType, ec.DataType) {
			diffs = append(diffs, fmt.Sprintf("type mismatch: %s.%s expected=%s actual=%s", table, ec.Name, ec.DataType, ac.DataType))
		}
		if ec.Nullable != "" && ac.IsNullable != ec.Nullable {
			diffs = append(diffs, fmt.Sprintf("nullable mismatch: %s.%s expected=%s actual=%s", table, ec.Name, ec.Nullable, ac.IsNullable))
		}
	}

	if len(diffs) > 0 {
		return fmt.Errorf("schema mismatch for table %q:\n- %s", table, strings.Join(diffs, "\n- "))
	}
	return nil
}

func matchesDataType(actual string, expected string) bool {
	if actual == expected {
		return true
	}
	// common aliases in information_schema
	if expected == "varchar" && actual == "character varying" {
		return true
	}
	if expected == "timestamp" && strings.Contains(actual, "timestamp") {
		return true
	}
	if expected == "bigint" && actual == "bigint" {
		return true
	}
	if expected == "double precision" && (actual == "double precision" || actual == "real") {
		return true
	}
	if expected == "bytea" && actual == "bytea" {
		return true
	}
	if expected == "text" && actual == "text" {
		return true
	}
	return false
}

