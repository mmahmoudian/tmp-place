package helpers

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
	Host     string         `json:"host"`
	Port     int            `json:"port"`
	Database DatabaseConfig `json:"database"`
	Logging  LoggingConfig  `json:"logging"`
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	DatabaseFile string `json:"database_file"`
}

// LoggingConfig contains logging-related configuration
type LoggingConfig struct {
	LogFile  string `json:"log_file"`
	LogLevel string `json:"log_level"`
}

// UploadsConfig contains upload-related configuration
type UploadsConfig struct {
	Path          string `json:"path"`
	MaxFileSize   int64  `json:"max_file_size"`
	MaxTTLSeconds int64  `json:"max_ttl_seconds"`
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
