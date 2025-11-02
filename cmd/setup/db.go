package setup

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// CreateDatabase creates the database file based on db_schema.sql.
func CreateDatabase(dbFilePath string) error {
	if dbFilePath == "" {
		return errors.New("database file path is empty")
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dbFilePath), 0o755); err != nil {
		return fmt.Errorf("creating database directory: %w", err)
	}

	// Open (create if not exists) the SQLite database
	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return fmt.Errorf("opening sqlite database: %w", err)
	}
	defer db.Close()

	// Read schema SQL from project root
	schemaPath := "db_schema.sql"
	f, err := os.Open(schemaPath)
	if err != nil {
		return fmt.Errorf("opening the schema file %s: %w", schemaPath, err)
	}
	defer f.Close()
	schemaBytes, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("reading schema file: %w", err)
	}
	schema := string(schemaBytes)

	// Execute schema; modernc SQLite supports multiple statements in one Exec
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("applying schema: %w", err)
	}

	return nil
}

// ExtractDatabaseSchema opens the SQLite database at dbPath and returns its schema as SQL.
// Returns the schema as a string.
func ExtractDatabaseSchema(dbPath string) (string, error) {
	if strings.TrimSpace(dbPath) == "" {
		return "", errors.New("dbPath is empty")
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", fmt.Errorf("database file does not exist: %s", dbPath)
	}

	// Prefer read-only mode; works even if the file is locked for writes.
	dsn := fmt.Sprintf("file:%s?mode=ro&_busy_timeout=5000", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return "", err
	}

	// Query all user-defined schema objects. `sql` can be NULL for virtual/internal entries.
	// We also filter out sqlite internal objects and ensure deterministic ordering.
	const q = `
		SELECT type, name, sql
		FROM sqlite_schema
		WHERE sql IS NOT NULL
		  AND name NOT LIKE 'sqlite_%'
		  AND type IN ('table','index','view','trigger')
		ORDER BY
		  CASE type WHEN 'table' THEN 1 WHEN 'index' THEN 2 WHEN 'view' THEN 3 WHEN 'trigger' THEN 4 ELSE 5 END,
		  name
	`

	rows, err := db.Query(q)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type stmt struct{ typ, name, sql string }
	var stmts []stmt
	for rows.Next() {
		var s stmt
		if err := rows.Scan(&s.typ, &s.name, &s.sql); err != nil {
			return "", err
		}
		// Normalize whitespace a bit; keep original SQL otherwise.
		s.sql = strings.TrimSpace(s.sql)
		if !strings.HasSuffix(s.sql, ";") {
			s.sql += ";"
		}
		stmts = append(stmts, s)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	// Some schemas may include indexes created automatically without SQL text; we already filtered sql IS NOT NULL.
	// To keep output stable, ensure deterministic secondary sort (by type then name already, but be safe).
	sort.SliceStable(stmts, func(i, j int) bool {
		if stmts[i].typ == stmts[j].typ {
			return stmts[i].name < stmts[j].name
		}
		return stmts[i].typ < stmts[j].typ
	})

	// Join with double newlines for readability.
	out := make([]string, 0, len(stmts))
	for _, s := range stmts {
		out = append(out, s.sql)
	}
	return strings.Join(out, "\n\n"), nil
}

// ReadSchemaTemplate reads the SQL schema from a file and returns it as a string.
// Trims leading/trailing whitespace but otherwise preserves the file contents.
// Returns the schema as a string.
func ReadSchemaTemplate(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is empty")
	}

	// Check if database file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("database schema template file does not exist: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// NormalizeSchema removes comments, extra whitespace, and normalizes SQL for comparison.
// Returns the normalized schema string.
func NormalizeSchema(schema string) string {
	// Remove SQL comments (-- style)
	re := regexp.MustCompile(`--[^\n]*`)
	schema = re.ReplaceAllString(schema, "")

	// Remove multi-line comments (/* ... */)
	re = regexp.MustCompile(`(?s)/\*.*?\*/`)
	schema = re.ReplaceAllString(schema, "")

	// Remove IF NOT EXISTS / IF EXISTS variations for comparison
	re = regexp.MustCompile(`(?i)\s+if\s+(not\s+)?exists\s+`)
	schema = re.ReplaceAllString(schema, " ")

	// remove the newlines for indented sections
	re = regexp.MustCompile(`\n\s+`)
	schema = re.ReplaceAllString(schema, " ")

	// remove newlines before closing parentheses of each sql statement
	re = regexp.MustCompile(`\n\s*\)`)
	schema = re.ReplaceAllString(schema, " )")

	// collapse multiple newlines to single space
	re = regexp.MustCompile(`\n+`)
	schema = re.ReplaceAllString(schema, "\n")

	// ensure each statement ends with a newline after semicolon
	re = regexp.MustCompile(`;\s+`)
	schema = re.ReplaceAllString(schema, ";\n")

	// Trim and convert to lowercase for case-insensitive comparison
	schema = strings.TrimSpace(schema)

	lines := strings.Split(schema, "\n")
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

// CheckDatabaseSchema compares the database schema and the project template schema
// Returns true if they match, false otherwise.
func CheckDatabaseSchema(dbPath string, templateSchemaPath string) (bool, error) {
	// get database schema
	dbSchema, err := ExtractDatabaseSchema(dbPath)
	if err != nil {
		return false, err
	}

	// get template schema
	templateSchema, err := ReadSchemaTemplate(templateSchemaPath)
	if err != nil {
		return false, err
	}

	// Normalize both schemas for comparison
	dbSchema = NormalizeSchema(dbSchema)
	templateSchema = NormalizeSchema(templateSchema)

	// Compare the normalized schemas
	return dbSchema == templateSchema, nil
}
