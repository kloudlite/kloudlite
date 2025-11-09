package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	zap2 "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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

// HostCommandExecutor implements CommandExecutor using nsenter to run commands on the host
// This is necessary for GPU detection and driver installation which must happen on the host system
type HostCommandExecutor struct{}

func (r *HostCommandExecutor) Execute(script string) ([]byte, error) {
	// Use nsenter to execute commands on the host by entering the mount namespace of PID 1
	// -t 1: target PID 1 (the host's init process)
	// -m: enter mount namespace
	// -u: enter UTS namespace
	// -i: enter IPC namespace
	// Set PATH explicitly to ensure standard binaries like nvidia-smi are found
	cmd := exec.Command("nsenter", "-t", "1", "-m", "-u", "-i", "bash", "-c", script)
	cmd.Env = append(os.Environ(), "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
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
	pkgReq := &packagesv1.PackageRequest{}
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
	if err := r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *packagesv1.PackageRequest) {
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
	if err := r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *packagesv1.PackageRequest) {
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
	Logger          *zap2.Logger
	FS              FileSystem
	WorkMachineName string
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
				return labels != nil &&
					labels["kloudlite.io/ssh-host-keys"] == "true" &&
					labels["kloudlite.io/workmachine"] == r.WorkMachineName
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				labels := e.ObjectNew.GetLabels()
				return labels != nil &&
					labels["kloudlite.io/ssh-host-keys"] == "true" &&
					labels["kloudlite.io/workmachine"] == r.WorkMachineName
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				labels := e.Object.GetLabels()
				return labels != nil &&
					labels["kloudlite.io/ssh-host-keys"] == "true" &&
					labels["kloudlite.io/workmachine"] == r.WorkMachineName
			},
		}).
		Complete(r)
}

// sanitizeLabelValue converts a string to a valid Kubernetes label value
// Kubernetes labels must match regex: (([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?
// and be at most 63 characters long
func sanitizeLabelValue(s string, maxLen int) string {
	if maxLen > 63 {
		maxLen = 63
	}
	if maxLen < 1 {
		maxLen = 1
	}

	// Replace spaces and invalid characters with hyphens
	sanitized := strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '-'
	}, s)

	// Truncate to max length
	if len(sanitized) > maxLen {
		sanitized = sanitized[:maxLen]
	}

	// Ensure it starts and ends with alphanumeric
	// Trim leading/trailing non-alphanumeric chars
	sanitized = strings.TrimFunc(sanitized, func(r rune) bool {
		return !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})

	// If empty after sanitization, return a default
	if sanitized == "" {
		return "unknown"
	}

	return sanitized
}

// GPUStatusReconciler monitors GPU hardware and updates node labels and resources
type GPUStatusReconciler struct {
	client.Client
	Logger          *zap2.Logger
	CmdExec         CommandExecutor
	NodeName        string
	LastGPUDetected bool
}

