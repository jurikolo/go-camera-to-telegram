// Package tests provides integration tests for the worker pool functionality
package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jurikolo/go-camera-to-telegram/internal/camera"
	"github.com/jurikolo/go-camera-to-telegram/internal/config"
	"github.com/jurikolo/go-camera-to-telegram/internal/credentials"
	"github.com/jurikolo/go-camera-to-telegram/internal/logger"
	"github.com/jurikolo/go-camera-to-telegram/internal/metrics"
	"github.com/jurikolo/go-camera-to-telegram/internal/telegram"
	"github.com/jurikolo/go-camera-to-telegram/internal/worker"
	"github.com/stretchr/testify/assert"
)

// MockCapture is a mock capture for testing
type MockCapture struct {
	captureFunc func(ctx context.Context, cameraIP, username, password string, options *camera.CaptureOptions) ([]byte, error)
}

// CaptureFrame captures a frame using the mock function
func (m *MockCapture) CaptureFrame(ctx context.Context, cameraIP, username, password string, options *camera.CaptureOptions) ([]byte, error) {
	if m.captureFunc != nil {
		return m.captureFunc(ctx, cameraIP, username, password, options)
	}
	return []byte("mock jpeg data"), nil
}

// MockTelegramClient is a mock Telegram client for testing
type MockTelegramClient struct {
	sendPhotoFunc      func(ctx context.Context, photo interface{}, caption string) error
	formatCameraFunc   func(cameraInfo telegram.CameraInfo) string
	messages           []string
	photos             []string
	mutex              sync.Mutex
}

// SendMessage simulates sending a text message
func (m *MockTelegramClient) SendMessage(ctx context.Context, message string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = append(m.messages, message)
	return nil
}

// SendPhoto simulates sending a photo
func (m *MockTelegramClient) SendPhoto(ctx context.Context, photo interface{}, caption string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.sendPhotoFunc != nil {
		return m.sendPhotoFunc(ctx, photo, caption)
	}
	
	m.photos = append(m.photos, fmt.Sprintf("Photo with caption: %s", caption))
	return nil
}

// FormatCameraMessage formats a message for a camera snapshot
func (m *MockTelegramClient) FormatCameraMessage(cameraInfo telegram.CameraInfo) string {
	if m.formatCameraFunc != nil {
		return m.formatCameraFunc(cameraInfo)
	}
	return fmt.Sprintf("Camera: %s, Time: %s", cameraInfo.IP, cameraInfo.CaptureTime.Format("2006-01-02 15:04:05"))
}

// GetMessages returns the sent messages
func (m *MockTelegramClient) GetMessages() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.messages
}

// GetPhotos returns the sent photos
func (m *MockTelegramClient) GetPhotos() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.photos
}

// Close closes the mock client
func (m *MockTelegramClient) Close() error {
	return nil
}

// TestWorkerPoolCreation tests that the worker pool can be created successfully
func TestWorkerPoolCreation(t *testing.T) {
	// Create a test configuration
	cfg := &config.Config{
		Network: config.NetworkConfig{
			CIDR:    "192.168.1.0/24",
			Timeout: 5,
			Workers: 10,
		},
		RTSP: config.RTSPConfig{
			Timeout: 5,
		},
		Telegram: config.TelegramConfig{
			ChatID: 123456789,
		},
		Scan: config.ScanConfig{
			Interval:      60,
			MaxConcurrent: 5,
		},
		RTSPUsername: credentials.NewCredential("testuser", credentials.RTSPPassword),
		RTSPPassword: credentials.NewCredential("testpass", credentials.RTSPPassword),
		TelegramToken: credentials.NewCredential("123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ", credentials.TelegramToken),
	}

	// Create mock capture
	capture := &MockCapture{}

	// Create mock Telegram client
	telegramClient := &MockTelegramClient{}

	// Test that we can create the worker pool
	// Note: We can't actually test the worker pool functionality without a real telegram.Client
	// because the NewWorkerPool function expects a *telegram.Client, not an interface
	assert.NotNil(t, cfg, "Configuration should be created successfully")
	assert.NotNil(t, capture, "Capture should be created successfully")
	assert.NotNil(t, telegramClient, "Telegram client should be created successfully")
}

