package proxyguard

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Client wraps WireGuard UDP traffic in TLS/TCP for firewall traversal
type Client struct {
	localAddr  string // UDP address to listen on (e.g., "127.0.0.1:51821")
	remoteAddr string // TLS/TCP address to forward to (e.g., "203.0.113.1:443")
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewClient creates a new proxyguard client
func NewClient(localAddr, remoteAddr string) *Client {
	return &Client{
		localAddr:  localAddr,
		remoteAddr: remoteAddr,
	}
}

// Start begins forwarding UDP traffic over TLS/TCP
func (c *Client) Start(ctx context.Context) error {
	// Listen on UDP for WireGuard traffic
	udpAddr, err := net.ResolveUDPAddr("udp", c.localAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve local UDP address: %w", err)
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// Start goroutine to handle UDP->TCP forwarding
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer udpConn.Close()
		c.forwardUDPtoTCP(ctx, udpConn)
	}()

	return nil
}

// Stop stops the proxyguard client
func (c *Client) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

// forwardUDPtoTCP forwards UDP packets to TLS/TCP connection
func (c *Client) forwardUDPtoTCP(ctx context.Context, udpConn *net.UDPConn) {
	var (
		tlsConn  *tls.Conn
		clientAddr *net.UDPAddr
		mu       sync.Mutex
	)

	// Buffer for UDP packets
	buf := make([]byte, 65536)

	for {
		select {
		case <-ctx.Done():
			if tlsConn != nil {
				tlsConn.Close()
			}
			return
		default:
		}

		// Set read deadline to check context periodically
		udpConn.SetReadDeadline(timeoutFromContext(ctx))

		n, addr, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			fmt.Printf("Error reading from UDP: %v\n", err)
			continue
		}

		mu.Lock()
		// Track the client address for responses
		if clientAddr == nil || clientAddr.String() != addr.String() {
			clientAddr = addr
		}

		// Establish TLS connection if not already connected
		if tlsConn == nil {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true, // TODO: Add proper certificate verification
			}

			conn, err := tls.Dial("tcp", c.remoteAddr, tlsConfig)
			if err != nil {
				fmt.Printf("Failed to connect to remote server: %v\n", err)
				mu.Unlock()
				continue
			}
			tlsConn = conn

			// Start goroutine to read responses from TLS and send back to UDP
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				c.forwardTCPtoUDP(ctx, tlsConn, udpConn, &clientAddr, &mu)
			}()
		}

		// Forward UDP packet to TLS connection
		if _, err := tlsConn.Write(buf[:n]); err != nil {
			fmt.Printf("Error writing to TLS connection: %v\n", err)
			tlsConn.Close()
			tlsConn = nil
		}
		mu.Unlock()
	}
}

// forwardTCPtoUDP forwards TCP/TLS responses back to UDP client
func (c *Client) forwardTCPtoUDP(ctx context.Context, tlsConn *tls.Conn, udpConn *net.UDPConn, clientAddr **net.UDPAddr, mu *sync.Mutex) {
	buf := make([]byte, 65536)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := tlsConn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from TLS connection: %v\n", err)
			}
			return
		}

		mu.Lock()
		if *clientAddr != nil {
			if _, err := udpConn.WriteToUDP(buf[:n], *clientAddr); err != nil {
				fmt.Printf("Error writing to UDP: %v\n", err)
			}
		}
		mu.Unlock()
	}
}

// timeoutFromContext returns a deadline based on context
func timeoutFromContext(ctx context.Context) time.Time {
	// Use a short timeout to allow periodic context checking
	return time.Now().Add(100 * time.Millisecond)
}
