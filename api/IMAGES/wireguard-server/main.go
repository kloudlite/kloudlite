package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/transport"
	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/tunnel"
	"go.uber.org/zap"
)

type Config struct {
	ListenAddr      string
	TLSCertFile     string
	TLSKeyFile      string
	WireguardTarget string
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.ListenAddr, "listen", ":443", "Listen address for TLS WebSocket server (e.g., :443)")
	flag.StringVar(&cfg.TLSCertFile, "tls-cert", "/certs/tls.crt", "Path to TLS certificate")
	flag.StringVar(&cfg.TLSKeyFile, "tls-key", "/certs/tls.key", "Path to TLS private key")
	flag.StringVar(&cfg.WireguardTarget, "wireguard-target", "127.0.0.1:51820", "WireGuard UDP target")
	version := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *version {
		fmt.Println("wireguard-tls-proxy v2.0.0 (UDP-over-WebSocket)")
		os.Exit(0)
	}

	// Setup logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		logger.Fatal("failed to load TLS certificate",
			zap.String("cert", cfg.TLSCertFile),
			zap.String("key", cfg.TLSKeyFile),
			zap.Error(err))
	}

	// Configure TLS with minimum version 1.3 for security
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		NextProtos:   []string{"http/1.1"},
	}

	// Create WebSocket listener
	transportConfig := transport.DefaultConfig()
	listener, err := transport.NewWebSocketListener(
		cfg.ListenAddr,
		tlsConfig,
		"", "", // Use certificates from tlsConfig
		transportConfig,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to create WebSocket listener", zap.Error(err))
	}

	// Create UDP tunnel server
	server := tunnel.NewUDPServer(listener, logger)

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Start server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info("starting UDP-over-WebSocket server",
			zap.String("listen", cfg.ListenAddr),
			zap.String("wireguard_target", cfg.WireguardTarget))
		serverErrChan <- server.Start(ctx)
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigChan:
		logger.Info("received shutdown signal, stopping...")
		cancel()
	case err := <-serverErrChan:
		if err != nil && err != context.Canceled {
			logger.Error("server error", zap.Error(err))
		}
		cancel()
	}

	// Graceful shutdown
	if err := server.Close(); err != nil {
		logger.Error("error closing server", zap.Error(err))
	}

	logger.Info("shutdown complete")
}
