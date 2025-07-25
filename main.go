package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"godedupe/utils"
)

func main() {
	start := time.Now()
	defer utils.CloseLog()

	// --- Configuration Loading ---
	var (
		configPath  string
		root        string
		ignoreDirs  utils.StringSlice
		ignoreExts  utils.StringSlice
		ignorePaths utils.StringSlice
		autoYes     bool
	)

	flag.StringVar(&configPath, "config", "config.json", "Path to the configuration file.")
	flag.StringVar(&root, "root", "", "Root directory to scan (overrides config file).")
	flag.Var(&ignoreDirs, "ignore-dir", "Directory to ignore (can be specified multiple times).")
	flag.Var(&ignoreExts, "ignore-ext", "File extension to ignore (can be specified multiple times).")
	flag.Var(&ignorePaths, "ignore-path", "Path to ignore (can be specified multiple times).")
	flag.BoolVar(&autoYes, "y", false, "Automatically answer yes to all prompts.")
	flag.BoolVar(&autoYes, "yes", false, "Automatically answer yes to all prompts.")
	flag.Parse()

	utils.SetAutoYes(autoYes)

	rawCfg, err := LoadConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		utils.PrintEf("loadConfig: %v", err)
		os.Exit(1)
	}

	// Override config with flags if they are set
	if root != "" {
		rawCfg.Root = root
	}
	if len(ignoreDirs) > 0 {
		rawCfg.IgnoreDirs = append(rawCfg.IgnoreDirs, ignoreDirs...)
	}
	if len(ignoreExts) > 0 {
		rawCfg.IgnoreExts = append(rawCfg.IgnoreExts, ignoreExts...)
	}
	if len(ignorePaths) > 0 {
		rawCfg.IgnorePaths = append(rawCfg.IgnorePaths, ignorePaths...)
	}

	cfg, err := NewConfig(rawCfg)
	if err != nil {
		utils.PrintEf("newConfig: %v", err)
		os.Exit(1)
	}
	// --- End Configuration Loading ---

	fmt.Printf("Scanning \"%s\" …\n", cfg.Root)

	// Phase 1: group by size
	filesBySize := findFilesBySize(cfg)

	// Phase 2: partial-hash + full-hash pipeline
	sem := make(chan struct{}, runtime.GOMAXPROCS(0))
	dupMap := make(map[uint64][]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, sameSize := range filesBySize {
		if len(sameSize) < 2 {
			continue
		}

		// Partial-hash pass
		partialGroups := make(map[uint64][]string)
		var pmu sync.Mutex
		for _, path := range sameSize {
			wg.Add(1)
			sem <- struct{}{}
			go func(p string) {
				defer wg.Done()
				defer func() { <-sem }()
				h, err := partialHash(p)
				if err != nil {
					utils.PrintEf("partialHash %s: %v", p, err)
					return
				}
				pmu.Lock()
				partialGroups[h] = append(partialGroups[h], p)
				pmu.Unlock()
			}(path)
		}
		wg.Wait()

		// Full-hash pass on collisions
		for _, group := range partialGroups {
			if len(group) < 2 {
				continue
			}
			for _, path := range group {
				wg.Add(1)
				sem <- struct{}{}
				go func(p string) {
					defer wg.Done()
					defer func() { <-sem }()
					h, err := fullHash(p)
					if err != nil {
						utils.PrintEf("fullHash %s: %v", p, err)
						return
					}
					mu.Lock()
					dupMap[h] = append(dupMap[h], p)
					mu.Unlock()
				}(path)
			}
		}
		wg.Wait()
	}

	// Report
	count := reportDuplicates(dupMap)

	elapsed := time.Since(start)
	utils.Printf("Elapsed: %.dms\n", elapsed.Milliseconds())

	if count > 0 && utils.AskStrict("Should I recycle the duplicates?") {
		recycle(dupMap)
	}
}

func recycle(dupMap map[uint64][]string) {
	f, err := os.OpenFile("recycled.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		utils.PrintEf("open log: %v", err)
		return
	}
	defer f.Close()
	logger := log.New(f, "", log.LstdFlags)

	for _, paths := range dupMap {
		if len(paths) <= 1 {
			continue
		}
		slices.SortFunc(paths, sortByModTime)

		s := paths[1:]
		fmt.Println("\n" + strings.Join(s, "\n\t"))
		if !utils.Ask("Going to recycle this files, are you sure?") {
			utils.WriteRed("\nFile set skipped!")
			continue
		}
		for _, path := range paths[1:] {
			fmt.Printf("Recycling '%s' \n", path)
			err := utils.RecycleFile(path)
			if err != nil {
				utils.PrintEf("RecycleFile %s: %v", path, err)
			} else {
				logger.Println(path)
			}
		}
	}
}

func sortByModTime(a, b string) int {
	info_a, err := os.Stat(a)
	if err != nil {
		utils.PrintEf("stat %s: %v", a, err)
		return 0
	}
	info_b, err := os.Stat(b)
	if err != nil {
		utils.PrintEf("stat %s: %v", b, err)
		return 0
	}
	ta := info_a.ModTime()
	tb := info_b.ModTime()
	if ta.Equal(tb) {
		return 0
	}
	if ta.Before(tb) {
		return -1
	}
	return 1
}
