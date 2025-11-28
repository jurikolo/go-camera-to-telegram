// Package network provides tests for edge cases in network scanning
package network

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestScannerEdgeCases tests various edge cases for the network scanner
func TestScannerEdgeCases(t *testing.T) {
	// Test with very small timeout
	scanner := NewScanner(2, 1*time.Millisecond, 1*time.Millisecond)
	
	ctx := context.Background()
	activeHosts, err := scanner.ScanCIDR(ctx, "127.0.0.1/32")
	
	// We don't assert specific results as they depend on the system
	// but we can check that the function doesn't panic
	assert.NoError(t, err)
	assert.NotNil(t, activeHosts)
}

// TestScannerWithInvalidInputs tests the scanner with various invalid inputs
func TestScannerWithInvalidInputs(t *testing.T) {
	// Test with zero workers
	scanner := NewScanner(0, time.Second, time.Millisecond)
	
	ctx := context.Background()
	activeHosts, err := scanner.ScanCIDR(ctx, "127.0.0.1/32")
	
	// With zero workers, we might still get results depending on implementation
	assert.NoError(t, err)
	assert.NotNil(t, activeHosts)
	
	// Test with negative timeout
	scannerNegative := NewScanner(2, -1*time.Second, time.Millisecond)
	
	activeHostsNegative, err := scannerNegative.ScanCIDR(ctx, "127.0.0.1/32")
	
	// Negative timeout might still work depending on implementation
	assert.NoError(t, err)
	assert.NotNil(t, activeHostsNegative)
	
	// Test with very small rate limit (but not zero)
	scannerSmallRate := NewScanner(2, time.Second, 1*time.Nanosecond)
	
	activeHostsSmallRate, err := scannerSmallRate.ScanCIDR(ctx, "127.0.0.1/32")
	
	// Very small rate limit might still work depending on implementation
	assert.NoError(t, err)
	assert.NotNil(t, activeHostsSmallRate)
}

// TestScannerWithLargeCIDR tests the scanner with a large CIDR range
func TestScannerWithLargeCIDR(t *testing.T) {
	// Test with a larger CIDR range but small timeout
	scanner := NewScanner(2, 100*time.Millisecond, 1*time.Millisecond)
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	
	// Test with a larger CIDR range
	activeHosts, err := scanner.ScanCIDR(ctx, "127.0.0.0/24")
	
	// We expect either success or timeout
	if err != nil {
		assert.Equal(t, context.DeadlineExceeded, err)
	} else {
		assert.NotNil(t, activeHosts)
	}
}

// TestScannerWithCancelledContext tests the scanner with a cancelled context
func TestScannerWithCancelledContext(t *testing.T) {
	scanner := NewScanner(2, time.Second, 10*time.Millisecond)
	
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	// Cancel the context immediately
	cancel()
	
	// Test with cancelled context
	activeHosts, err := scanner.ScanCIDR(ctx, "127.0.0.0/24")
	
	// We expect the context to be cancelled
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, activeHosts)
}

// TestScannerWithVeryLargeCIDR tests the scanner with a very large CIDR range
func TestScannerWithVeryLargeCIDR(t *testing.T) {
	scanner := NewScanner(2, 100*time.Millisecond, 1*time.Millisecond)
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	
	// Test with a very large CIDR range
	activeHosts, err := scanner.ScanCIDR(ctx, "10.0.0.0/8")
	
	// We expect either success or timeout
	if err != nil {
		assert.Equal(t, context.DeadlineExceeded, err)
	} else {
		assert.NotNil(t, activeHosts)
	}
}

// TestScannerWithSingleHost tests the scanner with a single host CIDR
func TestScannerWithSingleHost(t *testing.T) {
	scanner := NewScanner(2, time.Second, 10*time.Millisecond)
	
	ctx := context.Background()
	activeHosts, err := scanner.ScanCIDR(ctx, "127.0.0.1/32")
	
	// We expect success
	assert.NoError(t, err)
	assert.NotNil(t, activeHosts)
}

