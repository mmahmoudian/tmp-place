package shared

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSaveAndLoadConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	orig := Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
			Database: DatabaseConfig{
				DatabaseFile: filepath.Join(dir, "db.sqlite"),
			},
			Logging: LoggingConfig{
				LogFile:  filepath.Join(dir, "server.log"),
				LogLevel: "info",
			},
		},
		Uploads: UploadsConfig{
			Path:          filepath.Join(dir, "uploads"),
			MaxFileSize:   1024 * 1024,
			MaxTTLSeconds: 3600,
		},
	}

	if err := SaveConfig(cfgPath, orig); err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}

	got, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	if !reflect.DeepEqual(orig, got) {
		t.Fatalf("roundtrip mismatch:\norig=%+v\n got=%+v", orig, got)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")
	if _, err := LoadConfig(path); err == nil {
		t.Fatalf("expected error when loading missing file")
	}
}

func TestSaveConfig_InvalidPath(t *testing.T) {
	// Saving to a directory path should fail during rename
	dir := t.TempDir()
	// create a directory named 'config.json' to make rename fail
	badPath := filepath.Join(dir, "config.json")
	if err := os.MkdirAll(badPath, 0o755); err != nil {
		t.Fatalf("setup mkdir error: %v", err)
	}

	err := SaveConfig(badPath, Config{})
	if err == nil {
		t.Fatalf("expected error when saving to invalid path, got nil")
	}
}
