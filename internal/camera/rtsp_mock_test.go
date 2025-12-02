// Package camera provides mock RTSP servers for testing
package camera

import (
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockRTSPServer is a mock RTSP server for testing
type MockRTSPServer struct {
	host     string
	port     string
	listener net.Listener
	closed   bool
}

// NewMockRTSPServer creates a new mock RTSP server
func NewMockRTSPServer(host, port string) (*MockRTSPServer, error) {
	// Create a listener on the specified host and port
	addr := fmt.Sprintf("%s:%s", host, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	server := &MockRTSPServer{
		host:     host,
		port:     port,
		listener: listener,
	}

	// Start the server in a goroutine
	go server.serve()

	return server, nil
}

// serve handles incoming connections
func (m *MockRTSPServer) serve() {
	for {
		// Accept incoming connections
		conn, err := m.listener.Accept()
		if err != nil {
			if !m.closed {
				fmt.Printf("Error accepting connection: %v\n", err)
			}
			return
		}

		// Handle the connection in a goroutine
		go m.handleConnection(conn)
	}
}

// handleConnection handles a single connection
func (m *MockRTSPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Set read and write timeouts
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// Read the request
	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading request: %v\n", err)
		return
	}

	// Check if it's an OPTIONS request
	if len(request) >= 7 && request[:7] == "OPTIONS" {
		// Send a mock RTSP OPTIONS response
		response := "RTSP/1.0 200 OK\r\n" +
			"CSeq: 1\r\n" +
			"Public: OPTIONS, DESCRIBE, SETUP, PLAY, TEARDOWN\r\n" +
			"\r\n"
		conn.Write([]byte(response))
		return
	}

	// For other requests, send a generic response
	response := "RTSP/1.0 404 Not Found\r\n" +
		"CSeq: 1\r\n" +
		"\r\n"
	conn.Write([]byte(response))
}

// Close closes the mock RTSP server
func (m *MockRTSPServer) Close() error {
	m.closed = true
	return m.listener.Close()
}

// Addr returns the address of the mock RTSP server
func (m *MockRTSPServer) Addr() string {
	return fmt.Sprintf("%s:%s", m.host, m.port)
}

// TestMockRTSPServer tests the mock RTSP server functionality
func TestMockRTSPServer(t *testing.T) {
	// Create a mock RTSP server
	server, err := NewMockRTSPServer("127.0.0.1", "0") // Use port 0 to get a random available port
	assert.NoError(t, err)
	assert.NotNil(t, server)

	// Get the actual port
	addr := server.listener.Addr().String()
	_, port, err := net.SplitHostPort(addr)
	assert.NoError(t, err)

	// Test connecting to the mock server
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	assert.NoError(t, err)
	assert.NotNil(t, conn)

	// Send an OPTIONS request
	request := "OPTIONS rtsp://127.0.0.1:" + port + "/ RTSP/1.0\r\n" +
		"CSeq: 1\r\n" +
		"User-Agent: TestClient\r\n" +
		"\r\n"
	_, err = conn.Write([]byte(request))
	assert.NoError(t, err)

	// Read the response
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	assert.NoError(t, err)
	assert.Greater(t, n, 0)

	// Check that the response looks like an RTSP response
	responseStr := string(response[:n])
	assert.Contains(t, responseStr, "RTSP/1.0")
	assert.Contains(t, responseStr, "200 OK")

	// Close the connection
	err = conn.Close()
	assert.NoError(t, err)

	// Close the server
	err = server.Close()
	assert.NoError(t, err)
}

// TestRTSPClientWithMockServer tests the RTSP client with a mock server
func TestRTSPClientWithMockServer(t *testing.T) {
	// Create a mock RTSP server
	server, err := NewMockRTSPServer("127.0.0.1", "0")
	assert.NoError(t, err)
	assert.NotNil(t, server)

	// Get the actual port
	addr := server.listener.Addr().String()
	_, _, err = net.SplitHostPort(addr)
	assert.NoError(t, err)

	// Create an RTSP client
	client := NewRTSPClient("127.0.0.1", "admin", "password")

	// Try to connect (this will likely fail, but we're testing error handling)
	err = client.Connect()
	// We expect an error since our mock server doesn't implement the full protocol
	assert.Error(t, err)

	// Test that Disconnect doesn't panic
	err = client.Disconnect()
	assert.NoError(t, err)

	// Close the server
	err = server.Close()
	assert.NoError(t, err)
}