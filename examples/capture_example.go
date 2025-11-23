// Package main provides an example of how to use the camera capture functionality
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jurikolo/go-camera-to-telegram/internal/camera"
)

func main() {
	// Check command line arguments
	if len(os.Args) < 4 {
		fmt.Println("Usage: capture_example <camera_ip> <username> <password>")
		os.Exit(1)
	}
	
	cameraIP := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]
	
	fmt.Printf("Capturing frame from camera at %s...\n", cameraIP)
	
	// Create camera capture instance
	capture := camera.NewCapture()
	
	// Set capture options
	options := &camera.CaptureOptions{
		Timeout:   30 * time.Second,
		Quality:   75,
		MaxWidth:  1920,
		MaxHeight: 1080,
	}
	
	// Capture frame
	jpegData, err := capture.CaptureFrame(cameraIP, username, password, options)
	if err != nil {
		fmt.Printf("Failed to capture frame from camera at %s: %v\n", cameraIP, err)
		os.Exit(1)
	}
	
	fmt.Printf("Successfully captured frame from camera at %s (%d bytes)\n", cameraIP, len(jpegData))
	
	// Save to file
	filename := fmt.Sprintf("capture_%s.jpg", time.Now().Format("20060102_150405"))
	err = os.WriteFile(filename, jpegData, 0644)
	if err != nil {
		fmt.Printf("Failed to save captured frame to %s: %v\n", filename, err)
		os.Exit(1)
	}
	
	fmt.Printf("Saved captured frame to %s\n", filename)
}