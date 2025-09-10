package sqlite_base

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

func InitSchema(dbFileName string, dbConn *sqlx.DB, tableSchemas map[string]string, allExpectedColumns map[string]map[string]string, dbExists bool) {
	var err error

	for tableName, schema := range tableSchemas {
		if !dbExists { // Database file did not exist, so create the table
			log.Debug().Msgf("INIT: DB - Creating table '%s'...", tableName)
			_, err = dbConn.Exec(schema)
			if err != nil {
				log.Fatal().Err(err).Msgf("Error creating table '%s'", tableName)
			}

			log.Debug().Msgf("INIT: DB - Table '%s' created successfully!", tableName)
		} else { // Database file existed, validate its schema
			log.Debug().Msgf("INIT: DB - Database file '%s' found. Validating schema for table '%s'...", dbFileName, tableName)
			expectedCols, ok := allExpectedColumns[tableName]
			if !ok {
				log.Warn().Msgf("No expected column definitions for table '%s'. Skipping schema validation for this table.", tableName)
				continue
			}
			if err := validateSchema(dbConn, tableName, expectedCols); err != nil {
				log.Fatal().Err(err).Msgf("Schema validation failed for table '%s'", tableName)
			}

			log.Debug().Msgf("INIT: DB - Schema for table '%s' validated successfully.", tableName)
		}
	}
	log.Debug().Msg("INIT: DB - All tables processed successfully.")
}

func IsDBExists(dbFileName string) bool {
	dbExists := true
	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		dbExists = false
		log.Debug().Msgf("INIT: DB - Database file '%s' not found. It will be created.", dbFileName)
	} else if err != nil {
		log.Fatal().Err(err).Msg("Error checking database file status")
	}
	return dbExists
}

func InitDB(dbFileName string) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", dbFileName)
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening database connection")
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	return db
}

func validateSchema(db *sqlx.DB, tableName string, expectedColumns map[string]string) error {
	exists, err := tableExists(db, tableName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("table '%s' does not exist in the database", tableName)
	}

	// Query table info using PRAGMA to get column details
	rows, err := db.Queryx(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
	if err != nil {
		return fmt.Errorf("error querying table info for '%s': %w", tableName, err)
	}
	defer func(rows *sqlx.Rows) {
		err = rows.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Error closing rows for table '%s'", tableName)
		}
	}(rows)

	// Map to store found columns and their types
	foundColumns := make(map[string]string)
	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string // column type (e.g., TEXT, INTEGER)
			notnull    int
			dfltValue  sql.NullString // Default value, can be NULL
			pk         int            // Primary key flag
		)
		// Scan the results from PRAGMA table_info
		if err = rows.Scan(&cid, &name, &columnType, &notnull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("error scanning table info row: %w", err)
		}
		foundColumns[name] = columnType
	}

	// Validate each expected column against the found columns
	for colName, expectedType := range expectedColumns {
		foundType, ok := foundColumns[colName]
		if !ok {
			return fmt.Errorf("missing expected column: '%s'", colName)
		}
		// For simplicity, we'll check for an exact type match.
		// SQLite's type affinity can sometimes return slightly different names
		// (e.g., VARCHAR instead of TEXT), but for basic types, this is usually sufficient.
		if foundType != expectedType {
			return fmt.Errorf("column '%s' has unexpected type: expected '%s', got '%s'", colName, expectedType, foundType)
		}
	}

	// Optionally, you might want to check for extra columns not in expectedColumns,
	// but for now, we only ensure all expected columns are present and correct.

	return nil // Schema is valid
}

func tableExists(db *sqlx.DB, tableName string) (bool, error) {
	var count int

	// Query sqlite_master to check for the table's existence
	query := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`
	err := db.Get(&count, query, tableName)
	if err != nil {
		return false, fmt.Errorf("error checking if table '%s' exists: %w", tableName, err)
	}
	return count > 0, nil
}
