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
	"sync"
	"time"

	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
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
	workspaceHomePath         = "/var/lib/kloudlite/home"
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
	Scheme         *runtime.Scheme
	Logger         *zap2.Logger
	Namespace      string
	CmdExec        CommandExecutor
	ProfileManager *NixProfileManager

	// workspaceLocks prevents concurrent package installations on the same workspace
	workspaceLocks sync.Map // map[string]*sync.Mutex
}

// getWorkspaceLock returns a mutex for the given workspace, creating one if it doesn't exist
func (r *PackageManagerReconciler) getWorkspaceLock(workspace string) *sync.Mutex {
	lock, _ := r.workspaceLocks.LoadOrStore(workspace, &sync.Mutex{})
	return lock.(*sync.Mutex)
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

	workspace := pkgReq.Spec.WorkspaceRef

	// Check if the PackageRequest is being deleted
	if pkgReq.DeletionTimestamp != nil {
		if containsString(pkgReq.Finalizers, packageRequestFinalizer) {
			logger.Info("PackageRequest is being deleted, cleaning up profile",
				zap2.String("workspace", workspace))

			// Clean up the profile directory
			if err := r.ProfileManager.CleanupProfile(workspace); err != nil {
				logger.Error("Failed to cleanup profile", zap2.Error(err))
				// Continue anyway - profile may not exist
			}

			// Remove finalizer
			pkgReq.Finalizers = removeString(pkgReq.Finalizers, packageRequestFinalizer)
			if err := r.Update(ctx, pkgReq); err != nil {
				logger.Error("Failed to remove finalizer", zap2.Error(err))
				return reconcile.Result{}, err
			}

			logger.Info("Cleanup complete, finalizer removed")
		}
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
		return reconcile.Result{Requeue: true}, nil
	}

	// Acquire lock for this workspace to prevent concurrent builds
	workspaceLock := r.getWorkspaceLock(workspace)
	if !workspaceLock.TryLock() {
		logger.Info("Build already in progress for workspace, requeuing",
			zap2.String("workspace", workspace))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}
	defer workspaceLock.Unlock()

	// Debounce: wait briefly to batch multiple rapid package changes
	logger.Info("Debouncing package changes", zap2.Duration("wait", 2*time.Second))
	time.Sleep(2 * time.Second)

	// Re-fetch PackageRequest to get the latest spec after debounce
	if err := r.Get(ctx, req.NamespacedName, pkgReq); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("PackageRequest deleted during debounce, skipping")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to re-fetch PackageRequest after debounce", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Compute hash of current spec for change detection
	specHash := r.ProfileManager.ComputeSpecHash(pkgReq.Spec.Packages)

	// Check if already up to date
	if pkgReq.Status.SpecHash == specHash && pkgReq.Status.Phase == "Ready" {
		logger.Info("Packages already up to date, skipping build",
			zap2.String("workspace", workspace),
			zap2.String("specHash", specHash))
		return reconcile.Result{}, nil
	}

	// Extract package names for status
	packageNames := make([]string, len(pkgReq.Spec.Packages))
	for i, pkg := range pkgReq.Spec.Packages {
		packageNames[i] = pkg.Name
	}

	// Handle empty package list
	if len(pkgReq.Spec.Packages) == 0 {
		logger.Info("No packages specified, cleaning up profile",
			zap2.String("workspace", workspace))

		// Clean up existing profile if any
		_ = r.ProfileManager.CleanupProfile(workspace)

		// Update status
		if err := r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *packagesv1.PackageRequest) {
			latest.Status.ObservedGeneration = latest.Generation
			latest.Status.Phase = "Ready"
			latest.Status.Message = "No packages to install"
			latest.Status.SpecHash = specHash
			latest.Status.PackageCount = 0
			latest.Status.Packages = nil
			latest.Status.ProfileStorePath = ""
			latest.Status.PackagesPath = ""
			latest.Status.FailedPackage = ""
			latest.Status.LastUpdated = metav1.Now()
		}, logger); err != nil {
			logger.Error("Failed to update status", zap2.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// Set phase to Installing
	if err := r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *packagesv1.PackageRequest) {
		latest.Status.ObservedGeneration = latest.Generation
		latest.Status.Phase = "Installing"
		latest.Status.Message = fmt.Sprintf("Building %d packages...", len(pkgReq.Spec.Packages))
		latest.Status.FailedPackage = ""
		latest.Status.LastUpdated = metav1.Now()
	}, logger); err != nil {
		logger.Warn("Failed to update status to Installing phase", zap2.Error(err))
	}

	// Generate profile.nix
	nixPath, err := r.ProfileManager.GenerateProfileNix(workspace, pkgReq.Spec.Packages)
	if err != nil {
		logger.Error("Failed to generate profile.nix", zap2.Error(err))
		r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *packagesv1.PackageRequest) {
			latest.Status.Phase = "Failed"
			latest.Status.Message = fmt.Sprintf("Failed to generate profile: %v", err)
			latest.Status.LastUpdated = metav1.Now()
		}, logger)
		return reconcile.Result{}, err
	}

	logger.Info("Generated profile.nix",
		zap2.String("workspace", workspace),
		zap2.String("path", nixPath))

	// Build and activate the profile
	result, err := r.ProfileManager.BuildAndActivate(ctx, workspace)
	if err != nil {
		logger.Error("Build system error", zap2.Error(err))
		r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *packagesv1.PackageRequest) {
			latest.Status.Phase = "Failed"
			latest.Status.Message = fmt.Sprintf("Build error: %v", err)
			latest.Status.LastUpdated = metav1.Now()
		}, logger)
		return reconcile.Result{}, err
	}

	if !result.Success {
		logger.Error("Nix build failed",
			zap2.String("workspace", workspace),
			zap2.String("failedPackage", result.FailedPackage),
			zap2.String("error", result.Error))

		r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *packagesv1.PackageRequest) {
			latest.Status.Phase = "Failed"
			latest.Status.FailedPackage = result.FailedPackage
			if result.FailedPackage != "" {
				latest.Status.Message = fmt.Sprintf("Package '%s' failed to build", result.FailedPackage)
			} else {
				latest.Status.Message = "Build failed: " + truncateError(result.Error, 200)
			}
			latest.Status.LastUpdated = metav1.Now()
		}, logger)
		return reconcile.Result{}, nil
	}

	// Update status with success
	if err := r.updateStatusWithRetry(ctx, req.NamespacedName, func(latest *packagesv1.PackageRequest) {
		latest.Status.ObservedGeneration = latest.Generation
		latest.Status.Phase = "Ready"
		latest.Status.Message = fmt.Sprintf("Successfully installed %d packages", len(pkgReq.Spec.Packages))
		latest.Status.SpecHash = specHash
		latest.Status.PackageCount = len(pkgReq.Spec.Packages)
		latest.Status.Packages = packageNames
		latest.Status.ProfileStorePath = result.StorePath
		latest.Status.PackagesPath = result.PackagesPath
		latest.Status.FailedPackage = ""
		latest.Status.LastUpdated = metav1.Now()
	}, logger); err != nil {
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("PackageRequest reconciliation complete",
		zap2.String("workspace", workspace),
		zap2.Int("packages", len(pkgReq.Spec.Packages)),
		zap2.String("storePath", result.StorePath))

	return reconcile.Result{}, nil
}

