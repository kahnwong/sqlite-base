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

	// 2. Define initial migrations
	migrations := []sqlitebase.Migration{
		{
			ID:  "2023102701_create_users",
			SQL: "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT)",
		},
	}

	// 3. Run Migrations
	err = sqlitebase.RunMigrations(db, migrations)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	fmt.Println("Initial schema setup successfully!")
}
