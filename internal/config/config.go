// Package config provides configuration management for the application
package config

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jurikolo/go-camera-to-telegram/internal/credentials"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Network NetworkConfig `mapstructure:"network"`
	RTSP    RTSPConfig    `mapstructure:"rtsp"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Scan    ScanConfig    `mapstructure:"scan"`
	
	// Secure credentials
	RTSPUsername *credentials.Credential `mapstructure:"-"`
	RTSPPassword *credentials.Credential `mapstructure:"-"`
	TelegramToken *credentials.Credential `mapstructure:"-"`
}

// NetworkConfig holds network-related configuration
type NetworkConfig struct {
	CIDR     string `mapstructure:"cidr"`
	Timeout  int    `mapstructure:"timeout"`
	Workers  int    `mapstructure:"workers"`
}

// RTSPConfig holds RTSP-related configuration
type RTSPConfig struct {
	Timeout int `mapstructure:"timeout"`
}

// TelegramConfig holds Telegram-related configuration
type TelegramConfig struct {
	ChatID int64 `mapstructure:"chat_id"`
}

// ScanConfig holds scan-related configuration
type ScanConfig struct {
	Interval      int `mapstructure:"interval"`
	MaxConcurrent int `mapstructure:"max_concurrent"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	return LoadWithFile("")
}

// LoadWithFile loads configuration from a specific file and environment variables
func LoadWithFile(configPath string) (*Config, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}
	
	// Set environment variable prefix and enable automatic environment variable binding
	viper.SetEnvPrefix("CTT") // CTT = Camera To Telegram
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	
	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// If config file not found, continue with environment variables only
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Load credentials securely
	if err := config.loadCredentials(); err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
}

// loadCredentials loads credentials from environment variables, files, or encrypted files
func (c *Config) loadCredentials() error {
	var err error
	
	// Load RTSP credentials
	c.RTSPUsername, err = loadRTSPUsername()
	if err != nil {
		return fmt.Errorf("failed to load RTSP username: %w", err)
	}
	
	c.RTSPPassword, err = loadRTSPPassword()
	if err != nil {
		return fmt.Errorf("failed to load RTSP password: %w", err)
	}
	
	// Load Telegram credentials
	c.TelegramToken, err = loadTelegramToken()
	if err != nil {
		return fmt.Errorf("failed to load Telegram token: %w", err)
	}
	
	return nil
}

// loadRTSPUsername loads the RTSP username from environment variables or files
func loadRTSPUsername() (*credentials.Credential, error) {
	// Try environment variable first
	if username := os.Getenv("CTT_RTSP_USERNAME"); username != "" {
		cred := credentials.NewCredential(username, credentials.RTSPPassword)
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid RTSP username: %w", err)
		}
		return cred, nil
	}
	
	// Try file-based credential
	if filePath := os.Getenv("CTT_RTSP_USERNAME_FILE"); filePath != "" {
		cred, err := credentials.LoadFromFile(filePath, credentials.RTSPPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to load RTSP username from file: %w", err)
		}
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid RTSP username from file: %w", err)
		}
		return cred, nil
	}
	
	// Try encrypted file-based credential
	if filePath := os.Getenv("CTT_RTSP_USERNAME_ENCRYPTED_FILE"); filePath != "" {
		key := os.Getenv("CTT_ENCRYPTION_KEY")
		if key == "" {
			return nil, fmt.Errorf("encryption key not provided for encrypted RTSP username file")
		}
		cred, err := credentials.LoadFromEncryptedFile(filePath, key, credentials.RTSPPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to load RTSP username from encrypted file: %w", err)
		}
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid RTSP username from encrypted file: %w", err)
		}
		return cred, nil
	}
	
	// Fall back to config file value if available
	if viper.IsSet("rtsp.username") {
		username := viper.GetString("rtsp.username")
		cred := credentials.NewCredential(username, credentials.RTSPPassword)
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid RTSP username from config: %w", err)
		}
		return cred, nil
	}
	
	return nil, fmt.Errorf("RTSP username not found in environment, files, or config")
}

// loadRTSPPassword loads the RTSP password from environment variables or files
func loadRTSPPassword() (*credentials.Credential, error) {
	// Try environment variable first
	if password := os.Getenv("CTT_RTSP_PASSWORD"); password != "" {
		cred := credentials.NewCredential(password, credentials.RTSPPassword)
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid RTSP password: %w", err)
		}
		return cred, nil
	}
	
	// Try file-based credential
	if filePath := os.Getenv("CTT_RTSP_PASSWORD_FILE"); filePath != "" {
		cred, err := credentials.LoadFromFile(filePath, credentials.RTSPPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to load RTSP password from file: %w", err)
		}
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid RTSP password from file: %w", err)
		}
		return cred, nil
	}
	
	// Try encrypted file-based credential
	if filePath := os.Getenv("CTT_RTSP_PASSWORD_ENCRYPTED_FILE"); filePath != "" {
		key := os.Getenv("CTT_ENCRYPTION_KEY")
		if key == "" {
			return nil, fmt.Errorf("encryption key not provided for encrypted RTSP password file")
		}
		cred, err := credentials.LoadFromEncryptedFile(filePath, key, credentials.RTSPPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to load RTSP password from encrypted file: %w", err)
		}
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid RTSP password from encrypted file: %w", err)
		}
		return cred, nil
	}
	
	// Fall back to config file value if available
	if viper.IsSet("rtsp.password") {
		password := viper.GetString("rtsp.password")
		cred := credentials.NewCredential(password, credentials.RTSPPassword)
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid RTSP password from config: %w", err)
		}
		return cred, nil
	}
	
	return nil, fmt.Errorf("RTSP password not found in environment, files, or config")
}

