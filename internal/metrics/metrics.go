// Package metrics provides functionality for collecting and reporting application metrics
package metrics

import (
	"sync"
	"time"
)

// Metrics holds application metrics
type Metrics struct {
	// Camera metrics
	CamerasFound    int64
	SuccessfulCaptures int64
	FailedCaptures  int64
	
	// Timing metrics
	LastScanDuration time.Duration
	LastScanTime     time.Time
	
	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncrementCamerasFound increments the cameras found counter
func (m *Metrics) IncrementCamerasFound() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.CamerasFound++
}

// IncrementSuccessfulCaptures increments the successful captures counter
func (m *Metrics) IncrementSuccessfulCaptures() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.SuccessfulCaptures++
}

// IncrementFailedCaptures increments the failed captures counter
func (m *Metrics) IncrementFailedCaptures() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.FailedCaptures++
}

// SetScanDuration sets the last scan duration
func (m *Metrics) SetScanDuration(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.LastScanDuration = duration
	m.LastScanTime = time.Now()
}

// GetCamerasFound returns the number of cameras found
func (m *Metrics) GetCamerasFound() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.CamerasFound
}

// GetSuccessfulCaptures returns the number of successful captures
func (m *Metrics) GetSuccessfulCaptures() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.SuccessfulCaptures
}

// GetFailedCaptures returns the number of failed captures
func (m *Metrics) GetFailedCaptures() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.FailedCaptures
}

// GetScanDuration returns the last scan duration
func (m *Metrics) GetScanDuration() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.LastScanDuration
}

// GetLastScanTime returns the last scan time
func (m *Metrics) GetLastScanTime() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.LastScanTime
}

// GetTotalCaptures returns the total number of captures (successful + failed)
func (m *Metrics) GetTotalCaptures() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.SuccessfulCaptures + m.FailedCaptures
}

// GetSuccessRate returns the success rate as a percentage
func (m *Metrics) GetSuccessRate() float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	total := m.SuccessfulCaptures + m.FailedCaptures
	if total == 0 {
		return 0
	}
	
	return float64(m.SuccessfulCaptures) / float64(total) * 100
}

// Reset resets all metrics to zero
func (m *Metrics) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.CamerasFound = 0
	m.SuccessfulCaptures = 0
	m.FailedCaptures = 0
	m.LastScanDuration = 0
	m.LastScanTime = time.Time{}
}

// Snapshot creates a snapshot of the current metrics
func (m *Metrics) Snapshot() *MetricsSnapshot {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return &MetricsSnapshot{
		CamerasFound:       m.CamerasFound,
		SuccessfulCaptures: m.SuccessfulCaptures,
		FailedCaptures:    m.FailedCaptures,
		LastScanDuration:   m.LastScanDuration,
		LastScanTime:       m.LastScanTime,
	}
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	CamerasFound       int64
	SuccessfulCaptures int64
	FailedCaptures     int64
	LastScanDuration  time.Duration
	LastScanTime       time.Time
}