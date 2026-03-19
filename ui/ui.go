package ui

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/rlnorthcutt/cmdkit/logger"
)

// parseBoolEnv returns true for "true", "1", "y", or "yes" (case-insensitive).
func parseBoolEnv(s string) bool {
	switch strings.ToLower(s) {
	case "true", "1", "yes", "y":
		return true
	}
	return false
}

// UI holds interactive/TTY state for prompts and resolution. No global state.
// Ctx is always safe to pass to long-running operations; it is context.Background() until
// WithInterrupt is called, after which it cancels on SIGINT or SIGTERM.
type UI struct {
	Interactive bool
	Ctx         context.Context
	log         *logger.Logger
	reader      *bufio.Reader
	stopSignal  context.CancelFunc
}

// New builds a UI from the process environment. When nonInteractiveFlag is true or stdin is not a TTY, Interactive is false.
// The returned UI has no logger attached; use WithLogger to connect one.
func New(nonInteractiveFlag bool) *UI {
	stat, _ := os.Stdin.Stat()
	isTTY := (stat.Mode() & os.ModeCharDevice) != 0
	return &UI{
		Interactive: isTTY && !nonInteractiveFlag,
		Ctx:         context.Background(),
		reader:      bufio.NewReader(os.Stdin),
	}
}

// WithLogger attaches a logger to the UI and returns the same instance for chaining.
func (u *UI) WithLogger(log *logger.Logger) *UI {
	u.log = log
	return u
}

// WithInterrupt registers SIGINT/SIGTERM handlers and replaces ui.Ctx with a context that
// cancels when a signal is received. Pass ui.Ctx to long-running operations so they respect Ctrl+C.
// Call StopSignal (typically via defer) to release the signal handler when the command exits.
func (u *UI) WithInterrupt(parent context.Context) *UI {
	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	u.Ctx = ctx
	u.stopSignal = stop
	return u
}

// StopSignal releases the signal handler registered by WithInterrupt. Safe to call if WithInterrupt was not used.
func (u *UI) StopSignal() {
	if u.stopSignal != nil {
		u.stopSignal()
	}
}

func (u *UI) detail(msg string, v ...any) {
	if u.log != nil {
		u.log.Detail(msg, v...)
	}
}

// ResolveString applies precedence: Flag > Env > Prompt > Default.
// Pass the flag's current value and whether it was explicitly set by the user;
// these can come from any flag library (cobra, pflag, stdlib flag, etc).
// value is the default and is overwritten by the chosen source.
func (u *UI) ResolveString(flagValue string, flagSet bool, envKey, promptMsg string, value *string) {
	if flagSet {
		*value = flagValue
		u.detail("resolved from flag: %s", flagValue)
		return
	}

	if env := os.Getenv(envKey); env != "" {
		*value = env
		u.detail("resolved from env %s", envKey)
		return
	}

	if u.Interactive {
		prompt := promptMsg
		if *value != "" {
			prompt = fmt.Sprintf("%s (default: %s)", promptMsg, *value)
		}
		*value = u.Prompt(prompt, *value)
		u.detail("resolved from prompt")
		return
	}

	u.detail("using default: %s", *value)
}

// ResolveBool applies precedence: Flag > Env > Default.
// There is no prompt tier for booleans — use Confirm for interactive boolean input.
// Env values of "true", "1", or "yes" (case-insensitive) are treated as true; anything else is false.
func (u *UI) ResolveBool(flagValue bool, flagSet bool, envKey string, value *bool) {
	if flagSet {
		*value = flagValue
		u.detail("resolved from flag: %v", flagValue)
		return
	}

	if env := os.Getenv(envKey); env != "" {
		*value = parseBoolEnv(env)
		u.detail("resolved from env %s", envKey)
		return
	}

	u.detail("using default: %v", *value)
}

// Prompt reads a single line from stdin with an optional default. Returns defaultValue if input is empty.
func (u *UI) Prompt(message string, defaultValue string) string {
	fmt.Printf("%s: ", message)
	input, err := u.reader.ReadString('\n')
	if err != nil {
		if u.log != nil {
			u.log.Error("reading input: %v", err)
		}
		return defaultValue
	}
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

// Confirm prompts for y/n. When not interactive, returns true without prompting so automation proceeds unblocked.
func (u *UI) Confirm(message string) bool {
	if !u.Interactive {
		u.detail("auto-confirming '%s' because non-interactive", message)
		return true
	}
	res := u.Prompt(message+" (y/n)", "y")
	return strings.ToLower(res) == "y"
}
