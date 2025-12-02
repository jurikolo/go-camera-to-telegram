package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jurikolo/go-camera-to-telegram/internal/camera"
	"github.com/jurikolo/go-camera-to-telegram/internal/config"
	"github.com/jurikolo/go-camera-to-telegram/internal/logger"
	"github.com/jurikolo/go-camera-to-telegram/internal/metrics"
	"github.com/jurikolo/go-camera-to-telegram/internal/worker"
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
	log.Info("  Telegram Chat ID: %d", cfg.Telegram.ChatID)
	log.Info("  Scan Interval: %d minutes", cfg.Scan.Interval)

	// Create Telegram client
	telegramClient, err := telegram.NewClient(cfg.TelegramToken.Value(), cfg.Telegram.ChatID)
	if err != nil {
		log.Fatal("Failed to create Telegram client: %v", err)
	}
	defer func() {
		if err := telegramClient.Close(); err != nil {
			log.Error("Failed to close Telegram client: %v", err)
		}
	}()

	// Create camera scanner
	scanner := camera.NewScanner(cfg)

	// Create camera capture instance
	capture := camera.NewCapture()
	defer capture.Close() // Ensure capture connections are closed on exit

	// Create metrics instance
	metricsInstance := metrics.NewMetrics()

	// Create worker pool
	workerPool := worker.NewWorkerPool(cfg.Scan.MaxConcurrent, log, metricsInstance, cfg, capture, telegramClient)
	
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

	// Start worker pool
	workerPool.Start(ctx)
	defer workerPool.Stop()

	// Send startup message to Telegram
	startupMsg := fmt.Sprintf("Camera scanner started at %s", time.Now().Format("2006-01-02 15:04:05"))
	if err := telegramClient.SendMessage(ctx, startupMsg); err != nil {
		log.Error("Failed to send startup message to Telegram: %v", err)
	}

	// Run the main application loop
	runApplicationLoop(ctx, log, cfg, scanner, capture, telegramClient, workerPool, metricsInstance)

	// Send shutdown message to Telegram
	shutdownMsg := fmt.Sprintf("Camera scanner stopped at %s", time.Now().Format("2006-01-02 15:04:05"))
	if err := telegramClient.SendMessage(ctx, shutdownMsg); err != nil {
		log.Error("Failed to send shutdown message to Telegram: %v", err)
	}

	log.Info("Application shutdown complete")
}

// runApplicationLoop runs the main application loop with configurable scheduling
func runApplicationLoop(ctx context.Context, log *logger.Logger, cfg *config.Config, scanner *camera.Scanner, capture *camera.Capture, telegramClient *telegram.Client, workerPool *worker.WorkerPool, metricsInstance *metrics.Metrics) {
	ticker := time.NewTicker(time.Duration(cfg.Scan.Interval) * time.Minute)
	defer ticker.Stop()

	// Run immediately on startup
	log.Info("Starting initial scan...")
	if err := performScan(ctx, log, cfg, scanner, capture, telegramClient, workerPool, metricsInstance); err != nil {
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
			if err := performScan(ctx, log, cfg, scanner, capture, telegramClient, workerPool, metricsInstance); err != nil {
				log.Error("Scheduled scan failed: %v", err)
				// Send error message to Telegram
				errorMsg := fmt.Sprintf("Scheduled scan failed: %v", err)
				if err := telegramClient.SendMessage(ctx, errorMsg); err != nil {
					log.Error("Failed to send error message to Telegram: %v", err)
				}
			}
		}
	}
}

// performScan performs a complete scan, capture, and send cycle
func performScan(ctx context.Context, log *logger.Logger, cfg *config.Config, scanner *camera.Scanner, capture *camera.Capture, telegramClient *telegram.Client, workerPool *worker.WorkerPool, metricsInstance *metrics.Metrics) error {
	// Create a context with timeout for this scan operation
	scanCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Scan network for IP cameras
	log.Info("Scanning network for IP cameras...")
	cameras, err := scanner.ScanNetwork()
	if err != nil {
		return fmt.Errorf("failed to scan network: %w", err)
	}

	// Update metrics
	metricsInstance.IncrementCamerasFound()
	log.Info("Found %d cameras", len(cameras))
	if len(cameras) == 0 {
		// Send message to Telegram about no cameras found
		msg := "No cameras found during network scan"
		if err := telegramClient.SendMessage(ctx, msg); err != nil {
			log.Error("Failed to send no cameras found message to Telegram: %v", err)
		}
		return nil
	}

	// Log discovered cameras
	for _, cameraIP := range cameras {
		log.Info("Found camera at %s", cameraIP)
	}

	// Process cameras using worker pool
	var wg sync.WaitGroup

	// Submit jobs to worker pool
	for _, cameraIP := range cameras {
		// Check if context was cancelled
		select {
		case <-scanCtx.Done():
			log.Info("Scan context cancelled, stopping camera processing")
			return scanCtx.Err()
		default:
		}

		wg.Add(1)
		// Submit job to worker pool
		go func(ip string) {
			defer wg.Done()
			
			// Check if context was cancelled before submitting job
			select {
			case <-scanCtx.Done():
				log.Info("Scan context cancelled, not submitting job for camera %s", ip)
				return
			default:
			}
			
			workerPool.SubmitJob(worker.Job{CameraIP: ip})
		}(cameraIP)
	}

	// Wait for all jobs to be submitted
	wg.Wait()

	// Collect results (for now, we'll just log them)
	// In a more advanced implementation, we might want to collect specific results
	// For now, we'll rely on the worker pool's internal result collection

	// Check for errors (simplified approach for now)
	// In a real implementation, we'd want more sophisticated error collection
	log.Info("Scan completed - processing with worker pool")

	return nil
}
