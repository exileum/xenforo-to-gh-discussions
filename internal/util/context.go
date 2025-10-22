// Package util provides utility functions for common operations.
package util

import (
	"context"
	"fmt"
	"time"
)

// ContextSleep performs a context-aware sleep operation.
// Returns an error if the context is cancelled before the sleep completes.
func ContextSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("sleep cancelled: %w", ctx.Err())
	case <-time.After(d):
		return nil
	}
}

// WrapContextError wraps a context error with a descriptive message.
func WrapContextError(ctx context.Context, operation string) error {
	if ctx.Err() == nil {
		return nil
	}
	return fmt.Errorf("%s cancelled: %w", operation, ctx.Err())
}
