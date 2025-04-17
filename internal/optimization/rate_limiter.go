package optimization

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RateLimiter manages API rate limiting with exponential backoff
type RateLimiter struct {
	// Base delay for backoff (in milliseconds)
	BaseDelay time.Duration

	// Maximum delay for backoff (in milliseconds)
	MaxDelay time.Duration

	// Maximum number of retries
	MaxRetries int

	// Jitter factor (0-1) to add randomness to backoff timing
	Jitter float64

	// Time of last request
	LastRequestTime time.Time

	// Minimum time between requests (in milliseconds)
	MinRequestInterval time.Duration
}

// NewRateLimiter creates a new rate limiter with default settings
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		BaseDelay:          100 * time.Millisecond,
		MaxDelay:           30 * time.Second,
		MaxRetries:         5,
		Jitter:             0.2,
		MinRequestInterval: 200 * time.Millisecond,
	}
}

// Wait blocks until it's safe to make another request
func (r *RateLimiter) Wait() {
	// Calculate time since last request
	timeSinceLast := time.Since(r.LastRequestTime)
	
	// If we need to wait to respect the minimum interval
	if timeSinceLast < r.MinRequestInterval {
		waitTime := r.MinRequestInterval - timeSinceLast
		time.Sleep(waitTime)
	}
	
	// Update last request time
	r.LastRequestTime = time.Now()
}

// Execute executes a function with rate limiting and exponential backoff
func (r *RateLimiter) Execute(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for retry := 0; retry <= r.MaxRetries; retry++ {
		// Wait for rate limiting before attempting operation
		r.Wait()
		
		// Execute the operation
		err := operation()
		if err == nil {
			return nil // Success
		}
		
		// Store the error
		lastErr = err
		
		// Check if context is cancelled before retrying
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		default:
			// Continue with retry
		}
		
		// If this was the last retry, don't wait again
		if retry == r.MaxRetries {
			break
		}
		
		// Calculate backoff delay
		backoffDelay := r.calculateBackoff(retry)
		
		// Log or notify about the retry
		fmt.Printf("Rate limit exceeded or error occurred. Retrying in %.2f seconds. Error: %v\n", 
			backoffDelay.Seconds(), err)
		
		// Wait for backoff period
		select {
		case <-time.After(backoffDelay):
			// Continue with retry
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled during backoff: %w", ctx.Err())
		}
	}
	
	return fmt.Errorf("operation failed after %d retries: %w", r.MaxRetries, lastErr)
}

// calculateBackoff calculates the backoff duration with jitter
func (r *RateLimiter) calculateBackoff(retry int) time.Duration {
	// Calculate exponential backoff: baseDelay * 2^retry
	delay := float64(r.BaseDelay.Milliseconds()) * math.Pow(2, float64(retry))
	
	// Apply maximum cap
	if delay > float64(r.MaxDelay.Milliseconds()) {
		delay = float64(r.MaxDelay.Milliseconds())
	}
	
	// Apply jitter (random variance between delay and delay*(1+jitter))
	jitterFactor := 1.0 + (rand.Float64() * r.Jitter)
	delay = delay * jitterFactor
	
	return time.Duration(delay) * time.Millisecond
}

// CanMakeRequest checks if a request can be made without waiting
func (r *RateLimiter) CanMakeRequest() bool {
	return time.Since(r.LastRequestTime) >= r.MinRequestInterval
}

// SetRequestInterval sets the minimum time between requests
func (r *RateLimiter) SetRequestInterval(interval time.Duration) {
	r.MinRequestInterval = interval
}

// SetMaxRetries sets the maximum number of retries
func (r *RateLimiter) SetMaxRetries(retries int) {
	r.MaxRetries = retries
}

// SetBaseDelay sets the base delay for exponential backoff
func (r *RateLimiter) SetBaseDelay(delay time.Duration) {
	r.BaseDelay = delay
}

// SetMaxDelay sets the maximum delay for exponential backoff
func (r *RateLimiter) SetMaxDelay(delay time.Duration) {
	r.MaxDelay = delay
}