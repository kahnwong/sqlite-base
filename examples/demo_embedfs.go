//go:build ignore

package main

import (
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	sqlitebase "github.com/kahnwong/sqlite-base"
)

// This file lives in examples, so this embeds examples/migrations/*.sql.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

func main() {
	db, err := sqlitebase.Open(sqlitebase.Config{
		Path:         "demo_embedfs.db",
		MigrationDir: "migrations",
		MigrationFS:  migrationFiles,
	})
	if err != nil {
		log.Fatalf("open sqlite database: %v", err)
	}
	defer db.Close()

	if err := createUser(db); err != nil {
		log.Fatalf("create user: %v", err)
	}

	total, err := userCount(db)
	if err != nil {
		log.Fatalf("count users: %v", err)
	}

	fmt.Printf("database ready with embedded migrations, users total: %d\n", total)
}

func createUser(db *sqlx.DB) error {
	_, err := db.Exec(`
INSERT INTO users(name, email, role, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
`, "Alice", fmt.Sprintf("alice+%d@example.com", time.Now().UnixNano()), "member", time.Now().UTC(), time.Now().UTC())

	return err
}

func userCount(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(&count, `SELECT COUNT(1) FROM users`)

	return count, err
}
