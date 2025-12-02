// Package telegram provides mock Telegram API responses for testing
package telegram

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockBotAPI is a mock implementation of the Telegram Bot API
type MockBotAPI struct {
	Messages   []string
	Photos     []string
	LastChatID int64
}

// Send simulates sending a message
func (m *MockBotAPI) Send(c interface{}) (interface{}, error) {
	// Extract message content based on type
	switch msg := c.(type) {
	case struct {
		ChatID int64
		Text   string
	}:
		m.Messages = append(m.Messages, msg.Text)
		m.LastChatID = msg.ChatID
	case struct {
		ChatID int64
		Photo  struct {
			Name   string
			Reader io.Reader
		}
		Caption string
	}:
		// Read the photo content
		content, _ := io.ReadAll(msg.Photo.Reader)
		m.Photos = append(m.Photos, string(content))
		m.LastChatID = msg.ChatID
	}

	return nil, nil
}

// GetMe simulates getting bot information
func (m *MockBotAPI) GetMe() (interface{}, error) {
	return struct {
		UserName string
	}{UserName: "testbot"}, nil
}

// MockTelegramClient is a mock Telegram client for testing
type MockTelegramClient struct {
	bot         *MockBotAPI
	chatID      int64
	rateLimiter map[int64]<-chan time.Time
}

// NewMockTelegramClient creates a new mock Telegram client
func NewMockTelegramClient(chatID int64) *MockTelegramClient {
	return &MockTelegramClient{
		bot:         &MockBotAPI{},
		chatID:      chatID,
		rateLimiter: make(map[int64]<-chan time.Time),
	}
}

// SendMessage simulates sending a text message
func (c *MockTelegramClient) SendMessage(ctx context.Context, message string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Simulate rate limiting
	c.waitForRateLimit(c.chatID)

	// Check if context is cancelled after rate limiting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create a message struct
	msg := struct {
		ChatID int64
		Text   string
	}{
		ChatID: c.chatID,
		Text:   message,
	}

	_, err := c.bot.Send(msg)
	return err
}

// SendPhoto simulates sending a photo
func (c *MockTelegramClient) SendPhoto(ctx context.Context, photo io.Reader, caption string) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Simulate rate limiting
	c.waitForRateLimit(c.chatID)

	// Check if context is cancelled after rate limiting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create a photo struct
	photoMsg := struct {
		ChatID int64
		Photo  struct {
			Name   string
			Reader io.Reader
		}
		Caption string
	}{
		ChatID: c.chatID,
		Photo: struct {
			Name   string
			Reader io.Reader
		}{
			Name:   "photo.jpg",
			Reader: photo,
		},
		Caption: caption,
	}

	_, err := c.bot.Send(photoMsg)
	return err
}

// waitForRateLimit simulates waiting for rate limiting
func (c *MockTelegramClient) waitForRateLimit(chatID int64) {
	if c.rateLimiter[chatID] == nil {
		// Create a new rate limiter: 30 messages per second
		c.rateLimiter[chatID] = time.Tick(time.Second / 30)
	}

	// In a real implementation, we would wait here
	// For testing, we'll just return immediately
}

// FormatCameraMessage formats a message for a camera snapshot with markdown support
func (c *MockTelegramClient) FormatCameraMessage(cameraInfo CameraInfo) string {
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

// GetMessages returns the sent messages
func (c *MockTelegramClient) GetMessages() []string {
	return c.bot.Messages
}

// GetPhotos returns the sent photos
func (c *MockTelegramClient) GetPhotos() []string {
	return c.bot.Photos
}

// TestMockTelegramClient tests the mock Telegram client functionality
func TestMockTelegramClient(t *testing.T) {
	// Create a mock Telegram client
	client := NewMockTelegramClient(123456789)
	assert.NotNil(t, client)

	// Test sending a message
	ctx := context.Background()
	err := client.SendMessage(ctx, "Hello, World!")
	assert.NoError(t, err)

	// Check that the message was recorded
	messages := client.GetMessages()
	assert.Len(t, messages, 1)
	assert.Equal(t, "Hello, World!", messages[0])

	// Test sending a photo
	photoData := strings.NewReader("fake image data")
	err = client.SendPhoto(ctx, photoData, "Test photo")
	assert.NoError(t, err)

	// Check that the photo was recorded
	photos := client.GetPhotos()
	assert.Len(t, photos, 1)
	assert.Equal(t, "fake image data", photos[0])

	// Test with different chat ID
	client2 := NewMockTelegramClient(987654321)
	err = client2.SendMessage(ctx, "Different chat")
	assert.NoError(t, err)

	messages2 := client2.GetMessages()
	assert.Len(t, messages2, 1)
	assert.Equal(t, "Different chat", messages2[0])
}

// TestTelegramClientWithTimeout tests the Telegram client with timeout
func TestTelegramClientWithTimeout(t *testing.T) {
	// Create a mock Telegram client
	client := NewMockTelegramClient(123456789)
	assert.NotNil(t, client)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Test sending a message with timeout
	err := client.SendMessage(ctx, "Test message")
	assert.NoError(t, err)

	// Check that the message was recorded
	messages := client.GetMessages()
	assert.Len(t, messages, 1)
	assert.Equal(t, "Test message", messages[0])
}

// TestTelegramClientWithCancellation tests the Telegram client with context cancellation
func TestTelegramClientWithCancellation(t *testing.T) {
	// Create a mock Telegram client
	client := NewMockTelegramClient(123456789)
	assert.NotNil(t, client)

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	
	// Cancel the context immediately
	cancel()

	// Test sending a message with cancelled context
	err := client.SendMessage(ctx, "Test message")
	// We expect the context to be cancelled
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	// Check that no message was recorded
	messages := client.GetMessages()
	assert.Len(t, messages, 0)
}

// TestFormatCameraMessage tests the FormatCameraMessage function
func TestFormatCameraMessage(t *testing.T) {
	// Create a mock Telegram client
	client := NewMockTelegramClient(123456789)
	assert.NotNil(t, client)

	// Create camera info with name
	cameraInfo := CameraInfo{
		IP:          "192.168.1.100",
		Name:        "Front Door Camera",
		CaptureTime: time.Now(),
		Metadata: map[string]string{
			"Resolution": "1920x1080",
			"FPS":        "30",
		},
	}

	// Test formatting with camera name
	message := client.FormatCameraMessage(cameraInfo)
	assert.Contains(t, message, "Front Door Camera")
	assert.Contains(t, message, "Resolution")
	assert.Contains(t, message, "FPS")
	// When name is provided, IP should not be shown
	assert.NotContains(t, message, "192.168.1.100")

	// Create camera info without name
	cameraInfoNoName := CameraInfo{
		IP:          "192.168.1.101",
		Name:        "",
		CaptureTime: time.Now(),
		Metadata: map[string]string{
			"Resolution": "1280x720",
		},
	}

	// Test formatting without camera name
	messageNoName := client.FormatCameraMessage(cameraInfoNoName)
	assert.Contains(t, messageNoName, "192.168.1.101")
	assert.Contains(t, messageNoName, "Resolution")
	assert.NotContains(t, messageNoName, "Camera:")
	assert.Contains(t, messageNoName, "Camera IP:")
}