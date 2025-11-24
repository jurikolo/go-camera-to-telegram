package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jurikolo/go-camera-to-telegram/internal/camera"
	"github.com/jurikolo/go-camera-to-telegram/internal/config"
	"github.com/jurikolo/go-camera-to-telegram/internal/telegram"
)

func main() {
	fmt.Println("IP Camera Scanner with Telegram Integration")
	
	// Check for help flag
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Usage: scanner [options]")
		fmt.Println("Options:")
		fmt.Println("  --help     Show this help message")
		fmt.Println("  --config   Path to configuration file")
		return
	}
	
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
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Configuration loaded successfully:\n")
	fmt.Printf("  Network CIDR: %s\n", cfg.Network.CIDR)
	fmt.Printf("  RTSP Username: %s\n", cfg.RTSP.Username)
	fmt.Printf("  Telegram Bot Token: %s\n", cfg.Telegram.BotToken)
	fmt.Printf("  Telegram Chat ID: %d\n", cfg.Telegram.ChatID)
	fmt.Printf("  Scan Interval: %d minutes\n", cfg.Scan.Interval)
	
	// Create Telegram client
	telegramClient, err := telegram.NewClient(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Telegram client: %v\n", err)
		os.Exit(1)
	}
	
	// Create camera scanner
	scanner := camera.NewScanner(cfg)
	
	// Scan network for IP cameras
	fmt.Println("Scanning network for IP cameras...")
	cameras, err := scanner.ScanNetwork()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to scan network: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Found %d cameras:\n", len(cameras))
	for _, cameraIP := range cameras {
		fmt.Printf("  - %s\n", cameraIP)
	}
	
	// Create camera capture instance
	capture := camera.NewCapture()
	
	// Capture images from discovered cameras
	for _, cameraIP := range cameras {
		fmt.Printf("Capturing frame from camera at %s...\n", cameraIP)
		
		// Set capture options
		options := &camera.CaptureOptions{
			Timeout:   5 * time.Second,
			Quality:   75,
			MaxWidth:  1920,
			MaxHeight: 1080,
		}
		
		// Capture frame
		jpegData, err := capture.CaptureFrame(cameraIP, cfg.RTSP.Username, cfg.RTSP.Password, options)
		if err != nil {
			fmt.Printf("Failed to capture frame from camera at %s: %v\n", cameraIP, err)
			continue
		}
		
		fmt.Printf("Successfully captured frame from camera at %s (%d bytes)\n", cameraIP, len(jpegData))
		
		// Add watermark to the captured image
		processOptions := &camera.ProcessOptions{
			AddTimestamp: true,
			AddCameraID: true,
			CameraID:    cameraIP,
		}
		
		processedImage, err := camera.AddWatermark(jpegData, processOptions)
		if err != nil {
			fmt.Printf("Failed to add watermark to image from camera at %s: %v\n", cameraIP, err)
			continue
		}
		
		fmt.Printf("Successfully processed image from camera at %s (%d bytes)\n", cameraIP, len(processedImage.Data))
		
		// Send image to Telegram
		cameraInfo := telegram.CameraInfo{
			IP:          cameraIP,
			CaptureTime: time.Now(),
			Metadata: map[string]string{
				"Image Size": fmt.Sprintf("%d bytes", len(processedImage.Data)),
			},
		}
		caption := telegramClient.FormatCameraMessage(cameraInfo)
		err = telegramClient.SendPhoto(strings.NewReader(string(processedImage.Data)), caption)
		if err != nil {
			fmt.Printf("Failed to send image to Telegram from camera at %s: %v\n", cameraIP, err)
		} else {
			fmt.Printf("Successfully sent image to Telegram from camera at %s\n", cameraIP)
		}
	}
	
	fmt.Println("Scanner initialized successfully.")
}