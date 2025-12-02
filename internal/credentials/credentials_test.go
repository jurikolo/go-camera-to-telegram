// Package credentials provides tests for secure credential management
package credentials

import (
	"os"
	"testing"
)

func TestCredentialCreation(t *testing.T) {
	cred := NewCredential("test-password", RTSPPassword)
	if cred.Value() != "test-password" {
		t.Errorf("Expected value 'test-password', got '%s'", cred.Value())
	}
	
	if cred.Type() != RTSPPassword {
		t.Errorf("Expected type RTSPPassword, got %d", cred.Type())
	}
}

func TestCredentialString(t *testing.T) {
	cred := NewCredential("test-password", RTSPPassword)
	str := cred.String()
	if str != "[REDACTED_RTSPPassword]" {
		t.Errorf("Expected '[REDACTED_RTSPPassword]', got '%s'", str)
	}
}

func TestCredentialRedact(t *testing.T) {
	cred := NewCredential("test-password", RTSPPassword)
	redacted := cred.Redact()
	if redacted != "test[REDACTED]" {
		t.Errorf("Expected 'test[REDACTED]', got '%s'", redacted)
	}
	
	// Test with short credential
	shortCred := NewCredential("abc", RTSPPassword)
	shortRedacted := shortCred.Redact()
	if shortRedacted != "[REDACTED]" {
		t.Errorf("Expected '[REDACTED]', got '%s'", shortRedacted)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set a test environment variable
	os.Setenv("TEST_CREDENTIAL", "test-value")
	defer os.Unsetenv("TEST_CREDENTIAL")
	
	cred, err := LoadFromEnv("TEST_CREDENTIAL", RTSPPassword)
	if err != nil {
		t.Fatalf("Failed to load credential from environment: %v", err)
	}
	
	if cred.Value() != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", cred.Value())
	}
	
	// Test with non-existent environment variable
	_, err = LoadFromEnv("NON_EXISTENT_VAR", RTSPPassword)
	if err == nil {
		t.Error("Expected error for non-existent environment variable, got nil")
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary file with test content
	content := "test-file-password"
	tmpfile, err := os.CreateTemp("", "credential-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	
	// Set proper file permissions
	if err := os.Chmod(tmpfile.Name(), 0600); err != nil {
		t.Fatal(err)
	}
	
	cred, err := LoadFromFile(tmpfile.Name(), RTSPPassword)
	if err != nil {
		t.Fatalf("Failed to load credential from file: %v", err)
	}
	
	if cred.Value() != content {
		t.Errorf("Expected value '%s', got '%s'", content, cred.Value())
	}
}

func TestLoadFromFileWithBadPermissions(t *testing.T) {
	// Create a temporary file with test content
	content := "test-file-password"
	tmpfile, err := os.CreateTemp("", "credential-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	
	// Set bad file permissions (readable by others)
	if err := os.Chmod(tmpfile.Name(), 0644); err != nil {
		t.Fatal(err)
	}
	
	_, err = LoadFromFile(tmpfile.Name(), RTSPPassword)
	if err == nil {
		t.Error("Expected error for file with bad permissions, got nil")
	}
}

func TestLoadFromEncryptedFile(t *testing.T) {
	// Create a temporary file with test content
	content := "test-encrypted-password"
	tmpfile, err := os.CreateTemp("", "credential-test-encrypted")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	// Create an encrypted file
	key := "test-encryption-key-32-bytes-long!!"
	if err := CreateEncryptedFile(tmpfile.Name(), key, content, RTSPPassword); err != nil {
		t.Fatal(err)
	}
	
	cred, err := LoadFromEncryptedFile(tmpfile.Name(), key, RTSPPassword)
	if err != nil {
		t.Fatalf("Failed to load credential from encrypted file: %v", err)
	}
	
	if cred.Value() != content {
		t.Errorf("Expected value '%s', got '%s'", content, cred.Value())
	}
}

func TestCredentialValidation(t *testing.T) {
	// Test valid credentials
	validRTSP := NewCredential("valid-password", RTSPPassword)
	if err := validRTSP.Validate(); err != nil {
		t.Errorf("Valid RTSP credential failed validation: %v", err)
	}
	
	validTelegram := NewCredential("123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ", TelegramToken)
	if err := validTelegram.Validate(); err != nil {
		t.Errorf("Valid Telegram credential failed validation: %v", err)
	}
	
	// Test invalid credentials
	emptyCred := NewCredential("", RTSPPassword)
	if err := emptyCred.Validate(); err == nil {
		t.Error("Empty credential should fail validation")
	}
	
	shortTelegram := NewCredential("short", TelegramToken)
	if err := shortTelegram.Validate(); err == nil {
		t.Error("Short Telegram token should fail validation")
	}
	
	noColonTelegram := NewCredential("123456789ABCdefGhIJKlmNoPQRsTUVwxyZ", TelegramToken)
	if err := noColonTelegram.Validate(); err == nil {
		t.Error("Telegram token without colon should fail validation")
	}
}