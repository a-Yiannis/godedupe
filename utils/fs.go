package utils

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hymkor/trash-go"
)

var handleSlashes func(string) string

func init() {
	if runtime.GOOS == "windows" {
		handleSlashes = func(p string) string {
			return strings.ReplaceAll(p, "/", "\\")
		}
	} else {
		handleSlashes = func(p string) string {
			return filepath.ToSlash(p)
		}
	}
}

func NormalizePath(p string) string {
	p = filepath.Clean(p)
	p = handleSlashes(p)
	return strings.ToLower(p)
}

func RecycleFile(path string) error {
	err := trash.Throw(path)
	if err != nil {
		PrintEf("Failed to recycle %s: %v", path, err)
	}
	return err
}
