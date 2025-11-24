// Package circuitbreaker provides a generic circuit breaker implementation
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the state of the circuit breaker
type State int

const (
	// Closed state - requests are allowed
	Closed State = iota
	// Open state - requests are blocked
	Open
	// HalfOpen state - limited requests are allowed to test if service is recovering
	HalfOpen
)

// String returns the string representation of a State
func (s State) String() string {
	switch s {
	case Closed:
		return "Closed"
	case Open:
		return "Open"
	case HalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	// Configuration
	maxFailures    int
	timeout        time.Duration
	halfOpenDelay  time.Duration
	halfOpenMaxRequests int

	// State
	state          State
	failures       int
	lastFailure    time.Time
	halfOpenSuccess int
	mutex          sync.Mutex

	// Statistics
	totalRequests  int64
	totalFailures  int64
	totalSuccesses int64
}

// Config holds the configuration for a CircuitBreaker
type Config struct {
	// MaxFailures is the number of failures before the circuit opens
	MaxFailures int
	// Timeout is how long the circuit stays open before transitioning to half-open
	Timeout time.Duration
	// HalfOpenDelay is how long to wait in half-open state before allowing requests
	HalfOpenDelay time.Duration
	// HalfOpenMaxRequests is the maximum number of requests allowed in half-open state
	HalfOpenMaxRequests int
}

// New creates a new CircuitBreaker with the given configuration
func New(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:         config.MaxFailures,
		timeout:             config.Timeout,
		halfOpenDelay:       config.HalfOpenDelay,
		halfOpenMaxRequests: config.HalfOpenMaxRequests,
		state:               Closed,
	}
}

// Execute executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if we can execute the function
	if !cb.canExecute() {
		return errors.New("circuit breaker is open")
	}

	// Execute the function
	err := fn()

	// Update circuit breaker state based on result
	cb.updateState(err)

	return err
}

// canExecute checks if a request can be executed based on the current state
func (cb *CircuitBreaker) canExecute() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case Closed:
		// Always allow requests when closed
		return true
	case Open:
		// Check if timeout has elapsed
		if time.Since(cb.lastFailure) >= cb.timeout {
			// Transition to half-open
			cb.state = HalfOpen
			cb.halfOpenSuccess = 0
			return true
		}
		// Still open, block request
		return false
	case HalfOpen:
		// Allow limited requests
		if cb.halfOpenSuccess < cb.halfOpenMaxRequests {
			return true
		}
		// Reached max requests, block until state changes
		return false
	default:
		// Unknown state, block request
		return false
	}
}

// updateState updates the circuit breaker state based on the result of a request
func (cb *CircuitBreaker) updateState(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Update statistics
	cb.totalRequests++
	if err != nil {
		cb.totalFailures++
	} else {
		cb.totalSuccesses++
	}

	switch cb.state {
	case Closed:
		if err != nil {
			// Increment failure count
			cb.failures++
			cb.lastFailure = time.Now()
			
			// Check if we should open the circuit
			if cb.failures >= cb.maxFailures {
				cb.state = Open
			}
		} else {
			// Reset failure count on success
			cb.failures = 0
		}
	case Open:
		// Should not be possible to get here since canExecute would block requests
		// but handle it just in case
		if err == nil {
			// Transition to half-open on success
			cb.state = HalfOpen
			cb.halfOpenSuccess = 1
		}
	case HalfOpen:
		if err != nil {
			// Failed in half-open state, go back to open
			cb.state = Open
			cb.lastFailure = time.Now()
			cb.failures = cb.maxFailures // Ensure we stay open
		} else {
			// Succeeded in half-open state
			cb.halfOpenSuccess++
			
			// Check if we've had enough successes to close
			if cb.halfOpenSuccess >= cb.halfOpenMaxRequests {
				cb.state = Closed
				cb.failures = 0
			}
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() Stats {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	return Stats{
		State:          cb.state,
		Failures:       cb.failures,
		TotalRequests:   cb.totalRequests,
		TotalFailures:  cb.totalFailures,
		TotalSuccesses: cb.totalSuccesses,
	}
}

// Stats holds statistics about the circuit breaker
type Stats struct {
	State          State
	Failures       int
	TotalRequests  int64
	TotalFailures  int64
	TotalSuccesses int64
}