// truncateError truncates an error message to maxLen characters
func truncateError(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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

		// Change ownership to kl user (uid 1001) so workspace can use it for SSH
		if err := r.FS.Chown(targetPath, 1001, 1001); err != nil {
			logger.Error("Failed to chown ssh_host_rsa_key", zap2.Error(err))
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

		// If GPU was previously detected, clean up GPU resources from the node
		if r.LastGPUDetected {
			logger.Info("Cleaning up GPU resources from node (machine type changed to non-GPU)")
			if err := r.cleanupNodeGPU(ctx, node, logger); err != nil {
				logger.Error("Failed to cleanup node GPU resources", zap2.Error(err))
				return reconcile.Result{}, err
			}
			logger.Info("Successfully cleaned up GPU resources from node")
		}

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

// cleanupNodeGPU removes GPU labels and resources from the node
// Called when machine type changes from GPU to non-GPU
func (r *GPUStatusReconciler) cleanupNodeGPU(ctx context.Context, node *corev1.Node, logger *zap2.Logger) error {
	// Remove GPU labels
	if node.Labels != nil {
		delete(node.Labels, "nvidia.com/gpu")
		delete(node.Labels, "nvidia.com/gpu.count")
		delete(node.Labels, "nvidia.com/gpu.product")
		delete(node.Labels, "nvidia.com/gpu.driver-version")
		delete(node.Labels, "nvidia.com/gpu.driver-status")
		delete(node.Labels, "nvidia.com/gpu.driver-message")
	}

	// Update node to apply label deletions
	if err := r.Update(ctx, node); err != nil {
		return fmt.Errorf("failed to remove GPU labels from node: %w", err)
	}

	logger.Info("Removed GPU labels from node")

	// Fetch latest node to update status (capacity and allocatable)
	updatedNode := &corev1.Node{}
	if err := r.Get(ctx, client.ObjectKey{Name: node.Name}, updatedNode); err != nil {
		return fmt.Errorf("failed to get latest node: %w", err)
	}

	// Remove GPU resources from capacity and allocatable
	if updatedNode.Status.Capacity != nil {
		delete(updatedNode.Status.Capacity, corev1.ResourceName("nvidia.com/gpu"))
	}
	if updatedNode.Status.Allocatable != nil {
		delete(updatedNode.Status.Allocatable, corev1.ResourceName("nvidia.com/gpu"))
	}

	// Update node status
	if err := r.Status().Update(ctx, updatedNode); err != nil {
		return fmt.Errorf("failed to remove GPU resources from node status: %w", err)
	}

	logger.Info("Removed GPU capacity and allocatable from node")

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

// WorkspaceCleanupReconciler watches Workspace resources and manages workspace btrfs subvolumes
// It creates btrfs subvolumes when workspaces are created and deletes them when workspaces are deleted
type WorkspaceCleanupReconciler struct {
	client.Client
	Logger  *zap2.Logger
	FS      FileSystem
	CmdExec CommandExecutor
}

// workspaceStoragePath is the base path for workspace btrfs subvolumes
const workspaceStoragePath = "/var/lib/kloudlite/storage/workspaces"

func (r *WorkspaceCleanupReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("workspace", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling Workspace for btrfs storage management")

	// Fetch Workspace (namespaced)
	workspace := &workspacev1.Workspace{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, workspace); err != nil {
		// If workspace is not found, it's already fully deleted (including all finalizers)
		// Nothing to do - this is expected when workspace is deleted
		logger.Info("Workspace not found, already deleted")
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Check if workspace is being deleted
	if workspace.DeletionTimestamp != nil {
		// Workspace is being deleted
		if containsString(workspace.Finalizers, workspaceCleanupFinalizer) {
			logger.Info("Workspace is being deleted, cleaning up btrfs subvolume")

			// Clean up workspace btrfs subvolume
			workspaceDir := fmt.Sprintf("%s/%s", workspaceStoragePath, workspace.Name)
			logger.Info("Removing workspace btrfs subvolume", zap2.String("path", workspaceDir))

			// Delete btrfs subvolume (will fail if path doesn't exist or is not a subvolume)
			deleteScript := fmt.Sprintf("btrfs subvolume delete %s", workspaceDir)
			output, err := r.CmdExec.Execute(deleteScript)
			if err != nil {
				// Check if path exists - if not, nothing to clean up
				checkExistsScript := fmt.Sprintf("test -e %s", workspaceDir)
				if _, existsErr := r.CmdExec.Execute(checkExistsScript); existsErr != nil {
					logger.Info("Workspace subvolume doesn't exist, skipping cleanup", zap2.String("path", workspaceDir))
				} else {
					logger.Error("Failed to delete btrfs subvolume",
						zap2.String("path", workspaceDir),
						zap2.Error(err),
						zap2.String("output", string(output)))
					return reconcile.Result{}, fmt.Errorf("failed to delete btrfs subvolume: %w", err)
				}
			} else {
				logger.Info("Successfully deleted btrfs subvolume", zap2.String("path", workspaceDir))
			}

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

	// Workspace is NOT being deleted - ensure btrfs subvolume exists
	// This creates the subvolume BEFORE the pod starts
	workspaceDir := fmt.Sprintf("%s/%s", workspaceStoragePath, workspace.Name)

	// Check if btrfs subvolume already exists
	checkScript := fmt.Sprintf("btrfs subvolume show %s > /dev/null 2>&1", workspaceDir)
	if _, err := r.CmdExec.Execute(checkScript); err == nil {
		// Subvolume already exists
		logger.Debug("Workspace btrfs subvolume already exists", zap2.String("path", workspaceDir))
		return reconcile.Result{}, nil
	}

	// Create new btrfs subvolume
	logger.Info("Creating workspace btrfs subvolume", zap2.String("path", workspaceDir))
	createScript := fmt.Sprintf("btrfs subvolume create %s && chown 1001:1001 %s", workspaceDir, workspaceDir)

	if output, err := r.CmdExec.Execute(createScript); err != nil {
		logger.Error("Failed to create workspace btrfs subvolume",
			zap2.String("path", workspaceDir),
			zap2.Error(err),
			zap2.String("output", string(output)))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	logger.Info("Successfully created workspace btrfs subvolume", zap2.String("path", workspaceDir))
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

	// Cache for GPU metrics with periodic background updates
	cacheMutex    sync.RWMutex
	cachedMetrics *GPUMetricsResponse
	cacheInterval time.Duration
	stopChan      chan struct{}
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
	// Start background goroutine to poll GPU metrics at constant frequency
	s.startBackgroundPolling()

	http.HandleFunc("/metrics/gpu", s.handleGPUMetrics)
	http.HandleFunc("/healthz", s.handleHealthz)

	addr := fmt.Sprintf(":%d", s.Port)
	s.Logger.Info("Starting metrics server", zap2.Int("port", s.Port))
	return http.ListenAndServe(addr, nil)
}

// startBackgroundPolling starts a background goroutine that polls GPU metrics
// at a constant frequency and updates the in-memory cache
func (s *MetricsServer) startBackgroundPolling() {
	// Set default cache interval if not configured
	if s.cacheInterval == 0 {
		s.cacheInterval = 5 * time.Second
	}

	s.stopChan = make(chan struct{})

	go func() {
		s.Logger.Info("Starting background GPU metrics polling", zap2.Duration("interval", s.cacheInterval))

		// Poll immediately on startup
		s.updateCache()

		ticker := time.NewTicker(s.cacheInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.updateCache()
			case <-s.stopChan:
				s.Logger.Info("Stopping background GPU metrics polling")
				return
			}
		}
	}()
}

// updateCache polls GPU metrics and updates the in-memory cache
func (s *MetricsServer) updateCache() {
	// Check if GPU is detected
	gpuDetected := s.detectGPU()
	if !gpuDetected {
		s.cacheMutex.Lock()
		s.cachedMetrics = &GPUMetricsResponse{
			Detected: false,
		}
		s.cacheMutex.Unlock()
		return
	}

	// Collect GPU metrics
	metrics, err := s.collectGPUMetrics()
	if err != nil {
		s.Logger.Error("Failed to collect GPU metrics in background poll", zap2.Error(err))
		return
	}

	// Update cache with write lock
	s.cacheMutex.Lock()
	s.cachedMetrics = metrics
	s.cacheMutex.Unlock()

	s.Logger.Debug("GPU metrics cache updated",
		zap2.Bool("detected", metrics.Detected),
		zap2.String("model", metrics.Model),
		zap2.Int32("utilizationGpu", metrics.UtilizationGPU))
}

// Stop gracefully stops the background polling goroutine
func (s *MetricsServer) Stop() {
	if s.stopChan != nil {
		close(s.stopChan)
	}
}

func (s *MetricsServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *MetricsServer) handleGPUMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Serve from cache with read lock (fast, no nvidia-smi execution)
	s.cacheMutex.RLock()
	cachedMetrics := s.cachedMetrics
	s.cacheMutex.RUnlock()

	// If cache is not yet populated, return a default response
	if cachedMetrics == nil {
		json.NewEncoder(w).Encode(GPUMetricsResponse{
			Detected: false,
		})
		return
	}

	json.NewEncoder(w).Encode(cachedMetrics)
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

// SnapshotRequestReconciler watches SnapshotRequest resources and processes them on this node
type SnapshotRequestReconciler struct {
	client.Client
	Logger        *zap2.Logger
	HostCmdExec   CommandExecutor // For btrfs commands that must run on host
	LocalCmdExec  CommandExecutor // For tar/oras commands that run in container
	NodeName      string
}

const snapshotStoragePath = "/var/lib/kloudlite/storage/.snapshots"

func (r *SnapshotRequestReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("snapshotRequest", req.Name),
		zap2.String("namespace", req.Namespace),
	)

	// Fetch SnapshotRequest
	snapshotReq := &snapshotv1.SnapshotRequest{}
	if err := r.Get(ctx, req.NamespacedName, snapshotReq); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get SnapshotRequest", zap2.Error(err))
		return reconcile.Result{}, err
	}

	// Only process requests for this node
	if snapshotReq.Spec.NodeName != r.NodeName {
		return reconcile.Result{}, nil
	}

	// Handle completed or failed requests - no need to reprocess
	if snapshotReq.Status.State == snapshotv1.SnapshotRequestStateCompleted ||
		snapshotReq.Status.State == snapshotv1.SnapshotRequestStateFailed {
		return reconcile.Result{}, nil
	}

	logger.Info("Processing SnapshotRequest",
		zap2.String("state", string(snapshotReq.Status.State)),
		zap2.String("snapshotName", snapshotReq.Spec.SnapshotName))

	// Process based on current state
	switch snapshotReq.Status.State {
	case "", snapshotv1.SnapshotRequestStatePending:
		return r.handlePending(ctx, snapshotReq, logger)
	case snapshotv1.SnapshotRequestStateCreating:
		return r.handleCreating(ctx, snapshotReq, logger)
	case snapshotv1.SnapshotRequestStateUploading:
		return r.handleUploading(ctx, snapshotReq, logger)
	default:
		logger.Warn("Unknown snapshot request state", zap2.String("state", string(snapshotReq.Status.State)))
		return reconcile.Result{}, nil
	}
}

func (r *SnapshotRequestReconciler) handlePending(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Info("Starting snapshot request",
		zap2.String("sourcePath", req.Spec.SourcePath),
		zap2.String("owner", req.Spec.Owner))

	now := metav1.Now()
	req.Status.State = snapshotv1.SnapshotRequestStateCreating
	req.Status.Message = "Creating btrfs snapshot"
	req.Status.StartedAt = &now
	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *SnapshotRequestReconciler) handleCreating(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (reconcile.Result, error) {
	// Generate local snapshot path
	snapshotPath := fmt.Sprintf("%s/%s", snapshotStoragePath, req.Spec.SnapshotName)

	// Check if snapshot already exists (race condition protection)
	checkScript := fmt.Sprintf("test -d %s && echo exists", snapshotPath)
	checkOutput, _ := r.HostCmdExec.Execute(checkScript)
	if strings.TrimSpace(string(checkOutput)) == "exists" {
		logger.Info("Snapshot already exists, transitioning to Uploading", zap2.String("path", snapshotPath))
		req.Status.LocalSnapshotPath = snapshotPath
		req.Status.State = snapshotv1.SnapshotRequestStateUploading
		req.Status.Message = "Uploading to registry"
		if err := r.Status().Update(ctx, req); err != nil {
			if apierrors.IsConflict(err) {
				return reconcile.Result{Requeue: true}, nil
			}
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Ensure snapshot storage directory exists
	mkdirScript := fmt.Sprintf("mkdir -p %s", snapshotStoragePath)
	if _, err := r.HostCmdExec.Execute(mkdirScript); err != nil {
		logger.Warn("Failed to create snapshot storage directory", zap2.Error(err))
	}

	// Create btrfs snapshot
	snapshotScript := fmt.Sprintf("btrfs subvolume snapshot -r %s %s", req.Spec.SourcePath, snapshotPath)
	output, err := r.HostCmdExec.Execute(snapshotScript)
	if err != nil {
		logger.Error("Failed to create btrfs snapshot",
			zap2.String("sourcePath", req.Spec.SourcePath),
			zap2.String("snapshotPath", snapshotPath),
			zap2.Error(err),
			zap2.String("output", string(output)))
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to create btrfs snapshot: %v - %s", err, string(output)), logger)
	}

	// Update status
	req.Status.LocalSnapshotPath = snapshotPath
	req.Status.State = snapshotv1.SnapshotRequestStateUploading
	req.Status.Message = "Uploading to registry"

	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Created btrfs snapshot", zap2.String("path", snapshotPath))
	return reconcile.Result{Requeue: true}, nil
}

func (r *SnapshotRequestReconciler) handleUploading(ctx context.Context, req *snapshotv1.SnapshotRequest, logger *zap2.Logger) (reconcile.Result, error) {
	// Get the SnapshotStore
	store := &snapshotv1.SnapshotStore{}
	if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.Store}, store); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setFailed(ctx, req, fmt.Sprintf("SnapshotStore %q not found", req.Spec.Store), logger)
		}
		logger.Error("Failed to get SnapshotStore", zap2.Error(err))
		return reconcile.Result{}, err
	}

	if !store.Status.Ready {
		logger.Info("SnapshotStore not ready, waiting", zap2.String("store", store.Name))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Build OCI image reference
	// No DNS resolution needed since oras runs in container which can resolve K8s DNS
	imageRef := fmt.Sprintf("%s/%s/%s:%s",
		store.Spec.Registry.Endpoint,
		store.Spec.Registry.RepositoryPrefix,
		req.Spec.Owner,
		req.Spec.SnapshotName)

	// Use tar + oras to push snapshot to registry
	// oras is installed in the container image via Dockerfile
	// This runs in the container (not via nsenter) since oras is in the container
	pushScript := fmt.Sprintf(`
		set -e
		cd %s
		tar -cf - . | gzip > /tmp/%s.tar.gz
		oras push --insecure %s /tmp/%s.tar.gz:application/vnd.kloudlite.snapshot.v1.tar+gzip
		rm -f /tmp/%s.tar.gz
	`, req.Status.LocalSnapshotPath, req.Spec.SnapshotName, imageRef, req.Spec.SnapshotName, req.Spec.SnapshotName)

	output, err := r.LocalCmdExec.Execute(pushScript)
	if err != nil {
		logger.Error("Failed to push snapshot to registry",
			zap2.String("imageRef", imageRef),
			zap2.Error(err),
			zap2.String("output", string(output)))
		return r.setFailed(ctx, req, fmt.Sprintf("Failed to push to registry: %v", err), logger)
	}

	// Get snapshot size
	var sizeBytes int64
	sizeScript := fmt.Sprintf("du -sb %s | cut -f1", req.Status.LocalSnapshotPath)
	sizeOutput, err := r.LocalCmdExec.Execute(sizeScript)
	if err == nil {
		fmt.Sscanf(strings.TrimSpace(string(sizeOutput)), "%d", &sizeBytes)
	}

	// Build lineage from parent
	var lineage []string
	if req.Spec.ParentSnapshot != "" {
		parentSnapshot := &snapshotv1.Snapshot{}
		if err := r.Get(ctx, client.ObjectKey{Name: req.Spec.ParentSnapshot}, parentSnapshot); err == nil {
			lineage = append(parentSnapshot.Status.Lineage, parentSnapshot.Name)
		}
	}

	// Create the global Snapshot resource
	now := metav1.Now()
	snapshot := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Spec.SnapshotName,
			Labels: req.Labels,
		},
		Spec: snapshotv1.SnapshotSpec{
			Owner:           req.Spec.Owner,
			ParentSnapshot:  req.Spec.ParentSnapshot,
			Description:     req.Spec.Description,
			Artifacts:       req.Spec.Artifacts,
			RetentionPolicy: req.Spec.RetentionPolicy,
		},
		Status: snapshotv1.SnapshotStatus{
			State:     snapshotv1.SnapshotStateReady,
			Message:   "Snapshot ready",
			SizeBytes: sizeBytes,
			SizeHuman: formatSnapshotSize(sizeBytes),
			CreatedAt: &now,
			Lineage:   lineage,
			Registry: &snapshotv1.SnapshotRegistryInfo{
				ImageRef: imageRef,
				PushedAt: &now,
			},
		},
	}

	if err := r.Create(ctx, snapshot); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return r.setFailed(ctx, req, fmt.Sprintf("Failed to create Snapshot: %v", err), logger)
		}
		logger.Info("Snapshot already exists, updating", zap2.String("name", req.Spec.SnapshotName))
	}

	// Delete local snapshot to free space (btrfs operation runs on host)
	deleteScript := fmt.Sprintf("btrfs subvolume delete %s", req.Status.LocalSnapshotPath)
	if _, err := r.HostCmdExec.Execute(deleteScript); err != nil {
		logger.Warn("Failed to delete local snapshot", zap2.Error(err))
	}

	// Mark request as completed
	completedNow := metav1.Now()
	req.Status.State = snapshotv1.SnapshotRequestStateCompleted
	req.Status.Message = "Snapshot created successfully"
	req.Status.CompletedAt = &completedNow
	req.Status.CreatedSnapshot = req.Spec.SnapshotName

	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Snapshot created successfully",
		zap2.String("snapshot", req.Spec.SnapshotName),
		zap2.String("imageRef", imageRef))

	return reconcile.Result{}, nil
}

