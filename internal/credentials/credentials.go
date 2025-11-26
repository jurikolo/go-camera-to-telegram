// Package credentials provides secure credential management for the application
package credentials

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"
)

// CredentialType represents the type of credential
type CredentialType int

const (
	// RTSPPassword represents an RTSP password credential
	RTSPPassword CredentialType = iota
	// TelegramToken represents a Telegram bot token credential
	TelegramToken
)

// String returns the string representation of a CredentialType
func (c CredentialType) String() string {
	switch c {
	case RTSPPassword:
		return "RTSPPassword"
	case TelegramToken:
		return "TelegramToken"
	default:
		return "Unknown"
	}
}

// Credential represents a secure credential
type Credential struct {
	value string
	ctype CredentialType
}

// String implements the Stringer interface but redacts the credential value
func (c *Credential) String() string {
	// Never expose the actual credential value in logs
	return fmt.Sprintf("[REDACTED_%s]", c.ctype.String())
}

// Value returns the actual credential value
func (c *Credential) Value() string {
	return c.value
}

// Type returns the credential type
func (c *Credential) Type() CredentialType {
	return c.ctype
}

// NewCredential creates a new credential with the specified value and type
func NewCredential(value string, ctype CredentialType) *Credential {
	return &Credential{
		value: value,
		ctype: ctype,
	}
}

// LoadFromEnv loads a credential from an environment variable
func LoadFromEnv(envVar string, ctype CredentialType) (*Credential, error) {
	value := os.Getenv(envVar)
	if value == "" {
		return nil, fmt.Errorf("environment variable %s is not set", envVar)
	}
	return NewCredential(value, ctype), nil
}

// LoadFromFile loads a credential from a file with proper permissions
func LoadFromFile(filePath string, ctype CredentialType) (*Credential, error) {
	// Check file permissions (should be readable only by owner)
 fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	// On Unix systems, check if file permissions are too permissive
	mode := fileInfo.Mode()
	if mode&0077 != 0 {
		return nil, fmt.Errorf("file %s has insecure permissions (mode %s), should be readable only by owner", filePath, mode)
	}

	// Read the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Read the first line only (credentials should be single line)
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		value := strings.TrimSpace(scanner.Text())
		if value == "" {
			return nil, fmt.Errorf("file %s is empty or contains only whitespace", filePath)
		}
		return NewCredential(value, ctype), nil
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return nil, fmt.Errorf("file %s is empty", filePath)
}

// LoadFromEncryptedFile loads a credential from an encrypted file
func LoadFromEncryptedFile(filePath, key string, ctype CredentialType) (*Credential, error) {
	// Read the encrypted file
	encryptedData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read encrypted file %s: %w", filePath, err)
	}

	// Decrypt the data
	decryptedData, err := decrypt(encryptedData, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt file %s: %w", filePath, err)
	}

	value := strings.TrimSpace(string(decryptedData))
	if value == "" {
		return nil, fmt.Errorf("decrypted file %s is empty or contains only whitespace", filePath)
	}

	return NewCredential(value, ctype), nil
}

// encrypt encrypts data using AES-GCM
func encrypt(plaintext []byte, key string) ([]byte, error) {
	// Create cipher
	block, err := aes.NewCipher([]byte(key)[:32]) // Use first 32 bytes as key
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func decrypt(ciphertext []byte, key string) ([]byte, error) {
	// Create cipher
	block, err := aes.NewCipher([]byte(key)[:32]) // Use first 32 bytes as key
	if err != nil {
		return nil, err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get nonce size
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// CreateEncryptedFile creates an encrypted file with the given credential
func CreateEncryptedFile(filePath, key, credential string, ctype CredentialType) error {
	// Encrypt the credential
	encryptedData, err := encrypt([]byte(credential), key)
	if err != nil {
		return fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted file %s: %w", filePath, err)
	}

	return nil
}

// Validate ensures the credential is valid and not empty
func (c *Credential) Validate() error {
	if c.value == "" {
		return fmt.Errorf("credential value is empty")
	}
	
	// Additional validation based on credential type
	switch c.ctype {
	case RTSPPassword:
		// RTSP passwords can be any non-empty string
		if len(c.value) < 1 {
			return fmt.Errorf("RTSP password is too short")
		}
	case TelegramToken:
		// Telegram bot tokens are typically in the format 123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ
		if len(c.value) < 10 {
			return fmt.Errorf("Telegram token appears to be too short")
		}
		if !strings.Contains(c.value, ":") {
			return fmt.Errorf("Telegram token should contain a colon")
		}
	}
	
	return nil
}

// Redact returns a redacted version of the credential for logging
func (c *Credential) Redact() string {
	if len(c.value) <= 4 {
		return "[REDACTED]"
	}
	// Show only first 4 characters for debugging purposes
	return fmt.Sprintf("%s[REDACTED]", c.value[:4])
}