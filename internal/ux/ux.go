// Package ux provides terminal styling and user experience utilities
// for the DevForge CLI.
package ux

import (
	"fmt"
	"os"
)

// ANSI Color Codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Gray   = "\033[90m"
)

// Helper symbols
const (
	Check = "✔"
	Cross = "✖"
	Info  = "ℹ"
	Warn  = "⚠"
	Spin  = "⟳"
)

// Success prints a green success message.
func Success(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s%s %s%s\n", Green, Check, msg, Reset)
}

// Error prints a red error message. If the error is an actionable DevForgeError,
// it prints the structured format.
func Error(err error) {
	fmt.Printf("%s%s Error: %s%s\n", Red, Cross, err.Error(), Reset)
}

// Warning prints a yellow warning message.
func Warning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s%s %s%s\n", Yellow, Warn, msg, Reset)
}

// InfoMsg prints a blue informational message.
func InfoMsg(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s%s %s%s\n", Blue, Info, msg, Reset)
}

// Step prints a neutral step message with a spinner indicating progress.
func Step(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s...\n", Spin, msg)
}

// Print structed text.
func Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

func Println(a ...interface{}) {
	fmt.Println(a...)
}

func ExitWithMessage(code int, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s%s %s%s\n", Red, Cross, msg, Reset)
	os.Exit(code)
}
