package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/hymkor/trash-go"
	"golang.org/x/sys/windows"
)

const (
	// ANSI color codes
	red   = "\033[31m"
	cyan  = "\033[36m"
	reset = "\033[0m"
)

var emphaticPattern = regexp.MustCompile(`\*\*(.*?)\*\*|__(.*?)__`)
var emphaticSubPattern = "\x1b[31m$1\x1b[0m"

func println(s string) {
	colored := emphaticPattern.ReplaceAllString(s, emphaticSubPattern)
	fmt.Println(colored)
}
func printf(format string, args ...interface{}) {
	// First format the string with the provided arguments
	formatted := fmt.Sprintf(format, args...)
	// Then apply the emphasis coloring
	colored := emphaticPattern.ReplaceAllString(formatted, emphaticSubPattern)
	fmt.Println(colored)
}

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
		WriteRed(fmt.Sprintf("Failed to recycle %s: %v\n", path, err))
	}
	return err
}

func AskStrict(prompt string) bool {
	return askSimple(prompt, true)
}
func Ask(prompt string) bool {
	return askSimple(prompt, false)
}

// askSimple prompts the user with a yes/no question and waits for a single 'y' or 'n' keypress.
func askSimple(prompt string, strict bool) bool {
	h := windows.Handle(syscall.Stdin)
	var orig uint32
	if err := windows.GetConsoleMode(h, &orig); err != nil {
		panic(err)
	}
	// turn off line-buffering and echo
	noLine := orig &^ (windows.ENABLE_LINE_INPUT | windows.ENABLE_ECHO_INPUT)
	if err := windows.SetConsoleMode(h, noLine); err != nil {
		panic(err)
	}
	defer windows.SetConsoleMode(h, orig)

	fmt.Printf("%s ["+cyan+"y"+reset+"/"+cyan+"n"+reset+"] ", prompt)
	b := make([]byte, 1)
	for {
		_, err := os.Stdin.Read(b)
		if err != nil {
			panic(err)
		}
		switch b[0] {
		case 'y', 'Y':
			WriteCyan(string(b) + "\n")
			return true
		case 'n', 'N':
			WriteCyan(string(b) + "\n")
			return false
		default:
			if !strict {
				return false
			}
		}
	}
}

// GetUserInput prompts the user for input and returns the entered text.
func GetUserInput(prompt string) string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(prompt)
	if scanner.Scan() {
		return scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading from input:", err)
	}
	return ""
}

func WriteRed(msg string) {
	fmt.Print(red + msg + reset)
}

func WriteCyan(msg string) {
	fmt.Print(cyan + msg + reset)
}
