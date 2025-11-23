// Package main implements a TLS termination proxy for WireGuard using ProxyGuard library
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"codeberg.org/eduVPN/proxyguard"
)

// ProxyGuardLogger implements the proxyguard logger interface
type ProxyGuardLogger struct{}

func (l *ProxyGuardLogger) Logf(msg string, params ...interface{}) {
	log.Printf("[ProxyGuard] "+msg, params...)
}

func (l *ProxyGuardLogger) Log(msg string) {
	log.Printf("[ProxyGuard] %s", msg)
}

type Config struct {
	TLSListen       string
	TLSCertFile     string
	TLSKeyFile      string
	HTTPListen      string
	WireguardTarget string
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.TLSListen, "tls-listen", ":443", "TLS listen address (e.g., :443)")
	flag.StringVar(&cfg.TLSCertFile, "tls-cert", "/certs/tls.crt", "Path to TLS certificate")
	flag.StringVar(&cfg.TLSKeyFile, "tls-key", "/certs/tls.key", "Path to TLS private key")
	flag.StringVar(&cfg.HTTPListen, "http-listen", "127.0.0.1:51821", "HTTP listen address (internal)")
	flag.StringVar(&cfg.WireguardTarget, "wireguard-target", "127.0.0.1:51820", "WireGuard UDP target")
	version := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *version {
		fmt.Printf("wireguard-tls-proxy v1.0.0\n%s", proxyguard.Version())
		os.Exit(0)
	}

	// Setup ProxyGuard logger
	proxyguard.UpdateLogger(&ProxyGuardLogger{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Start ProxyGuard HTTP server in goroutine
	proxyguardErrChan := make(chan error, 1)
	go func() {
		log.Printf("Starting ProxyGuard HTTP server on %s -> %s", cfg.HTTPListen, cfg.WireguardTarget)
		err := proxyguard.Server(ctx, cfg.HTTPListen, cfg.WireguardTarget)
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("ProxyGuard server stopped")
			default:
				proxyguardErrChan <- fmt.Errorf("ProxyGuard error: %w", err)
			}
		}
	}()

	// Start TLS proxy server
	proxy := &TLSProxy{
		TLSAddr:     cfg.TLSListen,
		CertFile:    cfg.TLSCertFile,
		KeyFile:     cfg.TLSKeyFile,
		BackendAddr: cfg.HTTPListen,
	}

	tlsErrChan := make(chan error, 1)
	go func() {
		tlsErrChan <- proxy.Start(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		log.Println("Received shutdown signal, stopping...")
		cancel()
	case err := <-proxyguardErrChan:
		log.Printf("ProxyGuard error: %v", err)
		cancel()
	case err := <-tlsErrChan:
		if err != nil {
			log.Printf("TLS proxy error: %v", err)
		}
		cancel()
	}

	log.Println("Shutdown complete")
}

// TLSProxy terminates TLS and forwards traffic to ProxyGuard
type TLSProxy struct {
	TLSAddr     string
	CertFile    string
	KeyFile     string
	BackendAddr string
}

func (p *TLSProxy) Start(ctx context.Context) error {
	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Configure TLS with minimum version 1.3 for security
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		NextProtos:   []string{"http/1.1"},
	}

	// Create HTTP server with WebSocket upgrade support
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleWebSocketUpgrade)

	server := &http.Server{
		Addr:      p.TLSAddr,
		Handler:   mux,
		TLSConfig: tlsConfig,
		// Disable timeouts for long-lived WebSocket connections
		ReadTimeout:       0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("Starting TLS proxy on %s -> %s", p.TLSAddr, p.BackendAddr)

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.ListenAndServeTLS("", "") // Certs already in TLSConfig
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Println("Shutting down TLS proxy...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-serverErr:
		return err
	}
}

func (p *TLSProxy) handleWebSocketUpgrade(w http.ResponseWriter, r *http.Request) {
	// Check if this is an upgrade request (either WebSocket or UoTLV/1 for ProxyGuard)
	upgrade := r.Header.Get("Upgrade")
	if upgrade == "" {
		http.Error(w, "Expected upgrade header", http.StatusBadRequest)
		return
	}

	// Connect to backend ProxyGuard server
	backendConn, err := net.DialTimeout("tcp", p.BackendAddr, 5*time.Second)
	if err != nil {
		log.Printf("Failed to connect to ProxyGuard backend: %v", err)
		http.Error(w, "Backend unavailable", http.StatusBadGateway)
		return
	}
	defer backendConn.Close()

	// Write the HTTP request to backend
	if err := r.Write(backendConn); err != nil {
		log.Printf("Failed to write request to backend: %v", err)
		http.Error(w, "Failed to proxy request", http.StatusBadGateway)
		return
	}

	// Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Failed to hijack connection: %v", err)
		return
	}
	defer clientConn.Close()

	// Copy any buffered data from client to backend
	if clientBuf.Reader.Buffered() > 0 {
		buffered := make([]byte, clientBuf.Reader.Buffered())
		_, err := io.ReadFull(clientBuf.Reader, buffered)
		if err != nil {
			log.Printf("Failed to read buffered data: %v", err)
			return
		}
		if _, err := backendConn.Write(buffered); err != nil {
			log.Printf("Failed to write buffered data to backend: %v", err)
			return
		}
	}

	// Bidirectional copy between client and backend
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Backend
	go func() {
		defer wg.Done()
		_, err := io.Copy(backendConn, clientConn)
		if err != nil && err != io.EOF {
			log.Printf("Error copying client->backend: %v", err)
		}
		// Close write side to signal EOF
		if tcpConn, ok := backendConn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}()

	// Backend -> Client
	go func() {
		defer wg.Done()
		_, err := io.Copy(clientConn, backendConn)
		if err != nil && err != io.EOF {
			log.Printf("Error copying backend->client: %v", err)
		}
		// Close write side to signal EOF
		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}()

	wg.Wait()
}
