package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
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
	nixStorePath              = "/nix"
	workspaceHomePath         = "/var/lib/kloudlite/workspace-homes/kl"
	workspaceUserUID          = 1001
	workspaceUserGID          = 1001
	sshConfigPath             = "/var/lib/kloudlite/ssh-config"
	authorizedKeysFile        = "authorized_keys"
	packageRequestFinalizer   = "workspaces.kloudlite.io/package-cleanup"
	workspaceCleanupFinalizer = "workspaces.kloudlite.io/directory-cleanup"
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
	// Remove target file if it exists to ensure rename succeeds
	// On some systems/mounts, os.Rename may fail if target exists
	// This maintains reasonable atomicity since remove + rename is very fast
	_ = os.Remove(newpath)
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
	pkgReq := &workspacev1.PackageRequest{}
	if err := r.Get(ctx, req.NamespacedName, pkgReq); err != nil {
		logger.Error("Failed to get PackageRequest", zap2.Error(err))
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the PackageRequest is being deleted
	if pkgReq.DeletionTimestamp != nil {
		// Object is being deleted
		if containsString(pkgReq.Finalizers, packageRequestFinalizer) {
			// Our finalizer is present, perform cleanup
			logger.Info("PackageRequest is being deleted, cleaning up packages", zap2.String("profile", pkgReq.Spec.ProfileName))

			// Get all installed packages from the profile and remove them
			installedPkgs := r.getInstalledPackagesFromProfile(pkgReq.Spec.ProfileName, logger)
			for _, pkgName := range installedPkgs {
				logger.Info("Removing package from profile",
					zap2.String("package", pkgName),
					zap2.String("profile", pkgReq.Spec.ProfileName))

				if err := r.uninstallPackage(pkgName, pkgReq.Spec.ProfileName); err != nil {
					logger.Error("Failed to remove package during cleanup",
						zap2.String("package", pkgName),
						zap2.String("profile", pkgReq.Spec.ProfileName),
						zap2.Error(err))
					// Continue with other packages even if one fails
				} else {
					logger.Info("Successfully removed package",
						zap2.String("package", pkgName),
						zap2.String("profile", pkgReq.Spec.ProfileName))
				}
			}

			// Remove our finalizer
			pkgReq.Finalizers = removeString(pkgReq.Finalizers, packageRequestFinalizer)
			if err := r.Update(ctx, pkgReq); err != nil {
				logger.Error("Failed to remove finalizer", zap2.Error(err))
				return reconcile.Result{}, err
			}

			logger.Info("Cleanup complete, finalizer removed")
		}
		// Stop reconciliation as the object is being deleted
		return reconcile.Result{}, nil
	}

	// Add finalizer if it doesn't exist
	if !containsString(pkgReq.Finalizers, packageRequestFinalizer) {
		logger.Info("Adding finalizer to PackageRequest")
		pkgReq.Finalizers = append(pkgReq.Finalizers, packageRequestFinalizer)
		if err := r.Update(ctx, pkgReq); err != nil {
			logger.Error("Failed to add finalizer", zap2.Error(err))
			return reconcile.Result{}, err
		}
		// Requeue to continue reconciliation with finalizer in place
		return reconcile.Result{Requeue: true}, nil
	}

	logger.Info("Using profile", zap2.String("profile", pkgReq.Spec.ProfileName))

	// IMPORTANT: Reconcile based on ACTUAL state, not status
	// 1. Ensure profile directory exists (idempotent)
	// 2. Check what packages are ACTUALLY installed (check filesystem)
	// 3. Install missing packages, remove unwanted packages
	// 4. Update status to reflect actual state

	profilePath := fmt.Sprintf("%s/profiles/per-user/root/%s", nixStorePath, pkgReq.Spec.ProfileName)
	profileDir := fmt.Sprintf("%s/profiles/per-user/root", nixStorePath)

	// Ensure profile directory exists (idempotent)
	logger.Info("Ensuring profile directory exists", zap2.String("profileDir", profileDir))
	mkdirScript := fmt.Sprintf("mkdir -p %s", profileDir)
	if output, err := r.CmdExec.Execute(mkdirScript); err != nil {
		logger.Error("Failed to create profile directory",
			zap2.String("profileDir", profileDir),
			zap2.Error(err),
			zap2.String("output", string(output)))
		// Don't fail reconciliation, continue and let package installation fail if needed
	}

	// Build map of desired packages
	desiredPackages := make(map[string]bool)
	for _, pkg := range pkgReq.Spec.Packages {
		desiredPackages[pkg.Name] = true
	}

	// Remove packages that were previously installed but are no longer in spec
	// We use status.InstalledPackages as a hint for what might need removal
	// but we verify actual state before removing
	for _, prevInstalled := range pkgReq.Status.InstalledPackages {
		if !desiredPackages[prevInstalled.Name] {
			// Package was in previous spec but not in current spec
			// Check if it's actually installed before trying to remove
			if r.isPackageInstalled(prevInstalled.Name, pkgReq.Spec.ProfileName, logger) {
				logger.Info("Removing unwanted package",
					zap2.String("package", prevInstalled.Name),
					zap2.String("profile", pkgReq.Spec.ProfileName))

				if err := r.uninstallPackage(prevInstalled.Name, pkgReq.Spec.ProfileName); err != nil {
					logger.Error("Failed to remove package",
						zap2.String("package", prevInstalled.Name),
						zap2.String("profile", pkgReq.Spec.ProfileName),
						zap2.Error(err))
				} else {
					logger.Info("Successfully removed package",
						zap2.String("package", prevInstalled.Name),
						zap2.String("profile", pkgReq.Spec.ProfileName))
				}
			}
		}
	}

	// Set phase to Installing and clear old failed packages before starting installation
	// This ensures old failures are cleared when we retry
	if err := r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *workspacev1.PackageRequest) {
		latest.Status.Phase = "Installing"
		latest.Status.FailedPackages = []string{} // Clear old failed packages
		latest.Status.Message = "Installing packages..."
		latest.Status.LastUpdated = metav1.Now()
	}, logger); err != nil {
		logger.Warn("Failed to update status to Installing phase", zap2.Error(err))
		// Continue with installation even if status update fails
	}

	// Install missing packages and record installed ones
	installedPackages := []workspacev1.InstalledPackage{}
	failedPackages := []string{}

	for _, pkg := range pkgReq.Spec.Packages {
		// Check actual state: is package really installed on the filesystem?
		if r.isPackageInstalled(pkg.Name, pkgReq.Spec.ProfileName, logger) {
			logger.Info("Package already installed, querying info",
				zap2.String("package", pkg.Name),
				zap2.String("profile", pkgReq.Spec.ProfileName))

			// Query existing package info
			queryScript := fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -q --out-path %s", profilePath, pkg.Name)
			queryOutput, err := r.CmdExec.Execute(queryScript)

			storePath := nixStorePath + "/store"
			installedVersion := ""

			if err == nil && len(queryOutput) > 0 {
				parts := strings.Fields(string(queryOutput))
				if len(parts) >= 1 {
					pkgWithVersion := parts[0]
					if strings.HasPrefix(pkgWithVersion, pkg.Name+"-") {
						installedVersion = strings.TrimPrefix(pkgWithVersion, pkg.Name+"-")
					} else {
						installedVersion = pkgWithVersion
					}
				}
				if len(parts) >= 2 {
					storePath = parts[1]
				}
			}

			workspaceBinPath := fmt.Sprintf("/nix/profiles/per-user/root/%s/bin", pkgReq.Spec.ProfileName)
			installedPackages = append(installedPackages, workspacev1.InstalledPackage{
				Name:        pkg.Name,
				Version:     installedVersion,
				BinPath:     workspaceBinPath,
				StorePath:   storePath,
				InstalledAt: metav1.Now(),
			})
			continue
		}

		// Package not installed in actual filesystem, install it
		logger.Info("Installing missing package",
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

	// Update status to reflect actual state
	if err := r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *workspacev1.PackageRequest) {
		latest.Status.InstalledPackages = installedPackages
		latest.Status.FailedPackages = failedPackages
		latest.Status.LastUpdated = metav1.Now()
		if len(failedPackages) > 0 {
			latest.Status.Phase = "Failed"
			latest.Status.Message = fmt.Sprintf("Failed to install %d packages", len(failedPackages))
		} else {
			latest.Status.Phase = "Ready"
			latest.Status.Message = fmt.Sprintf("Successfully reconciled %d packages", len(installedPackages))
		}
	}, logger); err != nil {
		logger.Error("Failed to update status after retries", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("PackageRequest reconciliation complete",
		zap2.Int("installed", len(installedPackages)),
		zap2.Int("failed", len(failedPackages)))

	return reconcile.Result{}, nil
}

// isPackageInstalled checks if a specific package is installed in the profile
// Returns true if installed, false otherwise
func (r *PackageManagerReconciler) isPackageInstalled(packageName string, profileName string, logger *zap2.Logger) bool {
	profilePath := fmt.Sprintf("%s/profiles/per-user/root/%s", nixStorePath, profileName)

	// Check if profile exists
	checkScript := fmt.Sprintf("test -d %s", profilePath)
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		return false
	}

	// Query specific package in the profile
	// nix-env -q returns exit code 0 and outputs the package if installed, or exits with 0 but empty output if not
	queryScript := fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -q %s", profilePath, packageName)
	output, err := r.CmdExec.Execute(queryScript)

	// If query succeeds and returns non-empty output, package is installed
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		logger.Debug("Package is installed",
			zap2.String("package", packageName),
			zap2.String("profile", profileName),
			zap2.String("output", string(output)))
		return true
	}

	return false
}