func (r *SnapshotRequestReconciler) setFailed(ctx context.Context, req *snapshotv1.SnapshotRequest, message string, logger *zap2.Logger) (reconcile.Result, error) {
	logger.Error("Snapshot request failed", zap2.String("message", message))

	now := metav1.Now()
	req.Status.State = snapshotv1.SnapshotRequestStateFailed
	req.Status.Message = message
	req.Status.CompletedAt = &now

	if err := r.Status().Update(ctx, req); err != nil {
		if apierrors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		}
		logger.Error("Failed to update status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func formatSnapshotSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func (r *SnapshotRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&snapshotv1.SnapshotRequest{}).
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
	if err := snapshotv1.AddToScheme(scheme); err != nil {
		zapLogger.Fatal("Failed to add snapshot v1 scheme", zap2.Error(err))
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
				// Watch Snapshots globally (cluster-scoped)
				&snapshotv1.Snapshot{}: {},
				// Watch SnapshotStores globally (cluster-scoped)
				&snapshotv1.SnapshotStore{}: {},
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

	// Setup snapshot request reconciler (handles btrfs snapshots on this node)
	snapshotRequestReconciler := &SnapshotRequestReconciler{
		Client:       mgr.GetClient(),
		Logger:       zapLogger,
		HostCmdExec:  &HostCommandExecutor{}, // For btrfs commands on host
		LocalCmdExec: &RealCommandExecutor{}, // For tar/oras in container
		NodeName:     nodeName,
	}

	if err := snapshotRequestReconciler.SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("Failed to setup snapshot request controller", zap2.Error(err))
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