func (r *GPUStatusReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("node", req.Name),
	)

	// Only reconcile our own node
	if req.Name != r.NodeName {
		return reconcile.Result{}, nil
	}

	logger.Info("Reconciling GPU status for node")

	// Fetch the node
	node := &corev1.Node{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		logger.Error("Failed to get Node", zap2.Error(err))
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Detect GPU hardware
	gpuDetected := r.detectGPU(logger)

	// If GPU status changed, update node
	if gpuDetected != r.LastGPUDetected {
		logger.Info("GPU detection status changed",
			zap2.Bool("previouslyDetected", r.LastGPUDetected),
			zap2.Bool("currentlyDetected", gpuDetected))
		r.LastGPUDetected = gpuDetected
	}

	if !gpuDetected {
		logger.Info("No GPU detected on this node")
		return reconcile.Result{}, nil
	}

	// Ensure NVIDIA drivers are available and container runtime is configured
	setupErr := r.ensureNVIDIASetup(logger)
	if setupErr != nil {
		logger.Error("NVIDIA setup not ready", zap2.Error(setupErr))

		// Retry updating node with latest version on conflict
		for retries := 0; retries < 3; retries++ {
			// Refetch the latest node
			latestNode := &corev1.Node{}
			if err := r.Get(ctx, client.ObjectKey{Name: node.Name}, latestNode); err != nil {
				logger.Error("Failed to refetch node", zap2.Error(err))
				break
			}

			// Update labels with setup status
			if latestNode.Labels == nil {
				latestNode.Labels = make(map[string]string)
			}
			latestNode.Labels["nvidia.com/gpu.driver-status"] = "waiting"
			latestNode.Labels["nvidia.com/gpu.driver-message"] = sanitizeLabelValue(setupErr.Error(), 63)

			// Try to update node with status
			if updateErr := r.Update(ctx, latestNode); updateErr != nil {
				if strings.Contains(updateErr.Error(), "the object has been modified") {
					logger.Warn("Node was modified, retrying update", zap2.Int("retry", retries+1))
					time.Sleep(100 * time.Millisecond)
					continue
				}
				logger.Error("Failed to update node with driver status", zap2.Error(updateErr))
				break
			}

			logger.Info("Successfully updated node with driver status")
			break
		}

		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Get GPU information
	gpuInfo, err := r.getGPUInfo(logger)
	if err != nil {
		logger.Error("Failed to get GPU information", zap2.Error(err))

		// Retry updating node with latest version on conflict
		for retries := 0; retries < 3; retries++ {
			// Refetch the latest node
			latestNode := &corev1.Node{}
			if getErr := r.Get(ctx, client.ObjectKey{Name: node.Name}, latestNode); getErr != nil {
				logger.Error("Failed to refetch node", zap2.Error(getErr))
				break
			}

			// Update labels with status
			if latestNode.Labels == nil {
				latestNode.Labels = make(map[string]string)
			}
			latestNode.Labels["nvidia.com/gpu.driver-status"] = "error"
			latestNode.Labels["nvidia.com/gpu.driver-message"] = sanitizeLabelValue(err.Error(), 63)

			if updateErr := r.Update(ctx, latestNode); updateErr != nil {
				if strings.Contains(updateErr.Error(), "the object has been modified") {
					logger.Warn("Node was modified, retrying update", zap2.Int("retry", retries+1))
					time.Sleep(100 * time.Millisecond)
					continue
				}
				logger.Error("Failed to update node with GPU error status", zap2.Error(updateErr))
				break
			}

			logger.Info("Successfully updated node with GPU error status")
			break
		}

		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	logger.Info("GPU detected",
		zap2.Int("count", gpuInfo.Count),
		zap2.String("product", gpuInfo.Product),
		zap2.String("driverVersion", gpuInfo.DriverVersion))

	// Update node labels and resources
	if err := r.updateNodeGPU(ctx, node, gpuInfo, logger); err != nil {
		logger.Error("Failed to update node GPU status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Successfully updated node with GPU information")
	return reconcile.Result{}, nil
}

type GPUInfo struct {
	Count         int
	Product       string
	DriverVersion string
}

type GPUMetrics struct {
	Model             string
	DriverVersion     string
	Count             int
	MemoryTotal       int32
	MemoryUsed        int32
	MemoryFree        int32
	UtilizationGPU    int32
	UtilizationMemory int32
	Temperature       int32
	PowerDraw         float32
	PowerLimit        float32
}

// ensureNVIDIASetup checks if NVIDIA drivers are available
// Note: The Deep Learning AMI comes with drivers and container runtime pre-installed
func (r *GPUStatusReconciler) ensureNVIDIASetup(logger *zap2.Logger) error {
	// Check if nvidia-smi is available (should be pre-installed in Deep Learning AMI)
	// Use full path since nsenter may not inherit full PATH from the host
	checkScript := "/usr/bin/nvidia-smi > /dev/null 2>&1"
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		logger.Info("NVIDIA drivers not available (nvidia-smi failed)")
		return fmt.Errorf("nvidia-smi not available")
	}

	logger.Info("✓ NVIDIA drivers are available and working")
	return nil
}

func (r *GPUStatusReconciler) detectGPU(logger *zap2.Logger) bool {
	// Check for NVIDIA GPU by reading /sys/bus/pci/devices directly
	// This approach doesn't require lspci to be installed
	// Note: When using nsenter to enter host namespaces, paths like /sys are already host paths
	checkScript := `
		if [ -d /sys/bus/pci/devices ]; then
			for device in /sys/bus/pci/devices/*; do
				if [ -f "$device/vendor" ] && [ -f "$device/device" ]; then
					vendor=$(cat "$device/vendor" 2>/dev/null)
					# 0x10de is NVIDIA's PCI vendor ID
					if [ "$vendor" = "0x10de" ]; then
						exit 0
					fi
				fi
			done
		fi
		exit 1
	`
	_, err := r.CmdExec.Execute(checkScript)
	if err != nil {
		logger.Debug("No NVIDIA GPU detected in /sys/bus/pci/devices")
		return false
	}
	logger.Info("NVIDIA GPU detected via PCI device scan")
	return true
}

func (r *GPUStatusReconciler) getGPUInfo(logger *zap2.Logger) (*GPUInfo, error) {
	// Check if nvidia-smi is available
	checkScript := "nvidia-smi > /dev/null 2>&1"
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		return nil, fmt.Errorf("nvidia-smi not available or drivers not loaded")
	}

	// Get GPU count
	countScript := "nvidia-smi --query-gpu=name --format=csv,noheader | wc -l"
	countOutput, err := r.CmdExec.Execute(countScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU count: %w", err)
	}

	count := 1 // Default to 1
	if parsed, err := fmt.Sscanf(strings.TrimSpace(string(countOutput)), "%d", &count); err == nil && parsed == 1 {
		// Successfully parsed count
	}

	// Get GPU product name (normalized)
	productScript := "nvidia-smi --query-gpu=gpu_name --format=csv,noheader | head -1 | tr ' ' '-' | tr '[:upper:]' '[:lower:]'"
	productOutput, err := r.CmdExec.Execute(productScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU product: %w", err)
	}
	product := strings.TrimSpace(string(productOutput))

	// Get driver version
	driverScript := "nvidia-smi --query-gpu=driver_version --format=csv,noheader | head -1"
	driverOutput, err := r.CmdExec.Execute(driverScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver version: %w", err)
	}
	driverVersion := strings.TrimSpace(string(driverOutput))

	return &GPUInfo{
		Count:         count,
		Product:       product,
		DriverVersion: driverVersion,
	}, nil
}

func (r *GPUStatusReconciler) getGPUMetrics(logger *zap2.Logger) (*GPUMetrics, error) {
	// Query nvidia-smi for comprehensive metrics
	// Fields: name, driver_version, memory.total, memory.used, memory.free, utilization.gpu, utilization.memory, temperature.gpu, power.draw, power.limit
	metricsScript := "nvidia-smi --query-gpu=name,driver_version,memory.total,memory.used,memory.free,utilization.gpu,utilization.memory,temperature.gpu,power.draw,power.limit --format=csv,noheader,nounits | head -1"

	output, err := r.CmdExec.Execute(metricsScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU metrics: %w", err)
	}

	// Parse CSV output
	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) < 10 {
		return nil, fmt.Errorf("unexpected nvidia-smi output format: got %d fields, expected 10", len(parts))
	}

	// Helper function to parse int32
	parseInt32 := func(s string, fieldName string) (int32, error) {
		s = strings.TrimSpace(s)
		var val int32
		if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
			return 0, fmt.Errorf("failed to parse %s: %w", fieldName, err)
		}
		return val, nil
	}

	// Helper function to parse float32
	parseFloat32 := func(s string, fieldName string) (float32, error) {
		s = strings.TrimSpace(s)
		var val float32
		if _, err := fmt.Sscanf(s, "%f", &val); err != nil {
			return 0, fmt.Errorf("failed to parse %s: %w", fieldName, err)
		}
		return val, nil
	}

	memoryTotal, err := parseInt32(parts[2], "memory.total")
	if err != nil {
		return nil, err
	}

	memoryUsed, err := parseInt32(parts[3], "memory.used")
	if err != nil {
		return nil, err
	}

	memoryFree, err := parseInt32(parts[4], "memory.free")
	if err != nil {
		return nil, err
	}

	utilizationGPU, err := parseInt32(parts[5], "utilization.gpu")
	if err != nil {
		return nil, err
	}

	utilizationMemory, err := parseInt32(parts[6], "utilization.memory")
	if err != nil {
		return nil, err
	}

	temperature, err := parseInt32(parts[7], "temperature.gpu")
	if err != nil {
		return nil, err
	}

	powerDraw, err := parseFloat32(parts[8], "power.draw")
	if err != nil {
		return nil, err
	}

	powerLimit, err := parseFloat32(parts[9], "power.limit")
	if err != nil {
		return nil, err
	}

	// Get GPU count
	countScript := "nvidia-smi --query-gpu=name --format=csv,noheader | wc -l"
	countOutput, err := r.CmdExec.Execute(countScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU count: %w", err)
	}

	count := 1 // Default to 1
	if parsed, err := fmt.Sscanf(strings.TrimSpace(string(countOutput)), "%d", &count); err == nil && parsed == 1 {
		// Successfully parsed count
	}

	return &GPUMetrics{
		Model:             strings.TrimSpace(parts[0]),
		DriverVersion:     strings.TrimSpace(parts[1]),
		Count:             count,
		MemoryTotal:       memoryTotal,
		MemoryUsed:        memoryUsed,
		MemoryFree:        memoryFree,
		UtilizationGPU:    utilizationGPU,
		UtilizationMemory: utilizationMemory,
		Temperature:       temperature,
		PowerDraw:         powerDraw,
		PowerLimit:        powerLimit,
	}, nil
}

