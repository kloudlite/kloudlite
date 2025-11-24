package transport

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketTransport implements the Transport interface using WebSocket
type WebSocketTransport struct {
	conn   *websocket.Conn
	config *Config
	logger *zap.Logger

	// For implementing io.Reader/Writer
	readMu   sync.Mutex
	writeMu  sync.Mutex
	reader   io.Reader
	readType int

	// Ping/Pong handling
	pingTicker *time.Ticker
	pongCh     chan struct{}
	closeCh    chan struct{}
	closeOnce  sync.Once
}

// NewWebSocketTransport creates a new WebSocket transport
func NewWebSocketTransport(conn *websocket.Conn, config *Config, logger *zap.Logger) *WebSocketTransport {
	if config == nil {
		config = DefaultConfig()
	}

	wst := &WebSocketTransport{
		conn:    conn,
		config:  config,
		logger:  logger,
		pongCh:  make(chan struct{}, 1),
		closeCh: make(chan struct{}),
	}

	// Set up pong handler
	conn.SetPongHandler(func(string) error {
		select {
		case wst.pongCh <- struct{}{}:
		default:
		}
		return nil
	})

	// Start ping ticker if enabled
	if config.PingInterval > 0 {
		wst.startPingLoop()
	}

	return wst
}

func (wst *WebSocketTransport) startPingLoop() {
	wst.pingTicker = time.NewTicker(wst.config.PingInterval)

	go func() {
		defer wst.pingTicker.Stop()

		for {
			select {
			case <-wst.pingTicker.C:
				if err := wst.sendPing(); err != nil {
					wst.logger.Debug("failed to send ping", zap.Error(err))
					wst.Close()
					return
				}

				// Wait for pong with timeout
				select {
				case <-wst.pongCh:
					// Pong received
				case <-time.After(wst.config.PongTimeout):
					wst.logger.Debug("pong timeout")
					wst.Close()
					return
				case <-wst.closeCh:
					return
				}

			case <-wst.closeCh:
				return
			}
		}
	}()
}

func (wst *WebSocketTransport) sendPing() error {
	wst.writeMu.Lock()
	defer wst.writeMu.Unlock()

	if wst.config.WriteTimeout > 0 {
		wst.conn.SetWriteDeadline(time.Now().Add(wst.config.WriteTimeout))
	}

	return wst.conn.WriteMessage(websocket.PingMessage, nil)
}

// Read implements io.Reader
func (wst *WebSocketTransport) Read(p []byte) (int, error) {
	wst.readMu.Lock()
	defer wst.readMu.Unlock()

	for {
		if wst.reader == nil {
			var err error
			wst.readType, wst.reader, err = wst.conn.NextReader()
			if err != nil {
				return 0, err
			}

			// Only accept binary messages for data transfer
			if wst.readType != websocket.BinaryMessage {
				wst.reader = nil
				continue
			}
		}

		n, err := wst.reader.Read(p)
		if err == io.EOF {
			wst.reader = nil
			if n > 0 {
				return n, nil
			}
			continue
		}

		return n, err
	}
}

