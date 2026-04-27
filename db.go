package sqlite_base

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

type Config struct {
	Path         string
	MigrationDir string
	MigrationFS  fs.FS
}

var gooseMu sync.Mutex

func Open(config Config) (*sqlx.DB, error) {
	if strings.TrimSpace(config.Path) == "" {
		return nil, errors.New("path is required")
	}

	db, err := sqlx.Open("sqlite3", config.Path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	if config.MigrationFS != nil {
		err = ApplyMigrationsFS(db, config.MigrationFS, config.MigrationDir)
	} else {
		err = ApplyMigrations(db, config.MigrationDir)
	}
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
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

	if err := runGooseUp(db, nil, migrationDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func ApplyMigrationsFS(db *sqlx.DB, migrationFS fs.FS, migrationDir string) error {
	if migrationFS == nil || strings.TrimSpace(migrationDir) == "" {
		return nil
	}

	entries, err := fs.ReadDir(migrationFS, migrationDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read migration fs dir: %w", err)
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

	if err := runGooseUp(db, migrationFS, migrationDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func runGooseUp(db *sqlx.DB, migrationFS fs.FS, migrationDir string) error {
	gooseMu.Lock()
	defer gooseMu.Unlock()

	goose.SetBaseFS(migrationFS)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	return goose.Up(db.DB, migrationDir)
}
