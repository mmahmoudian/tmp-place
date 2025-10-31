package main

import (
    "encoding/json"
    "os"
)

type Config struct {
    MaxFileSize int64 `json:"max_file_size"`
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
