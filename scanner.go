package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// regex to catch %VAR% on Windows
var reEnvWindows = regexp.MustCompile(`%([^%]+)%`)

// expandEnv replaces $VAR/${VAR} (Unix) and %VAR% (Windows) in p.
func expandEnv(p string) string {
	// Unix‐style
	p = os.ExpandEnv(p)
	// Windows‐style
	return reEnvWindows.ReplaceAllStringFunc(p, func(m string) string {
		return os.Getenv(m[1 : len(m)-1])
	})
}

// findFilesBySize walks the directory tree starting from the root, and groups files by their size.
// It ignores directories and file extensions specified in the config.
func findFilesBySize(cfg Config) map[int64][]string {
	filesBySize := make(map[int64][]string)
	ignoreDirs := makeSet(cfg.IgnoreDirs)
	ignoreExts := makeSet(lowerSlice(cfg.IgnoreExts))
	// build an exact‐match set of cleaned, absolute, lowercase paths
	var ignorePathsSet map[string]struct{}
	if len(cfg.IgnorePaths) > 0 {
		ignorePathsSet = make(map[string]struct{}, len(cfg.IgnorePaths))
		for _, raw := range cfg.IgnorePaths {
			// 1) expand env vars
			p := expandEnv(raw)
			// 2) if relative, resolve against cfg.Root
			if !filepath.IsAbs(p) {
				p = filepath.Join(cfg.Root, p)
			}
			// 3) clean and lowercase
			p = normalizePath(p)
			ignorePathsSet[p] = struct{}{}
			fmt.Printf("Ignoring path: '%s'\n", p)
		}
	}

	lastUpdate := time.Now()
	minimumPeriod := 500 * time.Duration(time.Millisecond)
	count := 0

	filepath.WalkDir(cfg.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("walk: %v", err)
			return nil
		}
		count++
		if time.Since(lastUpdate) > minimumPeriod {
			fmt.Printf("Scanning [%d k]: %s...\n", count/1000.0, path)
			lastUpdate = time.Now()
		}
		// 0) if this exact path is in the ignore set, skip it (and its subtree)
		if ignorePathsSet != nil {
			key := strings.ToLower(path)
			if _, skip := ignorePathsSet[key]; skip {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if d.IsDir() {
			if ignoreDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if ignoreExts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		info, _ := d.Info()
		filesBySize[info.Size()] = append(filesBySize[info.Size()], path)
		return nil
	})

	return filesBySize
}

func normalizePath(path string) string {
	return strings.ReplaceAll(strings.ToLower(filepath.Clean(path)), "/", "\\")
}

func reportDuplicates(dupMap map[uint64][]string) {
	logFile, err := os.Create("duplicates.log")
	if err != nil {
		log.Fatalf("failed to create duplicates.log: %v", err)
	}
	defer logFile.Close()

	found := false
	for _, paths := range dupMap {
		if len(paths) > 1 {
			found = true
			for _, p := range paths {
				logFile.WriteString(p + "\n")
			}
		}
	}

	if found {
		fmt.Println("\nDuplicates found. See duplicates.log for a complete list.")
	} else {
		fmt.Println("\nNo duplicates found.")
	}
}
