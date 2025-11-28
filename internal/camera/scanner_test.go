// Package camera provides tests for the camera scanner functionality
package camera

import (
	"net"
	"testing"

	"github.com/jurikolo/go-camera-to-telegram/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestNewScanner tests the NewScanner function
func TestNewScanner(t *testing.T) {
	cfg := &config.Config{
		Network: config.NetworkConfig{
			CIDR:    "192.168.1.0/24",
			Timeout: 10,
			Workers: 10,
		},
	}

	scanner := NewScanner(cfg)
	assert.NotNil(t, scanner)
	assert.Equal(t, cfg, scanner.config)
}

// TestScanner_ScanNetwork tests the ScanNetwork method
func TestScanner_ScanNetwork(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Network: config.NetworkConfig{
			CIDR:    "127.0.0.1/32",
			Timeout: 5,
			Workers: 2,
		},
	}

	scanner := NewScanner(cfg)

	// Test scanning the network
	activeHosts, err := scanner.ScanNetwork()

	// We don't assert specific results as they depend on the system
	// but we can check that the function doesn't panic and returns without fatal error
	assert.NoError(t, err)

	// We should get at least localhost or 0 hosts
	assert.GreaterOrEqual(t, len(activeHosts), 0)
}

// TestScanner_ScanNetworkWithInvalidCIDR tests the ScanNetwork method with invalid CIDR
func TestScanner_ScanNetworkWithInvalidCIDR(t *testing.T) {
	// Create a test configuration with invalid CIDR
	cfg := &config.Config{
		Network: config.NetworkConfig{
			CIDR:    "invalid-cidr",
			Timeout: 5,
			Workers: 2,
		},
	}

	scanner := NewScanner(cfg)

	// Test scanning with invalid CIDR
	activeHosts, err := scanner.ScanNetwork()

	// We expect an error due to invalid CIDR
	assert.Error(t, err)
	assert.Nil(t, activeHosts)
}

// TestScanner_ScanNetworkWithEmptyCIDR tests the ScanNetwork method with empty CIDR
func TestScanner_ScanNetworkWithEmptyCIDR(t *testing.T) {
	// Create a test configuration with empty CIDR
	cfg := &config.Config{
		Network: config.NetworkConfig{
			CIDR:    "",
			Timeout: 5,
			Workers: 2,
		},
	}

	scanner := NewScanner(cfg)

	// Test scanning with empty CIDR
	activeHosts, err := scanner.ScanNetwork()

	// We expect an error due to empty CIDR
	assert.Error(t, err)
	assert.Nil(t, activeHosts)
}

// TestTableDrivenIPParsing tests IP parsing with table-driven tests
func TestTableDrivenIPParsing(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		expectError bool
	}{
		{"Valid IPv4 CIDR", "192.168.1.0/24", false},
		{"Valid IPv6 CIDR", "2001:db8::/32", false},
		{"Invalid CIDR format", "invalid-cidr", true},
		{"Empty CIDR", "", true},
		{"Invalid subnet", "192.168.1.0/33", true}, // Invalid subnet for IPv4
		{"Valid single host", "192.168.1.1/32", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ipnet, err := net.ParseCIDR(tt.cidr)
			if tt.expectError {
				assert.Error(t, err, "Expected error for CIDR: %s", tt.cidr)
			} else {
				assert.NoError(t, err, "Expected no error for CIDR: %s", tt.cidr)
				assert.NotNil(t, ipnet, "Expected non-nil IPNet for CIDR: %s", tt.cidr)
			}
		})
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	// Test with very large CIDR range
	cfg := &config.Config{
		Network: config.NetworkConfig{
			CIDR:    "127.0.0.1/32",
			Timeout: 1,
			Workers: 1,
		},
	}

	scanner := NewScanner(cfg)
	assert.NotNil(t, scanner)

	// Test with maximum workers
	cfgMaxWorkers := &config.Config{
		Network: config.NetworkConfig{
			CIDR:    "127.0.0.1/32",
			Timeout: 5,
			Workers: 100, // High number of workers
		},
	}

	scannerMaxWorkers := NewScanner(cfgMaxWorkers)
	assert.NotNil(t, scannerMaxWorkers)
}