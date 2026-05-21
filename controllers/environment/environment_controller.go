package environment

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/kloudlite/controllers/composition"
	"github.com/kloudlite/kloudlite/controllers/controllerconfig"
	"github.com/kloudlite/kloudlite/pkg/pagination"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	workmachinevl "github.com/kloudlite/kloudlite/types/workmachine/v1"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	environmentFinalizer = "environments.kloudlite.io/finalizer"
	// Kind constants for owner references
	workMachineKind = "WorkMachine"
)

// EnvironmentReconciler reconciles Environment objects and creates namespaces
type EnvironmentReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Logger          *zap.Logger
	Cfg             *controllerconfig.ControllerConfig // Controller configuration
	OwnNamespace    string
	WorkMachineName string
}

// Reconcile handles Environment events and ensures namespace exists
func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if r.Cfg == nil {
		r.Cfg = &controllerconfig.ControllerConfig{}
	}
	if r.Cfg.Environment.PodTerminationRetryInterval == 0 {
		r.Cfg.Environment.PodTerminationRetryInterval = 2 * time.Second
	}
	if r.Cfg.Environment.DeletionRetryInterval == 0 {
		r.Cfg.Environment.DeletionRetryInterval = 5 * time.Second
	}

	logger := r.Logger.With(
		zap.String("environment", req.Name),
		zap.String("namespace", req.Namespace),
	)

	logger.Info("reconciling environment")

	// Fetch the Environment instance (namespace-scoped)
	environment := &environmentsv1.Environment{}
	err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, environment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Environment has been deleted, nothing to do
			logger.Info("environment not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("failed to get environment", zap.Error(err))
		return reconcile.Result{}, err
	}
	if !r.shouldReconcileEnvironment(environment) {
		logger.Info("skipping environment outside controller scope",
			zap.String("own_namespace", r.OwnNamespace),
			zap.String("workmachine_name", r.WorkMachineName),
			zap.String("environment_namespace", environment.Namespace),
			zap.String("environment_workmachine", environment.Spec.WorkMachineName))
		return reconcile.Result{}, nil
	}

	if environment.DeletionTimestamp == nil {
		if !controllerutil.ContainsFinalizer(environment, environmentFinalizer) {
			logger.Info("adding finalizer to environment")
			result, err := addEnvironmentFinalizer(ctx, r.Client, environment)
			if err != nil {
				logger.Error("failed to update environment", zap.Error(err))
				return reconcile.Result{}, err
			}
			return result, nil
		}

		// Ensure hash label is set on the environment resource for efficient lookups
		envHash := generateHash(fmt.Sprintf("%s-%s", environment.Name, environment.Spec.OwnedBy))
		labels := environment.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		if labels["kloudlite.io/hash"] != envHash {
			labels["kloudlite.io/hash"] = envHash
			environment.Labels = labels
			logger.Info("adding hash label to environment", zap.String("hash", envHash))
			if err := r.Update(ctx, environment); err != nil {
				logger.Error("failed to update environment", zap.Error(err))
				return reconcile.Result{}, err
			}
		}

		// Set WorkMachine as owner if WorkMachineName is specified and owner reference not yet set
		if environment.Spec.WorkMachineName != "" {
			needsOwnerUpdate := true
			for _, ownerRef := range environment.OwnerReferences {
				if ownerRef.Kind == "WorkMachine" && ownerRef.Name == environment.Spec.WorkMachineName {
					needsOwnerUpdate = false
					break
				}
			}

			if needsOwnerUpdate {
				logger.Info("setting workmachine as owner of environment",
					zap.String("workmachine", environment.Spec.WorkMachineName))

				// Fetch WorkMachine to set as owner
				workmachine := &workmachinevl.WorkMachine{}
				if err := r.Get(ctx, client.ObjectKey{Name: environment.Spec.WorkMachineName}, workmachine); err != nil {
					logger.Error("failed to get workmachine for ownership",
						zap.String("workmachine", environment.Spec.WorkMachineName),
						zap.Error(err))
					// Don't fail reconciliation, just log the error
					// The ownership will be set on next reconciliation
				} else {
					// Set WorkMachine as owner for cascading deletion (without blockOwnerDeletion)
					// Note: TypeMeta isn't populated by controller-runtime's Get method,
					// so we set them explicitly using the GroupVersion constant
					blockOwnerDeletion := false
					ownerRef := metav1.OwnerReference{
						APIVersion:         workmachinevl.GroupVersion.String(),
						Kind:               workMachineKind,
						Name:               workmachine.Name,
						UID:                workmachine.UID,
						BlockOwnerDeletion: &blockOwnerDeletion,
					}
					environment.SetOwnerReferences([]metav1.OwnerReference{ownerRef})

					if err := r.Update(ctx, environment); err != nil {
						logger.Error("failed to update environment with owner reference", zap.Error(err))
						return reconcile.Result{}, err
					}
					logger.Info("successfully set workmachine as owner of environment")
					return reconcile.Result{Requeue: true}, nil
				}
			}
		}

		// Handle environment creation from snapshot if fromSnapshot is set
		// Snapshot restore takes precedence over normal environment reconciliation
		if environment.Spec.FromSnapshot != nil {
			logger.Info("environment has fromSnapshot set, handling snapshot restore",
				zap.String("snapshotName", environment.Spec.FromSnapshot.SnapshotName))
			return r.handleSnapshotRestore(ctx, environment, logger)
		}
	}

	original := environment.DeepCopy()
	session := NewEnvironmentStatusSession(environment)
	session.Touch()

	result, err := r.reconcileEnvironment(ctx, session)
	session.ComputeReady()

	patchResult, patchErr := PatchEnvironmentStatus(ctx, r.Client, original, session.Object())
	if finalResult, finalErr := reconcileEnvironmentReturn(result, err, patchResult, patchErr); finalErr != nil || !isZeroEnvironmentResult(finalResult) {
		return finalResult, finalErr
	}

	if environment.GetDeletionTimestamp() != nil && controllerutil.ContainsFinalizer(environment, environmentFinalizer) && cleanupComplete(environment) {
		logger.Info("all cleanup complete, removing finalizer from environment")
		result, err := removeEnvironmentFinalizer(ctx, r.Client, client.ObjectKeyFromObject(environment))
		if err != nil || !result.IsZero() {
			return result, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *EnvironmentReconciler) scopeConfigured() bool {
	return r.OwnNamespace != "" || r.WorkMachineName != ""
}

func (r *EnvironmentReconciler) shouldReconcileEnvironment(environment *environmentsv1.Environment) bool {
	if !r.scopeConfigured() {
		return true
	}
	if r.OwnNamespace == "" || r.WorkMachineName == "" {
		return false
	}
	return environment.Namespace == r.OwnNamespace && environment.Spec.WorkMachineName == r.WorkMachineName
}

// SetupWithManager sets up the controller with the Manager
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.Environment{}, builder.WithPredicates(r.environmentPredicate())).
		Owns(&networkingv1.NetworkPolicy{}). // Watch NetworkPolicies owned by Environments
		Watches(
			&appsv1.StatefulSet{},
			handler.EnqueueRequestsFromMapFunc(r.findEnvironmentForComposeResource),
		).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.findEnvironmentForComposeResource),
		).
		Complete(r)
	// Note: We don't watch WorkMachine here because Environment references WorkMachine by name
	// The Environment controller will handle WorkMachine ownership during reconciliation
}

