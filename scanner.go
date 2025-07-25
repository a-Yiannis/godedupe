package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"godedupe/utils"
)

// findFilesBySize walks the directory tree starting from the root, and groups files by their size.
// It ignores directories and file extensions specified in the config.
func findFilesBySize(cfg Config) map[int64][]string {
	filesBySize := make(map[int64][]string)

	ignoreDirs := cfg.IgnoreDirs
	ignoreExts := cfg.IgnoreExts
	ignorePaths := cfg.IgnorePaths

	lastUpdate := time.Now()
	minimumPeriod := 500 * time.Duration(time.Millisecond)
	count := 0

	filepath.WalkDir(cfg.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			utils.PrintEf("walk: %v", err)
			return nil
		}
		count++
		if time.Since(lastUpdate) > minimumPeriod {
			fmt.Printf("Scanning [%d k]: %s...\n", count/1000.0, path)
			lastUpdate = time.Now()
		}
		// 0) if this exact path is in the ignore set, skip it (and its subtree)
		key := utils.NormalizePath(path)
		if len(ignorePaths) != 0 {
			if _, skip := ignorePaths[key]; skip {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if d.IsDir() {
			if ignoreDirs[strings.ToLower(d.Name())] {
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

func reportDuplicates(dupMap map[uint64][]string) uint32 {
	logFile, err := os.Create("duplicates.log")
	if err != nil {
		utils.PrintEf("failed to create duplicates.log: %v", err)
		os.Exit(1)
	}
	defer logFile.Close()

	var count uint32 = 0
	for _, paths := range dupMap {
		if len(paths) > 1 {
			count++
			for _, p := range paths {
				logFile.WriteString(p + "\n")
			}
		}
	}

	if count > 0 {
		utils.Printf("\nDuplicates **%d** found. See **duplicates.log** for a complete list.\n", count)
	} else {
		utils.WriteCyan("\nNo duplicates found.\n")
	}
	return count
}