// Write implements io.Writer
func (wst *WebSocketTransport) Write(p []byte) (int, error) {
	wst.writeMu.Lock()
	defer wst.writeMu.Unlock()

	if wst.config.WriteTimeout > 0 {
		wst.conn.SetWriteDeadline(time.Now().Add(wst.config.WriteTimeout))
	}

	err := wst.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

// Close closes the WebSocket connection
func (wst *WebSocketTransport) Close() error {
	wst.closeOnce.Do(func() {
		close(wst.closeCh)
		if wst.pingTicker != nil {
			wst.pingTicker.Stop()
		}
	})

	return wst.conn.Close()
}

// LocalAddr returns the local network address
func (wst *WebSocketTransport) LocalAddr() net.Addr {
	return wst.conn.LocalAddr()
}

// RemoteAddr returns the remote network address
func (wst *WebSocketTransport) RemoteAddr() net.Addr {
	return wst.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines
func (wst *WebSocketTransport) SetDeadline(t time.Time) error {
	if err := wst.conn.SetReadDeadline(t); err != nil {
		return err
	}
	return wst.conn.SetWriteDeadline(t)
}

// SetReadDeadline sets the read deadline
func (wst *WebSocketTransport) SetReadDeadline(t time.Time) error {
	return wst.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (wst *WebSocketTransport) SetWriteDeadline(t time.Time) error {
	return wst.conn.SetWriteDeadline(t)
}

// WebSocketDialer implements Dialer for WebSocket connections
type WebSocketDialer struct {
	config    *Config
	tlsConfig *tls.Config
	header    http.Header
	logger    *zap.Logger
	dialer    *websocket.Dialer
}

// NewWebSocketDialer creates a new WebSocket dialer
func NewWebSocketDialer(config *Config, tlsConfig *tls.Config, header http.Header, logger *zap.Logger) *WebSocketDialer {
	if config == nil {
		config = DefaultConfig()
	}

	dialer := &websocket.Dialer{
		TLSClientConfig:   tlsConfig,
		HandshakeTimeout:  30 * time.Second,
		EnableCompression: config.EnableCompression,
	}

	return &WebSocketDialer{
		config:    config,
		tlsConfig: tlsConfig,
		header:    header,
		logger:    logger,
		dialer:    dialer,
	}
}

// Dial establishes a WebSocket connection to the remote server
func (wd *WebSocketDialer) Dial(ctx context.Context, serverURL string) (Transport, error) {
	conn, _, err := wd.dialer.DialContext(ctx, serverURL, wd.header)
	if err != nil {
		return nil, err
	}

	return NewWebSocketTransport(conn, wd.config, wd.logger), nil
}

// WebSocketListener implements Listener for WebSocket connections
type WebSocketListener struct {
	server    *http.Server
	upgrader  *websocket.Upgrader
	config    *Config
	logger    *zap.Logger
	acceptCh  chan *websocket.Conn
	errCh     chan error
	closeCh   chan struct{}
	closeOnce sync.Once
}

// NewWebSocketListener creates a new WebSocket listener
func NewWebSocketListener(addr string, tlsConfig *tls.Config, certFile, keyFile string, config *Config, logger *zap.Logger) (*WebSocketListener, error) {
	if config == nil {
		config = DefaultConfig()
	}

	listener := &WebSocketListener{
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Accept all origins
			},
			EnableCompression: config.EnableCompression,
		},
		config:   config,
		logger:   logger,
		acceptCh: make(chan *websocket.Conn),
		errCh:    make(chan error, 1),
		closeCh:  make(chan struct{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", listener.handleWebSocket)

	listener.server = &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
		// Disable timeouts for long-lived WebSocket connections
		ReadTimeout:       0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Start listening
	go func() {
		var err error
		if tlsConfig != nil || (certFile != "" && keyFile != "") {
			if certFile != "" && keyFile != "" {
				logger.Info("starting TLS WebSocket server",
					zap.String("addr", addr),
					zap.String("cert", certFile),
					zap.String("key", keyFile))
				err = listener.server.ListenAndServeTLS(certFile, keyFile)
			} else {
				logger.Info("starting TLS WebSocket server with configured certificates",
					zap.String("addr", addr))
				err = listener.server.ListenAndServeTLS("", "")
			}
		} else {
			logger.Info("starting non-TLS WebSocket server", zap.String("addr", addr))
			err = listener.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			select {
			case listener.errCh <- err:
			default:
			}
		}
	}()

	return listener, nil
}

func (wsl *WebSocketListener) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wsl.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsl.logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	select {
	case wsl.acceptCh <- conn:
	case <-wsl.closeCh:
		conn.Close()
	}
}

// Accept waits for and returns the next WebSocket connection
func (wsl *WebSocketListener) Accept(ctx context.Context) (Transport, error) {
	select {
	case conn := <-wsl.acceptCh:
		return NewWebSocketTransport(conn, wsl.config, wsl.logger), nil
	case err := <-wsl.errCh:
		return nil, err
	case <-wsl.closeCh:
		return nil, net.ErrClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close closes the WebSocket listener
func (wsl *WebSocketListener) Close() error {
	var err error
	wsl.closeOnce.Do(func() {
		close(wsl.closeCh)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = wsl.server.Shutdown(ctx)
	})
	return err
}

// Addr returns the listener's network address
func (wsl *WebSocketListener) Addr() net.Addr {
	// Parse the server address
	addr := wsl.server.Addr
	if addr == "" {
		addr = ":http"
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return &net.TCPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 443}
	}

	portNum := 443
	if port != "" {
		if p, err := net.LookupPort("tcp", port); err == nil {
			portNum = p
		}
	}

	ip := net.IPv4(0, 0, 0, 0)
	if host != "" {
		if parsedIP := net.ParseIP(host); parsedIP != nil {
			ip = parsedIP
		}
	}

	return &net.TCPAddr{IP: ip, Port: portNum}
}
