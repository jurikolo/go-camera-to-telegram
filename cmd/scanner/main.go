package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jurikolo/go-camera-to-telegram/internal/camera"
	"github.com/jurikolo/go-camera-to-telegram/internal/config"
	"github.com/jurikolo/go-camera-to-telegram/internal/logger"
	"github.com/jurikolo/go-camera-to-telegram/internal/telegram"
)

func main() {
	// Create logger
	log := logger.New(logger.InfoLevel)

	log.Info("IP Camera Scanner with Telegram Integration starting...")

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
		log.Info("Using config file: %s", configPath)
		cfg, err = config.LoadWithFile(configPath)
	} else {
		cfg, err = config.Load()
	}

	if err != nil {
		log.Fatal("Failed to load configuration: %v", err)
	}

	log.Info("Configuration loaded successfully")
	log.Info("  Network CIDR: %s", cfg.Network.CIDR)
	log.Info("  RTSP Username: %s", cfg.RTSP.Username)
	log.Info("  Telegram Bot Token: %s", cfg.Telegram.BotToken)
	log.Info("  Telegram Chat ID: %d", cfg.Telegram.ChatID)
	log.Info("  Scan Interval: %d minutes", cfg.Scan.Interval)

	// Create Telegram client
	telegramClient, err := telegram.NewClient(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
	if err != nil {
		log.Fatal("Failed to create Telegram client: %v", err)
	}

	// Create camera scanner
	scanner := camera.NewScanner(cfg)

	// Create camera capture instance
	capture := camera.NewCapture()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT and SIGTERM for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Info("Received signal %s, shutting down gracefully...", sig)
		cancel()
	}()

	// Send startup message to Telegram
	startupMsg := fmt.Sprintf("Camera scanner started at %s", time.Now().Format("2006-01-02 15:04:05"))
	if err := telegramClient.SendMessage(startupMsg); err != nil {
		log.Error("Failed to send startup message to Telegram: %v", err)
	}

	// Run the main application loop
	runApplicationLoop(ctx, log, cfg, scanner, capture, telegramClient)

	// Send shutdown message to Telegram
	shutdownMsg := fmt.Sprintf("Camera scanner stopped at %s", time.Now().Format("2006-01-02 15:04:05"))
	if err := telegramClient.SendMessage(shutdownMsg); err != nil {
		log.Error("Failed to send shutdown message to Telegram: %v", err)
	}

	log.Info("Application shutdown complete")
}

// runApplicationLoop runs the main application loop with configurable scheduling
func runApplicationLoop(ctx context.Context, log *logger.Logger, cfg *config.Config, scanner *camera.Scanner, capture *camera.Capture, telegramClient *telegram.Client) {
	ticker := time.NewTicker(time.Duration(cfg.Scan.Interval) * time.Minute)
	defer ticker.Stop()

	// Run immediately on startup
	log.Info("Starting initial scan...")
	if err := performScan(ctx, log, cfg, scanner, capture, telegramClient); err != nil {
		log.Error("Initial scan failed: %v", err)
	}

	// Main loop
	for {
		select {
		case <-ctx.Done():
			log.Info("Context cancelled, stopping application loop")
			return
		case <-ticker.C:
			log.Info("Starting scheduled scan...")
			if err := performScan(ctx, log, cfg, scanner, capture, telegramClient); err != nil {
				log.Error("Scheduled scan failed: %v", err)
				// Send error message to Telegram
				errorMsg := fmt.Sprintf("Scheduled scan failed: %v", err)
				if err := telegramClient.SendMessage(errorMsg); err != nil {
					log.Error("Failed to send error message to Telegram: %v", err)
				}
			}
		}
	}
}

