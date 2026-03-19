package ui_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rlnorthcutt/cmdkit/ui"
)

// newUIWithInput replaces os.Stdin with a pipe containing the given input,
// creates a UI (so its internal reader wraps the pipe), then restores os.Stdin.
// Must not be used with t.Parallel() as it temporarily replaces os.Stdin.
func newUIWithInput(nonInteractive bool, input string) *ui.UI {
	r, w, _ := os.Pipe()
	fmt.Fprint(w, input)
	w.Close()

	orig := os.Stdin
	os.Stdin = r
	u := ui.New(nonInteractive)
	os.Stdin = orig

	return u
}

// --- New ---

func TestNew_nonInteractiveFlag(t *testing.T) {
	u := ui.New(true)
	if u.Interactive {
		t.Error("expected Interactive=false when nonInteractiveFlag=true")
	}
}

func TestNew_ctxNotNil(t *testing.T) {
	u := ui.New(true)
	if u.Ctx == nil {
		t.Error("Ctx should not be nil after New")
	}
}

// --- WithInterrupt / StopSignal ---

func TestWithInterrupt_replacesCtx(t *testing.T) {
	u := ui.New(true)
	before := u.Ctx
	u.WithInterrupt(context.Background())
	if u.Ctx == before {
		t.Error("WithInterrupt should replace Ctx with a new context")
	}
}

func TestWithInterrupt_stopCancelsCtx(t *testing.T) {
	u := ui.New(true).WithInterrupt(context.Background())

	select {
	case <-u.Ctx.Done():
		t.Fatal("context should not be cancelled before StopSignal")
	default:
	}

	u.StopSignal()

	select {
	case <-u.Ctx.Done():
		// pass
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context should be cancelled after StopSignal")
	}
}

func TestStopSignal_safeWithoutWithInterrupt(t *testing.T) {
	u := ui.New(true)
	u.StopSignal() // must not panic
}

// --- Prompt ---

func TestPrompt_returnsInput(t *testing.T) {
	u := newUIWithInput(false, "my input\n")
	result := u.Prompt("Enter value", "default")
	if result != "my input" {
		t.Errorf("expected %q, got %q", "my input", result)
	}
}

func TestPrompt_emptyReturnsDefault(t *testing.T) {
	u := newUIWithInput(false, "\n")
	result := u.Prompt("Enter value", "default")
	if result != "default" {
		t.Errorf("expected default %q, got %q", "default", result)
	}
}

func TestPrompt_whitespaceIsTrimmed(t *testing.T) {
	u := newUIWithInput(false, "  trimmed  \n")
	result := u.Prompt("Enter value", "default")
	if result != "trimmed" {
		t.Errorf("expected trimmed input, got %q", result)
	}
}

func TestPrompt_eofReturnsDefault(t *testing.T) {
	// Pipe with no input — EOF immediately
	u := newUIWithInput(false, "")
	result := u.Prompt("Enter value", "fallback")
	if result != "fallback" {
		t.Errorf("expected fallback on EOF, got %q", result)
	}
}

// --- Confirm ---

func TestConfirm_nonInteractive_alwaysTrue(t *testing.T) {
	u := ui.New(true) // nonInteractive=true
	if !u.Confirm("Proceed?") {
		t.Error("Confirm should return true when non-interactive")
	}
}

func TestConfirm_y(t *testing.T) {
	u := newUIWithInput(false, "y\n")
	u.Interactive = true
	if !u.Confirm("Proceed?") {
		t.Error("expected true for 'y'")
	}
}

func TestConfirm_Y_caseInsensitive(t *testing.T) {
	u := newUIWithInput(false, "Y\n")
	u.Interactive = true
	if !u.Confirm("Proceed?") {
		t.Error("expected true for 'Y'")
	}
}

func TestConfirm_n(t *testing.T) {
	u := newUIWithInput(false, "n\n")
	u.Interactive = true
	if u.Confirm("Proceed?") {
		t.Error("expected false for 'n'")
	}
}

func TestConfirm_emptyDefaultsToY(t *testing.T) {
	u := newUIWithInput(false, "\n")
	u.Interactive = true
	if !u.Confirm("Proceed?") {
		t.Error("expected true when user presses Enter (default is 'y')")
	}
}

func TestConfirm_otherInputReturnsFalse(t *testing.T) {
	u := newUIWithInput(false, "q\n")
	u.Interactive = true
	if u.Confirm("Proceed?") {
		t.Error("expected false for non-y input")
	}
}

// --- ResolveString ---

func TestResolveString_flagWins(t *testing.T) {
	t.Setenv("INPUT_KEY", "from-env")

	u := ui.New(true)
	value := "default"
	u.ResolveString("from-flag", true, "INPUT_KEY", "Enter input", &value)

	if value != "from-flag" {
		t.Errorf("expected flag value, got %q", value)
	}
}

func TestResolveString_envWins(t *testing.T) {
	t.Setenv("INPUT_KEY", "from-env")

	u := ui.New(true)
	value := "default"
	u.ResolveString("", false, "INPUT_KEY", "Enter input", &value)

	if value != "from-env" {
		t.Errorf("expected env value, got %q", value)
	}
}

func TestResolveString_defaultUsed(t *testing.T) {
	t.Setenv("INPUT_KEY", "")

	u := ui.New(true) // non-interactive, no prompt
	value := "my-default"
	u.ResolveString("", false, "INPUT_KEY", "Enter input", &value)

	if value != "my-default" {
		t.Errorf("expected default value, got %q", value)
	}
}

func TestResolveString_promptUsed(t *testing.T) {
	t.Setenv("INPUT_KEY", "")

	u := newUIWithInput(false, "from-prompt\n")
	u.Interactive = true // force interactive since test runner stdin is not a TTY

	value := "default"
	u.ResolveString("", false, "INPUT_KEY", "Enter input", &value)

	if value != "from-prompt" {
		t.Errorf("expected prompt value, got %q", value)
	}
}

func TestResolveString_noLogger_noRace(t *testing.T) {
	u := ui.New(true) // no WithLogger
	value := "default"
	u.ResolveString("", false, "INPUT_KEY", "Enter input", &value) // must not panic
}
