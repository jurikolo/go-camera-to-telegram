// Package validation provides input validation functions for the application
package validation

import (
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a validation error with a descriptive message
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// ValidateIP validates an IP address (IPv4 or IPv6)
func ValidateIP(ip string) error {
	if ip == "" {
		return NewValidationError("IP", "IP address is required")
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return NewValidationError("IP", fmt.Sprintf("invalid IP address format: %s", ip))
	}

	return nil
}

// ValidateCIDR validates a CIDR notation string
func ValidateCIDR(cidr string) error {
	if cidr == "" {
		return NewValidationError("CIDR", "CIDR is required")
	}

	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return NewValidationError("CIDR", fmt.Sprintf("invalid CIDR format: %s", err.Error()))
	}

	return nil
}

// ValidatePort validates a port number (1-65535)
func ValidatePort(port string) error {
	if port == "" {
		return NewValidationError("Port", "port is required")
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return NewValidationError("Port", fmt.Sprintf("port must be a number: %s", port))
	}

	if portNum < 1 || portNum > 65535 {
		return NewValidationError("Port", fmt.Sprintf("port must be between 1 and 65535: %d", portNum))
	}

	return nil
}

// ValidatePortInt validates a port number as integer (1-65535)
func ValidatePortInt(port int) error {
	if port < 1 || port > 65535 {
		return NewValidationError("Port", fmt.Sprintf("port must be between 1 and 65535: %d", port))
	}

	return nil
}

// ValidateFilePath validates a file path for security and format
func ValidateFilePath(path string) error {
	if path == "" {
		return NewValidationError("FilePath", "file path is required")
	}

	// Check for null bytes which can be used for injection attacks
	if strings.Contains(path, "\x00") {
		return NewValidationError("FilePath", "file path contains null bytes")
	}

	// Check for directory traversal attempts
	if strings.Contains(path, "..") {
		// Resolve the path to check if it's actually trying to traverse
		cleanPath := filepath.Clean(path)
		if strings.Contains(cleanPath, "..") {
			return NewValidationError("FilePath", "file path contains directory traversal")
		}
	}

	// Check for invalid characters (platform-specific)
	if strings.ContainsAny(path, "<>:\"|?*") {
		return NewValidationError("FilePath", "file path contains invalid characters")
	}

	// Check path length (260 is a common limit on Windows)
	if len(path) > 260 {
		return NewValidationError("FilePath", "file path is too long")
	}

	return nil
}

// ValidateFilePathWithBase validates a file path against a base directory
func ValidateFilePathWithBase(path, baseDir string) error {
	// First validate the path format
	if err := ValidateFilePath(path); err != nil {
		return err
	}

	// Resolve the absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return NewValidationError("FilePath", fmt.Sprintf("failed to resolve absolute path: %s", err.Error()))
	}

	// Resolve the base directory
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return NewValidationError("BaseDir", fmt.Sprintf("failed to resolve base directory: %s", err.Error()))
	}

	// Check if the path is within the base directory
	if !strings.HasPrefix(absPath, absBaseDir) {
		return NewValidationError("FilePath", "file path is outside allowed directory")
	}

	return nil
}

// ValidateUserInput validates user input against common injection attacks
func ValidateUserInput(input string) error {
	return ValidateUserInputWithLength(input, 1000)
}

// ValidateUserInputWithLength validates user input with a custom maximum length
func ValidateUserInputWithLength(input string, maxLength int) error {
	if input == "" {
		return NewValidationError("Input", "input is required")
	}

	// Check length
	if len(input) > maxLength {
		return NewValidationError("Input", fmt.Sprintf("input exceeds maximum length of %d characters", maxLength))
	}

	// Check for null bytes
	if strings.Contains(input, "\x00") {
		return NewValidationError("Input", "input contains null bytes")
	}

	// Check for common injection patterns
	injectionPatterns := []string{
		`(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|UNION|SCRIPT)\b)`,
		`(?i)(<script.*?>)`,
		`(?i)(javascript:)`,
		`(?i)(on\w+\s*=)`,
		`(?i)(\b(alert|prompt|confirm)\s*\()`,
		`(?i)(\.\.\/)`,
		`(?i)(%2e%2e%2f)`,
	}

	for _, pattern := range injectionPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(input) {
			return NewValidationError("Input", "input contains potentially malicious content")
		}
	}

	return nil
}

