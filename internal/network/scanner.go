// Package network provides functionality for scanning networks for active hosts
package network

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// HostResult represents the result of scanning a single host
type HostResult struct {
	Host    string
	Open    bool
	Error   error
}

// Scanner handles network scanning with configurable parameters
type Scanner struct {
	workers   int
	timeout   time.Duration
	rateLimit time.Duration
}

// NewScanner creates a new network scanner with the specified configuration
func NewScanner(workers int, timeout time.Duration, rateLimit time.Duration) *Scanner {
	return &Scanner{
		workers:   workers,
		timeout:   timeout,
		rateLimit: rateLimit,
	}
}

// ScanCIDR scans a CIDR range for active hosts on port 554 (RTSP)
// It uses a worker pool pattern with context-based cancellation
func (s *Scanner) ScanCIDR(ctx context.Context, cidr string) ([]string, error) {
	// Parse the CIDR notation
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CIDR: %w", err)
	}

	// Generate all IP addresses in the range
	hosts := make([]string, 0)
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		hosts = append(hosts, ip.String())
	}

	// Remove network and broadcast addresses
	if len(hosts) > 2 {
		hosts = hosts[1 : len(hosts)-1]
	}

	// Create channels for the worker pool
	jobs := make(chan string, len(hosts))
	results := make(chan HostResult, len(hosts))

	// Create a context with cancellation for rate limiting
	rateCtx, rateCancel := context.WithCancel(ctx)
	defer rateCancel()

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go s.worker(rateCtx, &wg, jobs, results)
	}

	// Send jobs to workers
	go func() {
		defer close(jobs)
		ticker := time.NewTicker(s.rateLimit)
		defer ticker.Stop()

		for _, host := range hosts {
			select {
			case <-ctx.Done():
				return
			case <-rateCtx.Done():
				return
			case jobs <- host:
				// Wait for rate limit
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	activeHosts := make([]string, 0)
	for result := range results {
		if result.Error != nil {
			// Log error but continue scanning
			continue
		}
		if result.Open {
			activeHosts = append(activeHosts, result.Host)
		}
	}

	// Check if context was cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return activeHosts, nil
}

// worker is a goroutine that scans hosts for open ports
func (s *Scanner) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan string, results chan<- HostResult) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case host, ok := <-jobs:
			if !ok {
				return
			}

			// Check if port 554 is open
			open, err := s.isPortOpen(ctx, host, "554", s.timeout)
			results <- HostResult{
				Host:  host,
				Open:  open,
				Error: err,
			}
		}
	}
}

// isPortOpen checks if a specific port is open on a host
func (s *Scanner) isPortOpen(ctx context.Context, host, port string, timeout time.Duration) (bool, error) {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Attempt to establish a connection
	conn, err := (&net.Dialer{}).DialContext(timeoutCtx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		// Check if it's a context error (cancellation or timeout)
		if timeoutCtx.Err() != nil {
			return false, timeoutCtx.Err()
		}
		// Port is closed or unreachable
		return false, nil
	}
	defer conn.Close()

	// Port is open
	return true, nil
}

// incIP increments an IP address
func incIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}