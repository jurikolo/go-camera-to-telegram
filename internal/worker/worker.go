// Package worker provides a worker pool implementation for processing cameras concurrently
package worker

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jurikolo/go-camera-to-telegram/internal/camera"
	"github.com/jurikolo/go-camera-to-telegram/internal/circuitbreaker"
	"github.com/jurikolo/go-camera-to-telegram/internal/config"
	"github.com/jurikolo/go-camera-to-telegram/internal/logger"
	"github.com/jurikolo/go-camera-to-telegram/internal/metrics"
	"github.com/jurikolo/go-camera-to-telegram/internal/retry"
	"github.com/jurikolo/go-camera-to-telegram/internal/telegram"
)

// Job represents a camera processing job
type Job struct {
	CameraIP string
}

// Result represents the result of processing a camera
type Result struct {
	CameraIP string
	Success  bool
	Error    error
}

// WorkerPool manages a pool of workers for processing camera jobs
type WorkerPool struct {
	workers    int
	jobs       chan Job
	results    chan Result
	wg         sync.WaitGroup
	log        *logger.Logger
	metrics    *metrics.Metrics
	cfg        *config.Config
	capture    *camera.Capture
	telegram   *telegram.Client
	
	// Circuit breakers for each camera
	circuitBreakers map[string]*circuitbreaker.CircuitBreaker
	cbMutex       sync.RWMutex
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int, log *logger.Logger, metrics *metrics.Metrics, cfg *config.Config, capture *camera.Capture, telegramClient *telegram.Client) *WorkerPool {
	return &WorkerPool{
		workers:        workers,
		jobs:          make(chan Job, workers*2), // Buffer to prevent blocking
		results:       make(chan Result, workers*2),
		log:           log,
		metrics:       metrics,
		cfg:           cfg,
		capture:        capture,
		telegram:      telegramClient,
		circuitBreakers: make(map[string]*circuitbreaker.CircuitBreaker),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.log.Info("Starting worker pool with %d workers", wp.workers)
	
	// Start workers
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
	
	// Start result collector
	wp.wg.Add(1)
	go wp.collectResults(ctx)
}

// Stop stops the worker pool and waits for all workers to finish
func (wp *WorkerPool) Stop() {
	wp.log.Info("Stopping worker pool")
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
	wp.log.Info("Worker pool stopped")
}

// SubmitJob submits a job to the worker pool
func (wp *WorkerPool) SubmitJob(job Job) {
	wp.jobs <- job
}

// worker is a goroutine that processes camera jobs
func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()
	
	wp.log.Debug("Worker %d started", workerID)
	
	for {
		select {
		case <-ctx.Done():
			wp.log.Debug("Worker %d stopping due to context cancellation", workerID)
			return
		case job, ok := <-wp.jobs:
			if !ok {
				wp.log.Debug("Worker %d stopping due to closed jobs channel", workerID)
				return
			}
			
			wp.log.Debug("Worker %d processing camera %s", workerID, job.CameraIP)
			
			// Process the camera
			err := wp.processCamera(ctx, job.CameraIP)
			
			// Send result
			result := Result{
				CameraIP: job.CameraIP,
				Success:  err == nil,
				Error:    err,
			}
			
			select {
			case wp.results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

// collectResults collects results from workers and updates metrics
func (wp *WorkerPool) collectResults(ctx context.Context) {
	defer wp.wg.Done()
	
	wp.log.Debug("Result collector started")
	
	for {
		select {
		case <-ctx.Done():
			wp.log.Debug("Result collector stopping due to context cancellation")
			return
		case result, ok := <-wp.results:
			if !ok {
				wp.log.Debug("Result collector stopping due to closed results channel")
				return
			}
			
			if result.Success {
				wp.metrics.IncrementSuccessfulCaptures()
				wp.log.Info("Successfully processed camera %s", result.CameraIP)
			} else {
				wp.metrics.IncrementFailedCaptures()
				wp.log.Error("Failed to process camera %s: %v", result.CameraIP, result.Error)
			}
		}
	}
}

// processCamera handles the complete flow for a single camera
func (wp *WorkerPool) processCamera(ctx context.Context, cameraIP string) error {
	// Get or create circuit breaker for this camera
	cb := wp.getCircuitBreaker(cameraIP)
	
	// Check if circuit breaker is open
	if cb.GetState() == circuitbreaker.Open {
		wp.log.Warn("Circuit breaker is open for camera %s, skipping processing", cameraIP)
		return fmt.Errorf("circuit breaker is open for camera %s", cameraIP)
	}
	
	wp.log.Info("Processing camera at %s", cameraIP)

	// Check if context was cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Set capture options
	options := &camera.CaptureOptions{
		Timeout:   time.Duration(wp.cfg.RTSP.Timeout) * time.Second,
		Quality:   75,
		MaxWidth:  1920,
		MaxHeight: 1080,
	}

	// Capture frame with retry logic and circuit breaker
	var jpegData []byte
	captureErr := cb.Execute(func() error {
		return retry.WithRetry(retry.DefaultConfig(), func() error {
			var err error
			jpegData, err = wp.capture.CaptureFrame(cameraIP, wp.cfg.RTSP.Username, wp.cfg.RTSP.Password, options)
			return err
		})
	})
	
	if captureErr != nil {
		return fmt.Errorf("failed to capture frame after retries and circuit breaker: %w", captureErr)
	}

	wp.log.Info("Successfully captured frame from camera at %s (%d bytes)", cameraIP, len(jpegData))

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

	wp.log.Info("Successfully processed image from camera at %s (%d bytes)", cameraIP, len(processedImage.Data))

	// Check if context was cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Send image to Telegram with retry logic
	telegramErr := retry.WithRetry(retry.Config{
		MaxRetries:    3,
		InitialDelay:  2 * time.Second,
		MaxDelay:      30 * time.Second,
		Multiplier:    2.0,
		Jitter:        1 * time.Second,
	}, func() error {
		return wp.telegram.SendPhoto(strings.NewReader(string(processedImage.Data)), wp.telegram.FormatCameraMessage(telegram.CameraInfo{
			IP:          cameraIP,
			CaptureTime: time.Now(),
			Metadata: map[string]string{
				"Image Size": fmt.Sprintf("%d bytes", len(processedImage.Data)),
			},
		}))
	})
	
	if telegramErr != nil {
		return fmt.Errorf("failed to send image to Telegram after retries: %w", telegramErr)
	}

	wp.log.Info("Successfully sent image to Telegram from camera at %s", cameraIP)
	return nil
}

// getCircuitBreaker gets or creates a circuit breaker for a camera
func (wp *WorkerPool) getCircuitBreaker(cameraIP string) *circuitbreaker.CircuitBreaker {
	wp.cbMutex.Lock()
	defer wp.cbMutex.Unlock()
	
	if cb, exists := wp.circuitBreakers[cameraIP]; exists {
		return cb
	}
	
	// Create new circuit breaker
	cb := circuitbreaker.New(circuitbreaker.Config{
		MaxFailures:         3,
		Timeout:             5 * time.Minute,
		HalfOpenDelay:       1 * time.Minute,
		HalfOpenMaxRequests: 2,
	})
	
	wp.circuitBreakers[cameraIP] = cb
	return cb
}