// environmentPredicate returns a predicate that reconciles on:
// 1. Spec changes (generation changes)
// 2. Status state transitions that require reconciliation
func (r *EnvironmentReconciler) environmentPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true // Always reconcile on create
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldEnv, oldOk := e.ObjectOld.(*environmentsv1.Environment)
			newEnv, newOk := e.ObjectNew.(*environmentsv1.Environment)
			if !oldOk || !newOk {
				return true // Reconcile if we can't determine the type
			}

			// Reconcile on generation change (spec change)
			if oldEnv.Generation != newEnv.Generation {
				return true
			}

			// Reconcile when state transitions from snapping/deactivating to active/inactive
			// These transitions require scaling workloads up or down
			oldState := oldEnv.Status.State
			newState := newEnv.Status.State
			if oldState != newState {
				// Reconcile when transitioning OUT of snapping or deactivating
				if oldState == environmentsv1.EnvironmentStateSnapping ||
					oldState == environmentsv1.EnvironmentStateDeactivating {
					return true
				}
				// Reconcile when transitioning INTO snapping or deactivating
				if newState == environmentsv1.EnvironmentStateSnapping ||
					newState == environmentsv1.EnvironmentStateDeactivating {
					return true
				}
			}

			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true // Always reconcile on delete
		},
	}
}

