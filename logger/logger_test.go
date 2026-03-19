package logger_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/rlnorthcutt/cmdkit/logger"
)

func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func captureStderr(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stderr
	os.Stderr = w
	f()
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestInfo_stdout(t *testing.T) {
	log := logger.New(false)
	out := captureStdout(func() { log.Info("hello") })
	if !strings.Contains(out, "[INFO]") {
		t.Errorf("expected [INFO] in stdout, got: %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected message in stdout, got: %q", out)
	}
}

func TestInfo_notOnStderr(t *testing.T) {
	log := logger.New(false)
	err := captureStderr(func() { log.Info("hello") })
	if err != "" {
		t.Errorf("expected nothing on stderr, got: %q", err)
	}
}

func TestSuccess_stdout(t *testing.T) {
	log := logger.New(false)
	out := captureStdout(func() { log.Success("done") })
	if !strings.Contains(out, "[SUCCESS]") {
		t.Errorf("expected [SUCCESS] in stdout, got: %q", out)
	}
	if !strings.Contains(out, "done") {
		t.Errorf("expected message in stdout, got: %q", out)
	}
}

func TestWarn_stderr(t *testing.T) {
	log := logger.New(false)
	out := captureStdout(func() { log.Warn("watch out") })
	err := captureStderr(func() { log.Warn("watch out") })
	if out != "" {
		t.Errorf("expected nothing on stdout for Warn, got: %q", out)
	}
	if !strings.Contains(err, "[WARNING]") {
		t.Errorf("expected [WARNING] in stderr, got: %q", err)
	}
	if !strings.Contains(err, "watch out") {
		t.Errorf("expected message in stderr, got: %q", err)
	}
}

func TestError_stderr(t *testing.T) {
	log := logger.New(false)
	out := captureStdout(func() { log.Error("bad thing") })
	err := captureStderr(func() { log.Error("bad thing") })
	if out != "" {
		t.Errorf("expected nothing on stdout for Error, got: %q", out)
	}
	if !strings.Contains(err, "[ERROR]") {
		t.Errorf("expected [ERROR] in stderr, got: %q", err)
	}
	if !strings.Contains(err, "bad thing") {
		t.Errorf("expected message in stderr, got: %q", err)
	}
}

func TestDebug_alwaysPrints(t *testing.T) {
	logQuiet := logger.New(false)
	out := captureStdout(func() { logQuiet.Debug("dbg") })
	if !strings.Contains(out, "dbg") {
		t.Errorf("Debug should print even when not verbose, got: %q", out)
	}
}

func TestDetail_verbose(t *testing.T) {
	log := logger.New(true)
	out := captureStdout(func() { log.Detail("details") })
	if !strings.Contains(out, "details") {
		t.Errorf("Detail should print when verbose, got: %q", out)
	}
}

func TestDetail_nonVerbose(t *testing.T) {
	log := logger.New(false)
	out := captureStdout(func() { log.Detail("details") })
	if out != "" {
		t.Errorf("Detail should not print when not verbose, got: %q", out)
	}
}

func TestFormatString(t *testing.T) {
	log := logger.New(false)
	out := captureStdout(func() { log.Info("count: %d, name: %s", 3, "foo") })
	if !strings.Contains(out, "count: 3, name: foo") {
		t.Errorf("expected formatted message in output, got: %q", out)
	}
}

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
