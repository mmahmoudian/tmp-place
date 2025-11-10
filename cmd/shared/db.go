package shared

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// QueryOnDatabase opens the SQLite database at dbPath and performs the given query.
//   - For SELECT queries, it will open the database in read-only mode and return the result rows as a slice of maps (column name -> value).
//   - For INSERT queries, it will open the database in read-write mode, executes the statement and returns the last inserted ID.
func QueryOnDatabase(dbPath string, query string, args ...any) ([]map[string]any, error) {
	dbPath = strings.TrimSpace(dbPath)
	if dbPath == "" {
		return nil, errors.New("database path is empty")
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("query is empty")
	}

	// Ensure database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %s", dbPath)
	}

	// get the first word of the query to determine if it's a SELECT or non-SELECT
	queryType := strings.ToLower(strings.SplitN(query, " ", 2)[0])
	var dsn string
	switch queryType {
	case "select":
		// Open in read-only mode
		dsn = fmt.Sprintf("file:%s?mode=ro&_busy_timeout=5000", dbPath)
	case "insert", "update", "delete", "create", "drop", "alter":
		// Open in read-write mode
		dsn = fmt.Sprintf("file:%s?mode=rw&_busy_timeout=5000", dbPath)
	default:
		return nil, fmt.Errorf("unsupported query type: %s", queryType)
	}

	// Open with a modest busy timeout and read-write mode
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if queryType != "select" {
		res, err := db.Exec(query, args...)
		if err != nil {
			return nil, err
		}
		ra, _ := res.RowsAffected()
		li, _ := res.LastInsertId()
		return []map[string]any{{
			"rows_affected":  ra,
			"last_insert_id": li,
		}}, nil
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Prepare a slice of destinations for Scan
	dest := make([]any, len(cols))
	destPtrs := make([]any, len(cols))
	for i := range dest {
		destPtrs[i] = &dest[i]
	}

	var out []map[string]any
	for rows.Next() {
		if err := rows.Scan(destPtrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			v := dest[i]
			// Convert []byte to string for readability
			if b, ok := v.([]byte); ok {
				row[c] = string(b)
			} else {
				row[c] = v
			}
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
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