// findEnvironmentForComposeResource finds the environment that owns a compose resource
func (r *EnvironmentReconciler) findEnvironmentForComposeResource(ctx context.Context, obj client.Object) []reconcile.Request {
	// Check if this resource has the docker-composition label
	labels := obj.GetLabels()
	if labels == nil {
		return nil
	}
	if labels[composition.ManagedLabel] != "true" {
		return nil
	}

	envName, ok := labels[composition.DockerCompositionLabel]
	if !ok {
		return nil
	}

	envNamespace, ok := labels[composition.EnvironmentNamespaceLabel]
	if !ok {
		return nil
	}
	if r.scopeConfigured() && (r.OwnNamespace == "" || r.WorkMachineName == "" || envNamespace != r.OwnNamespace) {
		return nil
	}

	// Return a reconcile request for the environment
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      envName,
				Namespace: envNamespace,
			},
		},
	}
}

// waitForPodsTerminated waits for all pods in a namespace to be fully deleted
// This ensures databases have time to checkpoint and flush WAL before the
// environment is marked as inactive. Only Succeeded pods (completed Jobs) are ignored.
func (r *EnvironmentReconciler) waitForPodsTerminated(ctx context.Context, namespace string, logger *zap.Logger) bool {
	pods := &corev1.PodList{}
	if err := pagination.ListAll(ctx, r, pods, client.InNamespace(namespace)); err != nil {
		logger.Warn("failed to list pods", zap.Error(err))
		return false
	}

	// Wait for ALL pods to be deleted, except completed Job pods (Succeeded phase)
	for _, pod := range pods.Items {
		// Skip completed Job pods - they're finished and won't write to disk
		if pod.Status.Phase == corev1.PodSucceeded {
			continue
		}
		// Any other pod (Running, Pending, Failed, Unknown) means we should wait
		logger.Debug("pod still exists", zap.String("pod", pod.Name), zap.String("phase", string(pod.Status.Phase)))
		return false
	}
	return true
}

// hasActiveSnapshotOperation checks if there are any in-progress snapshot operations
// (EnvironmentSnapshotRequest or EnvironmentSnapshotRestore) for this environment
func (r *EnvironmentReconciler) hasActiveSnapshotOperation(ctx context.Context, environment *environmentsv1.Environment) (bool, error) {
	// Check for active EnvironmentSnapshotRequests
	snapshotRequests := &environmentsv1.EnvironmentSnapshotRequestList{}
	if err := pagination.ListAll(ctx, r, snapshotRequests); err != nil {
		return false, err
	}

	for _, req := range snapshotRequests.Items {
		if req.Spec.EnvironmentName != environment.Name || req.Namespace != environment.Spec.TargetNamespace {
			continue
		}
		if req.Spec.EnvironmentNamespace != environment.Namespace {
			continue
		}
		// Check if request is in-progress (not completed or failed)
		if req.Status.Phase != environmentsv1.EnvironmentSnapshotRequestPhaseCompleted &&
			req.Status.Phase != environmentsv1.EnvironmentSnapshotRequestPhaseFailed {
			return true, nil
		}
	}

	// Check for active EnvironmentSnapshotRestores
	snapshotRestores := &environmentsv1.EnvironmentSnapshotRestoreList{}
	if err := pagination.ListAll(ctx, r, snapshotRestores); err != nil {
		return false, err
	}

	for _, restore := range snapshotRestores.Items {
		if restore.Spec.EnvironmentName != environment.Name || restore.Namespace != environment.Spec.TargetNamespace {
			continue
		}
		if restore.Spec.EnvironmentNamespace != environment.Namespace {
			continue
		}
		// Check if restore is in-progress (not completed or failed)
		if restore.Status.Phase != environmentsv1.EnvironmentSnapshotRestorePhaseCompleted &&
			restore.Status.Phase != environmentsv1.EnvironmentSnapshotRestorePhaseFailed {
			return true, nil
		}
	}

	return false, nil
}
