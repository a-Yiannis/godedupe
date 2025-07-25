package utils

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

// StringSlice is a custom flag type to handle multiple string flags
type StringSlice []string

func (i *StringSlice) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *StringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var autoYes bool

func SetAutoYes(val bool) {
	autoYes = val
}

func AskStrict(prompt string) bool {
	return askSimple(prompt, true)
}
func Ask(prompt string) bool {
	return askSimple(prompt, false)
}

// askSimple prompts the user with a yes/no question and waits for a single 'y' or 'n' keypress.
func askSimple(prompt string, strict bool) bool {
	if autoYes {
		return true
	}
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
		PrintEf("Error reading from input: %v", err)
	}
	return ""
}

