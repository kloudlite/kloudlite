package workmachinenodemanager

import (
	"context"
	"fmt"
	"time"

	checkpointv1 "github.com/kloudlite/kloudlite/types/checkpoint/v1"
	environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/types/packages/v1"
	workspacev1 "github.com/kloudlite/kloudlite/types/workspace/v1"
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
	nixStorePath            = "/nix"
	workspaceHomePath       = "/var/lib/kloudlite/home"
	workspaceUserUID        = 1001
	workspaceUserGID        = 1001
	sshConfigPath           = "/var/lib/kloudlite/ssh-config"
	authorizedKeysFile      = "authorized_keys"
	packageRequestFinalizer = "workspaces.kloudlite.io/package-cleanup"
)

type Config struct {
	Namespace                string
	WorkMachineName          string
	SnapshotRegistryEndpoint string
	SnapshotRegistryPrefix   string
	SnapshotRegistryInsecure string
}

func nodeManagerReconcilerNames() []string {
	return []string{
		"package",
		"ssh-config",
		"gpu-status",
		"checkpoint",
		"checkpoint-restore",
		"storage-gc",
	}
}

func Start(ctx context.Context, cfg Config) error {
	// Setup logger using controller-runtime's zap logger
	opts := zap.Options{
		Development: false,
		Level:       zapcore.InfoLevel,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Create a native zap logger for our own use
	zapLogger, err := zap2.NewProduction()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer zapLogger.Sync()

	// Create filesystem interface for dependency injection
	fs := &RealFileSystem{}

	// Setup workspace home directory with correct ownership (system-level operation)
	if err := setupWorkspaceHome(zapLogger, fs); err != nil {
		return fmt.Errorf("failed to setup workspace home directory: %w", err)
	}

	// Setup SSH config directory
	if err := setupSSHConfigDirectory(zapLogger, fs); err != nil {
		return fmt.Errorf("failed to setup SSH config directory: %w", err)
	}

	namespace := cfg.Namespace
	if namespace == "" {
		zapLogger.Info("NAMESPACE not set, running in system setup mode only (not watching PackageRequests)")
		<-ctx.Done()
		return nil
	}

	workmachineName := cfg.WorkMachineName
	if workmachineName == "" {
		return fmt.Errorf("WORKMACHINE_NAME is required")
	}

	registryEndpoint := cfg.SnapshotRegistryEndpoint
	if registryEndpoint == "" {
		registryEndpoint = "image-registry.kloudlite.svc.cluster.local:5000" // Default
	}
	registryPrefix := cfg.SnapshotRegistryPrefix
	if registryPrefix == "" {
		registryPrefix = "snapshots" // Default
	}
	registryInsecure := cfg.SnapshotRegistryInsecure
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
		return fmt.Errorf("failed to add client-go scheme: %w", err)
	}
	if err := workspacev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add workspace v1 scheme: %w", err)
	}
	if err := packagesv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add packages v1 scheme: %w", err)
	}
	if err := checkpointv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add checkpoint v1 scheme: %w", err)
	}
	if err := environmentv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add environment v1 scheme: %w", err)
	}

	// Get in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get in-cluster config: %w", err)
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
				// Watch Checkpoints globally (all namespaces) since they can be in any env namespace
				&checkpointv1.Checkpoint{}: {
					Namespaces: map[string]cache.Config{
						cache.AllNamespaces: {},
					},
				},
				// Watch CheckpointRestores globally (all namespaces) since they can be in any env namespace
				&checkpointv1.CheckpointRestore{}: {
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
		return fmt.Errorf("failed to create manager: %w", err)
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
		return fmt.Errorf("failed to setup package controller: %w", err)
	}

	// Setup SSH config reconciler
	sshConfigReconciler := &SSHConfigReconciler{
		Client:          mgr.GetClient(),
		Logger:          zapLogger,
		FS:              fs,
		WorkMachineName: workmachineName,
	}

	if err := sshConfigReconciler.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup SSH config controller: %w", err)
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
		return fmt.Errorf("failed to setup GPU status controller: %w", err)
	}

	// Parse registry insecure setting
	registryInsecureBool := registryInsecure == "true"

	// Setup checkpoint reconciler (handles btrfs checkpoints on this node)
	checkpointReconciler := &CheckpointReconciler{
		Client:           mgr.GetClient(),
		Logger:           zapLogger,
		HostCmdExec:      &HostCommandExecutor{},
		NodeName:         nodeName,
		RegistryEndpoint: registryEndpoint,
		RegistryPrefix:   registryPrefix,
		RegistryInsecure: registryInsecureBool,
	}

	if err := checkpointReconciler.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup checkpoint controller: %w", err)
	}

	// Setup checkpoint restore reconciler (handles btrfs restore on this node)
	checkpointRestoreReconciler := &CheckpointRestoreReconciler{
		Client:           mgr.GetClient(),
		Logger:           zapLogger,
		HostCmdExec:      &HostCommandExecutor{},
		NodeName:         nodeName,
		RegistryInsecure: registryInsecureBool,
	}

	if err := checkpointRestoreReconciler.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed to setup checkpoint restore controller: %w", err)
	}

	zapLogger.Info("All reconcilers configured",
		zap2.String("nodeName", nodeName))

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
		return fmt.Errorf("failed to start manager: %w", err)
	}
	return nil
}
