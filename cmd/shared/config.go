package shared

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration structure
type Config struct {
	Server  ServerConfig  `json:"server"`
	Uploads UploadsConfig `json:"uploads"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	// name or address of the server
	Host string `json:"host"`
	// port number for the server
	Port     int            `json:"port"`
	Database DatabaseConfig `json:"database"`
	Logging  LoggingConfig  `json:"logging"`
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	// path to the database file
	DatabaseFile string `json:"database_file"`
}

// LoggingConfig contains logging-related configuration
type LoggingConfig struct {
	// path to the log file
	LogFile string `json:"log_file"`
	// logging level (e.g., debug, info, warn, error)
	LogLevel string `json:"log_level"`
}

// UploadsConfig contains upload-related configuration
type UploadsConfig struct {
	// path to the upload directory
	Path string `json:"path"`
	// maximum allowed file size for uploads in bytes
	MaxFileSize int64 `json:"max_file_size"`
	// maximum TTL (time-to-live) for uploaded files in seconds
	MaxTTLSeconds int64 `json:"max_ttl_seconds"`
}

// LoadConfig loads configuration from a JSON file.
// It returns a Config struct and any error encountered.
func LoadConfig(filename string) (Config, error) {
	var cfg Config
	file, err := os.Open(filename)
	if err != nil {
		return cfg, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	return cfg, err
}

// SaveConfig saves the given configuration to a JSON file.
// It writes the file with read/write permissions for the owner (0644).
func SaveConfig(filename string, cfg Config) error {
	// marshal with indentation for readability
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	// write atomically: write to temp file then rename (best-effort)
	tmp := filename + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filename)
}
