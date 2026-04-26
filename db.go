package sqlite_base

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

type TableDefinition struct {
	Name      string
	CreateSQL string
}

type Config struct {
	Path         string
	MigrationDir string
	Tables       []TableDefinition
}

func Open(config Config) (*sqlx.DB, error) {
	if strings.TrimSpace(config.Path) == "" {
		return nil, errors.New("path is required")
	}

	if err := validateDatabaseParentDir(config.Path); err != nil {
		return nil, err
	}

	wasExisting, err := databaseExists(config.Path)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open("sqlite3", config.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	if !wasExisting {
		if err := createTables(db, config.Tables); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	if err := ApplyMigrations(db, config.MigrationDir); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func createTables(db *sqlx.DB, tables []TableDefinition) error {
	if len(tables) == 0 {
		return nil
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin table creation transaction: %w", err)
	}

	for _, table := range tables {
		if strings.TrimSpace(table.CreateSQL) == "" {
			_ = tx.Rollback()
			return fmt.Errorf("create SQL is required for table %q", table.Name)
		}

		if _, err := tx.Exec(table.CreateSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("create table %q: %w", table.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit table creation transaction: %w", err)
	}

	return nil
}

func ApplyMigrations(db *sqlx.DB, migrationDir string) error {
	if strings.TrimSpace(migrationDir) == "" {
		return nil
	}

	info, err := os.Stat(migrationDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat migration dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("migration path is not a directory: %s", migrationDir)
	}

	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("read migration dir: %w", err)
	}

	hasSQL := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			hasSQL = true
			break
		}
	}
	if !hasSQL {
		return nil
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(db.DB, migrationDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func databaseExists(path string) (bool, error) {
	if path == ":memory:" || strings.HasPrefix(path, "file::memory:") {
		return false, nil
	}

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, fmt.Errorf("stat database path: %w", err)
}

func validateDatabaseParentDir(path string) error {
	if path == ":memory:" || strings.HasPrefix(path, "file::memory:") {
		return nil
	}

	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}

	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("database directory does not exist: %s", dir)
		}
		return fmt.Errorf("stat database directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("database parent path is not a directory: %s", dir)
	}

	return nil
}
