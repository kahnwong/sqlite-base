package sqlite_base

import (
	"testing"
)

func TestRunMigrations(t *testing.T) {
	_, db := createTestDB(t)

	migrations := []Migration{
		{
			ID:  "2023102701_create_users",
			SQL: "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)",
		},
		{
			ID:  "2023102702_add_email_to_users",
			SQL: "ALTER TABLE users ADD COLUMN email TEXT",
		},
	}

	// First run
	err := RunMigrations(db, migrations)
	if err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// Verify table and column exist
	exists, err := tableExists(db, "users")
	if err != nil || !exists {
		t.Fatalf("users table should exist")
	}

	expectedCols := map[string]string{
		"id":    "INTEGER",
		"name":  "TEXT",
		"email": "TEXT",
	}
	err = validateSchema(db, "users", expectedCols)
	if err != nil {
		t.Errorf("Schema validation failed after migration: %v", err)
	}

	// Second run (should be idempotent)
	err = RunMigrations(db, migrations)
	if err != nil {
		t.Fatalf("Second RunMigrations failed: %v", err)
	}

	// Add a new migration and run again
	migrations = append(migrations, Migration{
		ID:  "2023102703_create_posts",
		SQL: "CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT)",
	})

	err = RunMigrations(db, migrations)
	if err != nil {
		t.Fatalf("Third RunMigrations failed: %v", err)
	}

	exists, err = tableExists(db, "posts")
	if err != nil || !exists {
		t.Fatalf("posts table should exist")
	}
}

func TestRunMigrations_Rollback(t *testing.T) {
	_, db := createTestDB(t)

	migrations := []Migration{
		{
			ID:  "2023102701_valid",
			SQL: "CREATE TABLE valid (id INTEGER PRIMARY KEY)",
		},
		{
			ID:  "2023102702_invalid",
			SQL: "CREATE TABLE invalid (id INTEGER PRIMARY KEY, id INTEGER PRIMARY KEY)",
		},
	}

	err := RunMigrations(db, migrations)
	if err == nil {
		t.Fatal("Expected RunMigrations to fail")
	}

	// First migration should have been applied
	exists, err := tableExists(db, "valid")
	if err != nil || !exists {
		t.Fatalf("valid table should exist")
	}

	// Second migration should NOT have been applied
	exists, err = tableExists(db, "invalid")
	if err != nil || exists {
		t.Fatalf("invalid table should NOT exist")
	}

	// Check if migration 1 is recorded
	var count int
	err = db.Get(&count, "SELECT count(*) FROM schema_migrations WHERE id = ?", "2023102701_valid")
	if err != nil || count != 1 {
		t.Errorf("migration 1 should be recorded")
	}

	// Check if migration 2 is NOT recorded
	err = db.Get(&count, "SELECT count(*) FROM schema_migrations WHERE id = ?", "2023102702_invalid")
	if err != nil || count != 0 {
		t.Errorf("migration 2 should NOT be recorded")
	}
}
