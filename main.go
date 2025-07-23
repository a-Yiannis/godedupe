package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"
	"sync"
	"time"
)

func main() {
	start := time.Now()

	configFilePath := "config.json" // Default config file
	if len(os.Args) >= 2 {
		configFilePath = os.Args[1]
	}
	cfg, err := loadConfig(configFilePath)
	if err != nil {
		log.Fatalf("loadConfig: %v", err)
	}
	if cfg.Root == "" {
		cfg.Root, err = os.Getwd()
		if err != nil {
			log.Fatalf("Getwd: %v", err)
		}
	}
	fmt.Printf("Scanning %s …\n", cfg.Root)

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
					log.Printf("partialHash %s: %v", p, err)
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
						log.Printf("fullHash %s: %v", p, err)
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
	reportDuplicates(dupMap)

	elapsed := time.Since(start)
	fmt.Printf("Elapsed: %s\n", elapsed)

	if AskSimple("Should I recycle the duplicates?") {
		recycle(dupMap)
	}
}

func recycle(dupMap map[uint64][]string) {
	f, err := os.OpenFile("recycled.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("open log: %v", err)
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
		fmt.Println(s)
		if !AskSimple("Going to recycle this files, are you sure?") {
			fmt.Println("File set skipped!")
			return
		}
		for _, path := range paths[1:] {
			fmt.Printf("Recycling '%s' \n", path)
			err := RecycleFile(path)
			if err != nil {
				log.Printf("RecycleFile %s: %v", path, err)
			} else {
				logger.Println(path)
			}
		}
	}
}

func sortByModTime(a, b string) int {
	info_a, err := os.Stat(a)
	if err != nil {
		log.Printf("stat %s: %v", a, err)
		return 0
	}
	info_b, err := os.Stat(b)
	if err != nil {
		log.Printf("stat %s: %v", b, err)
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
