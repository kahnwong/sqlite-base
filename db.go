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

var migrationLock sync.Mutex

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

	if err := ApplyMigrationsFS(db, config.MigrationFS, config.MigrationDir); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func ApplyMigrations(db *sqlx.DB, migrationDir string) error {
	return ApplyMigrationsFS(db, nil, migrationDir)
}

func ApplyMigrationsFS(db *sqlx.DB, fsys fs.FS, migrationDir string) error {
	if strings.TrimSpace(migrationDir) == "" {
		return nil
	}

	if fsys == nil {
		fsys = os.DirFS(".")
		// If using os.DirFS("."), we need to adjust migrationDir if it's absolute
		// but typically it's relative in this context.
		// For simplicity, we can just use the previous logic if fsys is nil
		// or try to unify it.
	}

	// Unify by using fsys if provided, otherwise use os.
	var entries []fs.DirEntry

	if fsys != nil && fsys != os.DirFS(".") {
		info, err := fs.Stat(fsys, migrationDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("stat migration dir: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("migration path is not a directory: %s", migrationDir)
		}
		entries, err = fs.ReadDir(fsys, migrationDir)
		if err != nil {
			return fmt.Errorf("read migration dir: %w", err)
		}
	} else {
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
		entries, err = os.ReadDir(migrationDir)
		if err != nil {
			return fmt.Errorf("read migration dir: %w", err)
		}
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

	migrationLock.Lock()
	defer migrationLock.Unlock()

	if fsys != nil && fsys != os.DirFS(".") {
		goose.SetBaseFS(fsys)
		defer goose.SetBaseFS(nil)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.Up(db.DB, migrationDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
