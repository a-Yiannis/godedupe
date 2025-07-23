package main

import (
	"encoding/json"
	"os"
)

// Config defines the structure for the application's configuration.
// It holds the root directory to scan, and lists of directories and
// file extensions to ignore.
type Config struct {
	Root       string   `json:"rootDirectory"`
	IgnoreDirs []string `json:"directoriesToIgnore,omitempty"`
	IgnoreExts []string `json:"extensionsToIgnore,omitempty"`
	// full paths (files or dirs) to ignore; nil means “no path‐ignore check”
	IgnorePaths []string `json:"pathsToIgnore,omitempty"`
}

// loadConfig reads a configuration file from the given path and
// decodes it into a Config struct.
func loadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var cfg Config
	err = json.NewDecoder(f).Decode(&cfg)
	return cfg, err
}
