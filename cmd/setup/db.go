package setup

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
