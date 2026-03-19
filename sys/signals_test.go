package sys_test

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/rlnorthcutt/cmdkit/sys"
)

func TestContextWithInterrupt_stopCancels(t *testing.T) {
	ctx, stop := sys.ContextWithInterrupt(context.Background())
	defer stop()

	select {
	case <-ctx.Done():
		t.Fatal("context should not be cancelled before stop is called")
	default:
	}

	stop()

	select {
	case <-ctx.Done():
		// pass
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context should be cancelled after stop()")
	}
}

func TestContextWithInterrupt_parentCancelPropagates(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	ctx, stop := sys.ContextWithInterrupt(parent)
	defer stop()

	cancel()

	select {
	case <-ctx.Done():
		// pass
	case <-time.After(100 * time.Millisecond):
		t.Fatal("child context should cancel when parent is cancelled")
	}
}

func TestContextWithInterrupt_stopIdempotent(t *testing.T) {
	_, stop := sys.ContextWithInterrupt(context.Background())
	stop()
	stop() // must not panic
}

// Signal tests must not run in parallel — signal delivery is process-wide.

func TestContextWithInterrupt_SIGINT(t *testing.T) {
	ctx, stop := sys.ContextWithInterrupt(context.Background())
	defer stop()

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("could not find process: %v", err)
	}
	proc.Signal(os.Interrupt)

	select {
	case <-ctx.Done():
		// pass
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context not cancelled after SIGINT")
	}
}

func TestContextWithInterrupt_SIGTERM(t *testing.T) {
	ctx, stop := sys.ContextWithInterrupt(context.Background())
	defer stop()

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("could not find process: %v", err)
	}
	proc.Signal(syscall.SIGTERM)

	select {
	case <-ctx.Done():
		// pass
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context not cancelled after SIGTERM")
	}
}
