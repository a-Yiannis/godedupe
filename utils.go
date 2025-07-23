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

func RecycleFile(path string) error {
	return trash.Throw(path)
}

// AskSimple prompts the user with a yes/no question and waits for a single 'y' or 'n' keypress.

func AskSimple(prompt string) bool {
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

	fmt.Printf("%s [y/n] ", prompt)
	b := make([]byte, 1)
	if _, err := os.Stdin.Read(b); err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	return b[0] == 'y' || b[0] == 'Y'
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
