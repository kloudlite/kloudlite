package tunnel

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/transport"
	"go.uber.org/zap"
)

// UDPClient handles UDP tunneling from local UDP to remote server via WebSocket
type UDPClient struct {
	localAddr  string
	serverURL  string
	remoteAddr string
	logger     *zap.Logger
	dialer     transport.Dialer
	sessions   sync.Map // Map of client addresses to session info
}

// udpSession represents a UDP tunnel session
type udpSession struct {
	clientAddr *net.UDPAddr
	tunnelConn transport.Transport
	lastActive time.Time
	cancel     context.CancelFunc
	mu         sync.Mutex
}

// NewUDPClient creates a new UDP tunnel client
func NewUDPClient(localAddr, serverURL, remoteAddr string, dialer transport.Dialer, logger *zap.Logger) *UDPClient {
	return &UDPClient{
		localAddr:  localAddr,
		serverURL:  serverURL,
		remoteAddr: remoteAddr,
		logger:     logger,
		dialer:     dialer,
	}
}

// Start starts the UDP tunnel client
func (c *UDPClient) Start(ctx context.Context) error {
	udpAddr, err := net.ResolveUDPAddr("udp", c.localAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address %s: %w", c.localAddr, err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", c.localAddr, err)
	}
	defer conn.Close()

	c.logger.Info("UDP tunnel client listening",
		zap.String("local", c.localAddr),
		zap.String("server", c.serverURL),
		zap.String("remote", c.remoteAddr))

	// Start cleanup goroutine for expired sessions
	go c.cleanupSessions(ctx, 30*time.Second)

	buffer := make([]byte, 65536) // Max UDP packet size

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.logger.Error("failed to read UDP packet", zap.Error(err))
			continue
		}

		// Get or create session for this client
		session := c.getOrCreateSession(ctx, clientAddr, conn)
		if session == nil {
			continue
		}

		// Forward packet through tunnel
		go c.forwardPacket(session, buffer[:n])
	}
}

func (c *UDPClient) getOrCreateSession(ctx context.Context, clientAddr *net.UDPAddr, localConn *net.UDPConn) *udpSession {
	key := clientAddr.String()

	// Check if session exists
	if sess, ok := c.sessions.Load(key); ok {
		session := sess.(*udpSession)
		session.mu.Lock()
		session.lastActive = time.Now()
		session.mu.Unlock()
		return session
	}

	// Create new session
	sessionCtx, cancel := context.WithCancel(ctx)

	// Connect to tunnel server via WebSocket
	tunnelConn, err := c.dialer.Dial(sessionCtx, c.serverURL)
	if err != nil {
		c.logger.Error("failed to dial tunnel server", zap.Error(err))
		cancel()
		return nil
	}

	// Send tunnel request header: CONNECT_UDP <remote_addr>\n
	header := fmt.Sprintf("CONNECT_UDP %s\n", c.remoteAddr)
	if _, err := tunnelConn.Write([]byte(header)); err != nil {
		c.logger.Error("failed to send UDP tunnel request", zap.Error(err))
		tunnelConn.Close()
		cancel()
		return nil
	}

	// Read response: OK\n or ERR\n
	response := make([]byte, 3)
	if _, err := tunnelConn.Read(response); err != nil {
		c.logger.Error("failed to read UDP tunnel response", zap.Error(err))
		tunnelConn.Close()
		cancel()
		return nil
	}

	if string(response) != "OK\n" {
		c.logger.Error("UDP tunnel request rejected", zap.String("response", string(response)))
		tunnelConn.Close()
		cancel()
		return nil
	}

	session := &udpSession{
		clientAddr: clientAddr,
		tunnelConn: tunnelConn,
		lastActive: time.Now(),
		cancel:     cancel,
	}

	c.sessions.Store(key, session)

	c.logger.Debug("created UDP session",
		zap.String("client", clientAddr.String()),
		zap.String("remote", c.remoteAddr))

	// Start receiving responses from tunnel
	go c.receiveFromTunnel(sessionCtx, session, localConn)

	return session
}

func (c *UDPClient) forwardPacket(session *udpSession, data []byte) {
	session.mu.Lock()
	defer session.mu.Unlock()

	// Write packet length (2 bytes) + data
	length := uint16(len(data))
	packet := make([]byte, 2+len(data))
	packet[0] = byte(length >> 8)
	packet[1] = byte(length)
	copy(packet[2:], data)

	if _, err := session.tunnelConn.Write(packet); err != nil {
		c.logger.Error("failed to write to tunnel", zap.Error(err))
		return
	}

	session.lastActive = time.Now()
}

func (c *UDPClient) receiveFromTunnel(ctx context.Context, session *udpSession, localConn *net.UDPConn) {
	defer func() {
		session.tunnelConn.Close()
		c.sessions.Delete(session.clientAddr.String())
		session.cancel()
	}()

	buffer := make([]byte, 65536)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read packet length (2 bytes)
		lengthBuf := make([]byte, 2)
		if _, err := session.tunnelConn.Read(lengthBuf); err != nil {
			c.logger.Debug("tunnel read error", zap.Error(err))
			return
		}

		length := int(lengthBuf[0])<<8 | int(lengthBuf[1])
		if length > len(buffer) {
			c.logger.Error("packet too large", zap.Int("length", length))
			return
		}

		// Read packet data
		n, err := session.tunnelConn.Read(buffer[:length])
		if err != nil {
			c.logger.Debug("tunnel read error", zap.Error(err))
			return
		}

		// Send to client
		if _, err := localConn.WriteToUDP(buffer[:n], session.clientAddr); err != nil {
			c.logger.Error("failed to write to client", zap.Error(err))
			return
		}

		session.mu.Lock()
		session.lastActive = time.Now()
		session.mu.Unlock()
	}
}

func (c *UDPClient) cleanupSessions(ctx context.Context, timeout time.Duration) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if timeout == 0 {
				continue // No timeout
			}

			now := time.Now()
			c.sessions.Range(func(key, value interface{}) bool {
				session := value.(*udpSession)
				session.mu.Lock()
				expired := now.Sub(session.lastActive) > timeout
				session.mu.Unlock()

				if expired {
					c.logger.Debug("cleaning up expired UDP session", zap.String("client", key.(string)))
					session.cancel()
					session.tunnelConn.Close()
					c.sessions.Delete(key)
				}

				return true
			})
		}
	}
}
