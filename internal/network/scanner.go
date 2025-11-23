// Package network provides functionality for scanning networks for active hosts
package network

import (
	"context"
	"fmt"
	"net"
	"strings"
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

// VerifyRTSP verifies that a host is running an RTSP service on port 554
// It performs a more thorough check than just port scanning by sending an RTSP OPTIONS request
func (s *Scanner) VerifyRTSP(ctx context.Context, host string) (bool, error) {
	return s.verifyRTSPWithRetry(ctx, host, 3, time.Second)
}

// verifyRTSPWithRetry performs RTSP verification with exponential backoff retry logic
func (s *Scanner) verifyRTSPWithRetry(ctx context.Context, host string, maxRetries int, initialDelay time.Duration) (bool, error) {
	delay := initialDelay
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
		
		// Try to verify RTSP
		valid, err := s.verifyRTSP(ctx, host)
		if err == nil && valid {
			return true, nil
		}
		
		// If this was the last attempt, return the error
		if attempt == maxRetries {
			if err != nil {
				return false, fmt.Errorf("RTSP verification failed after %d attempts: %w", maxRetries+1, err)
			}
			return false, nil
		}
		
		// Wait before retrying with exponential backoff
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(delay):
			// Exponential backoff: double the delay for next attempt
			delay *= 2
		}
	}
	
	return false, nil
}

// verifyRTSP performs the actual RTSP verification by sending an OPTIONS request
func (s *Scanner) verifyRTSP(ctx context.Context, host string) (bool, error) {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	
	// Attempt to establish a connection to port 554
	conn, err := (&net.Dialer{}).DialContext(timeoutCtx, "tcp", net.JoinHostPort(host, "554"))
	if err != nil {
		return false, fmt.Errorf("failed to connect to RTSP port: %w", err)
	}
	defer conn.Close()
	
	// Send RTSP OPTIONS request
	optionsRequest := fmt.Sprintf("OPTIONS rtsp://%s:554/ RTSP/1.0\r\n"+
		"CSeq: 1\r\n"+
		"User-Agent: GoCameraToTelegram/1.0\r\n"+
		"\r\n", host)
	
	// Set write deadline
	if err := conn.SetWriteDeadline(time.Now().Add(s.timeout)); err != nil {
		return false, fmt.Errorf("failed to set write deadline: %w", err)
	}
	
	// Send the request
	if _, err := conn.Write([]byte(optionsRequest)); err != nil {
		return false, fmt.Errorf("failed to send OPTIONS request: %w", err)
	}
	
	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
		return false, fmt.Errorf("failed to set read deadline: %w", err)
	}
	
	// Read the response
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return false, fmt.Errorf("failed to read RTSP response: %w", err)
	}
	
	// Parse the response
	response := string(buffer[:n])
	
	// Check if it looks like a valid RTSP response
	// RTSP responses typically start with "RTSP/1.0" and include a status code
	if !s.isValidRTSPResponse(response) {
		return false, fmt.Errorf("invalid RTSP response received")
	}
	
	return true, nil
}

// isValidRTSPResponse checks if the response looks like a valid RTSP response
func (s *Scanner) isValidRTSPResponse(response string) bool {
	// Check if it starts with RTSP/1.0
	if !strings.HasPrefix(response, "RTSP/1.0") {
		return false
	}
	
	// Check if it contains a valid status code (200, 401, 403, etc.)
	// Common RTSP status codes:
	// 200 OK
	// 401 Unauthorized
	// 403 Forbidden
	// 404 Not Found
	// 500 Internal Server Error
	statusCodes := []string{"200", "401", "403", "404", "500"}
	for _, code := range statusCodes {
		if strings.Contains(response, "RTSP/1.0 "+code) {
			return true
		}
	}
	
	return false
}