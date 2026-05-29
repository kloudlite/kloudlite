package handlers

import (
	"sync"
	"sync/atomic"
	"time"
)

// ServerState holds the shared state for the tunnel server
type ServerState struct {
	startTime         time.Time
	activeConnections atomic.Int64
	totalConnections  atomic.Int64
	bytesReceived     atomic.Uint64
	bytesSent         atomic.Uint64
	mu                sync.RWMutex
}

// NewServerState creates a new ServerState
func NewServerState() *ServerState {
	return &ServerState{
		startTime: time.Now(),
	}
}

// IncrementConnections increments the connection counters
func (s *ServerState) IncrementConnections() {
	s.activeConnections.Add(1)
	s.totalConnections.Add(1)
}

// DecrementConnections decrements the active connections counter
func (s *ServerState) DecrementConnections() {
	s.activeConnections.Add(-1)
}

// GetActiveConnections returns the current number of active connections
func (s *ServerState) GetActiveConnections() int64 {
	return s.activeConnections.Load()
}

// GetTotalConnections returns the total number of connections since start
func (s *ServerState) GetTotalConnections() int64 {
	return s.totalConnections.Load()
}

// GetUptime returns the server uptime duration
func (s *ServerState) GetUptime() time.Duration {
	return time.Since(s.startTime)
}

// GetStartTime returns the server start time
func (s *ServerState) GetStartTime() time.Time {
	return s.startTime
}

// AddBytesReceived adds to the received bytes counter
func (s *ServerState) AddBytesReceived(bytes uint64) {
	s.bytesReceived.Add(bytes)
}

// AddBytesSent adds to the sent bytes counter
func (s *ServerState) AddBytesSent(bytes uint64) {
	s.bytesSent.Add(bytes)
}

// GetBytesReceived returns the total bytes received
func (s *ServerState) GetBytesReceived() uint64 {
	return s.bytesReceived.Load()
}

// GetBytesSent returns the total bytes sent
func (s *ServerState) GetBytesSent() uint64 {
	return s.bytesSent.Load()
}