func (r *GPUStatusReconciler) updateNodeGPU(ctx context.Context, node *corev1.Node, gpuInfo *GPUInfo, logger *zap2.Logger) error {
	// Update node labels
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	node.Labels["nvidia.com/gpu"] = "true"
	node.Labels["nvidia.com/gpu.count"] = fmt.Sprintf("%d", gpuInfo.Count)
	node.Labels["nvidia.com/gpu.product"] = gpuInfo.Product
	node.Labels["nvidia.com/gpu.driver-version"] = gpuInfo.DriverVersion
	node.Labels["nvidia.com/gpu.driver-status"] = "ready"
	node.Labels["nvidia.com/gpu.driver-message"] = "nvidia-drivers-operational"

	// Update node to apply labels
	if err := r.Update(ctx, node); err != nil {
		return fmt.Errorf("failed to update node labels: %w", err)
	}

	logger.Info("Updated node labels",
		zap2.String("gpu", "true"),
		zap2.Int("count", gpuInfo.Count),
		zap2.String("product", gpuInfo.Product),
		zap2.String("driverVersion", gpuInfo.DriverVersion))

	// Fetch latest node to update status (capacity and allocatable)
	updatedNode := &corev1.Node{}
	if err := r.Get(ctx, client.ObjectKey{Name: node.Name}, updatedNode); err != nil {
		return fmt.Errorf("failed to get latest node: %w", err)
	}

	// Update capacity and allocatable
	if updatedNode.Status.Capacity == nil {
		updatedNode.Status.Capacity = make(corev1.ResourceList)
	}
	if updatedNode.Status.Allocatable == nil {
		updatedNode.Status.Allocatable = make(corev1.ResourceList)
	}

	gpuQuantity := fmt.Sprintf("%d", gpuInfo.Count)
	updatedNode.Status.Capacity[corev1.ResourceName("nvidia.com/gpu")] = *parseQuantity(gpuQuantity)
	updatedNode.Status.Allocatable[corev1.ResourceName("nvidia.com/gpu")] = *parseQuantity(gpuQuantity)

	// Update node status
	if err := r.Status().Update(ctx, updatedNode); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	logger.Info("Updated node capacity and allocatable",
		zap2.String("nvidia.com/gpu", gpuQuantity))

	return nil
}

