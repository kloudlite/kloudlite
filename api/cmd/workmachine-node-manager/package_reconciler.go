package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	zap2 "go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

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
