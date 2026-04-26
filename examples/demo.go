package main

import (
	"fmt"
	"os"

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

	// 2. Run Migrations
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

	err = sqlitebase.RunMigrations(db, migrations)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	fmt.Println("Migrations applied successfully!")

	// 3. (Optional) Legacy InitSchema approach
	// This approach is still supported for simple use cases or backwards compatibility.
	dbExists, err := sqlitebase.IsDBExists(dbFileName)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to check if DB exists")
	}

	tableSchemas := map[string]string{
		"settings": "CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT)",
	}
	allExpectedColumns := map[string]map[string]string{
		"settings": {
			"key":   "TEXT",
			"value": "TEXT",
		},
	}

	err = sqlitebase.InitSchema(dbFileName, db, tableSchemas, allExpectedColumns, dbExists)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize schema")
	}
	fmt.Println("Legacy schema initialization/validation successful!")

	// Cleanup demo file
	err = os.Remove(dbFileName)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to remove demo database file")
	}
}
