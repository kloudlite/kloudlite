package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	packagesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/packages/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	zap2 "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	nixStorePath        = "/var/lib/kloudlite/nix-store"
	workspaceHomePath   = "/var/lib/kloudlite/workspace-homes/kl"
	workspaceUserUID    = 1001
	workspaceUserGID    = 1001
	sshConfigPath       = "/var/lib/kloudlite/ssh-config"
	authorizedKeysFile  = "authorized_keys"
)

// CommandExecutor defines an interface for executing shell commands
// This allows for mocking in tests
type CommandExecutor interface {
	Execute(script string) ([]byte, error)
}

// RealCommandExecutor implements CommandExecutor using os/exec
type RealCommandExecutor struct{}

func (r *RealCommandExecutor) Execute(script string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", script)
	// Pass through parent environment variables (PATH, etc.)
	// Don't set NIX_STATE_DIR since we're mounting /nix/store and /nix/var directly
	cmd.Env = os.Environ()
	return cmd.CombinedOutput()
}

// FileSystem defines an interface for filesystem operations
// This allows for mocking in tests
type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	Chown(name string, uid, gid int) error
	WriteFile(name string, data []byte, perm os.FileMode) error
	Rename(oldpath, newpath string) error
}

// RealFileSystem implements FileSystem using os package
type RealFileSystem struct{}

