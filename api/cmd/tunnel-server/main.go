package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kloudlite/kloudlite/api/cmd/tunnel-server/handlers"
	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/transport"
	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/tunnel"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Config struct {
	ListenAddr      string
	TLSSecretName   string // Kubernetes secret name containing tls.crt and tls.key
	WireguardTarget string
	WatchConfig     bool
	ConfigPath      string

	// WireGuard peer management config
	WgDevice        string
	WgCIDR          string
	WgServerAddress string
	WgEndpoint      string

	// CA certificate config
	CACertSecretName string // Kubernetes secret name containing ca.crt

	// Hosts endpoint config
	Namespace        string
	RouterServiceRef string
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.ListenAddr, "listen", ":443", "Listen address for TLS WebSocket server (e.g., :443)")
	flag.StringVar(&cfg.TLSSecretName, "tls-secret", "tunnel-server-tls", "Kubernetes secret name containing tls.crt and tls.key")
	flag.StringVar(&cfg.WireguardTarget, "wireguard-target", "127.0.0.1:51820", "WireGuard UDP target")
	flag.BoolVar(&cfg.WatchConfig, "watch-config", false, "Watch WireGuard config and reload peers dynamically")
	flag.StringVar(&cfg.ConfigPath, "config-path", "/etc/wireguard/wg0.conf", "Path to WireGuard config file to watch")
	flag.StringVar(&cfg.WgDevice, "wg-device", "wg0", "WireGuard device name")
	flag.StringVar(&cfg.WgCIDR, "wg-cidr", "10.17.0.0/24", "WireGuard CIDR for peer IP allocation")
	flag.StringVar(&cfg.WgServerAddress, "wg-server-address", "10.17.0.1", "WireGuard server address")
	flag.StringVar(&cfg.WgEndpoint, "wg-endpoint", os.Getenv("PUBLIC_HOST"), "WireGuard server public endpoint (e.g., tunnel.example.com:443), can also be set via PUBLIC_HOST env var")
	flag.StringVar(&cfg.CACertSecretName, "ca-cert-secret", "tunnel-server-ca", "Kubernetes secret name containing ca.crt")
	flag.StringVar(&cfg.Namespace, "namespace", os.Getenv("POD_NAMESPACE"), "Namespace to query for ingresses (defaults to POD_NAMESPACE env var)")
	flag.StringVar(&cfg.RouterServiceRef, "router-service", "wm-ingress-controller", "Name of the router service for hosts resolution")
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

	// Setup WireGuard interface
	if err := setupWireGuard(cfg.ConfigPath, cfg.WgDevice, logger); err != nil {
		logger.Fatal("failed to setup WireGuard", zap.Error(err))
	}

	// Create Kubernetes client (required for loading secrets)
	k8sClient, restConfig, scheme, err := createK8sClient()
	if err != nil {
		logger.Fatal("failed to create Kubernetes client", zap.Error(err))
	}

	// Load TLS certificate from Kubernetes secret
	cert, err := loadTLSFromSecret(context.Background(), k8sClient, cfg.Namespace, cfg.TLSSecretName)
	if err != nil {
		logger.Fatal("failed to load TLS certificate from secret",
			zap.String("secret", cfg.TLSSecretName),
			zap.String("namespace", cfg.Namespace),
			zap.Error(err))
	}
	logger.Info("loaded TLS certificate from Kubernetes secret",
		zap.String("secret", cfg.TLSSecretName),
		zap.String("namespace", cfg.Namespace))

	// Configure TLS with minimum version 1.3 for security
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		NextProtos:   []string{"http/1.1"},
	}

	// Create server state for tracking connections and stats
	serverState := handlers.NewServerState()

	// Create HTTP mux with multiple endpoints
	mux := http.NewServeMux()

	// Create WebSocket listener with custom handler
	transportConfig := transport.DefaultConfig()
	listener, err := transport.NewWebSocketListenerWithHandler(
		cfg.ListenAddr,
		tlsConfig,
		"", "", // Use certificates from tlsConfig
		mux,
		transportConfig,
		logger,
	)
	if err != nil {
		logger.Fatal("failed to create WebSocket listener", zap.Error(err))
	}

	// Register handlers
	mux.HandleFunc("/ws", listener.GetWebSocketUpgradeHandler())          // WebSocket endpoint
	mux.Handle("/health", handlers.NewHealthHandler(serverState, logger)) // Health check endpoint

	// WireGuard peer management handlers
	wgHandler := handlers.NewWireGuardHandler(logger, handlers.WireGuardHandlerConfig{
		Device:        cfg.WgDevice,
		CIDR:          cfg.WgCIDR,
		ServerAddress: cfg.WgServerAddress,
		Endpoint:      cfg.WgEndpoint,
	})
	mux.HandleFunc("/wg/public-key", wgHandler.GetPublicKeyHandler()) // GET
	mux.HandleFunc("/wg/peer", wgHandler.PeerHandler())               // POST (create), DELETE (delete)

	// CA certificate handler (loads from K8s secret)
	caCertHandler := handlers.NewCACertHandler(logger, handlers.CACertHandlerConfig{
		K8sClient:  k8sClient,
		Namespace:  cfg.Namespace,
		SecretName: cfg.CACertSecretName,
	})
	mux.Handle("/ca-cert", caCertHandler)

	// Create hosts cache with watch-based updates
	hostsCache, err := handlers.NewHostsCache(logger, k8sClient, handlers.HostsCacheConfig{
		Namespace:        cfg.Namespace,
		RouterServiceRef: cfg.RouterServiceRef,
		RestConfig:       restConfig,
		Scheme:           scheme,
	})
	if err != nil {
		logger.Fatal("failed to create hosts cache", zap.Error(err))
	}

	// Hosts handler (reads from cache)
	hostsHandler := handlers.NewHostsHandler(logger, hostsCache)
	mux.Handle("/hosts", hostsHandler)

	// VPN status handler (for vpn-check connectivity verification)
	vpnStatusHandler := handlers.NewVPNStatusHandler(logger)
	mux.Handle("/", vpnStatusHandler)

	logger.Info("registered HTTP endpoints",
		zap.String("websocket", "/ws"),
		zap.String("health", "/health"),
		zap.String("wg-public-key", "GET /wg/public-key"),
		zap.String("wg-peer", "POST|DELETE /wg/peer"),
		zap.String("ca-cert", "GET /ca-cert"),
		zap.String("hosts", "GET /hosts"),
		zap.String("vpn-status", "GET /"))

	// Create UDP tunnel server
	server := tunnel.NewUDPServer(listener, logger)

	// Setup context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Start hosts cache (runs informers in background)
	go func() {
		if err := hostsCache.Start(ctx); err != nil {
			logger.Error("hosts cache error", zap.Error(err))
		}
	}()

	// Start config watcher if enabled
	if cfg.WatchConfig {
		go func() {
			logger.Info("starting WireGuard config watcher", zap.String("config", cfg.ConfigPath))
			if err := runConfigWatcher(ctx, cfg.ConfigPath, logger); err != nil && err != context.Canceled {
				logger.Error("config watcher error", zap.Error(err))
			}
		}()
	}

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

