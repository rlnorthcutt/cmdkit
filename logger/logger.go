package logger

import (
	"fmt"
	"io"
	"os"
)

// Logger provides colored, level-aware output for CLI tools.
// No global state; create with New and pass through your execute flow.
type Logger struct {
	isVerbose bool
	isQuiet   bool
	out       io.Writer
	err       io.Writer
}

// New returns a Logger writing to os.Stdout and os.Stderr.
// When isVerbose is true, Detail calls produce output.
func New(isVerbose bool) *Logger {
	return &Logger{
		isVerbose: isVerbose,
		out:       os.Stdout,
		err:       os.Stderr,
	}
}

// WithQuiet suppresses Info, Success, and Detail output. Warn, Error, Fatal, and Debug are unaffected.
func (l *Logger) WithQuiet() *Logger {
	l.isQuiet = true
	return l
}

// WithWriters replaces the default os.Stdout and os.Stderr writers.
// Useful for testing or routing output to a file.
func (l *Logger) WithWriters(out, err io.Writer) *Logger {
	l.out = out
	l.err = err
	return l
}

// Print writes a plain unformatted line to stdout. Use for output that needs no prefix or color.
func (l *Logger) Print(msg string, v ...any) {
	if !l.isQuiet {
		fmt.Fprintf(l.out, "%s\n", fmt.Sprintf(msg, v...))
	}
}

// Detail prints only when the logger is verbose. Suppressed in quiet mode. Use for optional diagnostics.
func (l *Logger) Detail(msg string, v ...any) {
	if l.isVerbose && !l.isQuiet {
		fmt.Fprintf(l.out, "------- %s\n", fmt.Sprintf(msg, v...))
	}
}

// Info prints a cyan [INFO] line. Suppressed in quiet mode.
func (l *Logger) Info(msg string, v ...any) {
	if !l.isQuiet {
		fmt.Fprintf(l.out, "\u001B[0;36m[INFO]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
	}
}

// Warn prints a yellow [WARNING] line to stderr.
func (l *Logger) Warn(msg string, v ...any) {
	fmt.Fprintf(l.err, "\u001B[0;33m[WARNING]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
}

// Error prints a red [ERROR] line to stderr. Use Fatal when the process must exit.
func (l *Logger) Error(msg string, v ...any) {
	fmt.Fprintf(l.err, "\u001B[0;31m[ERROR]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
}

// Fatal logs the message to stderr and exits with code 1. Use in CLI orchestrator only.
func (l *Logger) Fatal(msg string, v ...any) {
	fmt.Fprintf(l.err, "\n\u001B[0;35m[FATAL]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
	os.Exit(1)
}

// Success prints a green [SUCCESS] line. Suppressed in quiet mode.
func (l *Logger) Success(msg string, v ...any) {
	if !l.isQuiet {
		fmt.Fprintf(l.out, "\u001B[1;32m[SUCCESS]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
	}
}

// Debug prints a red debug line. Always prints regardless of isVerbose or quiet mode —
// use Detail for verbose-gated output. Only use in development.
func (l *Logger) Debug(msg string, v ...any) {
	fmt.Fprintf(l.out, "\u001B[0;31m[--dEbUg--]\u001B[0;39m %s\n", fmt.Sprintf(msg, v...))
}
