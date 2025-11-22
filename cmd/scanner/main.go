package main

import (
	"fmt"
	"os"

	"github.com/jurikolo/go-camera-to-telegram/internal/config"
)

func main() {
	fmt.Println("IP Camera Scanner with Telegram Integration")
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Configuration loaded successfully:\n")
	fmt.Printf("  Network CIDR: %s\n", cfg.Network.CIDR)
	fmt.Printf("  RTSP Username: %s\n", cfg.RTSP.Username)
	fmt.Printf("  Telegram Bot Token: %s\n", cfg.Telegram.BotToken)
	fmt.Printf("  Scan Interval: %d minutes\n", cfg.Scan.Interval)
	
	// TODO: Implement main application logic
	// 1. Scan network for IP cameras
	// 2. Capture images from discovered cameras
	// 3. Send images to Telegram
	
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Usage: scanner [options]")
		fmt.Println("Options:")
		fmt.Println("  --help     Show this help message")
		fmt.Println("  --config   Path to configuration file")
		return
	}
	
	fmt.Println("Scanner initialized successfully.")
}