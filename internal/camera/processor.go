// Package camera provides functionality for processing captured images
package camera

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// ProcessedImage represents a processed image with metadata
type ProcessedImage struct {
	Data        []byte
	Width       int
	Height      int
	CameraIP    string
	CaptureTime time.Time
}

// ProcessOptions defines options for image processing
type ProcessOptions struct {
	AddTimestamp bool
	AddCameraID bool
	CameraID    string
	FontSize    int
}

// AddWatermark adds timestamp and camera labels to a JPEG image
func AddWatermark(jpegData []byte, options *ProcessOptions) (*ProcessedImage, error) {
	// Decode the JPEG image
	img, err := jpeg.Decode(bytes.NewReader(jpegData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode JPEG image: %w", err)
	}

	// Get image bounds
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new RGBA image for drawing
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Add watermark if requested
	if options.AddTimestamp || options.AddCameraID {
		// Create font drawer
		d := &font.Drawer{
			Dst:  rgba,
			Src:  image.NewUniform(image.White),
			Face: basicfont.Face7x13,
		}

		// Add timestamp if requested
		if options.AddTimestamp {
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			text := fmt.Sprintf("Captured: %s", timestamp)
			d.Dot = fixed.Point26_6{
				X: fixed.I(10),
				Y: fixed.I(height - 20),
			}
			d.DrawString(text)
		}

		// Add camera ID if requested
		if options.AddCameraID {
			text := fmt.Sprintf("Camera: %s", options.CameraID)
			d.Dot = fixed.Point26_6{
				X: fixed.I(10),
				Y: fixed.I(height - 5),
			}
			d.DrawString(text)
		}
	}

	// Encode the processed image back to JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: 75})
	if err != nil {
		return nil, fmt.Errorf("failed to encode processed image: %w", err)
	}

	return &ProcessedImage{
		Data:        buf.Bytes(),
		Width:       width,
		Height:      height,
		CameraIP:    options.CameraID,
		CaptureTime: time.Now(),
	}, nil
}

// AddTextOverlay adds text overlay to an image
func AddTextOverlay(jpegData []byte, text string, x, y int) ([]byte, error) {
	// Decode the JPEG image
	img, err := jpeg.Decode(bytes.NewReader(jpegData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode JPEG image: %w", err)
	}

	// Get image bounds
	bounds := img.Bounds()

	// Create a new RGBA image for drawing
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Create font drawer
	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(image.White),
		Face: basicfont.Face7x13,
		Dot: fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		},
	}

	// Draw the text
	d.DrawString(text)

	// Encode the processed image back to JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: 75})
	if err != nil {
		return nil, fmt.Errorf("failed to encode processed image: %w", err)
	}

	return buf.Bytes(), nil
}

// CreateTimestampOverlay creates a timestamp overlay for an image
func CreateTimestampOverlay() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// CreateCameraLabelOverlay creates a camera label overlay
func CreateCameraLabelOverlay(cameraIP string) string {
	return fmt.Sprintf("Camera: %s", cameraIP)
}