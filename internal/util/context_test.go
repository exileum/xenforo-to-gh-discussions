package util

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestContextSleep(t *testing.T) {
	t.Run("successful sleep", func(t *testing.T) {
		ctx := context.Background()
		start := time.Now()
		err := ContextSleep(ctx, 100*time.Millisecond)
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if elapsed < 100*time.Millisecond {
			t.Errorf("Sleep too short: %v", elapsed)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		start := time.Now()
		err := ContextSleep(ctx, 1*time.Second)
		elapsed := time.Since(start)

		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		if !strings.Contains(err.Error(), "sleep cancelled") {
			t.Errorf("Expected 'sleep cancelled' in error, got: %v", err)
		}

		// Should return immediately when context is already cancelled
		if elapsed > 100*time.Millisecond {
			t.Errorf("Should return immediately on cancelled context, took: %v", elapsed)
		}
	})

	t.Run("context cancelled during sleep", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		// Cancel after 50ms
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		start := time.Now()
		err := ContextSleep(ctx, 1*time.Second)
		elapsed := time.Since(start)

		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		// Should be cancelled around 50ms, not wait the full second
		if elapsed > 200*time.Millisecond {
			t.Errorf("Sleep should be interrupted, took: %v", elapsed)
		}
	})

	t.Run("context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := ContextSleep(ctx, 1*time.Second)
		elapsed := time.Since(start)

		if err == nil {
			t.Error("Expected error from timeout context")
		}

		// Should timeout around 50ms
		if elapsed > 200*time.Millisecond {
			t.Errorf("Sleep should timeout, took: %v", elapsed)
		}
	})
}

func TestWrapContextError(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		ctx := context.Background()
		err := WrapContextError(ctx, "test operation")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := WrapContextError(ctx, "test operation")
		if err == nil {
			t.Error("Expected error from cancelled context")
		}

		if !strings.Contains(err.Error(), "test operation cancelled") {
			t.Errorf("Expected 'test operation cancelled' in error, got: %v", err)
		}

		if !strings.Contains(err.Error(), "context canceled") {
			t.Errorf("Expected underlying context error in message, got: %v", err)
		}
	})

	t.Run("timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Wait for timeout
		time.Sleep(10 * time.Millisecond)

		err := WrapContextError(ctx, "database query")
		if err == nil {
			t.Error("Expected error from timeout context")
		}

		if !strings.Contains(err.Error(), "database query cancelled") {
			t.Errorf("Expected 'database query cancelled' in error, got: %v", err)
		}
	})
}
