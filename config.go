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

// LoadConfig reads JSON from path, decodes into rawConfig,
// then normalizes Root and IgnorePaths (absolute, lowercase,
// backslashes), and builds hash‐sets.
func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var raw rawConfig
	if err := json.NewDecoder(f).Decode(&raw); err != nil {
		return Config{}, err
	}

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
	root = NormalizeWindowsPath(root)

	cfg := Config{
		Root:        root,
		IgnoreDirs:  make(map[string]bool, len(raw.IgnoreDirs)),
		IgnoreExts:  make(map[string]bool, len(raw.IgnoreExts)),
		IgnorePaths: make(map[string]struct{}, len(raw.IgnorePaths)),
	}

	// Fill IgnoreDirs
	cfg.IgnoreDirs = makeSet(raw.IgnoreDirs)
	// Fill IgnoreExts
	cfg.IgnoreExts = makeSet(raw.IgnoreExts)

	// Normalize and fill IgnorePaths
	for _, p := range raw.IgnorePaths {
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		p = NormalizeWindowsPath(p)
		cfg.IgnorePaths[p] = struct{}{}
		fmt.Printf("Ignoring path: '%s'\n", p)
	}

	return cfg, nil
}

// makeSet creates a set (map[string]bool) from a slice of strings for efficient lookups.
func makeSet(keys []string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}
