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
	Host     string `json:"host"`
	Port     int    `json:"port"`
	LogLevel string `json:"log_level"`
}

// UploadsConfig contains upload-related configuration
type UploadsConfig struct {
	Path          string `json:"path"`
	MaxFileSize   int64  `json:"max_file_size"`
	MaxTTLSeconds int64  `json:"max_ttl_seconds"`
}

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
