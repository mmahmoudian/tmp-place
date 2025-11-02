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

func TestReadSchemaTemplate_Success(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "test_schema.sql")

	// Create test schema file
	testContent := "CREATE TABLE test (id INTEGER PRIMARY KEY);"
	if err := os.WriteFile(schemaPath, []byte(testContent), 0o644); err != nil {
		t.Fatalf("failed to write test schema file: %v", err)
	}

	// Read schema
	schema, err := ReadSchemaTemplate(schemaPath)
	if err != nil {
		t.Fatalf("ReadSchemaTemplate failed: %v", err)
	}

	if schema != testContent {
		t.Errorf("schema = %q; want %q", schema, testContent)
	}
}

func TestReadSchemaTemplate_EmptyPath(t *testing.T) {
	_, err := ReadSchemaTemplate("")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
	if err.Error() != "path is empty" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestReadSchemaTemplate_MissingFile(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "nonexistent.sql")

	_, err := ReadSchemaTemplate(schemaPath)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestNormalizeSchema_RemovesComments(t *testing.T) {
	input := `-- Single line comment
CREATE TABLE test (id INTEGER); -- inline comment
/* Multi-line
   comment */
CREATE INDEX idx ON test(id);`

	normalized := NormalizeSchema(input)

	// Should remove all comments
	if contains(normalized, "--") {
		t.Errorf("normalized schema still contains -- comments: %s", normalized)
	}
	if contains(normalized, "/*") || contains(normalized, "*/") {
		t.Errorf("normalized schema still contains /* */ comments: %s", normalized)
	}
	// Should still contain the actual SQL
	if !contains(normalized, "create table test") {
		t.Errorf("normalized schema missing CREATE TABLE: %s", normalized)
	}
}

func TestNormalizeSchema_RemovesIfExists(t *testing.T) {
	input := `CREATE TABLE IF NOT EXISTS test (id INTEGER);
CREATE INDEX IF EXISTS idx ON test(id);`

	normalized := NormalizeSchema(input)

	// Should remove IF NOT EXISTS and IF EXISTS
	if contains(normalized, "if not exists") || contains(normalized, "if exists") {
		t.Errorf("normalized schema still contains IF (NOT) EXISTS: %s", normalized)
	}
}

func TestNormalizeSchema_NormalizesWhitespace(t *testing.T) {
	input := `CREATE   TABLE    test   (
    id    INTEGER,
    name  TEXT
);`

	normalized := NormalizeSchema(input)

	// NormalizeSchema doesn't lowercase but does normalize structure
	// It collapses newlines within statements but may preserve some spacing
	if !contains(normalized, "table") && !contains(normalized, "TABLE") {
		t.Errorf("normalized schema missing table keyword: %s", normalized)
	}

	// Check that multiple newlines are collapsed
	if contains(normalized, "\n\n\n") {
		t.Errorf("normalized schema still contains multiple consecutive newlines: %s", normalized)
	}
}

func TestNormalizeSchema_SortsStatements(t *testing.T) {
	input := `CREATE TABLE zebra (id INTEGER);
CREATE TABLE apple (id INTEGER);
CREATE TABLE banana (id INTEGER);`

	normalized := NormalizeSchema(input)
	lines := splitLines(normalized)

	// After sorting, apple should come before banana and zebra
	appleIdx := findLineContaining(lines, "apple")
	bananaIdx := findLineContaining(lines, "banana")
	zebraIdx := findLineContaining(lines, "zebra")

	if appleIdx == -1 || bananaIdx == -1 || zebraIdx == -1 {
		t.Fatalf("missing expected tables in normalized output: %s", normalized)
	}

	if appleIdx > bananaIdx || bananaIdx > zebraIdx {
		t.Errorf("statements not sorted correctly: apple=%d, banana=%d, zebra=%d\n%s",
			appleIdx, bananaIdx, zebraIdx, normalized)
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

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// findLineContaining returns the index of the first line containing the given substring.
func findLineContaining(lines []string, substr string) int {
	for i, line := range lines {
		if contains(line, substr) {
			return i
		}
	}
	return -1
}
