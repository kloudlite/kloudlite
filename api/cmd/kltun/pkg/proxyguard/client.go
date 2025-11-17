package proxyguard

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// Client wraps WireGuard UDP traffic in HTTP/TCP using proxyguard UoTLV/1 protocol
type Client struct {
	localAddr  string // UDP address to listen on (e.g., "127.0.0.1:51821")
	remoteAddr string // HTTP address to forward to (e.g., "http://203.0.113.1:443")
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

// Start begins forwarding UDP traffic over HTTP with proxyguard protocol
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

	// Start goroutine to handle UDP->HTTP forwarding
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer udpConn.Close()
		c.forwardUDPtoHTTP(ctx, udpConn)
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

// forwardUDPtoHTTP forwards UDP packets to HTTP connection using proxyguard protocol
func (c *Client) forwardUDPtoHTTP(ctx context.Context, udpConn *net.UDPConn) {
	var (
		tcpConn    net.Conn
		clientAddr *net.UDPAddr
		mu         sync.Mutex
	)

	// Buffer for UDP packets
	buf := make([]byte, 65536)

	for {
		select {
		case <-ctx.Done():
			if tcpConn != nil {
				tcpConn.Close()
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

		// Establish HTTP connection with Upgrade if not already connected
		if tcpConn == nil {
			conn, err := c.httpUpgrade()
			if err != nil {
				fmt.Printf("Failed to establish proxyguard connection: %v\n", err)
				mu.Unlock()
				continue
			}
			tcpConn = conn

			// Start goroutine to read responses from HTTP and send back to UDP
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				c.forwardHTTPtoUDP(ctx, tcpConn, udpConn, &clientAddr, &mu)
			}()
		}

		// Write UDP packet to HTTP connection with UoTLV/1 protocol (2-byte length prefix)
		if err := c.writePacket(tcpConn, buf[:n]); err != nil {
			fmt.Printf("Error writing to HTTP connection: %v\n", err)
			tcpConn.Close()
			tcpConn = nil
		}
		mu.Unlock()
	}
}

// httpUpgrade performs HTTP Upgrade handshake to switch to UoTLV/1 protocol
func (c *Client) httpUpgrade() (net.Conn, error) {
	// Parse the remote address to extract host:port
	// remoteAddr is in format "http://host:port" or "host:port"
	addr := c.remoteAddr
	if len(addr) > 7 && addr[:7] == "http://" {
		addr = addr[7:]
	}

	// Connect to server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Send HTTP Upgrade request
	req := "GET /proxyguard HTTP/1.1\r\n" +
		"Host: " + addr + "\r\n" +
		"Connection: Upgrade\r\n" +
		"Upgrade: UoTLV/1\r\n" +
		"\r\n"

	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send upgrade request: %w", err)
	}

	// Read HTTP response
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, &http.Request{Method: "GET"})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read upgrade response: %w", err)
	}
	defer resp.Body.Close()

	// Check if protocol switch was accepted
	if resp.StatusCode != http.StatusSwitchingProtocols {
		conn.Close()
		return nil, fmt.Errorf("server did not accept protocol upgrade: %s", resp.Status)
	}

	// Connection is now upgraded to UoTLV/1
	return conn, nil
}

// writePacket writes a packet with UoTLV/1 protocol (2-byte big-endian length + data)
func (c *Client) writePacket(conn net.Conn, data []byte) error {
	// Write 2-byte length prefix (big-endian)
	length := uint16(len(data))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}

	// Write packet data
	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}

// readPacket reads a packet with UoTLV/1 protocol (2-byte big-endian length + data)
func (c *Client) readPacket(conn net.Conn) ([]byte, error) {
	// Read 2-byte length prefix (big-endian)
	var length uint16
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	// Read packet data
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	return data, nil
}

// forwardHTTPtoUDP forwards HTTP responses back to UDP client
func (c *Client) forwardHTTPtoUDP(ctx context.Context, tcpConn net.Conn, udpConn *net.UDPConn, clientAddr **net.UDPAddr, mu *sync.Mutex) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Read packet with UoTLV/1 protocol
		data, err := c.readPacket(tcpConn)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from HTTP connection: %v\n", err)
			}
			return
		}

		mu.Lock()
		if *clientAddr != nil {
			if _, err := udpConn.WriteToUDP(data, *clientAddr); err != nil {
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
