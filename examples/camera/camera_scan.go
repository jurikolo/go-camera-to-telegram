// Package main demonstrates how to use the camera scanning functionality
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jurikolo/go-camera-to-telegram/internal/camera"
	"github.com/jurikolo/go-camera-to-telegram/internal/config"
)

func main() {
	fmt.Println("Camera Scanner Example")

	// Check for config flag
	configPath := ""
	for i, arg := range os.Args {
		if arg == "--config" && i+1 < len(os.Args) {
			configPath = os.Args[i+1]
			break
		}
	}

	// Load configuration
	var cfg *config.Config
	var err error
	if configPath != "" {
		fmt.Printf("Using config file: %s\n", configPath)
		cfg, err = config.LoadWithFile(configPath)
	} else {
		cfg, err = config.Load()
	}
	
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Scanning network: %s\n", cfg.Network.CIDR)

	// Create camera scanner
	scanner := camera.NewScanner(cfg)

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
