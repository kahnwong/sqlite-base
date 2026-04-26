package sqlite_base

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestOpen_RequiresPath(t *testing.T) {
	t.Parallel()

	_, err := Open(Config{})
	if err == nil || err.Error() != "path is required" {
		t.Fatalf("expected path required error, got: %v", err)
	}
}

func TestOpen_RejectsMissingParentDir(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "db.sqlite")
	_, err := Open(Config{Path: path})
	if err == nil {
		t.Fatal("expected error for missing parent dir")
	}
}

func TestOpen_AppliesMigrations(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "app.sqlite")
	migrationDir := t.TempDir()
	migrationPath := filepath.Join(migrationDir, "00001_create_users.sql")
	migrationSQL := "-- +goose Up\nCREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL);\n-- +goose Down\nDROP TABLE users;\n"
	if err := os.WriteFile(migrationPath, []byte(migrationSQL), 0o600); err != nil {
		t.Fatalf("write migration failed: %v", err)
	}

	db, err := Open(Config{
		Path:         dbPath,
		MigrationDir: migrationDir,
	})
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec("INSERT INTO users (name) VALUES (?)", "alice"); err != nil {
		t.Fatalf("insert failed, table likely missing: %v", err)
	}
}

func TestApplyMigrations_AppliesSQLFiles(t *testing.T) {
	t.Parallel()

	db := sqlx.MustOpen("sqlite3", ":memory:")
	t.Cleanup(func() { _ = db.Close() })

	migrationDir := t.TempDir()
	migrationPath := filepath.Join(migrationDir, "00001_create_widgets.sql")
	migrationSQL := "-- +goose Up\nCREATE TABLE widgets (id INTEGER PRIMARY KEY, name TEXT NOT NULL);\n-- +goose Down\nDROP TABLE widgets;\n"
	if err := os.WriteFile(migrationPath, []byte(migrationSQL), 0o600); err != nil {
		t.Fatalf("write migration failed: %v", err)
	}

	if err := ApplyMigrations(db, migrationDir); err != nil {
		t.Fatalf("apply migrations failed: %v", err)
	}

	if _, err := db.Exec("INSERT INTO widgets (name) VALUES (?)", "w1"); err != nil {
		t.Fatalf("insert failed, migration not applied: %v", err)
	}
}

func TestApplyMigrations_NonDirectoryPath(t *testing.T) {
	t.Parallel()

	db := sqlx.MustOpen("sqlite3", ":memory:")
	t.Cleanup(func() { _ = db.Close() })

	f := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	if err := ApplyMigrations(db, f); err == nil {
		t.Fatal("expected error when migration path is a file")
	}
}

func TestApplyMigrations_EmptyOrMissingNoOp(t *testing.T) {
	t.Parallel()

	db := sqlx.MustOpen("sqlite3", ":memory:")
	t.Cleanup(func() { _ = db.Close() })

	if err := ApplyMigrations(db, ""); err != nil {
		t.Fatalf("empty migration dir should be noop: %v", err)
	}

	missing := filepath.Join(t.TempDir(), "missing")
	if err := ApplyMigrations(db, missing); err != nil {
		t.Fatalf("missing migration dir should be noop: %v", err)
	}
}
