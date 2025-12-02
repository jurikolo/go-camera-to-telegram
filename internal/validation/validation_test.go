// Package validation provides tests for input validation functions
package validation

import (
	"testing"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		expectError bool
	}{
		{"Valid IPv4", "192.168.1.1", false},
		{"Valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"Empty IP", "", true},
		{"Invalid IP", "999.999.999.999", true},
		{"Invalid format", "not.an.ip.address", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIP(tt.ip)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateIP(%s) error = %v, expectError %v", tt.ip, err, tt.expectError)
			}
		})
	}
}

func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		expectError bool
	}{
		{"Valid CIDR IPv4", "192.168.1.0/24", false},
		{"Valid CIDR IPv6", "2001:db8::/32", false},
		{"Empty CIDR", "", true},
		{"Invalid CIDR", "192.168.1.0/99", true},
		{"Invalid format", "not.a.cidr", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCIDR(tt.cidr)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateCIDR(%s) error = %v, expectError %v", tt.cidr, err, tt.expectError)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name        string
		port        string
		expectError bool
	}{
		{"Valid port string", "8080", false},
		{"Minimum port", "1", false},
		{"Maximum port", "65535", false},
		{"Empty port", "", true},
		{"Invalid port string", "not_a_number", true},
		{"Port too low", "0", true},
		{"Port too high", "65536", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidatePort(%s) error = %v, expectError %v", tt.port, err, tt.expectError)
			}
		})
	}
}

func TestValidatePortInt(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		expectError bool
	}{
		{"Valid port", 8080, false},
		{"Minimum port", 1, false},
		{"Maximum port", 65535, false},
		{"Port too low", 0, true},
		{"Port too high", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortInt(tt.port)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidatePortInt(%d) error = %v, expectError %v", tt.port, err, tt.expectError)
			}
		})
	}
}

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{"Valid path", "/tmp/test.txt", false},
		{"Valid relative path", "test.txt", false},
		{"Empty path", "", true},
		{"Path with null bytes", "test\x00.txt", true},
		{"Path with directory traversal", "../test.txt", true},
		{"Path with invalid characters", "test<>.txt", true},
		{"Path too long", string(make([]byte, 300)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateFilePath(%s) error = %v, expectError %v", tt.path, err, tt.expectError)
			}
		})
	}
}

func TestValidateFilePathWithBase(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		baseDir     string
		expectError bool
	}{
		{"Valid path within base", "/tmp/test.txt", "/tmp", false},
		{"Path outside base", "/etc/passwd", "/tmp", true},
		{"Empty path", "", "/tmp", true},
		{"Empty base", "/tmp/test.txt", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePathWithBase(tt.path, tt.baseDir)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateFilePathWithBase(%s, %s) error = %v, expectError %v", tt.path, tt.baseDir, err, tt.expectError)
			}
		})
	}
}

func TestValidateUserInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"Valid input", "This is a test", false},
		{"Empty input", "", true},
		{"Input too long", string(make([]byte, 1500)), true},
		{"Input with null bytes", "test\x00input", true},
		{"Input with SQL injection", "SELECT * FROM users", true},
		{"Input with XSS", "<script>alert('xss')</script>", true},
		{"Input with directory traversal", "../../../etc/passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserInput(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateUserInput(%s) error = %v, expectError %v", tt.input, err, tt.expectError)
			}
		})
	}
}

func TestValidateUserInputWithLength(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		maxLength   int
		expectError bool
	}{
		{"Valid input", "test", 10, false},
		{"Input at max length", "12345", 5, false},
		{"Input exceeding length", "123456", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserInputWithLength(tt.input, tt.maxLength)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateUserInputWithLength(%s, %d) error = %v, expectError %v", tt.input, tt.maxLength, err, tt.expectError)
			}
		})
	}
}

