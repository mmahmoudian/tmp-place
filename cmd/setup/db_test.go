package setup

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
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