// TestScannerWithIPv6 tests the scanner with IPv6
func TestScannerWithIPv6(t *testing.T) {
	scanner := NewScanner(2, time.Second, 10*time.Millisecond)
	
	ctx := context.Background()
	activeHosts, err := scanner.ScanCIDR(ctx, "::1/128")
	
	// We expect success
	assert.NoError(t, err)
	assert.NotNil(t, activeHosts)
}

// TestTableDrivenNetworkInputs tests various network inputs with table-driven tests
func TestTableDrivenNetworkInputs(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		workers     int
		timeout     time.Duration
		rateLimit   time.Duration
		expectError bool
	}{
		{"Valid IPv4", "127.0.0.1/32", 2, time.Second, 10 * time.Millisecond, false},
		{"Valid IPv6", "::1/128", 2, time.Second, 10 * time.Millisecond, false},
		{"Single host IPv4", "127.0.0.1/32", 1, time.Second, 1 * time.Millisecond, false},
		{"Single host IPv6", "::1/128", 1, time.Second, 1 * time.Millisecond, false},
		{"Invalid CIDR", "invalid", 2, time.Second, 10 * time.Millisecond, true},
		{"Empty CIDR", "", 2, time.Second, 10 * time.Millisecond, true},
		{"Large CIDR", "127.0.0.0/24", 2, 100 * time.Millisecond, 1 * time.Millisecond, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner(tt.workers, tt.timeout, tt.rateLimit)
			
			ctx := context.Background()
			activeHosts, err := scanner.ScanCIDR(ctx, tt.cidr)
			
			if tt.expectError {
				assert.Error(t, err, "Expected error for CIDR: %s", tt.cidr)
			} else {
				assert.NoError(t, err, "Expected no error for CIDR: %s", tt.cidr)
				assert.NotNil(t, activeHosts, "Expected non-nil activeHosts for CIDR: %s", tt.cidr)
			}
		})
	}
}

// TestIncIPEdgeCases tests edge cases for the incIP function
func TestIncIPEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Increment from 1", "192.168.1.1", "192.168.1.2"},
		{"Increment to next octet", "192.168.1.255", "192.168.2.0"},
		{"Increment from 254", "192.168.1.254", "192.168.1.255"},
		{"Increment from 0", "192.168.1.0", "192.168.1.1"},
		{"Increment IPv6", "2001:db8::1", "2001:db8::2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.input)
			assert.NotNil(t, ip, "Failed to parse IP: %s", tt.input)
			
			incIP(ip)
			
			expectedIP := net.ParseIP(tt.expected)
			assert.NotNil(t, expectedIP, "Failed to parse expected IP: %s", tt.expected)
			
			assert.True(t, ip.Equal(expectedIP), "Expected %s, got %s", tt.expected, ip.String())
		})
	}
}

// TestScannerVerifyRTSPEdgeCases tests edge cases for RTSP verification
func TestScannerVerifyRTSPEdgeCases(t *testing.T) {
	scanner := NewScanner(2, 100*time.Millisecond, 10*time.Millisecond)
	
	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	valid, err := scanner.VerifyRTSP(ctx, "127.0.0.1")
	assert.False(t, valid, "Expected false for cancelled context")
	assert.Error(t, err, "Expected error for cancelled context")
	assert.Equal(t, context.Canceled, err, "Expected context.Canceled error")
	
	// Test with timeout context
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancelTimeout()
	
	validTimeout, errTimeout := scanner.VerifyRTSP(ctxTimeout, "127.0.0.1")
	// We expect either false with timeout error or false with no error
	if errTimeout != nil {
		assert.Equal(t, context.DeadlineExceeded, errTimeout, "Expected context.DeadlineExceeded error")
	}
	assert.False(t, validTimeout, "Expected false for timeout context")
	
	// Test with invalid IP
	validInvalid, errInvalid := scanner.VerifyRTSP(context.Background(), "invalid-ip")
	// We expect false with some error
	assert.False(t, validInvalid, "Expected false for invalid IP")
	assert.Error(t, errInvalid, "Expected error for invalid IP")
}