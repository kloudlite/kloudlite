package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/transport"
	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/tunnel"
	"go.uber.org/zap"
)

type Config struct {
	LocalAddr   string // Local UDP address to listen on (e.g., 127.0.0.1:51820)
	ServerURL   string // WebSocket server URL (e.g., wss://tunnel.example.com/ws)
	RemoteAddr  string // Remote UDP address on server side (e.g., 127.0.0.1:51820)
	InsecureTLS bool   // Skip TLS certificate verification
	AuthToken   string // Optional auth token for server
	Verbose     bool   // Enable verbose logging
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.LocalAddr, "local", "127.0.0.1:51820", "Local UDP address to listen on for WireGuard traffic")
	flag.StringVar(&cfg.ServerURL, "server", "", "WebSocket server URL (e.g., wss://tunnel.example.com/ws)")
	flag.StringVar(&cfg.RemoteAddr, "remote", "127.0.0.1:51820", "Remote UDP address on server side")
	flag.BoolVar(&cfg.InsecureTLS, "insecure", false, "Skip TLS certificate verification")
	flag.StringVar(&cfg.AuthToken, "auth-token", os.Getenv("AUTH_TOKEN"), "Auth token for server (can also use AUTH_TOKEN env var)")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	version := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *version {
		fmt.Println("kl-tun-proxy v1.0.0 (WireGuard UDP Tunnel Client)")
		os.Exit(0)
	}

	if cfg.ServerURL == "" {
		log.Fatal("server URL is required (use -server flag)")
	}

	// Setup logger
	var logger *zap.Logger
	var err error
	if cfg.Verbose {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Configure TLS (server requires TLS 1.3)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureTLS,
		MinVersion:         tls.VersionTLS13,
	}

	// Setup headers (for auth if needed)
	headers := http.Header{}
	if cfg.AuthToken != "" {
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AuthToken))
	}

	// Create WebSocket dialer
	transportConfig := transport.DefaultConfig()
	dialer := transport.NewWebSocketDialer(transportConfig, tlsConfig, headers, logger)

	// Create UDP tunnel client
	client := tunnel.NewUDPClient(cfg.LocalAddr, cfg.ServerURL, cfg.RemoteAddr, dialer, logger)

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Start client in goroutine
	clientErrChan := make(chan error, 1)
	go func() {
		logger.Info("starting UDP tunnel client",
			zap.String("local", cfg.LocalAddr),
			zap.String("server", cfg.ServerURL),
			zap.String("remote", cfg.RemoteAddr))
		clientErrChan <- client.Start(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		logger.Info("received shutdown signal, stopping...")
		cancel()
	case err := <-clientErrChan:
		if err != nil && err != context.Canceled {
			logger.Error("client error", zap.Error(err))
		}
		cancel()
	}

	logger.Info("shutdown complete")
}
