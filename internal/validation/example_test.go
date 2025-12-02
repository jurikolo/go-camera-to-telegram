// Package validation provides examples of how to use the validation functions
package validation

import (
	"fmt"
)

// Example of validating an IP address
func ExampleValidateIP() {
	// Valid IP address
	err := ValidateIP("192.168.1.1")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid IP address")
	}
	// Output: Valid IP address
}

// Example of validating a CIDR range
func ExampleValidateCIDR() {
	// Valid CIDR
	err := ValidateCIDR("192.168.1.0/24")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid CIDR range")
	}
	// Output: Valid CIDR range
}

// Example of validating a port
func ExampleValidatePort() {
	// Valid port
	err := ValidatePort("8080")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid port")
	}
	// Output: Valid port
}

// Example of validating a file path
func ExampleValidateFilePath() {
	// Valid file path
	err := ValidateFilePath("/tmp/test.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid file path")
	}
	// Output: Valid file path
}

// Example of validating user input
func ExampleValidateUserInput() {
	// Valid user input
	err := ValidateUserInput("This is a test input")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid user input")
	}
	// Output: Valid user input
}

// Example of validating a username
func ExampleValidateUsername() {
	// Valid username
	err := ValidateUsername("testuser")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid username")
	}
	// Output: Valid username
}

// Example of validating a password
func ExampleValidatePassword() {
	// Valid password
	err := ValidatePassword("SecurePass123!")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid password")
	}
	// Output: Valid password
}

// Example of validating an email
func ExampleValidateEmail() {
	// Valid email
	err := ValidateEmail("test@example.com")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid email")
	}
	// Output: Valid email
}

// Example of validating a URL
func ExampleValidateURL() {
	// Valid URL
	err := ValidateURL("https://example.com")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid URL")
	}
	// Output: Valid URL
}

// Example of validating a MAC address
func ExampleValidateMACAddress() {
	// Valid MAC address
	err := ValidateMACAddress("00:11:22:33:44:55")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid MAC address")
	}
	// Output: Valid MAC address
}

// Example of validating a subnet mask
func ExampleValidateSubnetMask() {
	// Valid subnet mask
	err := ValidateSubnetMask("255.255.255.0")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid subnet mask")
	}
	// Output: Valid subnet mask
}

// Example of validating a hostname
func ExampleValidateHostname() {
	// Valid hostname
	err := ValidateHostname("example.com")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid hostname")
	}
	// Output: Valid hostname
}

// Example of validating a domain
func ExampleValidateDomain() {
	// Valid domain
	err := ValidateDomain("example.com")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Valid domain")
	}
	// Output: Valid domain
}

// Example of handling validation errors
func ExampleValidationError() {
	// Invalid IP address
	err := ValidateIP("invalid-ip")
	if err != nil {
		// Type assert to get the validation error
		if validationErr, ok := err.(ValidationError); ok {
			fmt.Printf("Field: %s, Message: %s\n", validationErr.Field, validationErr.Message)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
	// Output: Field: IP, Message: invalid IP address format: invalid-ip
}