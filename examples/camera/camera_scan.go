// Package main demonstrates how to use the camera scanning functionality
package main

import (
	"fmt"
	"log"

	"github.com/jurikolo/go-camera-to-telegram/internal/camera"
	"github.com/jurikolo/go-camera-to-telegram/internal/config"
)

func main() {
	fmt.Println("Camera Scanner Example")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Scanning network: %s\n", cfg.Network.CIDR)

	// Create camera scanner
	scanner := camera.NewScanner()

	// Scan for cameras
	cameras, err := scanner.ScanNetwork()
	if err != nil {
		log.Fatalf("Failed to scan network: %v", err)
	}

	fmt.Printf("Found %d cameras:\n", len(cameras))
	for _, cameraIP := range cameras {
		fmt.Printf("  - %s\n", cameraIP)

		// Example of capturing an image from a camera
		capture := camera.NewCapture()
		imageData, err := capture.CaptureFrame(cameraIP)
		if err != nil {
			fmt.Printf("    Failed to capture image: %v\n", err)
			continue
		}

		fmt.Printf("    Captured image (%d bytes)\n", len(imageData))
	}

	fmt.Println("Camera scan complete")
}
