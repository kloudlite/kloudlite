package test

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

// MockLogger creates a mock logger for testing
func MockLogger(t *testing.T) *zap.Logger {
	return zaptest.NewLogger(t)
}

// MockConfig represents a mock configuration
type MockConfig struct {
	Port        string
	Environment string
	LogLevel    string
	Namespace   string
}

// DefaultMockConfig returns a default mock configuration
func DefaultMockConfig() MockConfig {
	return MockConfig{
		Port:        "8080",
		Environment: "test",
		LogLevel:    "debug",
		Namespace:   "test-namespace",
	}
}

// MockTimeProvider provides mock time for testing
type MockTimeProvider struct {
	CurrentTime time.Time
}

// Now returns the mock current time
func (m *MockTimeProvider) Now() time.Time {
	if m.CurrentTime.IsZero() {
		return time.Now()
	}
	return m.CurrentTime
}

// SetTime sets the mock time
func (m *MockTimeProvider) SetTime(t time.Time) {
	m.CurrentTime = t
}