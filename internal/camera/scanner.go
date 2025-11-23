// Package camera provides functionality for discovering IP cameras on the network
package camera

import (
	"context"
	"time"

	"github.com/jurikolo/go-camera-to-telegram/internal/config"
	"github.com/jurikolo/go-camera-to-telegram/internal/network"
)

// Scanner handles the discovery of IP cameras on the network
type Scanner struct {
	config *config.Config
}

// NewScanner creates a new camera scanner instance
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		config: cfg,
	}
}

// ScanNetwork scans the network for IP cameras
func (s *Scanner) ScanNetwork() ([]string, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.Network.Timeout)*time.Second)
	defer cancel()

	// Create a network scanner with the configured parameters
	scanner := network.NewScanner(
		s.config.Network.Workers,
		time.Duration(s.config.Network.Timeout)*time.Second,
		time.Duration(1000/s.config.Network.Workers)*time.Millisecond, // Rate limit
	)

	// Scan the network CIDR for active hosts on port 554
	activeHosts, err := scanner.ScanCIDR(ctx, s.config.Network.CIDR)
	if err != nil {
		return nil, err
	}

	return activeHosts, nil
}