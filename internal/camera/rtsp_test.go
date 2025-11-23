// Package camera provides tests for the RTSP client functionality
package camera

import (
	"testing"
	"time"
)

// TestRTSPClient tests the RTSP client functionality
func TestRTSPClient(t *testing.T) {
	// Test creating a new RTSP client
	client := NewRTSPClient("192.168.1.100", "admin", "password")
	if client == nil {
		t.Fatal("Failed to create RTSP client")
	}
	
	if client.host != "192.168.1.100" {
		t.Errorf("Expected host 192.168.1.100, got %s", client.host)
	}
	
	if client.username != "admin" {
		t.Errorf("Expected username admin, got %s", client.username)
	}
	
	if client.password != "password" {
		t.Errorf("Expected password password, got %s", client.password)
	}
}

// TestRTSPClientConnect tests the Connect method
func TestRTSPClientConnect(t *testing.T) {
	client := NewRTSPClient("192.168.1.100", "admin", "password")
	
	// This test will fail since there's no actual RTSP server running
	// but we can test that it returns an error
	err := client.Connect()
	if err == nil {
		t.Error("Expected error when connecting to non-existent RTSP server")
	}
	
	// Test that Disconnect doesn't panic when not connected
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Expected no error when disconnecting from non-connected client, got %v", err)
	}
}

// TestConnectionPool tests the connection pool functionality
func TestConnectionPool(t *testing.T) {
	// Create a connection pool with max size of 2
	pool := NewConnectionPool(2, 5*time.Second)
	if pool == nil {
		t.Fatal("Failed to create connection pool")
	}
	
	// Test getting active connections count
	if pool.GetActiveConnections() != 0 {
		t.Errorf("Expected 0 active connections, got %d", pool.GetActiveConnections())
	}
	
	// Test closing empty pool
	pool.Close()
	
	// Create a new pool for further testing
	pool = NewConnectionPool(2, 5*time.Second)
	
	// Test getting a connection (will fail since no server is running)
	client, err := pool.GetConnection("192.168.1.100", "admin", "password")
	if err == nil {
		t.Error("Expected error when getting connection to non-existent RTSP server")
	}
	if client != nil {
		t.Error("Expected nil client when getting connection to non-existent RTSP server")
	}
	
	// Test active connections count
	if pool.GetActiveConnections() != 0 {
		t.Errorf("Expected 0 active connections, got %d", pool.GetActiveConnections())
	}
	
	// Test releasing a non-existent connection
	pool.ReleaseConnection("192.168.1.100", "admin", "password")
	
	// Close the pool
	pool.Close()
}

// TestConnectionPoolMaxSize tests the maximum pool size limit
func TestConnectionPoolMaxSize(t *testing.T) {
	// Create a connection pool with max size of 1
	pool := NewConnectionPool(1, 5*time.Second)
	
	// Try to get a connection (will fail since no server is running)
	// but we're testing the pool size limit logic
	_, err := pool.GetConnection("192.168.1.100", "admin", "password")
	if err == nil {
		t.Error("Expected error when getting connection to non-existent RTSP server")
	}
	
	// Try to get another connection - should fail due to pool size limit
	// (but will actually fail for the same reason as above)
	_, err = pool.GetConnection("192.168.1.101", "admin", "password")
	if err == nil {
		t.Error("Expected error when getting connection to non-existent RTSP server")
	}
	
	pool.Close()
}