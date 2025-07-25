package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the application configuration.
// Root is an absolute, lowercase, backslash‐only path.
// IgnoreDirs, IgnoreExts, IgnorePaths are hash‐sets.
type Config struct {
	Root        string
	IgnoreDirs  map[string]bool
	IgnoreExts  map[string]bool
	IgnorePaths map[string]struct{}
}

type rawConfig struct {
	Root        string   `json:"rootDirectory"`
	IgnoreDirs  []string `json:"directoriesToIgnore,omitempty"`
	IgnoreExts  []string `json:"extensionsToIgnore,omitempty"`
	IgnorePaths []string `json:"pathsToIgnore,omitempty"`
}

// LoadConfig reads JSON from path and decodes it into a rawConfig struct.
func LoadConfig(path string) (rawConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return rawConfig{}, err
	}
	defer f.Close()

	var raw rawConfig
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return rawConfig{}, err
	}
	return raw, nil
}

// NewConfig processes a rawConfig to produce a final, validated Config.
func NewConfig(raw rawConfig) (Config, error) {
	// Determine Root
	root := strings.TrimSpace(raw.Root)
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return Config{}, err
		}
		root = cwd
	}
	if !filepath.IsAbs(root) {
		// make absolute relative to cwd
		cwd, _ := os.Getwd()
		root = filepath.Join(cwd, root)
	}
	root = normalizeConfigPath(root)
	root = normalizeConfigPath(root)

	cfg := Config{
		Root:        root,
		IgnoreDirs:  makeSet(raw.IgnoreDirs),
		IgnoreExts:  makeSet(raw.IgnoreExts),
		IgnorePaths: make(map[string]struct{}, len(raw.IgnorePaths)),
	}

	// Normalize and fill IgnorePaths
	for _, p := range raw.IgnorePaths {
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		p = normalizeConfigPath(p)
		p = normalizeConfigPath(p)
		cfg.IgnorePaths[p] = struct{}{}
		fmt.Printf("Ignoring path: '%s'\n", p)
	}

	return cfg, nil
}

// normalizeConfigPath same as Normalize Path with the added bonus of expanding env variables
func normalizeConfigPath(path string) string {
	path = os.ExpandEnv(path)
	return NormalizePath(path)
}

// normalizeConfigPath same as Normalize Path with the added bonus of expanding env variables
func normalizeConfigPath(path string) string {
	os.ExpandEnv(path)
	return NormalizePath(path)
}

// makeSet creates a set (map[string]bool) from a slice of strings for efficient lookups.
func makeSet(keys []string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}
