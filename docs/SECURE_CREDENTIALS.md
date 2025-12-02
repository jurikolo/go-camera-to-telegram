# Secure Credential Management

This document explains how to use the secure credential management system in the go-camera-to-telegram application.

## Overview

The secure credential management system provides multiple ways to store and load sensitive credentials such as RTSP passwords and Telegram bot tokens:

1. Environment variables (most secure)
2. Files with proper permissions (secure)
3. Encrypted files (secure)
4. Configuration files (less secure, but convenient for development)

## Credential Types

The system supports the following credential types:

- `RTSPPassword` - For RTSP camera passwords
- `TelegramToken` - For Telegram bot tokens

## Loading Credentials

### Environment Variables

The most secure way to provide credentials is through environment variables:

```bash
# RTSP credentials
export CTT_RTSP_USERNAME="admin"
export CTT_RTSP_PASSWORD="secretpassword"

# Telegram credentials
export CTT_TELEGRAM_BOT_TOKEN="123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"

# Encryption key for encrypted files (if used)
export CTT_ENCRYPTION_KEY="your-32-byte-encryption-key-here!!"
```

### File-based Credentials

You can store credentials in files with proper permissions. The system will check that the files have secure permissions (readable only by the owner):

```bash
# Create credential files
echo "admin" > /etc/camera-to-telegram/rtsp-username
echo "secretpassword" > /etc/camera-to-telegram/rtsp-password
echo "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ" > /etc/camera-to-telegram/telegram-token

# Set secure permissions
chmod 600 /etc/camera-to-telegram/rtsp-username
chmod 600 /etc/camera-to-telegram/rtsp-password
chmod 600 /etc/camera-to-telegram/telegram-token

# Set environment variables to point to the files
export CTT_RTSP_USERNAME_FILE="/etc/camera-to-telegram/rtsp-username"
export CTT_RTSP_PASSWORD_FILE="/etc/camera-to-telegram/rtsp-password"
export CTT_TELEGRAM_BOT_TOKEN_FILE="/etc/camera-to-telegram/telegram-token"
```

### Encrypted File-based Credentials

For maximum security, you can store credentials in encrypted files:

```bash
# Create encrypted credential files (this would be done by an admin script)
# The application provides functions to create these files:
# credentials.CreateEncryptedFile("/etc/camera-to-telegram/rtsp-password.enc", encryptionKey, "secretpassword", credentials.RTSPPassword)

# Set environment variables to point to the encrypted files
export CTT_ENCRYPTION_KEY="your-32-byte-encryption-key-here!!"
export CTT_RTSP_PASSWORD_ENCRYPTED_FILE="/etc/camera-to-telegram/rtsp-password.enc"
export CTT_TELEGRAM_BOT_TOKEN_ENCRYPTED_FILE="/etc/camera-to-telegram/telegram-token.enc"
```

## Configuration File Fallback

If credentials are not provided through environment variables or files, the system will fall back to values in the configuration file. This is less secure and should only be used for development:

```yaml
rtsp:
  username: "admin"
  password: "secretpassword"
  timeout: 5

telegram:
  bot_token: "123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ"
  chat_id: 123456789
```

## Security Best Practices

1. **Never** store credentials in configuration files in production
2. Use environment variables for containerized deployments
3. Use file-based credentials with proper permissions for traditional deployments
4. Use encrypted files for maximum security
5. Always set proper file permissions (600) for credential files
6. Use different credentials for different environments
7. Rotate credentials regularly

## Preventing Credential Leaks

The credential system automatically prevents credential leaks in logs:

```go
// When logging credentials, they are automatically redacted
cred := credentials.NewCredential("secretpassword", credentials.RTSPPassword)
fmt.Printf("Credential: %s\n", cred) // Outputs: [REDACTED_RTSPPassword]

// To get the actual value, use the Value() method
actualValue := cred.Value()

// For debugging, you can use the Redact() method to show partial values
debugValue := cred.Redact() // Shows first 4 characters + [REDACTED]
```

## Using Credentials in Code

The application automatically loads credentials during configuration loading. You can access them through the config object:

```go
cfg, err := config.Load()
if err != nil {
    log.Fatal("Failed to load configuration: %v", err)
}

// Access credentials
rtspUsername := cfg.RTSPUsername.Value()
rtspPassword := cfg.RTSPPassword.Value()
telegramToken := cfg.TelegramToken.Value()
```

## Testing

The credential system includes comprehensive tests to ensure proper functionality and security:

```bash
go test ./internal/credentials
```

## Troubleshooting

### Permission Errors

If you get permission errors when loading file-based credentials, ensure the files have proper permissions:

```bash
chmod 600 /path/to/credential/file
```

### Encryption Key Issues

If using encrypted files, ensure the encryption key is properly set:

```bash
export CTT_ENCRYPTION_KEY="your-32-byte-encryption-key-here!!"
```

### Credential Validation Errors

If credentials fail validation, check that they meet the required format:

- RTSP passwords: Non-empty string
- Telegram tokens: Format `123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ` (numbers, colon, then alphanumeric characters)