//go:build ignore

package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	sqlitebase "github.com/kahnwong/sqlite-base"
	"github.com/kahnwong/sqlite-base/examples/store"
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

	ctx := context.Background()
	queries := store.New(db)

	if err := createUser(ctx, queries); err != nil {
		log.Fatalf("create user: %v", err)
	}

	total, err := queries.CountUsers(ctx)
	if err != nil {
		log.Fatalf("count users: %v", err)
	}

	fmt.Printf("database ready with embedded migrations, users total: %d\n", total)
}

func createUser(ctx context.Context, queries *store.Queries) error {
	return queries.CreateUser(ctx, store.CreateUserParams{
		Name:  "Alice",
		Email: fmt.Sprintf("alice+%d@example.com", time.Now().UnixNano()),
		Role:  "member",
	})
}
