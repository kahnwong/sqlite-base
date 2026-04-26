package main

import (
	"fmt"

	sqlitebase "github.com/kahnwong/sqlite-base"
	"github.com/rs/zerolog/log"
)

func main() {
	dbFileName := "demo.db"

	// 1. Initialize DB
	db, err := sqlitebase.InitDB(dbFileName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// 2. Define migrations (including the new one)
	migrations := []sqlitebase.Migration{
		{
			ID:  "2023102701_create_users",
			SQL: "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)",
		},
		{
			ID:  "2023102702_add_email_to_users",
			SQL: "ALTER TABLE users ADD COLUMN email TEXT",
		},
	}

	// 3. Run Migrations
	// This will apply both migrations if it's a new DB,
	// or only the second one if the first was already applied.
	err = sqlitebase.RunMigrations(db, migrations)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	fmt.Println("Schema migrated successfully!")
}
