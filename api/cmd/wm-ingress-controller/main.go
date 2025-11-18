package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/kloudlite/kloudlite/api/internal/controllers/wmingress"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	var (
		healthProbeAddr  string
		namespace        string
		httpPort         int
		httpsPort        int
		ingressClassName string
	)

	flag.StringVar(&healthProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&namespace, "namespace", "", "The namespace this controller is running in.")
	flag.IntVar(&httpPort, "http-port", 8000, "The HTTP port for the ingress server.")
	flag.IntVar(&httpsPort, "https-port", 8443, "The HTTPS port for the ingress server.")
	flag.StringVar(&ingressClassName, "ingress-class", "", "The ingress class name to watch for.")
	flag.Parse()

	// Validate required flags
	if namespace == "" {
		namespace = os.Getenv("POD_NAMESPACE")
		if namespace == "" {
			fmt.Println("Error: namespace must be provided via --namespace flag or POD_NAMESPACE env var")
			os.Exit(1)
		}
	}

	if ingressClassName == "" {
		fmt.Println("Error: ingress-class must be provided via --ingress-class flag")
		os.Exit(1)
	}

	// Setup logger
	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.TimeKey = "timestamp"
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := logConfig.Build()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Ingress Controller",
		zap.String("namespace", namespace),
		zap.String("ingress-class", ingressClassName),
		zap.Int("http-port", httpPort),
		zap.Int("https-port", httpsPort),
	)

	// Setup controller-runtime manager
	// No leader election needed - this is a read-only controller
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// Watch all namespaces for Ingress resources
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{},
		},
	})
	if err != nil {
		logger.Fatal("Unable to create manager", zap.Error(err))
	}

	// Create and setup the Ingress Reconciler
	reconciler := &wmingress.IngressReconciler{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		Logger:           logger,
		Namespace:        namespace,
		IngressClassName: ingressClassName,
		HTTPPort:         httpPort,
		HTTPSPort:        httpsPort,
	}

	if err = reconciler.SetupWithManager(mgr); err != nil {
		logger.Fatal("Unable to setup IngressReconciler", zap.Error(err))
	}

	// Add health check endpoints
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logger.Fatal("Unable to setup health check", zap.Error(err))
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logger.Fatal("Unable to setup ready check", zap.Error(err))
	}

	// Start HTTP/HTTPS servers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := reconciler.StartServers(ctx); err != nil {
		logger.Fatal("Failed to start HTTP/HTTPS servers", zap.Error(err))
	}

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Received termination signal, shutting down...")
		cancel()
	}()

	// Start the manager
	logger.Info("Starting controller manager")
	if err := mgr.Start(ctx); err != nil {
		logger.Fatal("Controller manager exited with error", zap.Error(err))
	}

	logger.Info("Controller manager stopped")
}