// getInstalledPackagesFromProfile lists all packages installed in the profile
// Used to detect packages that need to be removed
func (r *PackageManagerReconciler) getInstalledPackagesFromProfile(profileName string, logger *zap2.Logger) []string {
	profilePath := fmt.Sprintf("%s/profiles/per-user/root/%s", nixStorePath, profileName)

	// Check if profile exists
	checkScript := fmt.Sprintf("test -d %s", profilePath)
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		logger.Info("Profile does not exist yet", zap2.String("profile", profileName))
		return []string{}
	}

	// Query all installed packages in the profile
	// nix-env -q lists all packages with their full names (package-version)
	queryScript := fmt.Sprintf(". /root/.nix-profile/etc/profile.d/nix.sh && nix-env -p %s -q", profilePath)
	output, err := r.CmdExec.Execute(queryScript)
	if err != nil {
		logger.Warn("Failed to query installed packages from profile",
			zap2.String("profile", profileName),
			zap2.Error(err),
			zap2.String("output", string(output)))
		return []string{}
	}

	// Parse output - each line is "package-version" (e.g., "nodejs-24.5.0")
	// We return the full package name as-is for use with nix-env -e
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	packages := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			packages = append(packages, line)
		}
	}

	return packages
}

// updateStatusWithRetry retries status updates with optimistic concurrency control
// It fetches the latest version before each update attempt
func (r *PackageManagerReconciler) updateStatusWithRetry(
	ctx context.Context,
	namespacedName client.ObjectKey,
	updateFn func(*workspacev1.PackageRequest),
	logger *zap2.Logger,
) error {
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		// Fetch the latest version
		latest := &workspacev1.PackageRequest{}
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

func (r *PackageManagerReconciler) installPackage(pkg workspacev1.PackageSpec, profileName string) (workspacev1.InstalledPackage, error) {
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
		return workspacev1.InstalledPackage{}, fmt.Errorf("nix-env failed: %w, output: %s", err, string(output))
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

	// BinPath should use the shared mount path at /nix
	// Both workmachine-host-manager and workspace pods mount the hostPath at /nix
	// This ensures packages installed by workmachine-host-manager are accessible in workspaces
	workspaceBinPath := fmt.Sprintf("/nix/profiles/per-user/root/%s/bin", profileName)

	return workspacev1.InstalledPackage{
		Name:        pkg.Name,
		Version:     installedVersion,
		BinPath:     workspaceBinPath,
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

// containsString checks if a string is present in a slice
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// removeString removes a string from a slice
func removeString(slice []string, s string) []string {
	result := []string{}
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

func (r *PackageManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workspacev1.PackageRequest{}).
		WithEventFilter(predicate.Funcs{
			// Reconcile on Create and Update (spec changes)
			// The Reconcile function itself will check if reconciliation is needed
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldPR, okOld := e.ObjectOld.(*workspacev1.PackageRequest)
				newPR, okNew := e.ObjectNew.(*workspacev1.PackageRequest)
				if !okOld || !okNew {
					return false
				}

				// Only reconcile if spec changed (not just status)
				// This prevents infinite loops from status-only updates
				return oldPR.Generation != newPR.Generation
			},
			// Reconcile on delete to clean up packages
			DeleteFunc: func(e event.DeleteEvent) bool {
				return true
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
	if err := fs.MkdirAll(workspaceHomePath, 0o755); err != nil {
		return fmt.Errorf("failed to create workspace home directory: %w", err)
	}

	// Set ownership to workspace user
	if err := fs.Chown(workspaceHomePath, workspaceUserUID, workspaceUserGID); err != nil {
		return fmt.Errorf("failed to set ownership on workspace home directory: %w", err)
	}

	// Create workspaces subdirectory with correct ownership
	workspacesPath := workspaceHomePath + "/workspaces"
	if err := fs.MkdirAll(workspacesPath, 0o755); err != nil {
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
	if err := fs.MkdirAll(sshConfigPath, 0o755); err != nil {
		return fmt.Errorf("failed to create SSH config directory: %w", err)
	}

	logger.Info("Successfully set up SSH config directory", zap2.String("path", sshConfigPath))
	return nil
}

func writeAuthorizedKeys(logger *zap2.Logger, content string, fs FileSystem) error {
	targetPath := filepath.Join(sshConfigPath, authorizedKeysFile)
	tempPath := targetPath + ".tmp"

	// Write to temporary file first (atomic operation)
	if err := fs.WriteFile(tempPath, []byte(content), 0o644); err != nil {
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

// SSHConfigReconciler watches the ssh-host-keys Secret and writes authorized_keys to the host filesystem
type SSHConfigReconciler struct {
	client.Client
	Logger *zap2.Logger
	FS     FileSystem
}

func (r *SSHConfigReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("secret", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling SSH config from Secret")

	// Fetch Secret
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Secret deleted or not found")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Secret", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Write authorized_keys
	if authorizedKeysBytes, ok := secret.Data["authorized_keys"]; ok {
		if err := writeAuthorizedKeys(logger, string(authorizedKeysBytes), r.FS); err != nil {
			logger.Error("Failed to write authorized_keys", zap2.Error(err))
			return reconcile.Result{}, err
		}
	}

	// Write SSH host keys
	if rsaKeyBytes, ok := secret.Data["ssh_host_rsa_key"]; ok {
		targetPath := filepath.Join(sshConfigPath, "ssh_host_rsa_key")
		tempPath := targetPath + ".tmp"

		// Write to temp file
		if err := r.FS.WriteFile(tempPath, rsaKeyBytes, 0o600); err != nil {
			logger.Error("Failed to write ssh_host_rsa_key temp file", zap2.Error(err))
			return reconcile.Result{}, err
		}

		// Atomic rename
		if err := r.FS.Rename(tempPath, targetPath); err != nil {
			logger.Error("Failed to rename ssh_host_rsa_key", zap2.Error(err))
			return reconcile.Result{}, err
		}
	}

	if rsaPubKeyBytes, ok := secret.Data["ssh_host_rsa_key.pub"]; ok {
		targetPath := filepath.Join(sshConfigPath, "ssh_host_rsa_key.pub")
		tempPath := targetPath + ".tmp"

		// Write to temp file
		if err := r.FS.WriteFile(tempPath, rsaPubKeyBytes, 0o644); err != nil {
			logger.Error("Failed to write ssh_host_rsa_key.pub temp file", zap2.Error(err))
			return reconcile.Result{}, err
		}

		// Atomic rename
		if err := r.FS.Rename(tempPath, targetPath); err != nil {
			logger.Error("Failed to rename ssh_host_rsa_key.pub", zap2.Error(err))
			return reconcile.Result{}, err
		}
	}

	logger.Info("Successfully updated SSH config files")
	return reconcile.Result{}, nil
}

func (r *SSHConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				labels := e.Object.GetLabels()
				return labels != nil && labels["kloudlite.io/ssh-host-keys"] == "true"
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				labels := e.ObjectNew.GetLabels()
				return labels != nil && labels["kloudlite.io/ssh-host-keys"] == "true"
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				labels := e.Object.GetLabels()
				return labels != nil && labels["kloudlite.io/ssh-host-keys"] == "true"
			},
		}).
		Complete(r)
}

// WorkspaceCleanupReconciler watches Workspace resources and cleans up workspace directories
type WorkspaceCleanupReconciler struct {
	client.Client
	Logger *zap2.Logger
	FS     FileSystem
}

func (r *WorkspaceCleanupReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("workspace", req.Name),
	)

	logger.Info("Reconciling Workspace for directory cleanup")

	// Fetch Workspace (cluster-scoped, no namespace)
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name}, workspace); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logger.Info("Workspace deleted or not found")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Workspace", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Check if workspace is being deleted
	if workspace.DeletionTimestamp != nil {
		// Workspace is being deleted
		if containsString(workspace.Finalizers, workspaceCleanupFinalizer) {
			logger.Info("Workspace is being deleted, cleaning up directory")

			// Clean up workspace directory
			workspaceDir := fmt.Sprintf("%s/workspaces/%s", workspaceHomePath, workspace.Name)
			logger.Info("Removing workspace directory", zap2.String("path", workspaceDir))

			// Use rm -rf to remove directory and all contents
			removeScript := fmt.Sprintf("rm -rf %s", workspaceDir)
			cmd := exec.Command("sh", "-c", removeScript)
			output, err := cmd.CombinedOutput()
			if err != nil {
				logger.Error("Failed to remove workspace directory",
					zap2.String("path", workspaceDir),
					zap2.Error(err),
					zap2.String("output", string(output)))
				return reconcile.Result{}, fmt.Errorf("failed to remove workspace directory: %w", err)
			}

			logger.Info("Successfully removed workspace directory", zap2.String("path", workspaceDir))

			// Remove finalizer
			workspace.Finalizers = removeString(workspace.Finalizers, workspaceCleanupFinalizer)
			if err := r.Update(ctx, workspace); err != nil {
				logger.Error("Failed to remove finalizer", zap2.Error(err))
				return reconcile.Result{}, err
			}

			logger.Info("Cleanup complete, finalizer removed")
		}
		return reconcile.Result{}, nil
	}

	// Add finalizer if not present
	if !containsString(workspace.Finalizers, workspaceCleanupFinalizer) {
		logger.Info("Adding cleanup finalizer to Workspace")
		workspace.Finalizers = append(workspace.Finalizers, workspaceCleanupFinalizer)
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to add finalizer", zap2.Error(err))
			return reconcile.Result{}, err
		}
		logger.Info("Successfully added cleanup finalizer")
	}

	return reconcile.Result{}, nil
}

func (r *WorkspaceCleanupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workspacev1.Workspace{}).
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
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		zapLogger.Fatal("Failed to add client-go scheme", zap2.Error(err))
	}
	if err := workspacev1.AddToScheme(scheme); err != nil {
		zapLogger.Fatal("Failed to add workspace v1 scheme", zap2.Error(err))
	}

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

	// Setup workspace cleanup reconciler
	workspaceCleanupReconciler := &WorkspaceCleanupReconciler{
		Client: mgr.GetClient(),
		Logger: zapLogger,
		FS:     fs,
	}

	if err := workspaceCleanupReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup workspace cleanup controller", zap2.Error(err))
	}

	ctx := ctrl.SetupSignalHandler()
	zapLogger.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		zapLogger.Fatal("Failed to start manager", zap2.Error(err))
	}
}
