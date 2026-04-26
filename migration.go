package sqlite_base

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// Migration represents a single database migration
type Migration struct {
	ID  string
	SQL string
}

// RunMigrations executes a list of migrations in order
func RunMigrations(db *sqlx.DB, migrations []Migration) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	for _, m := range migrations {
		var exists bool
		err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE id = ?)", m.ID)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", m.ID, err)
		}

		if exists {
			log.Debug().Msgf("Migration %s already applied", m.ID)
			continue
		}

		log.Info().Msgf("Applying migration %s", m.ID)

		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", m.ID, err)
		}

		// Execute the migration SQL
		_, err = tx.Exec(m.SQL)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", m.ID, err)
		}

		// Record the migration
		_, err = tx.Exec("INSERT INTO schema_migrations (id) VALUES (?)", m.ID)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", m.ID, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", m.ID, err)
		}

		log.Info().Msgf("Migration %s applied successfully", m.ID)
	}

	return nil
}