// loadTLSFromSecret loads TLS certificate and key from a Kubernetes secret
func loadTLSFromSecret(ctx context.Context, k8sClient client.Client, namespace, secretName string) (tls.Certificate, error) {
	secret := &corev1.Secret{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretName}, secret); err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to get secret %s/%s: %w", namespace, secretName, err)
	}

	certPEM, ok := secret.Data["tls.crt"]
	if !ok {
		return tls.Certificate{}, fmt.Errorf("secret %s/%s missing tls.crt", namespace, secretName)
	}

	keyPEM, ok := secret.Data["tls.key"]
	if !ok {
		return tls.Certificate{}, fmt.Errorf("secret %s/%s missing tls.key", namespace, secretName)
	}

	return tls.X509KeyPair(certPEM, keyPEM)
}

func runConfigWatcher(ctx context.Context, configPath string, logger *zap.Logger) error {
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create initial config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info("creating initial WireGuard config", zap.String("path", configPath))
		if err := createInitialConfig(configPath); err != nil {
			return fmt.Errorf("failed to create initial config: %w", err)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Watch the config file
	if err := watcher.Add(configPath); err != nil {
		return err
	}

	logger.Info("watching WireGuard config file", zap.String("path", configPath))

	// Debounce timer to avoid too frequent reloads
	var debounceTimer *time.Timer
	debounceDuration := 2 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only reload on write/create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			logger.Info("config file changed", zap.String("operation", event.Op.String()))

			// Debounce: cancel existing timer and start new one
			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			debounceTimer = time.AfterFunc(debounceDuration, func() {
				if err := syncWireGuardPeers(configPath, logger); err != nil {
					logger.Error("failed to sync WireGuard peers", zap.Error(err))
				} else {
					logger.Info("successfully synced WireGuard peers")
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			logger.Error("watcher error", zap.Error(err))
		}
	}
}

func syncWireGuardPeers(configPath string, logger *zap.Logger) error {
	logger.Info("syncing WireGuard configuration")

	// Run: wg-quick strip /etc/wireguard/wg0.conf
	stripCmd := exec.Command("wg-quick", "strip", configPath)
	strippedConfig, err := stripCmd.Output()
	if err != nil {
		logger.Error("failed to strip config", zap.Error(err))
		return err
	}

	// Create pipe to pass stripped config to wg syncconf
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	defer r.Close()

	syncCmd := exec.Command("wg", "syncconf", "wg0", "/dev/stdin")
	syncCmd.Stdin = r

	// Write stripped config to pipe in goroutine
	go func() {
		defer w.Close()
		w.Write(strippedConfig)
	}()

	// Run sync command
	output, err := syncCmd.CombinedOutput()
	if err != nil {
		logger.Error("wg syncconf failed", zap.String("output", string(output)), zap.Error(err))
		return err
	}

	logger.Info("WireGuard peers synced successfully")
	return nil
}

func createInitialConfig(configPath string) error {
	// Generate WireGuard private key
	genKeyCmd := exec.Command("wg", "genkey")
	privateKey, err := genKeyCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create initial config with proper PostUp/PostDown scripts
	// Note: ip_forward sysctl is set at pod level via securityContext
	config := fmt.Sprintf(`# WireGuard Server Configuration
[Interface]
PrivateKey = %s
Address = 10.17.0.1/24
ListenPort = 51820

# NAT and forwarding rules (ip_forward set at pod level)
PostUp = iptables -A FORWARD -i %%i -j ACCEPT
PostUp = iptables -A FORWARD -o %%i -j ACCEPT
PostUp = iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE

PostDown = iptables -D FORWARD -i %%i -j ACCEPT
PostDown = iptables -D FORWARD -o %%i -j ACCEPT
PostDown = iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

# Add peers below - they will be dynamically reloaded
# [Peer]
# PublicKey = client_public_key_here
# AllowedIPs = 10.17.0.2/32
`, string(privateKey))

	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// createK8sClient creates a Kubernetes client using in-cluster config
// Returns the client, rest config, and scheme for use by other components
func createK8sClient() (client.Client, *rest.Config, *runtime.Scheme, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	// Create a scheme with all required types
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(networkingv1.AddToScheme(scheme))
	utilruntime.Must(domainrequestv1.AddToScheme(scheme))
	utilruntime.Must(workspacev1.AddToScheme(scheme))

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return k8sClient, cfg, scheme, nil
}

// setupWireGuard initializes the WireGuard interface
func setupWireGuard(configPath, device string, logger *zap.Logger) error {
	// Check if interface already exists
	checkCmd := exec.Command("ip", "link", "show", device)
	if err := checkCmd.Run(); err == nil {
		logger.Info("WireGuard interface already exists", zap.String("device", device))
		return nil
	}

	// Create config file if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		logger.Info("creating initial WireGuard config", zap.String("path", configPath))
		if err := createInitialConfig(configPath); err != nil {
			return fmt.Errorf("failed to create initial config: %w", err)
		}
	}

	// Bring up WireGuard interface using wg-quick
	logger.Info("bringing up WireGuard interface", zap.String("device", device))
	upCmd := exec.Command("wg-quick", "up", configPath)
	output, err := upCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg-quick up failed: %s: %w", string(output), err)
	}

	logger.Info("WireGuard interface is up", zap.String("device", device))
	return nil
}
