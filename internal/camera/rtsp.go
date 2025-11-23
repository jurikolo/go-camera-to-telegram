// Package camera provides functionality for connecting to RTSP cameras
package camera

import (
	"fmt"
	"sync"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/format/rtsp"
)

// RTSPClient handles connections to RTSP cameras
type RTSPClient struct {
	host     string
	username string
	password string
	stream   int
	client   *rtsp.Client
	mutex    sync.Mutex
}

// NewRTSPClient creates a new RTSP client
func NewRTSPClient(host, username, password string) *RTSPClient {
	return &RTSPClient{
		host:     host,
		username: username,
		password: password,
		stream:   0,
	}
}

// Connect establishes a connection to the RTSP camera
func (r *RTSPClient) Connect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Try both stream=0 and stream=1
	for stream := 0; stream <= 1; stream++ {
		url := fmt.Sprintf("rtsp://%s:554/user=%s_password=%s_channel=0_stream=%d.sdp",
			r.host, r.username, r.password, stream)
		
		fmt.Printf("Connecting to RTSP camera at %s (stream=%d)\n", r.host, stream)
		
		// Create RTSP client
		client, err := rtsp.Dial(url)
		if err != nil {
			fmt.Printf("Failed to connect to RTSP camera at %s (stream=%d): %v\n", r.host, stream, err)
			continue
		}
		
		// Explicitly send DESCRIBE command
		_, err = client.Describe()
		if err != nil {
			fmt.Printf("Failed to send DESCRIBE command to RTSP camera at %s (stream=%d): %v\n", r.host, stream, err)
			client.Close()
			continue
		}
		
		// Explicitly send SETUP command for all streams
		err = client.SetupAll()
		if err != nil {
			fmt.Printf("Failed to send SETUP command to RTSP camera at %s (stream=%d): %v\n", r.host, stream, err)
			client.Close()
			continue
		}
		
		// Explicitly send PLAY command
		err = client.Play()
		if err != nil {
			fmt.Printf("Failed to send PLAY command to RTSP camera at %s (stream=%d): %v\n", r.host, stream, err)
			client.Close()
			continue
		}
		
		// Test the connection by reading stream info
		streams, err := client.Streams()
		if err != nil {
			fmt.Printf("Failed to get stream info from RTSP camera at %s (stream=%d): %v\n", r.host, stream, err)
			client.Teardown()
			client.Close()
			continue
		}
		
		fmt.Printf("Successfully connected to RTSP camera at %s (stream=%d) with %d streams\n",
			r.host, stream, len(streams))
		
		// Store the successful client and stream number
		r.client = client
		r.stream = stream
		return nil
	}
	
	return fmt.Errorf("failed to connect to RTSP camera at %s with either stream=0 or stream=1", r.host)
}

// Disconnect closes the connection to the RTSP camera
func (r *RTSPClient) Disconnect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if r.client != nil {
		fmt.Printf("Disconnecting from RTSP camera at %s (stream=%d)\n", r.host, r.stream)
		// Send TEARDOWN command before closing
		err := r.client.Teardown()
		if err != nil {
			fmt.Printf("Failed to send TEARDOWN command to RTSP camera at %s: %v\n", r.host, err)
		}
		err = r.client.Close()
		r.client = nil
		return err
	}
	
	return nil
}

// CaptureFrame captures a single frame from the RTSP stream
func (r *RTSPClient) CaptureFrame() ([]byte, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if r.client == nil {
		return nil, fmt.Errorf("not connected to RTSP camera at %s", r.host)
	}
	
	// Read packets until we get a video keyframe
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for video frame from RTSP camera at %s", r.host)
		case <-ticker.C:
			// Try to read a packet
			packet, err := r.client.ReadPacket()
			if err != nil {
				return nil, fmt.Errorf("failed to read packet from RTSP camera at %s: %v", r.host, err)
			}
			
			// Check if this is a video keyframe
			if packet.IsKeyFrame {
				// Return the packet data as our "frame"
				return packet.Data, nil
			}
		}
	}
}

// GetStreamInfo returns information about the RTSP stream
func (r *RTSPClient) GetStreamInfo() ([]av.CodecData, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if r.client == nil {
		return nil, fmt.Errorf("not connected to RTSP camera at %s", r.host)
	}
	
	return r.client.Streams()
}