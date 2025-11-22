// Package config provides configuration management for the application
package config

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Network NetworkConfig `mapstructure:"network"`
	RTSP    RTSPConfig    `mapstructure:"rtsp"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	Scan    ScanConfig    `mapstructure:"scan"`
}

// NetworkConfig holds network-related configuration
type NetworkConfig struct {
	CIDR     string `mapstructure:"cidr"`
	Timeout  int    `mapstructure:"timeout"`
	Workers  int    `mapstructure:"workers"`
}

// RTSPConfig holds RTSP-related configuration
type RTSPConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Timeout  int    `mapstructure:"timeout"`
}

// TelegramConfig holds Telegram-related configuration
type TelegramConfig struct {
	BotToken string `mapstructure:"bot_token"`
	ChatID   string `mapstructure:"chat_id"`
}

// ScanConfig holds scan-related configuration
type ScanConfig struct {
	Interval      int `mapstructure:"interval"`
	MaxConcurrent int `mapstructure:"max_concurrent"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	
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
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return &config, nil
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
	if r.Username == "" {
		return fmt.Errorf("username is required")
	}
	
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	
	if r.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	
	return nil
}

// Validate validates the Telegram configuration
func (t *TelegramConfig) Validate() error {
	if t.BotToken == "" {
		return fmt.Errorf("bot_token is required")
	}
	
	if t.ChatID == "" {
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