func TestValidateUserInputWithWhitelist(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		pattern     string
		expectError bool
	}{
		{"Valid input with alphanumeric pattern", "abc123", "^[a-zA-Z0-9]+$", false},
		{"Invalid input with alphanumeric pattern", "abc-123", "^[a-zA-Z0-9]+$", true},
		{"Empty input", "", "^[a-zA-Z0-9]+$", true},
		{"Invalid pattern", "test", "[", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUserInputWithWhitelist(tt.input, tt.pattern)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateUserInputWithWhitelist(%s, %s) error = %v, expectError %v", tt.input, tt.pattern, err, tt.expectError)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		expectError bool
	}{
		{"Valid username", "testuser", false},
		{"Username with underscore", "test_user", false},
		{"Username with hyphen", "test-user", false},
		{"Empty username", "", true},
		{"Username too short", "ab", true},
		{"Username too long", "a12345678901234567890123456789012", true},
		{"Username with invalid characters", "test user", true},
		{"Username with SQL injection", "test'; DROP TABLE users; --", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateUsername(%s) error = %v, expectError %v", tt.username, err, tt.expectError)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{"Valid password", "SecurePass123!", false},
		{"Minimum length password", "12345678", false},
		{"Empty password", "", true},
		{"Password too short", "1234567", true},
		{"Weak password", "password", true},
		{"Password with null bytes", "test\x00pass", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidatePassword(%s) error = %v, expectError %v", tt.password, err, tt.expectError)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{"Valid email", "test@example.com", false},
		{"Valid email with subdomain", "user@mail.example.com", false},
		{"Empty email", "", true},
		{"Invalid email format", "not.an.email", true},
		{"Email too long", "a" + string(make([]byte, 255)) + "@example.com", true},
		{"Email with injection", "test@example.com'; DROP TABLE users; --", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateEmail(%s) error = %v, expectError %v", tt.email, err, tt.expectError)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{"Valid HTTP URL", "http://example.com", false},
		{"Valid HTTPS URL", "https://example.com", false},
		{"Empty URL", "", true},
		{"Invalid URL format", "not.a.url", true},
		{"URL with injection", "http://example.com'; DROP TABLE users; --", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateURL(%s) error = %v, expectError %v", tt.url, err, tt.expectError)
			}
		})
	}
}

func TestValidateMACAddress(t *testing.T) {
	tests := []struct {
		name        string
		mac         string
		expectError bool
	}{
		{"Valid MAC address", "00:11:22:33:44:55", false},
		{"Valid MAC address with hyphens", "00-11-22-33-44-55", false},
		{"Empty MAC", "", true},
		{"Invalid MAC format", "00:11:22:33:44", true},
		{"Invalid MAC characters", "GG:HH:II:JJ:KK:LL", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMACAddress(tt.mac)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateMACAddress(%s) error = %v, expectError %v", tt.mac, err, tt.expectError)
			}
		})
	}
}

func TestValidateSubnetMask(t *testing.T) {
	tests := []struct {
		name        string
		mask        string
		expectError bool
	}{
		{"Valid subnet mask IP", "255.255.255.0", false},
		{"Valid subnet mask CIDR", "/24", false},
		{"Valid subnet mask number", "24", false},
		{"Empty mask", "", true},
		{"Invalid mask", "255.255.255.1", true},
		{"Invalid CIDR", "/33", true},
		{"Invalid number", "33", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubnetMask(tt.mask)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateSubnetMask(%s) error = %v, expectError %v", tt.mask, err, tt.expectError)
			}
		})
	}
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		name        string
		hostname    string
		expectError bool
	}{
		{"Valid hostname", "example.com", false},
		{"Valid hostname with subdomain", "www.example.com", false},
		{"Valid hostname with hyphens", "test-server.local", false},
		{"Empty hostname", "", true},
		{"Hostname too long", string(make([]byte, 254)) + ".com", true},
		{"Invalid hostname format", "test..server", true},
		{"Hostname with injection", "example.com'; DROP TABLE users; --", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostname(tt.hostname)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateHostname(%s) error = %v, expectError %v", tt.hostname, err, tt.expectError)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		expectError bool
	}{
		{"Valid domain", "example.com", false},
		{"Valid domain with subdomain", "www.example.com", false},
		{"Valid domain with hyphens", "test-domain.local", false},
		{"Empty domain", "", true},
		{"Domain too long", string(make([]byte, 254)) + ".com", true},
		{"Invalid domain format", "test..domain", true},
		{"Domain with injection", "example.com'; DROP TABLE users; --", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDomain(tt.domain)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateDomain(%s) error = %v, expectError %v", tt.domain, err, tt.expectError)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("test", "test error")
	expected := "test: test error"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}

	errNoField := ValidationError{Message: "test error"}
	expected = "test error"
	if errNoField.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, errNoField.Error())
	}
}