// ValidateUserInputWithWhitelist validates user input against a whitelist of allowed characters
func ValidateUserInputWithWhitelist(input, whitelistPattern string) error {
	if input == "" {
		return NewValidationError("Input", "input is required")
	}

	// Check for null bytes
	if strings.Contains(input, "\x00") {
		return NewValidationError("Input", "input contains null bytes")
	}

	// Compile whitelist pattern
	re, err := regexp.Compile(whitelistPattern)
	if err != nil {
		return fmt.Errorf("invalid whitelist pattern: %w", err)
	}

	// Check if input matches whitelist
	if !re.MatchString(input) {
		return NewValidationError("Input", "input contains disallowed characters")
	}

	return nil
}

// ValidateUsername validates a username with common restrictions
func ValidateUsername(username string) error {
	if username == "" {
		return NewValidationError("Username", "username is required")
	}

	// Check length
	if len(username) < 3 || len(username) > 32 {
		return NewValidationError("Username", "username must be between 3 and 32 characters")
	}

	// Check for valid characters (alphanumeric, underscore, hyphen)
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(username) {
		return NewValidationError("Username", "username contains invalid characters")
	}

	// Check for injection patterns
	if err := ValidateUserInput(username); err != nil {
		return err
	}

	return nil
}

// ValidatePassword validates a password with common security requirements
func ValidatePassword(password string) error {
	if password == "" {
		return NewValidationError("Password", "password is required")
	}

	// Check length
	if len(password) < 8 {
		return NewValidationError("Password", "password must be at least 8 characters")
	}

	// Check for common weak passwords
	weakPasswords := []string{
		"password", "qwertyui", "admin123",
	}
	for _, weak := range weakPasswords {
		if strings.ToLower(password) == weak {
			return NewValidationError("Password", "password is too common")
		}
	}

	// Check for null bytes
	if strings.Contains(password, "\x00") {
		return NewValidationError("Password", "password contains null bytes")
	}

	return nil
}

// ValidateEmail validates an email address format
func ValidateEmail(email string) error {
	if email == "" {
		return NewValidationError("Email", "email is required")
	}

	// Basic email format validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return NewValidationError("Email", "invalid email format")
	}

	// Check length
	if len(email) > 254 {
		return NewValidationError("Email", "email is too long")
	}

	// Check for injection patterns
	if err := ValidateUserInput(email); err != nil {
		return err
	}

	return nil
}

// ValidateURL validates a URL format
func ValidateURL(url string) error {
	if url == "" {
		return NewValidationError("URL", "URL is required")
	}

	// Basic URL format validation
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return NewValidationError("URL", "invalid URL format")
	}

	// Check for injection patterns
	if err := ValidateUserInput(url); err != nil {
		return err
	}

	return nil
}

// ValidateMACAddress validates a MAC address format
func ValidateMACAddress(mac string) error {
	if mac == "" {
		return NewValidationError("MAC", "MAC address is required")
	}

	// Parse MAC address
	_, err := net.ParseMAC(mac)
	if err != nil {
		return NewValidationError("MAC", fmt.Sprintf("invalid MAC address format: %s", err.Error()))
	}

	return nil
}

