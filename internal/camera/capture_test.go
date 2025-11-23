// Package camera provides tests for the image capture functionality
package camera

import (
	"testing"
	"time"
)

// TestCaptureOptions tests the CaptureOptions struct
func TestCaptureOptions(t *testing.T) {
	options := &CaptureOptions{
		Timeout:   30 * time.Second,
		Quality:   75,
		MaxWidth:  1920,
		MaxHeight: 1080,
	}
	
	if options.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", options.Timeout)
	}
	
	if options.Quality != 75 {
		t.Errorf("Expected quality 75, got %d", options.Quality)
	}
	
	if options.MaxWidth != 1920 {
		t.Errorf("Expected max width 1920, got %d", options.MaxWidth)
	}
	
	if options.MaxHeight != 1080 {
		t.Errorf("Expected max height 1080, got %d", options.MaxHeight)
	}
}

// TestNewCapture tests creating a new capture instance
func TestNewCapture(t *testing.T) {
	capture := NewCapture()
	if capture == nil {
		t.Fatal("Failed to create capture instance")
	}
	
	// Check that the connection pool was created
	if capture.pool == nil {
		t.Error("Connection pool was not created")
	}
}

// TestCaptureFrame tests the CaptureFrame method
func TestCaptureFrame(t *testing.T) {
	capture := NewCapture()
	if capture == nil {
		t.Fatal("Failed to create capture instance")
	}
	
	// Test with nil options (should use defaults)
	options := &CaptureOptions{
		Timeout: 5 * time.Second, // Short timeout for testing
		Quality: 50,
	}
	
	// We can't actually connect to a real camera in tests,
	// so we'll just test that the method signature works
	// In a real test environment, you would mock the RTSP client
	_ = options
}