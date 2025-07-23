package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

func RecycleFile(path string) error {
	return trash.Throw(path)
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

// NormalizeWindowsPath cleans p, replaces "/"→"\\", lowercases.
func NormalizeWindowsPath(p string) string {
	p = filepath.Clean(p)
	p = strings.ReplaceAll(p, "/", "\\")
	return strings.ToLower(p)
}

func WriteRed(msg string) {
	fmt.Print(red + msg + reset)
}

func WriteCyan(msg string) {
	fmt.Print(cyan + msg + reset)
}