// ValidateSubnetMask validates a subnet mask
func ValidateSubnetMask(mask string) error {
	if mask == "" {
		return NewValidationError("Mask", "subnet mask is required")
	}

	// Try parsing as CIDR notation first
	if strings.Contains(mask, "/") {
		// Validate full CIDR notation like "192.168.1.0/24"
		_, _, err := net.ParseCIDR("192.168.1.0" + mask)
		if err != nil {
			// Try just the mask part
			parts := strings.Split(mask, "/")
			if len(parts) != 2 {
				return NewValidationError("Mask", fmt.Sprintf("invalid CIDR format: %s", mask))
			}
			
			// Validate the CIDR part
			cidrMask := "/" + parts[1]
			_, ipnet, err := net.ParseCIDR("0.0.0.0" + cidrMask)
			if err != nil {
				return NewValidationError("Mask", fmt.Sprintf("invalid subnet mask format: %s", err.Error()))
			}
			if ipnet == nil {
				return NewValidationError("Mask", "invalid subnet mask")
			}
		}
		return nil
	}

	// Try parsing as IP address
	ip := net.ParseIP(mask)
	if ip == nil {
		// Try parsing as a plain number (0-32)
		if num, err := strconv.Atoi(mask); err == nil {
			if num >= 0 && num <= 32 {
				return nil
			}
			return NewValidationError("Mask", fmt.Sprintf("subnet mask must be between 0 and 32: %s", mask))
		}
		return NewValidationError("Mask", fmt.Sprintf("invalid subnet mask format: %s", mask))
	}

	// Check if it's a valid subnet mask
	if !isSubnetMask(ip) {
		// For now, let's be more permissive and accept common subnet masks
		// This is a simplified check for common valid subnet masks
		validMasks := []string{
			"255.255.255.255", "255.255.255.254", "255.255.255.252", "255.255.255.248",
			"255.255.255.240", "255.255.255.224", "255.255.255.192", "255.255.255.128",
			"255.255.255.0", "255.255.254.0", "255.255.252.0", "255.255.248.0",
			"255.255.240.0", "255.255.224.0", "255.255.192.0", "255.255.128.0",
			"255.255.0.0", "255.254.0.0", "255.252.0.0", "255.248.0.0",
			"255.240.0.0", "255.224.0.0", "255.192.0.0", "255.128.0.0",
			"255.0.0.0", "254.0.0.0", "252.0.0.0", "248.0.0.0",
			"240.0.0.0", "224.0.0.0", "192.0.0.0", "128.0.0.0", "0.0.0.0",
		}
		
		for _, valid := range validMasks {
			if mask == valid {
				return nil
			}
		}
		
		return NewValidationError("Mask", fmt.Sprintf("invalid subnet mask: %s", mask))
	}

	return nil
}

// isSubnetMask checks if an IP address represents a valid subnet mask
func isSubnetMask(ip net.IP) bool {
	// Convert to 4-byte representation
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	// Convert to integer
	mask := uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])

	// Check if it's a valid subnet mask (contiguous 1s followed by 0s)
	// This works by checking if (mask & (mask + 1)) == 0
	return (mask & (mask + 1)) == 0
}

// ValidateHostname validates a hostname
func ValidateHostname(hostname string) error {
	if hostname == "" {
		return NewValidationError("Hostname", "hostname is required")
	}

	// Check length
	if len(hostname) > 253 {
		return NewValidationError("Hostname", "hostname is too long")
	}

	// Check for valid characters and format
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !hostnameRegex.MatchString(hostname) {
		return NewValidationError("Hostname", "invalid hostname format")
	}

	// Check for injection patterns
	if err := ValidateUserInput(hostname); err != nil {
		return err
	}

	return nil
}

// ValidateDomain validates a domain name
func ValidateDomain(domain string) error {
	if domain == "" {
		return NewValidationError("Domain", "domain is required")
	}

	// Check length
	if len(domain) > 253 {
		return NewValidationError("Domain", "domain is too long")
	}

	// Check for valid characters and format
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(domain) {
		return NewValidationError("Domain", "invalid domain format")
	}

	// Check for injection patterns
	if err := ValidateUserInput(domain); err != nil {
		return err
	}

	return nil
}