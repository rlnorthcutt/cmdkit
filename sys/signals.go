package sys

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// ContextWithInterrupt returns a context cancelled when SIGINT or SIGTERM is received.
// Use this when you want signal-aware cancellation without a UI (e.g. background workers or daemons).
// For interactive CLI tools, prefer ui.WithInterrupt which integrates with the UI session.
func ContextWithInterrupt(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}