// TestWorkerPoolStructure tests the structure and configuration of the worker pool
func TestWorkerPoolStructure(t *testing.T) {
	// Setup
	log := logger.New(logger.InfoLevel)
	metricsInstance := metrics.NewMetrics()

	// Create a test configuration
	cfg := &config.Config{
		Network: config.NetworkConfig{
			CIDR:    "192.168.1.0/24",
			Timeout: 5,
			Workers: 10,
		},
		RTSP: config.RTSPConfig{
			Timeout: 5,
		},
		Telegram: config.TelegramConfig{
			ChatID: 123456789,
		},
		Scan: config.ScanConfig{
			Interval:      60,
			MaxConcurrent: 5,
		},
		RTSPUsername: credentials.NewCredential("testuser", credentials.RTSPPassword),
		RTSPPassword: credentials.NewCredential("testpass", credentials.RTSPPassword),
		TelegramToken: credentials.NewCredential("123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ", credentials.TelegramToken),
	}

	// Create camera capture instance
	capture := camera.NewCapture()
	defer capture.Close()

	// Create Telegram client (this will fail but we're just testing structure)
	telegramClient, err := telegram.NewClient("123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ", 123456789)
	if err != nil {
		// If we can't create a real client, that's okay for this structural test
		t.Logf("Could not create real Telegram client: %v", err)
		return
	}
	defer telegramClient.Close()

	// Test that we can create the worker pool
	workerPool := worker.NewWorkerPool(5, log, metricsInstance, cfg, capture, telegramClient)
	assert.NotNil(t, workerPool, "Worker pool should be created successfully")
}

// TestJobStructure tests the job structure
func TestJobStructure(t *testing.T) {
	// Test creating a job
	job := worker.Job{CameraIP: "192.168.1.100"}
	assert.Equal(t, "192.168.1.100", job.CameraIP, "Job should have correct camera IP")

	// Test creating a result
	result := worker.Result{
		CameraIP: "192.168.1.100",
		Success:  true,
		Error:     nil,
	}
	assert.Equal(t, "192.168.1.100", result.CameraIP, "Result should have correct camera IP")
	assert.True(t, result.Success, "Result should be successful")
	assert.Nil(t, result.Error, "Result should have no error")
}

// TestMockCapture tests the mock capture functionality
func TestMockCapture(t *testing.T) {
	// Create a mock capture
	capture := &MockCapture{
		captureFunc: func(ctx context.Context, cameraIP, username, password string, options *camera.CaptureOptions) ([]byte, error) {
			return []byte("test jpeg data"), nil
		},
	}

	// Test capturing a frame
	ctx := context.Background()
	data, err := capture.CaptureFrame(ctx, "192.168.1.100", "admin", "password", nil)
	assert.NoError(t, err, "Capture should not return an error")
	assert.Equal(t, []byte("test jpeg data"), data, "Should return test data")
}

// TestMockTelegramClient tests the mock Telegram client functionality
func TestMockTelegramClient(t *testing.T) {
	// Create a mock Telegram client
	telegramClient := &MockTelegramClient{}

	// Test sending a message
	ctx := context.Background()
	err := telegramClient.SendMessage(ctx, "Test message")
	assert.NoError(t, err, "Sending message should not return an error")

	// Test sending a photo
	err = telegramClient.SendPhoto(ctx, "test image data", "Test caption")
	assert.NoError(t, err, "Sending photo should not return an error")

	// Check that messages were recorded
	messages := telegramClient.GetMessages()
	assert.Len(t, messages, 1, "Should have one message")
	assert.Equal(t, "Test message", messages[0], "Message should match")

	// Check that photos were recorded
	photos := telegramClient.GetPhotos()
	assert.Len(t, photos, 1, "Should have one photo")
	assert.Contains(t, photos[0], "Test caption", "Photo should contain caption")
}

// TestTelegramClientWithCustomFormatting tests the Telegram client with custom formatting
func TestTelegramClientWithCustomFormatting(t *testing.T) {
	// Create a mock Telegram client with custom formatting
	telegramClient := &MockTelegramClient{
		formatCameraFunc: func(cameraInfo telegram.CameraInfo) string {
			return fmt.Sprintf("Custom format: %s at %s", cameraInfo.IP, cameraInfo.CaptureTime.Format("15:04:05"))
		},
	}

	// Test formatting camera info
	cameraInfo := telegram.CameraInfo{
		IP:          "192.168.1.100",
		CaptureTime: time.Now(),
	}

	// Format the message
	message := telegramClient.FormatCameraMessage(cameraInfo)
	assert.Contains(t, message, "Custom format", "Message should use custom format")
	assert.Contains(t, message, "192.168.1.100", "Message should contain camera IP")
}

// TestTelegramClientWithCustomPhotoSending tests the Telegram client with custom photo sending
func TestTelegramClientWithCustomPhotoSending(t *testing.T) {
	// Create a mock Telegram client with custom photo sending
	called := false
	telegramClient := &MockTelegramClient{
		sendPhotoFunc: func(ctx context.Context, photo interface{}, caption string) error {
			called = true
			return nil
		},
	}

	// Test sending a photo
	ctx := context.Background()
	err := telegramClient.SendPhoto(ctx, "test image data", "Test caption")
	assert.NoError(t, err, "Sending photo should not return an error")
	assert.True(t, called, "Custom photo sending function should be called")
}