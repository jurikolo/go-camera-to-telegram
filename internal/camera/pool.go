// Package camera provides functionality for connecting to RTSP cameras
package camera

import (
	"fmt"
	"sync"
	"time"
)

// ConnectionPool manages a pool of RTSP connections
type ConnectionPool struct {
	connections map[string]*RTSPClient
	mutex       sync.RWMutex
	maxSize     int
	timeout     time.Duration
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxSize int, timeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		connections: make(map[string]*RTSPClient),
		maxSize:     maxSize,
		timeout:     timeout,
	}
}

// GetConnection gets or creates a connection to an RTSP camera
func (p *ConnectionPool) GetConnection(host, username, password string) (*RTSPClient, error) {
	key := fmt.Sprintf("%s:%s:%s", host, username, password)
	
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Check if we already have a connection
	if client, exists := p.connections[key]; exists {
		return client, nil
	}
	
	// Check if we've reached the maximum pool size
	if len(p.connections) >= p.maxSize {
		return nil, fmt.Errorf("connection pool is full (max %d connections)", p.maxSize)
	}
	
	// Create a new connection
	client := NewRTSPClient(host, username, password)
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RTSP camera at %s: %w", host, err)
	}
	
	// Add to pool
	p.connections[key] = client
	
	return client, nil
}

// ReleaseConnection releases a connection back to the pool
func (p *ConnectionPool) ReleaseConnection(host, username, password string) {
	key := fmt.Sprintf("%s:%s:%s", host, username, password)
	
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Remove from pool and close connection
	if client, exists := p.connections[key]; exists {
		client.Disconnect()
		delete(p.connections, key)
	}
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	for key, client := range p.connections {
		client.Disconnect()
		delete(p.connections, key)
	}
}

// GetActiveConnections returns the number of active connections in the pool
func (p *ConnectionPool) GetActiveConnections() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	return len(p.connections)
}