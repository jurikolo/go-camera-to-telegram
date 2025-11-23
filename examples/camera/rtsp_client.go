// Package main demonstrates how to use the RTSP client functionality
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jurikolo/go-camera-to-telegram/internal/camera"
)

func main() {
	fmt.Println("RTSP Client Example")

	// Check command line arguments
	if len(os.Args) < 4 {
		fmt.Println("Usage: rtsp_client <host> <username> <password>")
		fmt.Println("Example: rtsp_client 192.168.1.100 admin password")
		os.Exit(1)
	}

	host := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]

	// Create an RTSP client
	client := camera.NewRTSPClient(host, username, password)

	// Connect to the camera
	fmt.Printf("Connecting to RTSP camera at %s...\n", host)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect to RTSP camera: %v", err)
	}
	defer client.Disconnect()

	fmt.Println("Successfully connected to RTSP camera")

	// Get stream information
	streams, err := client.GetStreamInfo()
	if err != nil {
		log.Printf("Failed to get stream info: %v", err)
	} else {
		fmt.Printf("Stream info: %d streams\n", len(streams))
		for i, stream := range streams {
			fmt.Printf("  Stream %d: %s\n", i, stream.Type().String())
		}
	}

	// Try to capture a frame (this will likely timeout on most cameras)
	fmt.Println("Attempting to capture a frame...")
	frame, err := client.CaptureFrame()
	if err != nil {
		log.Printf("Failed to capture frame: %v", err)
	} else {
		fmt.Printf("Captured frame: %d bytes\n", len(frame))
	}

	// Demonstrate connection pooling
	fmt.Println("\nDemonstrating connection pooling...")
	pool := camera.NewConnectionPool(5, 10*time.Second)
	defer pool.Close()

	// Get a connection from the pool
	_, err = pool.GetConnection(host, username, password)
	if err != nil {
		log.Printf("Failed to get connection from pool: %v", err)
	} else {
		fmt.Printf("Got connection from pool. Active connections: %d\n", pool.GetActiveConnections())
		
		// Release the connection back to the pool
		pool.ReleaseConnection(host, username, password)
		fmt.Printf("Released connection back to pool. Active connections: %d\n", pool.GetActiveConnections())
	}

	fmt.Println("RTSP client example completed")
}