// loadTelegramToken loads the Telegram token from environment variables or files
func loadTelegramToken() (*credentials.Credential, error) {
	// Try environment variable first
	if token := os.Getenv("CTT_TELEGRAM_BOT_TOKEN"); token != "" {
		cred := credentials.NewCredential(token, credentials.TelegramToken)
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid Telegram token: %w", err)
		}
		return cred, nil
	}
	
	// Try file-based credential
	if filePath := os.Getenv("CTT_TELEGRAM_BOT_TOKEN_FILE"); filePath != "" {
		cred, err := credentials.LoadFromFile(filePath, credentials.TelegramToken)
		if err != nil {
			return nil, fmt.Errorf("failed to load Telegram token from file: %w", err)
		}
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid Telegram token from file: %w", err)
		}
		return cred, nil
	}
	
	// Try encrypted file-based credential
	if filePath := os.Getenv("CTT_TELEGRAM_BOT_TOKEN_ENCRYPTED_FILE"); filePath != "" {
		key := os.Getenv("CTT_ENCRYPTION_KEY")
		if key == "" {
			return nil, fmt.Errorf("encryption key not provided for encrypted Telegram token file")
		}
		cred, err := credentials.LoadFromEncryptedFile(filePath, key, credentials.TelegramToken)
		if err != nil {
			return nil, fmt.Errorf("failed to load Telegram token from encrypted file: %w", err)
		}
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid Telegram token from encrypted file: %w", err)
		}
		return cred, nil
	}
	
	// Fall back to config file value if available
	if viper.IsSet("telegram.bot_token") {
		token := viper.GetString("telegram.bot_token")
		cred := credentials.NewCredential(token, credentials.TelegramToken)
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf("invalid Telegram token from config: %w", err)
		}
		return cred, nil
	}
	
	return nil, fmt.Errorf("Telegram token not found in environment, files, or config")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.Network.Validate(); err != nil {
		return fmt.Errorf("network config validation failed: %w", err)
	}
	
	if err := c.RTSP.Validate(); err != nil {
		return fmt.Errorf("rtsp config validation failed: %w", err)
	}
	
	if err := c.Telegram.Validate(); err != nil {
		return fmt.Errorf("telegram config validation failed: %w", err)
	}
	
	if err := c.Scan.Validate(); err != nil {
		return fmt.Errorf("scan config validation failed: %w", err)
	}
	
	// Validate credentials are loaded
	if c.RTSPUsername == nil {
		return fmt.Errorf("RTSP username is required")
	}
	
	if c.RTSPPassword == nil {
		return fmt.Errorf("RTSP password is required")
	}
	
	if c.TelegramToken == nil {
		return fmt.Errorf("Telegram token is required")
	}
	
	// Validate credentials
	if err := c.RTSPUsername.Validate(); err != nil {
		return fmt.Errorf("RTSP username validation failed: %w", err)
	}
	
	if err := c.RTSPPassword.Validate(); err != nil {
		return fmt.Errorf("RTSP password validation failed: %w", err)
	}
	
	if err := c.TelegramToken.Validate(); err != nil {
		return fmt.Errorf("Telegram token validation failed: %w", err)
	}
	
	return nil
}

// Validate validates the network configuration
func (n *NetworkConfig) Validate() error {
	if n.CIDR == "" {
		return fmt.Errorf("cidr is required")
	}
	
	if _, _, err := net.ParseCIDR(n.CIDR); err != nil {
		return fmt.Errorf("invalid cidr format: %w", err)
	}
	
	if n.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	if n.Workers <= 0 {
		return fmt.Errorf("workers must be positive")
	}
	
	return nil
}

// Validate validates the RTSP configuration
func (r *RTSPConfig) Validate() error {
	if r.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	return nil
}

// Validate validates the Telegram configuration
func (t *TelegramConfig) Validate() error {
	if t.ChatID == 0 {
		return fmt.Errorf("chat_id is required")
	}
	
	return nil
}

// Validate validates the scan configuration
func (s *ScanConfig) Validate() error {
	if s.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	
	if s.MaxConcurrent <= 0 {
		return fmt.Errorf("max_concurrent must be positive")
	}
	
	return nil
}