// Package retry provides retry logic with exponential backoff and jitter
package retry

import (
	"math/rand"
	"time"
)

// Config holds the configuration for retry logic
type Config struct {
	// MaxRetries is the maximum number of retries
	MaxRetries int
	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the multiplier for exponential backoff
	Multiplier float64
	// Jitter is the maximum jitter to add to the delay
	Jitter time.Duration
}

// DefaultConfig returns a default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		Multiplier:    2.0,
		Jitter:        1 * time.Second,
	}
}

// WithRetry executes the given function with retry logic
func WithRetry(config Config, fn func() error) error {
	delay := config.InitialDelay
	
	for i := 0; i <= config.MaxRetries; i++ {
		// Execute the function
		err := fn()
		if err == nil {
			// Success, no need to retry
			return nil
		}
		
		// If this was the last attempt, return the error
		if i == config.MaxRetries {
			return err
		}
		
		// Add jitter to the delay
		jitter := time.Duration(rand.Int63n(int64(config.Jitter*2))) - config.Jitter
		delayWithJitter := delay + jitter
		
		// Ensure delay is within bounds
		if delayWithJitter < 0 {
			delayWithJitter = 0
		}
		if delayWithJitter > config.MaxDelay {
			delayWithJitter = config.MaxDelay
		}
		
		// Wait before retrying
		time.Sleep(delayWithJitter)
		
		// Calculate next delay
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}
	
	return nil
}