func parseQuantity(value string) *resource.Quantity {
	q, err := resource.ParseQuantity(value)
	if err != nil {
		// Fallback to 0 if parsing fails
		return resource.NewQuantity(0, resource.DecimalSI)
	}
	return &q
}

func (r *GPUStatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithEventFilter(predicate.Funcs{
			// Only watch our own node
			CreateFunc: func(e event.CreateEvent) bool {
				return e.Object.GetName() == r.NodeName
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectNew.GetName() == r.NodeName
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false // Don't reconcile on delete
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

// MetricsServer provides HTTP endpoints for GPU and host metrics
type MetricsServer struct {
	CmdExec CommandExecutor
	Logger  *zap2.Logger
	Port    int
}

// GPUMetricsResponse is the JSON response for GPU metrics
type GPUMetricsResponse struct {
	Detected          bool    `json:"detected"`
	Model             string  `json:"model,omitempty"`
	DriverVersion     string  `json:"driverVersion,omitempty"`
	Count             int     `json:"count,omitempty"`
	MemoryTotal       int32   `json:"memoryTotal,omitempty"`
	MemoryUsed        int32   `json:"memoryUsed,omitempty"`
	MemoryFree        int32   `json:"memoryFree,omitempty"`
	UtilizationGPU    int32   `json:"utilizationGpu,omitempty"`
	UtilizationMemory int32   `json:"utilizationMemory,omitempty"`
	Temperature       int32   `json:"temperature,omitempty"`
	PowerDraw         float32 `json:"powerDraw,omitempty"`
	PowerLimit        float32 `json:"powerLimit,omitempty"`
}

func (s *MetricsServer) Start() error {
	http.HandleFunc("/metrics/gpu", s.handleGPUMetrics)
	http.HandleFunc("/healthz", s.handleHealthz)

	addr := fmt.Sprintf(":%d", s.Port)
	s.Logger.Info("Starting metrics server", zap2.Int("port", s.Port))
	return http.ListenAndServe(addr, nil)
}

func (s *MetricsServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *MetricsServer) handleGPUMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check if GPU is detected
	gpuDetected := s.detectGPU()
	if !gpuDetected {
		json.NewEncoder(w).Encode(GPUMetricsResponse{
			Detected: false,
		})
		return
	}

	// Collect GPU metrics
	metrics, err := s.collectGPUMetrics()
	if err != nil {
		s.Logger.Error("Failed to collect GPU metrics", zap2.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

func (s *MetricsServer) detectGPU() bool {
	checkScript := `
		if [ -d /sys/bus/pci/devices ]; then
			for device in /sys/bus/pci/devices/*; do
				if [ -f "$device/vendor" ] && [ -f "$device/device" ]; then
					vendor=$(cat "$device/vendor" 2>/dev/null)
					if [ "$vendor" = "0x10de" ]; then
						exit 0
					fi
				fi
			done
		fi
		exit 1
	`
	_, err := s.CmdExec.Execute(checkScript)
	return err == nil
}

func (s *MetricsServer) collectGPUMetrics() (*GPUMetricsResponse, error) {
	// Query nvidia-smi for comprehensive metrics
	metricsScript := "nvidia-smi --query-gpu=name,driver_version,memory.total,memory.used,memory.free,utilization.gpu,utilization.memory,temperature.gpu,power.draw,power.limit --format=csv,noheader,nounits | head -1"

	output, err := s.CmdExec.Execute(metricsScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU metrics: %w", err)
	}

	// Parse CSV output
	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) < 10 {
		return nil, fmt.Errorf("unexpected nvidia-smi output format: got %d fields, expected 10", len(parts))
	}

	// Helper to parse int32
	parseInt32 := func(s string) (int32, error) {
		s = strings.TrimSpace(s)
		var val int32
		if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
			return 0, err
		}
		return val, nil
	}

	// Helper to parse float32
	parseFloat32 := func(s string) (float32, error) {
		s = strings.TrimSpace(s)
		var val float32
		if _, err := fmt.Sscanf(s, "%f", &val); err != nil {
			return 0, err
		}
		return val, nil
	}

	memoryTotal, _ := parseInt32(parts[2])
	memoryUsed, _ := parseInt32(parts[3])
	memoryFree, _ := parseInt32(parts[4])
	utilizationGPU, _ := parseInt32(parts[5])
	utilizationMemory, _ := parseInt32(parts[6])
	temperature, _ := parseInt32(parts[7])
	powerDraw, _ := parseFloat32(parts[8])
	powerLimit, _ := parseFloat32(parts[9])

	// Get GPU count
	countScript := "nvidia-smi --query-gpu=name --format=csv,noheader | wc -l"
	countOutput, err := s.CmdExec.Execute(countScript)
	count := 1
	if err == nil {
		fmt.Sscanf(strings.TrimSpace(string(countOutput)), "%d", &count)
	}

	return &GPUMetricsResponse{
		Detected:          true,
		Model:             strings.TrimSpace(parts[0]),
		DriverVersion:     strings.TrimSpace(parts[1]),
		Count:             count,
		MemoryTotal:       memoryTotal,
		MemoryUsed:        memoryUsed,
		MemoryFree:        memoryFree,
		UtilizationGPU:    utilizationGPU,
		UtilizationMemory: utilizationMemory,
		Temperature:       temperature,
		PowerDraw:         powerDraw,
		PowerLimit:        powerLimit,
	}, nil
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

	// Get WorkMachine name from environment
	workmachineName := os.Getenv("WORKMACHINE_NAME")
	if workmachineName == "" {
		zapLogger.Fatal("WORKMACHINE_NAME environment variable not set")
	}

	zapLogger.Info("Starting Package Manager",
		zap2.String("namespace", namespace),
		zap2.String("workmachineName", workmachineName),
		zap2.String("nixStorePath", nixStorePath))

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
				namespace: {}, // Watch namespace-scoped resources in this namespace
			},
			// Cluster-scoped resources (Nodes, Workspaces) are watched globally by default
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
		Client:          mgr.GetClient(),
		Logger:          zapLogger,
		FS:              fs,
		WorkMachineName: workmachineName,
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

	zapLogger.Info("All reconcilers configured",
		zap2.String("nodeName", nodeName))

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

	ctx := ctrl.SetupSignalHandler()
	zapLogger.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		zapLogger.Fatal("Failed to start manager", zap2.Error(err))
	}
}
