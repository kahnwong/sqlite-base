package sqlite_base

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
)

// Helper function to create a temporary test database
func createTestDB(t *testing.T) (string, *sqlx.DB) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		if err = db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	})

	return dbPath, db
}

func TestInitDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	t.Cleanup(func() {
		if err = db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	})

	// Verify connection works
	if err = db.Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestIsDBExists(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string
		expected bool
	}{
		{
			name: "database exists",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "test.db")

				db, err := InitDB(dbPath)
				if err != nil {
					t.Fatalf("Failed to create test database: %v", err)
				}

				if err := db.Close(); err != nil {
					t.Fatalf("Failed to close database: %v", err)
				}

				return dbPath
			},
			expected: true,
		},
		{
			name: "database does not exist",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent.db")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := tt.setup(t)
			exists, err := IsDBExists(dbPath)
			if err != nil {
				t.Fatalf("IsDBExists failed: %v", err)
			}
			if exists != tt.expected {
				t.Errorf("Expected exists=%v, got %v", tt.expected, exists)
			}
		})
	}
}

func TestTableExists(t *testing.T) {
	_, db := createTestDB(t)

	// Create a test table
	_, err := db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	tests := []struct {
		name      string
		tableName string
		expected  bool
	}{
		{"existing table", "test_table", true},
		{"non-existing table", "nonexistent_table", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := tableExists(db, tt.tableName)
			if err != nil {
				t.Fatalf("tableExists failed: %v", err)
			}
			if exists != tt.expected {
				t.Errorf("Expected exists=%v, got %v", tt.expected, exists)
			}
		})
	}
}

func TestValidateSchema(t *testing.T) {
	_, db := createTestDB(t)

	// Create a test table
	_, err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	tests := []struct {
		name            string
		tableName       string
		expectedColumns map[string]string
		shouldError     bool
	}{
		{
			name:      "valid schema",
			tableName: "users",
			expectedColumns: map[string]string{
				"id":   "INTEGER",
				"name": "TEXT",
				"age":  "INTEGER",
			},
			shouldError: false,
		},
		{
			name:      "missing column",
			tableName: "users",
			expectedColumns: map[string]string{
				"id":    "INTEGER",
				"name":  "TEXT",
				"email": "TEXT",
			},
			shouldError: true,
		},
		{
			name:      "wrong type",
			tableName: "users",
			expectedColumns: map[string]string{
				"id":   "INTEGER",
				"name": "INTEGER",
			},
			shouldError: true,
		},
		{
			name:      "non-existing table",
			tableName: "nonexistent",
			expectedColumns: map[string]string{
				"id": "INTEGER",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSchema(db, tt.tableName, tt.expectedColumns)
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestInitSchema_NewDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "new.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database: %v", err)
		}
	})

	tableSchemas := map[string]string{
		"users": "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)",
		"posts": "CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, content TEXT)",
	}

	expectedColumns := map[string]map[string]string{
		"users": {
			"id":    "INTEGER",
			"name":  "TEXT",
			"email": "TEXT",
		},
		"posts": {
			"id":      "INTEGER",
			"title":   "TEXT",
			"content": "TEXT",
		},
	}

	err = InitSchema(dbPath, db, tableSchemas, expectedColumns, false)
	if err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	// Verify tables were created
	exists, err := tableExists(db, "users")
	if err != nil {
		t.Fatalf("tableExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected users table to exist")
	}

	exists, err = tableExists(db, "posts")
	if err != nil {
		t.Fatalf("tableExists failed: %v", err)
	}
	if !exists {
		t.Error("Expected posts table to exist")
	}
}

func TestInitSchema_ExistingDatabase(t *testing.T) {
	dbPath, db := createTestDB(t)

	// Create tables manually
	_, err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	tableSchemas := map[string]string{
		"users": "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)",
	}

	expectedColumns := map[string]map[string]string{
		"users": {
			"id":   "INTEGER",
			"name": "TEXT",
		},
	}

	// Should validate successfully
	err = InitSchema(dbPath, db, tableSchemas, expectedColumns, true)
	if err != nil {
		t.Fatalf("InitSchema validation failed: %v", err)
	}
}

func TestInitSchema_ValidationFailure(t *testing.T) {
	dbPath, db := createTestDB(t)

	// Create table with different schema
	_, err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	tableSchemas := map[string]string{
		"users": "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)",
	}

	expectedColumns := map[string]map[string]string{
		"users": {
			"id":    "INTEGER",
			"name":  "TEXT",
			"email": "TEXT", // This column doesn't exist
		},
	}

	// Should fail validation
	err = InitSchema(dbPath, db, tableSchemas, expectedColumns, true)
	if err == nil {
		t.Error("Expected validation to fail but it succeeded")
	}
}
