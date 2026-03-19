package logger_test

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/rlnorthcutt/cmdkit/logger"
)

// newLogger returns a logger with buffers wired for stdout and stderr.
func newLogger(verbose bool) (*logger.Logger, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}
	return logger.New(verbose).WithWriters(out, err), out, err
}

// --- Destination: stdout vs stderr ---

func TestInfo_stdout(t *testing.T) {
	log, out, err := newLogger(false)
	log.Info("hello")
	if !strings.Contains(out.String(), "[INFO]") || !strings.Contains(out.String(), "hello") {
		t.Errorf("expected [INFO] and message on stdout, got: %q", out.String())
	}
	if err.String() != "" {
		t.Errorf("expected nothing on stderr, got: %q", err.String())
	}
}

func TestSuccess_stdout(t *testing.T) {
	log, out, err := newLogger(false)
	log.Success("done")
	if !strings.Contains(out.String(), "[SUCCESS]") || !strings.Contains(out.String(), "done") {
		t.Errorf("expected [SUCCESS] and message on stdout, got: %q", out.String())
	}
	if err.String() != "" {
		t.Errorf("expected nothing on stderr, got: %q", err.String())
	}
}

func TestWarn_stderr(t *testing.T) {
	log, out, err := newLogger(false)
	log.Warn("watch out")
	if out.String() != "" {
		t.Errorf("expected nothing on stdout for Warn, got: %q", out.String())
	}
	if !strings.Contains(err.String(), "[WARNING]") || !strings.Contains(err.String(), "watch out") {
		t.Errorf("expected [WARNING] and message on stderr, got: %q", err.String())
	}
}

func TestError_stderr(t *testing.T) {
	log, out, err := newLogger(false)
	log.Error("bad thing")
	if out.String() != "" {
		t.Errorf("expected nothing on stdout for Error, got: %q", out.String())
	}
	if !strings.Contains(err.String(), "[ERROR]") || !strings.Contains(err.String(), "bad thing") {
		t.Errorf("expected [ERROR] and message on stderr, got: %q", err.String())
	}
}

// --- Verbose / quiet gating ---

func TestDetail_verbose(t *testing.T) {
	log, out, _ := newLogger(true)
	log.Detail("details")
	if !strings.Contains(out.String(), "details") {
		t.Errorf("Detail should print when verbose, got: %q", out.String())
	}
}

func TestDetail_nonVerbose(t *testing.T) {
	log, out, _ := newLogger(false)
	log.Detail("details")
	if out.String() != "" {
		t.Errorf("Detail should not print when not verbose, got: %q", out.String())
	}
}

func TestDebug_alwaysPrints(t *testing.T) {
	log, out, _ := newLogger(false)
	log.Debug("dbg")
	if !strings.Contains(out.String(), "dbg") {
		t.Errorf("Debug should print even when not verbose, got: %q", out.String())
	}
}

func TestQuiet_suppressesInfoSuccessDetail(t *testing.T) {
	log, out, err := newLogger(true)
	log.WithQuiet()
	log.Info("info msg")
	log.Success("success msg")
	log.Detail("detail msg")
	if out.String() != "" {
		t.Errorf("quiet mode should suppress Info/Success/Detail, got stdout: %q", out.String())
	}
	if err.String() != "" {
		t.Errorf("unexpected stderr output: %q", err.String())
	}
}

func TestQuiet_doesNotSuppressWarnErrorDebug(t *testing.T) {
	log, out, err := newLogger(false)
	log.WithQuiet()
	log.Warn("warn msg")
	log.Error("error msg")
	log.Debug("debug msg")
	if !strings.Contains(err.String(), "[WARNING]") {
		t.Errorf("Warn should not be suppressed in quiet mode, got: %q", err.String())
	}
	if !strings.Contains(err.String(), "[ERROR]") {
		t.Errorf("Error should not be suppressed in quiet mode, got: %q", err.String())
	}
	if !strings.Contains(out.String(), "[--dEbUg--]") {
		t.Errorf("Debug should not be suppressed in quiet mode, got: %q", out.String())
	}
}

// --- Print ---

func TestPrint_noPrefix(t *testing.T) {
	log, out, _ := newLogger(false)
	log.Print("plain line")
	if out.String() != "plain line\n" {
		t.Errorf("expected plain output, got: %q", out.String())
	}
}

func TestPrint_suppressedInQuiet(t *testing.T) {
	log, out, _ := newLogger(false)
	log.WithQuiet()
	log.Print("should be hidden")
	if out.String() != "" {
		t.Errorf("Print should be suppressed in quiet mode, got: %q", out.String())
	}
}

func TestPrint_format(t *testing.T) {
	log, out, _ := newLogger(false)
	log.Print("value: %d", 42)
	if !strings.Contains(out.String(), "value: 42") {
		t.Errorf("expected formatted output, got: %q", out.String())
	}
}

// --- Format string safety ---

func TestFormatString(t *testing.T) {
	log, out, _ := newLogger(false)
	log.Info("count: %d, name: %s", 3, "foo")
	if !strings.Contains(out.String(), "count: 3, name: foo") {
		t.Errorf("expected formatted message, got: %q", out.String())
	}
}

// --- Fatal (subprocess) ---

// TestFatal runs Fatal in a subprocess to safely test os.Exit(1) behavior.
func TestFatal(t *testing.T) {
	if os.Getenv("CMDKIT_TEST_FATAL") == "1" {
		logger.New(false).Fatal("boom")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFatal")
	cmd.Env = append(os.Environ(), "CMDKIT_TEST_FATAL=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "[FATAL]") {
		t.Errorf("expected [FATAL] in stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "boom") {
		t.Errorf("expected message in stderr, got: %q", stderr.String())
	}
}
