package transport

import (
	"context"
	"io"
	"net"
	"time"
)

// Transport represents the underlying transport protocol (WebSocket)
type Transport interface {
	io.ReadWriteCloser

	// LocalAddr returns the local network address
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address
	RemoteAddr() net.Addr

	// SetDeadline sets the read and write deadlines
	SetDeadline(t time.Time) error

	// SetReadDeadline sets the read deadline
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the write deadline
	SetWriteDeadline(t time.Time) error
}

// Dialer creates outbound connections to a remote server
type Dialer interface {
	// Dial establishes a connection to the remote server
	Dial(ctx context.Context, serverURL string) (Transport, error)
}

// Listener accepts incoming connections
type Listener interface {
	// Accept waits for and returns the next connection
	Accept(ctx context.Context) (Transport, error)

	// Close closes the listener
	Close() error

	// Addr returns the listener's network address
	Addr() net.Addr
}

// Config contains common configuration for transports
type Config struct {
	// PingInterval is the interval between ping messages (0 to disable)
	PingInterval time.Duration

	// PongTimeout is the timeout for receiving pong messages
	PongTimeout time.Duration

	// WriteTimeout is the timeout for write operations
	WriteTimeout time.Duration

	// ReadTimeout is the timeout for read operations
	ReadTimeout time.Duration

	// MaxMessageSize is the maximum message size in bytes
	MaxMessageSize int64

	// EnableCompression enables per-message compression
	EnableCompression bool
}

// DefaultConfig returns the default transport configuration
func DefaultConfig() *Config {
	return &Config{
		PingInterval:      30 * time.Second,
		PongTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       0,        // No timeout
		MaxMessageSize:    32 << 20, // 32 MB
		EnableCompression: false,
	}
}
