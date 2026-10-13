package main

import (
	"context"
	"fmt"
	"log"
	"time"

	sqlitebase "github.com/kahnwong/sqlite-base"
	"github.com/kahnwong/sqlite-base/examples/db"
)

func main() {
	database, err := sqlitebase.Open(sqlitebase.Config{
		Path:         "demo.db",
		MigrationDir: "migrations",
	})
	if err != nil {
		log.Fatalf("open sqlite database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	queries := db.New(database)

	if err := createUser(ctx, queries); err != nil {
		log.Fatalf("create user: %v", err)
	}

	total, err := queries.CountUsers(ctx)
	if err != nil {
		log.Fatalf("count users: %v", err)
	}

	fmt.Printf("database ready, users total: %d\n", total)
}

func createUser(ctx context.Context, queries *db.Queries) error {
	return queries.CreateUser(ctx, db.CreateUserParams{
		Name:  "Alice",
		Email: fmt.Sprintf("alice+%d@example.com", time.Now().UnixNano()),
		Role:  "member",
	})
}
