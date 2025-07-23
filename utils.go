package main

import (
	"bufio"
	"fmt"
	"os"
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

// makeSet creates a set (map[string]bool) from a slice of strings for efficient lookups.
func makeSet(keys []string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}

// lowerSlice converts all strings in a slice to lowercase.
func lowerSlice(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = strings.ToLower(s)
	}
	return out
}
