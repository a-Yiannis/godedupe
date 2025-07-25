package utils

import (
	"fmt"
	"regexp"
)

const (
	// ANSI color codes
	red   = "\033[31m"
	cyan  = "\033[36m"
	reset = "\033[0m"
)

var emphaticPattern = regexp.MustCompile(`\*\*(.*?)\*\*|__(.*?)__`)
var emphaticSubPattern = "\x1b[31m$1\x1b[0m"

func Println(s string) {
	colored := emphaticPattern.ReplaceAllString(s, emphaticSubPattern)
	fmt.Println(colored)
}
func Printf(format string, args ...any) {
	// First format the string with the provided arguments
	formatted := fmt.Sprintf(format, args...)
	// Then apply the emphasis coloring
	colored := emphaticPattern.ReplaceAllString(formatted, emphaticSubPattern)
	fmt.Println(colored)
}

func WriteRed(msg string) {
	fmt.Print(red + msg + reset)
}

func WriteCyan(msg string) {
	fmt.Print(cyan + msg + reset)
}