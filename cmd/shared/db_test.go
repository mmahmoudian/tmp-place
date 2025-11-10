package shared

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

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
	if Contains(normalized, "--") {
		t.Errorf("normalized schema still contains -- comments: %s", normalized)
	}
	if Contains(normalized, "/*") || Contains(normalized, "*/") {
		t.Errorf("normalized schema still contains /* */ comments: %s", normalized)
	}
	// Should still contain the actual SQL
	if !Contains(normalized, "create table test") {
		t.Errorf("normalized schema missing CREATE TABLE: %s", normalized)
	}
}

func TestNormalizeSchema_RemovesIfExists(t *testing.T) {
	input := `CREATE TABLE IF NOT EXISTS test (id INTEGER);
CREATE INDEX IF EXISTS idx ON test(id);`

	normalized := NormalizeSchema(input)

	// Should remove IF NOT EXISTS and IF EXISTS
	if Contains(normalized, "if not exists") || Contains(normalized, "if exists") {
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
	if !Contains(normalized, "table") && !Contains(normalized, "TABLE") {
		t.Errorf("normalized schema missing table keyword: %s", normalized)
	}

	// Check that multiple newlines are collapsed
	if Contains(normalized, "\n\n\n") {
		t.Errorf("normalized schema still contains multiple consecutive newlines: %s", normalized)
	}
}

func TestNormalizeSchema_SortsStatements(t *testing.T) {
	input := `CREATE TABLE zebra (id INTEGER);
CREATE TABLE apple (id INTEGER);
CREATE TABLE banana (id INTEGER);`

	normalized := NormalizeSchema(input)
	lines := SplitLines(normalized)

	// After sorting, apple should come before banana and zebra
	appleIdx := FindLineContaining(lines, "apple")
	bananaIdx := FindLineContaining(lines, "banana")
	zebraIdx := FindLineContaining(lines, "zebra")

	if appleIdx == -1 || bananaIdx == -1 || zebraIdx == -1 {
		t.Fatalf("missing expected tables in normalized output: %s", normalized)
	}

	if appleIdx > bananaIdx || bananaIdx > zebraIdx {
		t.Errorf("statements not sorted correctly: apple=%d, banana=%d, zebra=%d\n%s",
			appleIdx, bananaIdx, zebraIdx, normalized)
	}
}

// Helper functions for tests
func Contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			FindSubstring(s, substr))
}

// findSubstring checks if substr exists in s (case-insensitive).
func FindSubstring(s, substr string) bool {
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
func SplitLines(s string) []string {
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
func FindLineContaining(lines []string, substr string) int {
	for i, line := range lines {
		if Contains(line, substr) {
			return i
		}
	}
	return -1
}
