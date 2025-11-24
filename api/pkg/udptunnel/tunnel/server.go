package tunnel

import (
	"bufio"
	"context"
	"net"
	"strings"
	"sync"

	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/transport"
	"go.uber.org/zap"
)

// UDPServer handles incoming UDP tunnel requests via WebSocket
type UDPServer struct {
	listener transport.Listener
	logger   *zap.Logger
}

// NewUDPServer creates a new UDP tunnel server
func NewUDPServer(listener transport.Listener, logger *zap.Logger) *UDPServer {
	return &UDPServer{
		listener: listener,
		logger:   logger,
	}
}

// Start starts the UDP tunnel server
func (s *UDPServer) Start(ctx context.Context) error {
	s.logger.Info("UDP tunnel server started", zap.String("addr", s.listener.Addr().String()))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Accept incoming WebSocket connection
		tunnelConn, err := s.listener.Accept(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			s.logger.Error("failed to accept connection", zap.Error(err))
			continue
		}

		// Handle connection in goroutine
		go s.handleConnection(ctx, tunnelConn)
	}
}

func (s *UDPServer) handleConnection(ctx context.Context, tunnelConn transport.Transport) {
	defer tunnelConn.Close()

	// Read tunnel request header: CONNECT_UDP <remote_addr>\n
	reader := bufio.NewReader(tunnelConn)
	line, err := reader.ReadString('\n')
	if err != nil {
		s.logger.Error("failed to read tunnel request", zap.Error(err))
		return
	}

	parts := strings.Fields(line)
	if len(parts) != 2 || parts[0] != "CONNECT_UDP" {
		s.logger.Error("invalid tunnel request", zap.String("request", line))
		tunnelConn.Write([]byte("ERR\n"))
		return
	}

	remoteAddr := parts[1]
	s.logger.Debug("handling UDP tunnel request", zap.String("remote", remoteAddr))

	// Connect to the actual remote UDP destination
	destConn, err := net.Dial("udp", remoteAddr)
	if err != nil {
		s.logger.Error("failed to connect to UDP destination",
			zap.String("dest", remoteAddr),
			zap.Error(err))
		tunnelConn.Write([]byte("ERR\n"))
		return
	}
	defer destConn.Close()

	// Send success response
	if _, err := tunnelConn.Write([]byte("OK\n")); err != nil {
		return
	}

	s.logger.Debug("connected to UDP destination", zap.String("dest", remoteAddr))

	// Bidirectional forwarding
	var wg sync.WaitGroup
	wg.Add(2)

	// Forward from tunnel to destination
	go func() {
		defer wg.Done()
		s.forwardTunnelToDestination(tunnelConn, destConn)
	}()

	// Forward from destination to tunnel
	go func() {
		defer wg.Done()
		s.forwardDestinationToTunnel(destConn, tunnelConn)
	}()

	wg.Wait()
	s.logger.Debug("UDP tunnel connection closed", zap.String("dest", remoteAddr))
}

func (s *UDPServer) forwardTunnelToDestination(tunnelConn transport.Transport, destConn net.Conn) {
	buffer := make([]byte, 65536)

	for {
		// Read packet length (2 bytes)
		lengthBuf := make([]byte, 2)
		if _, err := tunnelConn.Read(lengthBuf); err != nil {
			return
		}

		length := int(lengthBuf[0])<<8 | int(lengthBuf[1])
		if length > len(buffer) {
			s.logger.Error("packet too large", zap.Int("length", length))
			return
		}

		// Read packet data
		n, err := tunnelConn.Read(buffer[:length])
		if err != nil {
			return
		}

		// Write to destination
		if _, err := destConn.Write(buffer[:n]); err != nil {
			return
		}
	}
}

func (s *UDPServer) forwardDestinationToTunnel(destConn net.Conn, tunnelConn transport.Transport) {
	buffer := make([]byte, 65536)

	for {
		n, err := destConn.Read(buffer)
		if err != nil {
			return
		}

		// Write packet length (2 bytes) + data
		length := uint16(n)
		packet := make([]byte, 2+n)
		packet[0] = byte(length >> 8)
		packet[1] = byte(length)
		copy(packet[2:], buffer[:n])

		if _, err := tunnelConn.Write(packet); err != nil {
			return
		}
	}
}

// Close closes the UDP server
func (s *UDPServer) Close() error {
	return s.listener.Close()
}
