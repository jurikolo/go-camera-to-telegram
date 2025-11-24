// Package camera provides tests for the image processing functionality
package camera

import (
	"image"
	"image/color"
	"image/jpeg"
	"bytes"
	"testing"
)

// TestCreateTimestampOverlay tests the CreateTimestampOverlay function
func TestCreateTimestampOverlay(t *testing.T) {
	timestamp := CreateTimestampOverlay()
	if timestamp == "" {
		t.Error("Expected non-empty timestamp")
	}
}

// TestCreateCameraLabelOverlay tests the CreateCameraLabelOverlay function
func TestCreateCameraLabelOverlay(t *testing.T) {
	cameraIP := "192.168.1.100"
	label := CreateCameraLabelOverlay(cameraIP)
	if label != "Camera: 192.168.1.100" {
		t.Errorf("Expected 'Camera: 192.168.1.100', got '%s'", label)
	}
}

// TestProcessOptions tests the ProcessOptions struct
func TestProcessOptions(t *testing.T) {
	options := &ProcessOptions{
		AddTimestamp: true,
		AddCameraID: true,
		CameraID:    "192.168.1.100",
		FontSize:   12,
	}
	
	if !options.AddTimestamp {
		t.Error("Expected AddTimestamp to be true")
	}
	
	if !options.AddCameraID {
		t.Error("Expected AddCameraID to be true")
	}
	
	if options.CameraID != "192.168.1.100" {
		t.Errorf("Expected CameraID '192.168.1.100', got '%s'", options.CameraID)
	}
	
	if options.FontSize != 12 {
		t.Errorf("Expected FontSize 12, got %d", options.FontSize)
	}
}

// TestAddWatermark tests the AddWatermark function with a simple image
func TestAddWatermark(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	// Fill with blue color
	blue := color.RGBA{0, 0, 255, 255}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, blue)
		}
	}
	
	// Encode to JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75})
	if err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}
	
	// Test AddWatermark
	options := &ProcessOptions{
		AddTimestamp: true,
		AddCameraID: true,
		CameraID:    "192.168.1.100",
	}
	
	processedImage, err := AddWatermark(buf.Bytes(), options)
	if err != nil {
		t.Fatalf("AddWatermark failed: %v", err)
	}
	
	if processedImage == nil {
		t.Fatal("AddWatermark returned nil")
	}
	
	if len(processedImage.Data) == 0 {
		t.Error("Processed image data is empty")
	}
	
	if processedImage.Width != 100 {
		t.Errorf("Expected width 100, got %d", processedImage.Width)
	}
	
	if processedImage.Height != 100 {
		t.Errorf("Expected height 100, got %d", processedImage.Height)
	}
	
	if processedImage.CameraIP != "192.168.1.100" {
		t.Errorf("Expected CameraIP '192.168.1.100', got '%s'", processedImage.CameraIP)
	}
}