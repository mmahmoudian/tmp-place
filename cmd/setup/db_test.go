package setup

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// findSchema locates db_schema.sql by walking up from the given directory.
// TODO: I'm not sure if such function is necessary. I'll keep it for now
// in case I restructure the project and to prevent from this test failing
// for now.
func findSchema(startDir string) ([]byte, error) {
	dir := startDir
	// search up to 3 levels
	for i := 0; i < 3; i++ {
		candidate := filepath.Join(dir, "db_schema.sql")
		if b, err := os.ReadFile(candidate); err == nil {
			return b, nil
		}
		parent := filepath.Dir(dir)
		// reached filesystem root
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, errors.New("db_schema.sql not found by walking up directories")
}

func TestCreateDatabase_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Change to temp dir so db_schema.sql can be found
	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Copy db_schema.sql to temp dir (locate it by walking up from the package dir)
	schemaContent, err := findSchema(origDir)
	if err != nil {
		t.Fatalf("failed to read db_schema.sql: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "db_schema.sql"), schemaContent, 0o644); err != nil {
		t.Fatalf("failed to copy db_schema.sql: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	// Create database
	if err := CreateDatabase(dbPath); err != nil {
		t.Fatalf("CreateDatabase failed: %v", err)
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}

	// Verify schema was applied by checking tables exist
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open created database: %v", err)
	}
	defer db.Close()

	// Check uploads table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='uploads'").Scan(&tableName)
	if err != nil {
		t.Fatalf("uploads table not found: %v", err)
	}
	if tableName != "uploads" {
		t.Errorf("expected table 'uploads', got %q", tableName)
	}

	// Check log table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='log'").Scan(&tableName)
	if err != nil {
		t.Fatalf("log table not found: %v", err)
	}
	if tableName != "log" {
		t.Errorf("expected table 'log', got %q", tableName)
	}

	// Check index exists
	var indexName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_tagged_filename'").Scan(&indexName)
	if err != nil {
		t.Fatalf("index not found: %v", err)
	}
	if indexName != "idx_tagged_filename" {
		t.Errorf("expected index 'idx_tagged_filename', got %q", indexName)
	}
}

func TestCreateDatabase_EmptyPath(t *testing.T) {
	err := CreateDatabase("")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
	if err.Error() != "database file path is empty" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCreateDatabase_MissingSchemaFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Change to temp dir where db_schema.sql doesn't exist
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	// Attempt to create database
	err = CreateDatabase(dbPath)
	if err == nil {
		t.Fatal("expected error for missing schema file, got nil")
	}
}

func TestCreateDatabase_CreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nested", "dirs", "test.db")

	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()

	// Copy db_schema.sql to temp dir (locate it by walking up from the package dir)
	schemaContent, err := findSchema(origDir)
	if err != nil {
		t.Fatalf("failed to read db_schema.sql: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "db_schema.sql"), schemaContent, 0o644); err != nil {
		t.Fatalf("failed to copy db_schema.sql: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to temp dir: %v", err)
	}

	// Create database
	if err := CreateDatabase(dbPath); err != nil {
		t.Fatalf("CreateDatabase failed: %v", err)
	}

	// Verify nested directories were created
	parentDir := filepath.Dir(dbPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Fatal("parent directory was not created")
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestExtractDatabaseSchema_Success(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Create a test database with known schema
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test schema
	testSchema := `
		CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
		CREATE INDEX idx_name ON users(name);
	`
	if _, err := db.Exec(testSchema); err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}
	db.Close()

	// Extract schema
	schema, err := ExtractDatabaseSchema(dbPath)
	if err != nil {
		t.Fatalf("ExtractDatabaseSchema failed: %v", err)
	}

	// Verify schema contains expected elements
	if !contains(schema, "CREATE TABLE users") {
		t.Errorf("schema missing CREATE TABLE users, got: %s", schema)
	}
	if !contains(schema, "CREATE INDEX idx_name") {
		t.Errorf("schema missing CREATE INDEX, got: %s", schema)
	}
}

func TestExtractDatabaseSchema_EmptyPath(t *testing.T) {
	_, err := ExtractDatabaseSchema("")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
	if err.Error() != "dbPath is empty" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExtractDatabaseSchema_MissingFile(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "nonexistent.db")

	_, err := ExtractDatabaseSchema(dbPath)
	if err == nil {
		t.Fatal("expected error for missing database file, got nil")
	}
}

func TestCheckDatabaseSchema_Matching(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	schemaPath := filepath.Join(dir, "schema.sql")

	// Create schema template
	schemaContent := `CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT);
CREATE INDEX idx_name ON test(name);`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0o644); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	// Create database with same schema
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	if _, err := db.Exec(schemaContent); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}
	db.Close()

	// Check schema
	match, err := CheckDatabaseSchema(dbPath, schemaPath)
	if err != nil {
		t.Fatalf("CheckDatabaseSchema failed: %v", err)
	}
	if !match {
		t.Fatal("expected schemas to match")
	}
}

func TestCheckDatabaseSchema_Mismatched(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	schemaPath := filepath.Join(dir, "schema.sql")

	// Create schema template
	schemaContent := `CREATE TABLE test (id INTEGER PRIMARY KEY);`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0o644); err != nil {
		t.Fatalf("failed to write schema file: %v", err)
	}

	// Create database with different schema
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	differentSchema := `CREATE TABLE different (id INTEGER PRIMARY KEY);`
	if _, err := db.Exec(differentSchema); err != nil {
		t.Fatalf("failed to apply schema: %v", err)
	}
	db.Close()

	// Check schema
	match, err := CheckDatabaseSchema(dbPath, schemaPath)
	if err != nil {
		t.Fatalf("CheckDatabaseSchema failed: %v", err)
	}
	if match {
		t.Fatal("expected schemas to NOT match")
	}
}

func TestCheckDatabaseSchema_EmptyPaths(t *testing.T) {
	// Test empty database path
	_, err := CheckDatabaseSchema("", "schema.sql")
	if err == nil {
		t.Fatal("expected error for empty database path")
	}

	// Test empty schema path
	_, err = CheckDatabaseSchema("db.db", "")
	if err == nil {
		t.Fatal("expected error for empty schema path")
	}
}

// Helper functions for tests
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			findSubstring(s, substr))
}

// findSubstring checks if substr exists in s (case-insensitive).
func findSubstring(s, substr string) bool {
	sLower := strings.ToLower(s)
	subLower := strings.ToLower(substr)
	for i := 0; i <= len(sLower)-len(subLower); i++ {
		if sLower[i:i+len(subLower)] == subLower {
			return true
		}
	}
	return false
}
