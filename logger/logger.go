package logger

import (
	"fmt"
	"os"
)

// Logger provides colored, level-aware output for CLI tools.
// No global state; create with New and pass through your execute flow.
type Logger struct {
	isVerbose bool
}

// New returns a Logger. When isVerbose is true, Detail calls produce output.
func New(isVerbose bool) *Logger {
	return &Logger{isVerbose: isVerbose}
}

// Detail prints only when the logger is verbose. Use for optional diagnostics.
func (l *Logger) Detail(msg string, v ...any) {
	if l.isVerbose {
		fmt.Printf("------- %s\n", fmt.Sprintf(msg, v...))
	}
}

// Info prints a cyan [INFO] line.
func (l *Logger) Info(msg string, v ...any) {
	fmt.Printf("\u001B[0;36m[INFO]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
}

// Warn prints a yellow [WARNING] line to stderr.
func (l *Logger) Warn(msg string, v ...any) {
	fmt.Fprintf(os.Stderr, "\u001B[0;33m[WARNING]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
}

// Error prints a red [ERROR] line to stderr. Use Fatal when the process must exit.
func (l *Logger) Error(msg string, v ...any) {
	fmt.Fprintf(os.Stderr, "\u001B[0;31m[ERROR]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
}

// Fatal logs the message to stderr and exits with code 1. Use in CLI orchestrator only.
func (l *Logger) Fatal(msg string, v ...any) {
	fmt.Fprintf(os.Stderr, "\n\u001B[0;35m[FATAL]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
	os.Exit(1)
}

// Success prints a green [SUCCESS] line.
func (l *Logger) Success(msg string, v ...any) {
	fmt.Printf("\u001B[1;32m[SUCCESS]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
}

// Debug prints a red debug line. Always prints regardless of isVerbose — use Detail for verbose-gated output. Only use in development.
func (l *Logger) Debug(msg string, v ...any) {
	fmt.Printf("\u001B[0;31m[--dEbUg--]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
}
