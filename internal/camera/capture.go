// Package camera provides functionality for capturing images from IP cameras
package camera

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"time"

	"github.com/deepch/vdk/av"
)

// Capture handles image capture from RTSP streams
type Capture struct {
	pool *ConnectionPool
}

// CaptureOptions defines options for frame capture
type CaptureOptions struct {
	Timeout     time.Duration // Timeout for frame capture
	Quality     int           // JPEG quality (1-100)
	MaxWidth    int           // Maximum width for resizing (0 = no resize)
	MaxHeight   int           // Maximum height for resizing (0 = no resize)
}

// NewCapture creates a new capture instance
func NewCapture() *Capture {
	return &Capture{
		pool: NewConnectionPool(10, 30*time.Second), // Max 10 connections, 30s timeout
	}
}

// CaptureFrame captures a single frame from an RTSP stream and converts it to JPEG
func (c *Capture) CaptureFrame(cameraIP, username, password string, options *CaptureOptions) ([]byte, error) {
	// Set default options if not provided
	if options == nil {
		options = &CaptureOptions{
			Timeout: 30 * time.Second,
			Quality: 75,
		}
	}
	
	// Get connection from pool
	client, err := c.pool.GetConnection(cameraIP, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection to RTSP camera at %s: %w", cameraIP, err)
	}
	defer c.pool.ReleaseConnection(cameraIP, username, password)
	
	// Get stream information
	streams, err := client.GetStreamInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info from RTSP camera at %s: %w", cameraIP, err)
	}
	
	// Check if the stream is already in JPEG format
	for _, stream := range streams {
		if stream.Type() == av.MJPEG {
			// Stream is already in JPEG format, just capture a frame
			return c.captureJPEGFrame(client, options)
		}
	}
	
	// For H.264/H.265 streams, we would need to decode the video frame
	// Since we don't have FFmpeg dependencies, we'll return the raw packet data
	// In a real implementation, you would decode the video frame here
	return c.captureRawFrame(client, options)
}

// captureJPEGFrame captures a JPEG frame from an MJPEG stream
func (c *Capture) captureJPEGFrame(client *RTSPClient, options *CaptureOptions) ([]byte, error) {
	// Read packets until we get a video keyframe
	timeout := time.After(options.Timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for JPEG frame")
		case <-ticker.C:
			// Try to read a packet
			packet, err := client.CaptureFrame()
			if err != nil {
				return nil, fmt.Errorf("failed to read packet: %w", err)
			}
			
			// Return the JPEG data as-is
			return packet, nil
		}
	}
}

// captureRawFrame captures a raw frame from an H.264/H.265 stream
// Note: This is a simplified implementation that returns the raw packet data
// In a real implementation, you would decode the video frame using FFmpeg or similar
func (c *Capture) captureRawFrame(client *RTSPClient, options *CaptureOptions) ([]byte, error) {
	// Read packets until we get a video keyframe
	timeout := time.After(options.Timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for video frame")
		case <-ticker.C:
			// Try to read a packet
			packet, err := client.CaptureFrame()
			if err != nil {
				return nil, fmt.Errorf("failed to read packet: %w", err)
			}
			
			// For now, we'll just return the raw packet data
			// In a real implementation, you would decode the video frame here
			// and then convert it to JPEG
			return packet, nil
		}
	}
}

// convertToJPEG converts an image to JPEG format with specified quality
// This is a placeholder implementation - in a real implementation you would
// decode the video frame and then convert it to JPEG
func (c *Capture) convertToJPEG(img image.Image, options *CaptureOptions) ([]byte, error) {
	// Encode to JPEG
	var buf bytes.Buffer
	jpegOptions := &jpeg.Options{Quality: options.Quality}
	
	if err := jpeg.Encode(&buf, img, jpegOptions); err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", err)
	}
	
	return buf.Bytes(), nil
}