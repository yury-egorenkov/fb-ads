package optimization

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRateLimiter_Wait(t *testing.T) {
	limiter := NewRateLimiter()
	limiter.SetRequestInterval(100 * time.Millisecond)

	// Record start time
	start := time.Now()

	// First wait should not block since LastRequestTime is zero
	limiter.Wait()
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Errorf("First wait took %v, expected it to be nearly instant", elapsed)
	}

	// Second wait should block for approximately the interval time
	start = time.Now()
	limiter.Wait()
	elapsed := time.Since(start)
	if elapsed < 90*time.Millisecond || elapsed > 120*time.Millisecond {
		t.Errorf("Second wait took %v, expected around 100ms", elapsed)
	}
}

func TestRateLimiter_Execute_Success(t *testing.T) {
	ctx := context.Background()
	limiter := NewRateLimiter()
	// Set a very short interval for testing
	limiter.SetRequestInterval(1 * time.Millisecond)

	// Operation that succeeds on first try
	callCount := 0
	operation := func() error {
		callCount++
		return nil
	}

	err := limiter.Execute(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected operation to be called 1 time, got %d", callCount)
	}
}

func TestRateLimiter_Execute_Retries(t *testing.T) {
	ctx := context.Background()
	limiter := NewRateLimiter()
	// Set short times for testing
	limiter.SetRequestInterval(1 * time.Millisecond)
	limiter.SetBaseDelay(10 * time.Millisecond)
	limiter.SetMaxRetries(3)

	// Operation that fails twice, then succeeds
	callCount := 0
	operation := func() error {
		callCount++
		if callCount <= 2 {
			return errors.New("simulated error")
		}
		return nil
	}

	err := limiter.Execute(ctx, operation)
	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected operation to be called 3 times, got %d", callCount)
	}
}

func TestRateLimiter_Execute_MaxRetriesExceeded(t *testing.T) {
	ctx := context.Background()
	limiter := NewRateLimiter()
	// Set short times for testing
	limiter.SetRequestInterval(1 * time.Millisecond)
	limiter.SetBaseDelay(10 * time.Millisecond)
	limiter.SetMaxRetries(2)

	// Operation that always fails
	callCount := 0
	operation := func() error {
		callCount++
		return errors.New("simulated error")
	}

	err := limiter.Execute(ctx, operation)
	if err == nil {
		t.Error("Expected an error after max retries, got nil")
	}

	// Should be called MaxRetries + 1 times (initial attempt + retries)
	if callCount != 3 {
		t.Errorf("Expected operation to be called 3 times, got %d", callCount)
	}
}

func TestRateLimiter_Execute_CancelContext(t *testing.T) {
	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	limiter := NewRateLimiter()
	// Set long times to ensure we're testing cancellation
	limiter.SetRequestInterval(10 * time.Millisecond)
	limiter.SetBaseDelay(500 * time.Millisecond)

	// Operation that always fails so it will retry
	callCount := 0
	operation := func() error {
		callCount++
		// Cancel after first call
		if callCount == 1 {
			cancel()
		}
		return errors.New("simulated error")
	}

	err := limiter.Execute(ctx, operation)
	if err == nil {
		t.Error("Expected an error after context cancellation, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	// Should be called only once before cancellation
	if callCount != 1 {
		t.Errorf("Expected operation to be called once, got %d", callCount)
	}
}

func TestRateLimiter_CalculateBackoff(t *testing.T) {
	limiter := NewRateLimiter()
	limiter.SetBaseDelay(100 * time.Millisecond)
	limiter.SetMaxDelay(10 * time.Second)
	limiter.Jitter = 0 // Disable jitter for predictable testing

	// Test exponential growth
	backoff0 := limiter.calculateBackoff(0)
	if expected := 100 * time.Millisecond; backoff0 != expected {
		t.Errorf("Expected backoff of %v for retry 0, got %v", expected, backoff0)
	}

	backoff1 := limiter.calculateBackoff(1)
	if expected := 200 * time.Millisecond; backoff1 != expected {
		t.Errorf("Expected backoff of %v for retry 1, got %v", expected, backoff1)
	}

	backoff2 := limiter.calculateBackoff(2)
	if expected := 400 * time.Millisecond; backoff2 != expected {
		t.Errorf("Expected backoff of %v for retry 2, got %v", expected, backoff2)
	}

	// Test max delay cap
	backoff10 := limiter.calculateBackoff(10) // 100ms * 2^10 = 102.4s, should be capped at 10s
	if expected := 10 * time.Second; backoff10 != expected {
		t.Errorf("Expected backoff of %v for retry 10 (capped), got %v", expected, backoff10)
	}
}

func TestRateLimiter_CanMakeRequest(t *testing.T) {
	limiter := NewRateLimiter()
	limiter.SetRequestInterval(100 * time.Millisecond)

	// Initial state should allow a request
	if !limiter.CanMakeRequest() {
		t.Error("Initially, CanMakeRequest should return true")
	}

	// After a request, should return false temporarily
	limiter.Wait()
	if limiter.CanMakeRequest() {
		t.Error("Immediately after a request, CanMakeRequest should return false")
	}

	// After waiting the full interval, should return true again
	time.Sleep(100 * time.Millisecond)
	if !limiter.CanMakeRequest() {
		t.Error("After waiting the interval, CanMakeRequest should return true")
	}
}