# Validation Package

This package provides comprehensive input validation functions for the application, with security features to prevent injection attacks and ensure data integrity.

## Features

- **IP Address Validation**: Validate IPv4 and IPv6 addresses
- **CIDR Range Validation**: Validate CIDR notation for network ranges
- **Port Number Validation**: Validate TCP/UDP port numbers (1-65535)
- **File Path Validation**: Secure file path validation with directory traversal prevention
- **User Input Validation**: Protection against injection attacks (SQL, XSS, etc.)
- **Common Data Types**: Username, password, email, URL, MAC address, hostname, domain validation
- **Clear Error Messages**: Descriptive error messages for better debugging
- **Security Focused**: Protection against common injection attacks and security vulnerabilities

## Usage Examples

```go
import "github.com/jurikolo/go-camera-to-telegram/internal/validation"

// Validate an IP address
err := validation.ValidateIP("192.168.1.1")
if err != nil {
    // Handle error
}

// Validate a CIDR range
err := validation.ValidateCIDR("192.168.1.0/24")
if err != nil {
    // Handle error
}

// Validate a port number
err := validation.ValidatePort("8080")
if err != nil {
    // Handle error
}

// Validate a file path
err := validation.ValidateFilePath("/tmp/test.txt")
if err != nil {
    // Handle error
}

// Validate user input with injection prevention
err := validation.ValidateUserInput("user provided input")
if err != nil {
    // Handle error
}
```

## Security Features

- **Injection Attack Prevention**: Built-in protection against SQL injection, XSS, and other common attacks
- **Directory Traversal Prevention**: Secure file path validation
- **Input Sanitization**: Automatic detection of malicious input patterns
- **Whitelist Validation**: Support for whitelist-based input validation
- **Length Restrictions**: Configurable maximum length limits

## Error Handling

All validation functions return a `ValidationError` type that provides clear, descriptive error messages:

```go
err := validation.ValidateIP("invalid-ip")
if validationErr, ok := err.(validation.ValidationError); ok {
    fmt.Printf("Field: %s, Message: %s\n", validationErr.Field, validationErr.Message)
}
```

## Testing

The package includes comprehensive tests for all validation functions:

```bash
go test ./internal/validation/
```

## Integration with Existing Code

The validation functions can be easily integrated into existing configuration validation, user input handling, and data processing code to enhance security and data integrity.