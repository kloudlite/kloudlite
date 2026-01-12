package main

import (
	"fmt"
	"os"
	"time"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	zap2 "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	nixStorePath              = "/nix"
	workspaceHomePath         = "/var/lib/kloudlite/home"
	workspaceUserUID          = 1001
	workspaceUserGID          = 1001
	sshConfigPath             = "/var/lib/kloudlite/ssh-config"
	authorizedKeysFile        = "authorized_keys"
	packageRequestFinalizer   = "workspaces.kloudlite.io/package-cleanup"
	workspaceCleanupFinalizer = "workspaces.kloudlite.io/directory-cleanup"
)

func main() {
	// Setup logger using controller-runtime's zap logger
	opts := zap.Options{
		Development: false,
		Level:       zapcore.InfoLevel,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Create a native zap logger for our own use
	zapLogger, err := zap2.NewProduction()
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer zapLogger.Sync()

	// Create filesystem interface for dependency injection
	fs := &RealFileSystem{}

	// Setup workspace home directory with correct ownership (system-level operation)
	if err := setupWorkspaceHome(zapLogger, fs); err != nil {
		zapLogger.Fatal("Failed to setup workspace home directory", zap2.Error(err))
	}

	// Setup SSH config directory
	if err := setupSSHConfigDirectory(zapLogger, fs); err != nil {
		zapLogger.Fatal("Failed to setup SSH config directory", zap2.Error(err))
	}

	// Get namespace from environment
	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		zapLogger.Info("NAMESPACE not set, running in system setup mode only (not watching PackageRequests)")
		// Keep running but don't start the controller
		select {} // Block forever
	}

	// Get WorkMachine name from environment
	workmachineName := os.Getenv("WORKMACHINE_NAME")
	if workmachineName == "" {
		zapLogger.Fatal("WORKMACHINE_NAME environment variable not set")
	}

	// Get snapshot registry configuration from environment
	registryEndpoint := os.Getenv("SNAPSHOT_REGISTRY_ENDPOINT")
	if registryEndpoint == "" {
		registryEndpoint = "image-registry.kloudlite.svc.cluster.local:5000" // Default
	}
	registryPrefix := os.Getenv("SNAPSHOT_REGISTRY_PREFIX")
	if registryPrefix == "" {
		registryPrefix = "snapshots" // Default
	}
	registryInsecure := os.Getenv("SNAPSHOT_REGISTRY_INSECURE")
	if registryInsecure == "" {
		registryInsecure = "true" // Default
	}

	zapLogger.Info("Starting Package Manager",
		zap2.String("namespace", namespace),
		zap2.String("workmachineName", workmachineName),
		zap2.String("nixStorePath", nixStorePath),
		zap2.String("registryEndpoint", registryEndpoint),
		zap2.String("registryPrefix", registryPrefix))

	// Setup scheme
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		zapLogger.Fatal("Failed to add client-go scheme", zap2.Error(err))
	}
	if err := workspacev1.AddToScheme(scheme); err != nil {
		zapLogger.Fatal("Failed to add workspace v1 scheme", zap2.Error(err))
	}
	if err := packagesv1.AddToScheme(scheme); err != nil {
		zapLogger.Fatal("Failed to add packages v1 scheme", zap2.Error(err))
	}
	if err := snapshotv1.AddToScheme(scheme); err != nil {
		zapLogger.Fatal("Failed to add snapshot v1 scheme", zap2.Error(err))
	}
	if err := environmentv1.AddToScheme(scheme); err != nil {
		zapLogger.Fatal("Failed to add environment v1 scheme", zap2.Error(err))
	}

	// Get in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		zapLogger.Fatal("Failed to get in-cluster config", zap2.Error(err))
	}

	// Create manager watching namespace-scoped resources in namespace and cluster-scoped resources (Nodes)
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:         scheme,
		LeaderElection: false, // Each Deployment manages only its namespace
		Metrics: server.Options{
			BindAddress: ":8080",
		},
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				namespace: {}, // Watch PackageRequests and Secrets in workmachine namespace
			},
			ByObject: map[client.Object]cache.ByObject{
				// Watch Workspaces in workmachine namespace (e.g., wm-karthik)
				&workspacev1.Workspace{}: {
					Namespaces: map[string]cache.Config{
						workmachineName: {},
					},
				},
				// Watch SnapshotRequests globally (all namespaces) since they can be in any env namespace
				&snapshotv1.SnapshotRequest{}: {
					Namespaces: map[string]cache.Config{
						cache.AllNamespaces: {},
					},
				},
				// Watch SnapshotRestores globally (all namespaces) since they can be in any env namespace
				&snapshotv1.SnapshotRestore{}: {
					Namespaces: map[string]cache.Config{
						cache.AllNamespaces: {},
					},
				},
				// Watch Snapshots globally (all namespaces) since they are now namespaced
				&snapshotv1.Snapshot{}: {
					Namespaces: map[string]cache.Config{
						cache.AllNamespaces: {},
					},
				},
				// Note: Environments are read via GetAPIReader() to bypass cache
				// This avoids needing watch/list RBAC permissions for the cache
			},
			// Cluster-scoped resources (Nodes) are watched globally by default
		},
	})
	if err != nil {
		zapLogger.Fatal("Failed to create manager", zap2.Error(err))
	}

	// Setup command executor for Nix operations
	cmdExec := &RealCommandExecutor{}

	// Setup Nix profile manager
	profileManager := NewNixProfileManager(zapLogger, cmdExec)

	// Setup package reconciler
	packageReconciler := &PackageManagerReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		Logger:         zapLogger,
		Namespace:      namespace,
		CmdExec:        cmdExec,
		ProfileManager: profileManager,
	}

	if err := packageReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup package controller", zap2.Error(err))
	}

	// Setup SSH config reconciler
	sshConfigReconciler := &SSHConfigReconciler{
		Client:          mgr.GetClient(),
		Logger:          zapLogger,
		FS:              fs,
		WorkMachineName: workmachineName,
	}

	if err := sshConfigReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup SSH config controller", zap2.Error(err))
	}

	// Setup workspace cleanup reconciler (manages btrfs subvolumes for workspaces)
	workspaceCleanupReconciler := &WorkspaceCleanupReconciler{
		Client:  mgr.GetClient(),
		Logger:  zapLogger,
		FS:      fs,
		CmdExec: &HostCommandExecutor{}, // Use host executor for btrfs commands
	}

	if err := workspaceCleanupReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup workspace cleanup controller", zap2.Error(err))
	}

	// Setup GPU status reconciler
	// Get node name from environment (should match the K3s node name)
	nodeName := workmachineName // Use workmachine name as node name
	gpuStatusReconciler := &GPUStatusReconciler{
		Client:   mgr.GetClient(),
		Logger:   zapLogger,
		CmdExec:  &HostCommandExecutor{}, // Use HostCommandExecutor to run commands on host for GPU detection
		NodeName: nodeName,
	}

	if err := gpuStatusReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup GPU status controller", zap2.Error(err))
	}

	// Parse registry insecure setting
	registryInsecureBool := registryInsecure == "true"

	// Setup snapshot request reconciler (handles btrfs snapshots on this node)
	snapshotRequestReconciler := &SnapshotRequestReconciler{
		Client:           mgr.GetClient(),
		Logger:           zapLogger,
		HostCmdExec:      &HostCommandExecutor{}, // For btrfs commands on host
		NodeName:         nodeName,
		RegistryEndpoint: registryEndpoint,
		RegistryPrefix:   registryPrefix,
		RegistryInsecure: registryInsecureBool,
	}

	if err := snapshotRequestReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup snapshot request controller", zap2.Error(err))
	}

	// Setup snapshot restore reconciler (handles btrfs restore on this node)
	snapshotRestoreReconciler := &SnapshotRestoreReconciler{
		Client:           mgr.GetClient(),
		Logger:           zapLogger,
		HostCmdExec:      &HostCommandExecutor{}, // For btrfs commands on host
		NodeName:         nodeName,
		RegistryInsecure: registryInsecureBool,
	}

	if err := snapshotRestoreReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup snapshot restore controller", zap2.Error(err))
	}

	zapLogger.Info("All reconcilers configured",
		zap2.String("nodeName", nodeName))

	// Setup signal handler once
	ctx := ctrl.SetupSignalHandler()

	// Start storage garbage collector in a goroutine
	// Use GetAPIReader() to bypass cache - avoids needing watch permissions
	storageGC := &StorageGarbageCollector{
		Reader:      mgr.GetAPIReader(),
		Logger:      zapLogger,
		HostCmdExec: &HostCommandExecutor{},
		Interval:    5 * time.Minute, // Run every 5 minutes
	}
	go storageGC.Run(ctx)

	// Start metrics HTTP server in a goroutine
	metricsServer := &MetricsServer{
		CmdExec: &HostCommandExecutor{},
		Logger:  zapLogger,
		Port:    8081,
	}
	go func() {
		if err := metricsServer.Start(); err != nil {
			zapLogger.Error("Metrics server failed", zap2.Error(err))
		}
	}()

	zapLogger.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		zapLogger.Fatal("Failed to start manager", zap2.Error(err))
	}
}