// performScan performs a complete scan, capture, and send cycle
func performScan(ctx context.Context, log *logger.Logger, cfg *config.Config, scanner *camera.Scanner, capture *camera.Capture, telegramClient *telegram.Client) error {
	// Create a context with timeout for this scan operation
	scanCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Scan network for IP cameras
	log.Info("Scanning network for IP cameras...")
	cameras, err := scanner.ScanNetwork()
	if err != nil {
		return fmt.Errorf("failed to scan network: %w", err)
	}

	log.Info("Found %d cameras", len(cameras))
	if len(cameras) == 0 {
		// Send message to Telegram about no cameras found
		msg := "No cameras found during network scan"
		if err := telegramClient.SendMessage(msg); err != nil {
			log.Error("Failed to send no cameras found message to Telegram: %v", err)
		}
		return nil
	}

	// Log discovered cameras
	for _, cameraIP := range cameras {
		log.Info("Found camera at %s", cameraIP)
	}

	// Process cameras concurrently with a limit
	semaphore := make(chan struct{}, cfg.Scan.MaxConcurrent)
	var wg sync.WaitGroup
	errors := make(chan error, len(cameras))

	// Process each camera
	for _, cameraIP := range cameras {
		// Check if context was cancelled
		select {
		case <-scanCtx.Done():
			log.Info("Scan context cancelled, stopping camera processing")
			return scanCtx.Err()
		default:
		}

		// Acquire semaphore
		semaphore <- struct{}{}
		wg.Add(1)

		// Process camera in goroutine
		go func(ip string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			if err := processCamera(scanCtx, log, cfg, capture, telegramClient, ip); err != nil {
				log.Error("Failed to process camera at %s: %v", ip, err)
				errors <- fmt.Errorf("camera %s: %w", ip, err)
			}
		}(cameraIP)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)

	// Collect errors
	var errorList []string
	for err := range errors {
		errorList = append(errorList, err.Error())
	}

	// If we have errors, return them as a combined error
	if len(errorList) > 0 {
		// Send error summary to Telegram
		errorMsg := fmt.Sprintf("Scan completed with %d errors:\n%s", len(errorList), strings.Join(errorList, "\n"))
		if err := telegramClient.SendMessage(errorMsg); err != nil {
			log.Error("Failed to send error summary to Telegram: %v", err)
		}
		log.Warn("Scan completed with %d errors", len(errorList))
	} else {
		log.Info("Scan completed successfully with no errors")
	}

	return nil
}

// processCamera handles the complete flow for a single camera
func processCamera(ctx context.Context, log *logger.Logger, cfg *config.Config, capture *camera.Capture, telegramClient *telegram.Client, cameraIP string) error {
	log.Info("Processing camera at %s", cameraIP)

	// Check if context was cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Set capture options
	options := &camera.CaptureOptions{
		Timeout:   time.Duration(cfg.RTSP.Timeout) * time.Second,
		Quality:   75,
		MaxWidth:  1920,
		MaxHeight: 1080,
	}

	// Capture frame
	log.Info("Capturing frame from camera at %s", cameraIP)
	jpegData, err := capture.CaptureFrame(cameraIP, cfg.RTSP.Username, cfg.RTSP.Password, options)
	if err != nil {
		return fmt.Errorf("failed to capture frame: %w", err)
	}

	log.Info("Successfully captured frame from camera at %s (%d bytes)", cameraIP, len(jpegData))

	// Check if context was cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Add watermark to the captured image
	processOptions := &camera.ProcessOptions{
		AddTimestamp: true,
		AddCameraID:  true,
		CameraID:     cameraIP,
	}

	processedImage, err := camera.AddWatermark(jpegData, processOptions)
	if err != nil {
		return fmt.Errorf("failed to add watermark: %w", err)
	}

	log.Info("Successfully processed image from camera at %s (%d bytes)", cameraIP, len(processedImage.Data))

	// Check if context was cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

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
		return fmt.Errorf("failed to send image to Telegram: %w", err)
	}

	log.Info("Successfully sent image to Telegram from camera at %s", cameraIP)
	return nil
}