func (r *RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (r *RealFileSystem) Chown(name string, uid, gid int) error {
	return os.Chown(name, uid, gid)
}

func (r *RealFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (r *RealFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

type PackageManagerReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Logger    *zap2.Logger
	Namespace string
	CmdExec   CommandExecutor
}

func (r *PackageManagerReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("packageRequest", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling PackageRequest")

	// Fetch PackageRequest
	pkgReq := &packagesv1.PackageRequest{}
	if err := r.Get(ctx, req.NamespacedName, pkgReq); err != nil {
		logger.Error("Failed to get PackageRequest", zap2.Error(err))
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Check if packages changed (by comparing spec packages with installed packages)
	// If packages changed, we need to reconcile even if in terminal state
	packagesChanged := len(pkgReq.Spec.Packages) != len(pkgReq.Status.InstalledPackages)
	if !packagesChanged {
		// Check if any package changed
		installedMap := make(map[string]bool)
		for _, installed := range pkgReq.Status.InstalledPackages {
			installedMap[installed.Name] = true
		}
		for _, pkg := range pkgReq.Spec.Packages {
			if !installedMap[pkg.Name] {
				packagesChanged = true
				break
			}
		}
	}

	// Skip if already in terminal state (Ready or Failed) AND packages haven't changed
	if !packagesChanged && (pkgReq.Status.Phase == "Ready" || pkgReq.Status.Phase == "Failed") {
		logger.Info("PackageRequest already in terminal state and packages haven't changed",
			zap2.String("phase", pkgReq.Status.Phase))
		return reconcile.Result{}, nil
	}

	if packagesChanged {
		logger.Info("Package list changed, will reconcile",
			zap2.Int("spec_packages", len(pkgReq.Spec.Packages)),
			zap2.Int("installed_packages", len(pkgReq.Status.InstalledPackages)))
	}

	// Update status to Installing
	pkgReq.Status.Phase = "Installing"
	pkgReq.Status.Message = "Installing packages"
	pkgReq.Status.LastUpdated = metav1.Now()
	if err := r.Status().Update(ctx, pkgReq); err != nil {
		logger.Error("Failed to update status to Installing", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Using profile", zap2.String("profile", pkgReq.Spec.ProfileName))

	// Detect and remove packages that are no longer in the spec
	specPackageNames := make(map[string]bool)
	for _, pkg := range pkgReq.Spec.Packages {
		specPackageNames[pkg.Name] = true
	}

	for _, installedPkg := range pkgReq.Status.InstalledPackages {
		if !specPackageNames[installedPkg.Name] {
			logger.Info("Removing package",
				zap2.String("package", installedPkg.Name),
				zap2.String("profile", pkgReq.Spec.ProfileName))

			if err := r.uninstallPackage(installedPkg.Name, pkgReq.Spec.ProfileName); err != nil {
				logger.Error("Failed to remove package",
					zap2.String("package", installedPkg.Name),
					zap2.String("profile", pkgReq.Spec.ProfileName),
					zap2.Error(err))
			} else {
				logger.Info("Successfully removed package",
					zap2.String("package", installedPkg.Name),
					zap2.String("profile", pkgReq.Spec.ProfileName))
			}
		}
	}

	// Install packages
	installedPackages := []workspacesv1.InstalledPackage{}
	failedPackages := []string{}

	for _, pkg := range pkgReq.Spec.Packages {
		logger.Info("Installing package",
			zap2.String("package", pkg.Name),
			zap2.String("profile", pkgReq.Spec.ProfileName))

		installedPkg, err := r.installPackage(pkg, pkgReq.Spec.ProfileName)
		if err != nil {
			logger.Error("Failed to install package",
				zap2.String("package", pkg.Name),
				zap2.String("profile", pkgReq.Spec.ProfileName),
				zap2.Error(err))
			failedPackages = append(failedPackages, pkg.Name)
			continue
		}

		installedPackages = append(installedPackages, installedPkg)
		logger.Info("Successfully installed package",
			zap2.String("package", pkg.Name),
			zap2.String("profile", pkgReq.Spec.ProfileName),
			zap2.String("storePath", installedPkg.StorePath))
	}

	// Update PackageRequest status
	pkgReq.Status.InstalledPackages = installedPackages
	pkgReq.Status.FailedPackages = failedPackages
	pkgReq.Status.LastUpdated = metav1.Now()

	if len(failedPackages) > 0 {
		pkgReq.Status.Phase = "Failed"
		pkgReq.Status.Message = fmt.Sprintf("Failed to install %d packages", len(failedPackages))
	} else {
		pkgReq.Status.Phase = "Ready"
		pkgReq.Status.Message = fmt.Sprintf("Successfully installed %d packages", len(installedPackages))
	}

	// Retry status update with optimistic concurrency control
	if err := r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *packagesv1.PackageRequest) {
		latest.Status.InstalledPackages = installedPackages
		latest.Status.FailedPackages = failedPackages
		latest.Status.LastUpdated = metav1.Now()
		if len(failedPackages) > 0 {
			latest.Status.Phase = "Failed"
			latest.Status.Message = fmt.Sprintf("Failed to install %d packages", len(failedPackages))
		} else {
			latest.Status.Phase = "Ready"
			latest.Status.Message = fmt.Sprintf("Successfully installed %d packages", len(installedPackages))
		}
	}, logger); err != nil {
		logger.Error("Failed to update status after retries", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("PackageRequest reconciliation complete",
		zap2.String("phase", pkgReq.Status.Phase),
		zap2.Int("installed", len(installedPackages)),
		zap2.Int("failed", len(failedPackages)))

	return reconcile.Result{}, nil
}

// updateStatusWithRetry retries status updates with optimistic concurrency control
// It fetches the latest version before each update attempt
func (r *PackageManagerReconciler) updateStatusWithRetry(
	ctx context.Context,
	namespacedName client.ObjectKey,
	updateFn func(*packagesv1.PackageRequest),
	logger *zap2.Logger,
) error {
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		// Fetch the latest version
		latest := &packagesv1.PackageRequest{}
		if err := r.Get(ctx, namespacedName, latest); err != nil {
			return fmt.Errorf("failed to fetch latest PackageRequest: %w", err)
		}

		// Apply the update function
		updateFn(latest)

		// Try to update status
		if err := r.Status().Update(ctx, latest); err != nil {
			if apierrors.IsConflict(err) && i < maxRetries-1 {
				logger.Info("Status update conflict, retrying",
					zap2.Int("attempt", i+1),
					zap2.Int("maxRetries", maxRetries))
				continue
			}
			return fmt.Errorf("failed to update status: %w", err)
		}

		// Success
		return nil
	}

	return fmt.Errorf("failed to update status after %d retries", maxRetries)
}

func (r *PackageManagerReconciler) installPackage(pkg workspacesv1.PackageSpec, profileName string) (workspacesv1.InstalledPackage, error) {
	// Determine package source and install command
	// Store profiles in nix-store so they're accessible via hostPath mount in workspace pods
	profilePath := fmt.Sprintf("%s/profiles/per-user/root/%s", nixStorePath, profileName)

	var installScript string
	var pkgSource string

	// Priority: NixpkgsCommit > Channel > latest unstable
	if pkg.NixpkgsCommit != "" {
		// Install from specific nixpkgs commit using nix-env with tarball
		// We use nix-env instead of nix profile because not all commits have flake.nix
		nixpkgsTarball := fmt.Sprintf("https://github.com/nixos/nixpkgs/archive/%s.tar.gz", pkg.NixpkgsCommit)
		pkgAttr := pkg.Name
		installScript = fmt.Sprintf(
			`. /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -f %s -iA %s`,
			profilePath, nixpkgsTarball, pkgAttr,
		)
		r.Logger.Info("Installing package from nixpkgs commit",
			zap2.String("package", pkg.Name),
			zap2.String("commit", pkg.NixpkgsCommit))
	} else if pkg.Channel != "" {
		// Install from specific channel/release (e.g., nixos-24.05, nixos-23.11, unstable)
		pkgSource = fmt.Sprintf("nixpkgs/%s#%s", pkg.Channel, pkg.Name)
		installScript = fmt.Sprintf(
			`. /root/.nix-profile/etc/profile.d/nix.sh && nix --extra-experimental-features "nix-command flakes" profile install --profile %s '%s'`,
			profilePath, pkgSource,
		)
		r.Logger.Info("Installing package from channel",
			zap2.String("package", pkg.Name),
			zap2.String("channel", pkg.Channel))
	} else {
		// Install latest version from nixpkgs unstable using nix-env (legacy, more compatible)
		pkgAttr := "nixpkgs." + pkg.Name
		installScript = fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -iA %s", profilePath, pkgAttr)
		r.Logger.Info("Installing package from nixpkgs unstable",
			zap2.String("package", pkg.Name))
	}

	output, err := r.CmdExec.Execute(installScript)
	if err != nil {
		return workspacesv1.InstalledPackage{}, fmt.Errorf("nix-env failed: %w, output: %s", err, string(output))
	}

	// Query package info to get store path and actual installed version from the named profile
	queryScript := fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -q --out-path %s", profilePath, pkg.Name)

	queryOutput, err := r.CmdExec.Execute(queryScript)
	if err != nil {
		r.Logger.Warn("Failed to query package path, using default",
			zap2.String("package", pkg.Name),
			zap2.Error(err))
	}

	// Parse store path and version from output
	storePath := nixStorePath + "/store"
	installedVersion := ""

	// Determine version source for status
	if pkg.NixpkgsCommit != "" {
		installedVersion = "commit:" + pkg.NixpkgsCommit[:8] // Short commit hash
	} else if pkg.Channel != "" {
		installedVersion = "channel:" + pkg.Channel
	}

	if len(queryOutput) > 0 {
		// Output format is typically: "package-version  /nix/store/hash-package-version"
		parts := strings.Fields(string(queryOutput))
		if len(parts) >= 1 {
			// First part is "package-version", extract version
			pkgWithVersion := parts[0]
			// Remove package name prefix to get version
			if strings.HasPrefix(pkgWithVersion, pkg.Name+"-") {
				actualVersion := strings.TrimPrefix(pkgWithVersion, pkg.Name+"-")
				// Append actual version to source info
				if installedVersion != "" {
					installedVersion = actualVersion + " (" + installedVersion + ")"
				} else {
					installedVersion = actualVersion
				}
			} else if installedVersion == "" {
				installedVersion = pkgWithVersion
			}
		}
		if len(parts) >= 2 {
			storePath = parts[1]
		}
	}

	binPath := fmt.Sprintf("%s/bin", profilePath)

	return workspacesv1.InstalledPackage{
		Name:        pkg.Name,
		Version:     installedVersion,
		BinPath:     binPath,
		StorePath:   storePath,
		InstalledAt: metav1.Now(),
	}, nil
}

func (r *PackageManagerReconciler) uninstallPackage(pkgName string, profileName string) error {
	// Remove package from the profile (doesn't delete from Nix store)
	// Using nix-env -e only removes the package from the user environment
	profilePath := fmt.Sprintf("%s/profiles/per-user/root/%s", nixStorePath, profileName)
	uninstallScript := fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -e %s", profilePath, pkgName)

	output, err := r.CmdExec.Execute(uninstallScript)
	if err != nil {
		return fmt.Errorf("nix-env uninstall failed: %w, output: %s", err, string(output))
	}

	return nil
}

func (r *PackageManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&packagesv1.PackageRequest{}).
		WithEventFilter(predicate.Funcs{
			// Reconcile on Create and Update (spec changes)
			// The Reconcile function itself will check if reconciliation is needed
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldPR, okOld := e.ObjectOld.(*packagesv1.PackageRequest)
				newPR, okNew := e.ObjectNew.(*packagesv1.PackageRequest)
				if !okOld || !okNew {
					return false
				}

				// Only reconcile if spec changed (not just status)
				// This prevents infinite loops from status-only updates
				return oldPR.Generation != newPR.Generation
			},
			// Don't reconcile on delete
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
		}).
		Complete(r)
}

func setupWorkspaceHome(logger *zap2.Logger, fs FileSystem) error {
	logger.Info("Setting up workspace home directory",
		zap2.String("path", workspaceHomePath),
		zap2.Int("uid", workspaceUserUID),
		zap2.Int("gid", workspaceUserGID))

	// Create directory if it doesn't exist
	if err := fs.MkdirAll(workspaceHomePath, 0755); err != nil {
		return fmt.Errorf("failed to create workspace home directory: %w", err)
	}

	// Set ownership to workspace user
	if err := fs.Chown(workspaceHomePath, workspaceUserUID, workspaceUserGID); err != nil {
		return fmt.Errorf("failed to set ownership on workspace home directory: %w", err)
	}

	// Create workspaces subdirectory with correct ownership
	workspacesPath := workspaceHomePath + "/workspaces"
	if err := fs.MkdirAll(workspacesPath, 0755); err != nil {
		return fmt.Errorf("failed to create workspaces subdirectory: %w", err)
	}

	// Set ownership to workspace user
	if err := fs.Chown(workspacesPath, workspaceUserUID, workspaceUserGID); err != nil {
		return fmt.Errorf("failed to set ownership on workspaces subdirectory: %w", err)
	}

	logger.Info("Successfully set up workspace home directory with workspaces subdirectory",
		zap2.String("path", workspaceHomePath),
		zap2.String("workspacesPath", workspacesPath))

	return nil
}

func setupSSHConfigDirectory(logger *zap2.Logger, fs FileSystem) error {
	logger.Info("Setting up SSH config directory", zap2.String("path", sshConfigPath))

	// Create directory if it doesn't exist
	if err := fs.MkdirAll(sshConfigPath, 0755); err != nil {
		return fmt.Errorf("failed to create SSH config directory: %w", err)
	}

	logger.Info("Successfully set up SSH config directory", zap2.String("path", sshConfigPath))
	return nil
}

func writeAuthorizedKeys(logger *zap2.Logger, content string, fs FileSystem) error {
	targetPath := filepath.Join(sshConfigPath, authorizedKeysFile)
	tempPath := targetPath + ".tmp"

	// Write to temporary file first (atomic operation)
	if err := fs.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temporary authorized_keys file: %w", err)
	}

	// Atomically rename temp file to target (atomic on POSIX systems)
	if err := fs.Rename(tempPath, targetPath); err != nil {
		return fmt.Errorf("failed to rename temporary authorized_keys file: %w", err)
	}

	logger.Info("Successfully wrote authorized_keys file",
		zap2.String("path", targetPath),
		zap2.Int("size", len(content)))
	return nil
}

// SSHConfigReconciler watches the ssh-authorized-keys ConfigMap and writes it to the host filesystem
type SSHConfigReconciler struct {
	client.Client
	Logger *zap2.Logger
	FS     FileSystem
}

func (r *SSHConfigReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	// Only reconcile the ssh-authorized-keys ConfigMap
	if req.Name != "ssh-authorized-keys" {
		return reconcile.Result{}, nil
	}

	logger := r.Logger.With(
		zap2.String("configMap", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling SSH authorized_keys ConfigMap")

	// Fetch ConfigMap
	cm := &corev1.ConfigMap{}
	if err := r.Get(ctx, req.NamespacedName, cm); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("ConfigMap deleted or not found")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get ConfigMap", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Get authorized_keys content
	authorizedKeys, ok := cm.Data["authorized_keys"]
	if !ok {
		logger.Warn("ConfigMap does not contain authorized_keys key")
		return reconcile.Result{}, nil
	}

	// Write to host filesystem
	if err := writeAuthorizedKeys(logger, authorizedKeys, r.FS); err != nil {
		logger.Error("Failed to write authorized_keys", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Successfully updated authorized_keys file")
	return reconcile.Result{}, nil
}

func (r *SSHConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return e.Object.GetName() == "ssh-authorized-keys"
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectNew.GetName() == "ssh-authorized-keys"
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return e.Object.GetName() == "ssh-authorized-keys"
			},
		}).
		Complete(r)
}

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

	zapLogger.Info("Starting Package Manager",
		zap2.String("namespace", namespace),
		zap2.String("nixStorePath", nixStorePath))

	// Setup scheme
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = packagesv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)

	// Get in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		zapLogger.Fatal("Failed to get in-cluster config", zap2.Error(err))
	}

	// Create manager watching only the specified namespace
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:         scheme,
		LeaderElection: false, // Each Deployment manages only its namespace
		Metrics: server.Options{
			BindAddress: ":8080",
		},
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				namespace: {},
			},
		},
	})
	if err != nil {
		zapLogger.Fatal("Failed to create manager", zap2.Error(err))
	}

	// Setup package reconciler
	packageReconciler := &PackageManagerReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Logger:    zapLogger,
		Namespace: namespace,
		CmdExec:   &RealCommandExecutor{},
	}

	if err := packageReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup package controller", zap2.Error(err))
	}

	// Setup SSH config reconciler
	sshConfigReconciler := &SSHConfigReconciler{
		Client: mgr.GetClient(),
		Logger: zapLogger,
		FS:     fs,
	}

	if err := sshConfigReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup SSH config controller", zap2.Error(err))
	}

	ctx := ctrl.SetupSignalHandler()
	zapLogger.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		zapLogger.Fatal("Failed to start manager", zap2.Error(err))
	}
}
