// Package validation provides an example of how to integrate the validation package with existing code
package validation

import (
	"fmt"
)

// NetworkConfig represents a network configuration (similar to the one in internal/config)
type NetworkConfig struct {
	CIDR    string
	Timeout int
	Workers int
}

// ValidateWithValidationPackage demonstrates how to use the new validation package
// to enhance existing validation logic
func (n *NetworkConfig) ValidateWithValidationPackage() error {
	// Use the new validation package for better error messages and security
	if err := ValidateCIDR(n.CIDR); err != nil {
		return fmt.Errorf("network config validation failed: %w", err)
	}

	// Validate timeout with clear error messages
	if n.Timeout <= 0 {
		return NewValidationError("Timeout", "timeout must be positive")
	}

	// Validate workers with clear error messages
	if n.Workers <= 0 {
		return NewValidationError("Workers", "workers must be positive")
	}

	return nil
}

// EnhancedRTSPConfigValidation shows how to enhance RTSP configuration validation
func EnhancedRTSPConfigValidation(timeout int) error {
	// Validate timeout with clear error messages
	if timeout <= 0 {
		return NewValidationError("Timeout", "RTSP timeout must be positive")
	}

	// Additional validation could be added here
	return nil
}

// EnhancedTelegramConfigValidation shows how to enhance Telegram configuration validation
func EnhancedTelegramConfigValidation(chatID int64) error {
	// Validate chat ID
	if chatID == 0 {
		return NewValidationError("ChatID", "Telegram chat ID is required")
	}

	// Additional validation could be added here
	return nil
}

// EnhancedScanConfigValidation shows how to enhance scan configuration validation
func EnhancedScanConfigValidation(interval, maxConcurrent int) error {
	// Validate interval with clear error messages
	if interval <= 0 {
		return NewValidationError("Interval", "scan interval must be positive")
	}

	// Validate max concurrent with clear error messages
	if maxConcurrent <= 0 {
		return NewValidationError("MaxConcurrent", "max concurrent scans must be positive")
	}

	// Additional validation could be added here
	return nil
}

// Example of validating user-provided network configuration
func ExampleValidateNetworkConfig() {
	config := &NetworkConfig{
		CIDR:    "192.168.1.0/24",
		Timeout: 5,
		Workers: 10,
	}

	err := config.ValidateWithValidationPackage()
	if err != nil {
		// Handle validation error with detailed information
		if validationErr, ok := err.(ValidationError); ok {
			fmt.Printf("Validation failed for field %s: %s\n", validationErr.Field, validationErr.Message)
		} else {
			fmt.Printf("Validation failed: %v\n", err)
		}
		return
	}

	fmt.Println("Network configuration is valid")
	// Output: Network configuration is valid
}

// Example of handling invalid network configuration
func ExampleValidateInvalidNetworkConfig() {
	config := &NetworkConfig{
		CIDR:    "invalid-cidr",
		Timeout: -1,
		Workers: 0,
	}

	err := config.ValidateWithValidationPackage()
	if err != nil {
		// Handle validation error with detailed information
		if validationErr, ok := err.(ValidationError); ok {
			fmt.Printf("Field: %s, Message: %s\n", validationErr.Field, validationErr.Message)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
	// Output: Field: CIDR, Message: invalid CIDR format: invalid CIDR address: invalid-cidr
}