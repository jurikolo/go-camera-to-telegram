// Package credentials provides examples for secure credential management
package credentials

import (
	"fmt"
)

// Example of how to use secure credentials in your application
func Example_secureCredentials() {
	// In a real application, you would load credentials from environment variables,
	// files, or encrypted files as shown in the config package
	
	// For this example, we'll create credentials directly
	rtspUsername := NewCredential("admin", RTSPPassword)
	rtspPassword := NewCredential("secretpassword", RTSPPassword)
	telegramToken := NewCredential("123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ", TelegramToken)
	
	// When logging, credentials are automatically redacted
	fmt.Printf("RTSP Username: %s\n", rtspUsername)
	fmt.Printf("RTSP Password: %s\n", rtspPassword)
	fmt.Printf("Telegram Token: %s\n", telegramToken)
	
	// To get the actual values, use the Value() method
	fmt.Printf("RTSP Username Value: %s\n", rtspUsername.Value())
	fmt.Printf("RTSP Password Value: %s\n", rtspPassword.Value())
	fmt.Printf("Telegram Token Value: %s\n", telegramToken.Value())
	
	// For debugging, you can use the Redact() method to show partial values
	fmt.Printf("RTSP Username Redacted: %s\n", rtspUsername.Redact())
	fmt.Printf("RTSP Password Redacted: %s\n", rtspPassword.Redact())
	fmt.Printf("Telegram Token Redacted: %s\n", telegramToken.Redact())
	
	// Output:
	// RTSP Username: [REDACTED_RTSPPassword]
	// RTSP Password: [REDACTED_RTSPPassword]
	// Telegram Token: [REDACTED_TelegramToken]
	// RTSP Username Value: admin
	// RTSP Password Value: secretpassword
	// Telegram Token Value: 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ
	// RTSP Username Redacted: admi[REDACTED]
	// RTSP Password Redacted: secr[REDACTED]
	// Telegram Token Redacted: 1234[REDACTED]
}