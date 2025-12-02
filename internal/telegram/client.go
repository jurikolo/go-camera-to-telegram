// Package telegram provides functionality for sending messages to Telegram
package telegram

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Client handles communication with the Telegram Bot API
type Client struct {
	bot    *tgbotapi.BotAPI
	chatID int64
	// Rate limiting: 30 messages per second to different chats
	rateLimiter map[int64]<-chan time.Time
}

// CameraInfo holds information about a camera for message formatting
type CameraInfo struct {
	IP          string
	Name        string
	CaptureTime time.Time
	Metadata    map[string]string
}

// NewClient creates a new Telegram client with rate limiting
func NewClient(botToken string, chatID int64) (*Client, error) {
	// Validate bot token by making a test API call
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API client: %w", err)
	}

	// Test the token by getting bot info
	_, err = bot.GetMe()
	if err != nil {
		return nil, fmt.Errorf("invalid bot token: %w", err)
	}

	return &Client{
		bot:         bot,
		chatID:      chatID,
		rateLimiter: make(map[int64]<-chan time.Time),
	}, nil
}

// Close closes the Telegram client and cleans up resources
func (c *Client) Close() error {
	// The tgbotapi.BotAPI doesn't have an explicit Close method,
	// but we can clear our rate limiter
	c.rateLimiter = nil
	return nil
}

// SendMessage sends a text message to the configured chat
func (c *Client) SendMessage(ctx context.Context, message string) error {
	// Wait for rate limiter
	c.waitForRateLimit(c.chatID)

	msg := tgbotapi.NewMessage(c.chatID, message)
	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendPhoto sends a photo with caption to the configured chat with retry mechanism
func (c *Client) SendPhoto(ctx context.Context, photo io.Reader, caption string) error {
	return c.SendPhotoWithRetry(ctx, photo, caption, 3, 5*time.Second)
}

// SendPhotoWithRetry sends a photo with caption to the configured chat with retry mechanism
func (c *Client) SendPhotoWithRetry(ctx context.Context, photo io.Reader, caption string, maxRetries int, retryDelay time.Duration) error {
	var lastErr error
	
	for i := 0; i <= maxRetries; i++ {
		// Check if context was cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Wait for rate limiter
		c.waitForRateLimit(c.chatID)
		
		// Create a new photo config for each attempt since the reader might be consumed
		photoConfig := tgbotapi.NewPhoto(c.chatID, tgbotapi.FileReader{
			Name:   "photo.jpg",
			Reader: photo,
		})
		photoConfig.Caption = caption
		
		_, lastErr = c.bot.Send(photoConfig)
		if lastErr == nil {
			return nil // Success
		}
		
		// If we have more retries, wait before trying again
		if i < maxRetries {
			// Wait for retry delay or until context is cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
			}
		}
	}
	
	return fmt.Errorf("failed to send photo after %d attempts: %w", maxRetries+1, lastErr)
}

// waitForRateLimit waits for the rate limiter for the given chat ID
func (c *Client) waitForRateLimit(chatID int64) {
	if c.rateLimiter[chatID] == nil {
		// Create a new rate limiter: 30 messages per second
		c.rateLimiter[chatID] = time.Tick(time.Second / 30)
	}
	
	<-c.rateLimiter[chatID]
}

// FormatCameraMessage formats a message for a camera snapshot with markdown support
func (c *Client) FormatCameraMessage(cameraInfo CameraInfo) string {
	var message strings.Builder
	
	// Add camera name or IP
	if cameraInfo.Name != "" {
		message.WriteString(fmt.Sprintf("*Camera:* %s\n", cameraInfo.Name))
	} else {
		message.WriteString(fmt.Sprintf("*Camera IP:* %s\n", cameraInfo.IP))
	}
	
	// Add capture time
	message.WriteString(fmt.Sprintf("*Capture Time:* %s\n", cameraInfo.CaptureTime.Format("2006-01-02 15:04:05")))
	
	// Add metadata if available
	if len(cameraInfo.Metadata) > 0 {
		message.WriteString("\n*Metadata:*\n")
		for key, value := range cameraInfo.Metadata {
			message.WriteString(fmt.Sprintf("• %s: %s\n", key, value))
		}
	}
	
	return message.String()
}