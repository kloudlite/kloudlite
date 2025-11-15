package environment

import (
	"context"
	"fmt"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workmachinevl "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"github.com/kloudlite/kloudlite/api/pkg/utils"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	environmentFinalizer = "environments.kloudlite.io/finalizer"
)

// EnvironmentReconciler reconciles Environment objects and creates namespaces
type EnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// applyLabelsAndAnnotations applies labels and annotations from environment spec to namespace
func (r *EnvironmentReconciler) applyLabelsAndAnnotations(namespace *corev1.Namespace, environment *environmentsv1.Environment) {
	// Initialize maps if nil
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string)
	}
	if namespace.Annotations == nil {
		namespace.Annotations = make(map[string]string)
	}

	// Add standard environment label
	namespace.Labels["kloudlite.io/environment"] = environment.Name

	// Add created-by annotation (emails contain invalid label characters)
	namespace.Annotations["kloudlite.io/created-by"] = environment.Spec.OwnedBy

	// Add custom labels from environment spec (move to annotations for invalid characters)
	if environment.Spec.Labels != nil {
		for k, v := range environment.Spec.Labels {
			if utils.IsValidLabel(k) && utils.IsValidLabel(v) {
				namespace.Labels[k] = v
			} else {
				// Move invalid labels to annotations
				namespace.Annotations[k] = v
			}
		}
	}

	// Add custom annotations from environment spec
	if environment.Spec.Annotations != nil {
		for k, v := range environment.Spec.Annotations {
			namespace.Annotations[k] = v
		}
	}
}

// updateEnvironmentStatus safely updates environment status with retry logic
func (r *EnvironmentReconciler) updateEnvironmentStatus(ctx context.Context, environment *environmentsv1.Environment, state environmentsv1.EnvironmentState, message string, logger *zap.Logger) error {
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.State = state
		environment.Status.Message = message

		now := metav1.Now()
		if state == environmentsv1.EnvironmentStateActive {
			environment.Status.LastActivatedTime = &now
		} else if state == environmentsv1.EnvironmentStateInactive {
			environment.Status.LastDeactivatedTime = &now
		}

		return nil
	}, logger)
}

// addOrUpdateCondition adds or updates a condition in the environment status
func (r *EnvironmentReconciler) addOrUpdateCondition(environment *environmentsv1.Environment, conditionType environmentsv1.EnvironmentConditionType, status metav1.ConditionStatus, reason, message string) {
	if environment.Status.Conditions == nil {
		environment.Status.Conditions = []environmentsv1.EnvironmentCondition{}
	}

	now := metav1.Now()
	newCondition := environmentsv1.EnvironmentCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: &now,
		Reason:             reason,
		Message:            message,
	}

	// Find and update existing condition or add new one
	found := false
	for i, condition := range environment.Status.Conditions {
		if condition.Type == conditionType {
			environment.Status.Conditions[i] = newCondition
			found = true
			break
		}
	}
	if !found {
		environment.Status.Conditions = append(environment.Status.Conditions, newCondition)
	}
}

// handleEnvironmentDeactivation disconnects workspaces and removes service intercepts when environment is deactivated
func (r *EnvironmentReconciler) handleEnvironmentDeactivation(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	logger.Info("Handling environment deactivation cleanup",
		zap.String("environment", environment.Name),
		zap.String("targetNamespace", environment.Spec.TargetNamespace))

	disconnectedWorkspaces := 0

	// 1. Find and disconnect all workspaces connected to this environment
	// Workspaces are cluster-scoped, so list without namespace filter
	logger.Info("Finding workspaces connected to this environment")
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(ctx, workspaceList); err != nil {
		logger.Error("Failed to list workspaces", zap.Error(err))
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	for i := range workspaceList.Items {
		workspace := &workspaceList.Items[i]

		// Check if this workspace is connected to the environment being deactivated
		if workspace.Status.ConnectedEnvironment != nil && workspace.Status.ConnectedEnvironment.Name == environment.Name {
			logger.Info("Disconnecting workspace from environment",
				zap.String("workspace", workspace.Name),
				zap.String("environment", environment.Name))

			// Clear the connected environment from workspace status
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, workspace, func() error {
				workspace.Status.ConnectedEnvironment = nil
				return nil
			}, logger); err != nil {
				logger.Error("Failed to disconnect workspace",
					zap.String("workspace", workspace.Name),
					zap.Error(err))
				// Continue with other workspaces instead of failing
				continue
			}

			disconnectedWorkspaces++
			logger.Info("Successfully disconnected workspace",
				zap.String("workspace", workspace.Name),
				zap.String("environment", environment.Name))
		}
	}

	// Note: Service intercepts are now managed as part of Composition.spec.intercepts
	// The composition controller handles their lifecycle automatically

	logger.Info("Environment deactivation cleanup completed",
		zap.String("environment", environment.Name),
		zap.Int("disconnectedWorkspaces", disconnectedWorkspaces))

	return nil
}

// Reconcile handles Environment events and ensures namespace exists
func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("environment", req.Name),
	)

	logger.Info("Reconciling Environment")

	// Fetch the Environment instance (cluster-scoped)
	environment := &environmentsv1.Environment{}
	err := r.Get(ctx, client.ObjectKey{Name: req.Name}, environment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Environment has been deleted, nothing to do
			logger.Info("Environment not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Environment", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if environment is being deleted
	if environment.DeletionTimestamp != nil {
		logger.Info("Environment is being deleted, starting cleanup")
		return r.handleDeletion(ctx, environment, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(environment, environmentFinalizer) {
		logger.Info("Adding finalizer to environment")
		controllerutil.AddFinalizer(environment, environmentFinalizer)
		if err := r.Update(ctx, environment); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
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
			logger.Info("Setting WorkMachine as owner of Environment",
				zap.String("workmachine", environment.Spec.WorkMachineName))

			// Fetch WorkMachine to set as owner
			workmachine := &workmachinevl.WorkMachine{}
			if err := r.Get(ctx, client.ObjectKey{Name: environment.Spec.WorkMachineName}, workmachine); err != nil {
				logger.Error("Failed to get WorkMachine for ownership",
					zap.String("workmachine", environment.Spec.WorkMachineName),
					zap.Error(err))
				// Don't fail reconciliation, just log the error
				// The ownership will be set on next reconciliation
			} else {
				// Set WorkMachine as owner for cascading deletion (without blockOwnerDeletion)
				blockOwnerDeletion := false
				ownerRef := metav1.OwnerReference{
					APIVersion:         workmachine.APIVersion,
					Kind:               workmachine.Kind,
					Name:               workmachine.Name,
					UID:                workmachine.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
				}
				environment.SetOwnerReferences([]metav1.OwnerReference{ownerRef})

				if err := r.Update(ctx, environment); err != nil {
					logger.Error("Failed to update Environment with owner reference", zap.Error(err))
					return reconcile.Result{}, err
				}
				logger.Info("Successfully set WorkMachine as owner of Environment")
				return reconcile.Result{Requeue: true}, nil
			}
		}
	}

	// Check if cloning is requested
	if environment.Spec.CloneFrom != "" {
		logger.Info("Cloning requested from source environment",
			zap.String("source", environment.Spec.CloneFrom))
		return r.handleCloning(ctx, environment, logger)
	}

	// Note: TargetNamespace is always set by the mutation webhook.
	// The webhook generates it as "env-{name}" if not provided by the user.
	// The webhook also validates that the namespace doesn't already exist.
	// Controller's responsibility is to create the actual Kubernetes namespace resource.

	// Check if namespace already exists
	namespace := &corev1.Namespace{}
	err = r.Get(ctx, client.ObjectKey{Name: environment.Spec.TargetNamespace}, namespace)

	if err == nil {
		// Namespace already exists
		logger.Info("Namespace already exists for environment",
			zap.String("namespace", environment.Spec.TargetNamespace))

		// Apply labels and annotations using helper function
		r.applyLabelsAndAnnotations(namespace, environment)

		// Update the namespace
		if err := r.Update(ctx, namespace); err != nil {
			logger.Warn("Failed to update namespace labels/annotations", zap.Error(err))
		}

		// Update environment status based on activation state
		// The actual scaling is handled by the composition controller
		desiredState := environmentsv1.EnvironmentStateInactive
		if environment.Spec.Activated {
			desiredState = environmentsv1.EnvironmentStateActive
		}

		// Detect deactivation transition (from active to inactive)
		wasActive := environment.Status.State == environmentsv1.EnvironmentStateActive
		willBeInactive := desiredState == environmentsv1.EnvironmentStateInactive

		if wasActive && willBeInactive {
			// Environment is being deactivated - disconnect workspaces and remove service intercepts
			logger.Info("Environment is being deactivated, cleaning up connections")
			if err := r.handleEnvironmentDeactivation(ctx, environment, logger); err != nil {
				logger.Error("Failed to complete environment deactivation cleanup", zap.Error(err))
				// Continue with status update even if cleanup partially fails
			}
		}

		if environment.Status.State != desiredState {
			message := "Environment is inactive"
			if desiredState == environmentsv1.EnvironmentStateActive {
				message = "Environment is active"
			}

			if err := r.updateEnvironmentStatus(ctx, environment, desiredState, message, logger); err != nil {
				logger.Error("Failed to update environment status after retries", zap.Error(err))
			}
		}

		return reconcile.Result{}, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check existing namespace", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Create namespace for the Environment
	logger.Info("Creating namespace for environment",
		zap.String("namespace", environment.Spec.TargetNamespace))

	namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: environment.Spec.TargetNamespace,
			Labels: map[string]string{
				"kloudlite.io/environment": environment.Name,
			},
			Annotations: map[string]string{
				"kloudlite.io/environment-uid": string(environment.UID),
				"kloudlite.io/creation-reason": "auto-created-for-environment",
				"kloudlite.io/created-by":      environment.Spec.OwnedBy, // Move email to annotations
			},
		},
	}

	// Apply labels and annotations using helper function
	r.applyLabelsAndAnnotations(namespace, environment)

	// Set Environment as the owner of the Namespace using OwnerReferences
	// Note: This might not work for namespace as it's cluster-scoped
	// and Environment is also cluster-scoped. Owner references typically
	// work for namespace-scoped resources owned by cluster-scoped resources.
	if err := controllerutil.SetControllerReference(environment, namespace, r.Scheme); err != nil {
		logger.Warn("Failed to set owner reference on namespace (expected for cluster-scoped resources)", zap.Error(err))
		// Continue anyway, as this is expected behavior for cluster-scoped resources
	}

	// Create the namespace
	if err := r.Create(ctx, namespace); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Another reconciliation might have created it
			logger.Info("Namespace already exists (race condition)")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to create namespace", zap.Error(err))
		// Retry after a delay
		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	logger.Info("Successfully created namespace for environment",
		zap.String("namespace", environment.Spec.TargetNamespace))

	// Update environment status to indicate namespace has been created
	desiredState := environmentsv1.EnvironmentStateInactive
	if environment.Spec.Activated {
		desiredState = environmentsv1.EnvironmentStateActive
	}

	// Update status with retry logic
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		environment.Status.State = desiredState
		environment.Status.Message = "Namespace created successfully"

		// Add condition for namespace creation
		r.addOrUpdateCondition(environment, environmentsv1.EnvironmentConditionNamespaceCreated, metav1.ConditionTrue, "NamespaceCreated", "Namespace has been created successfully")

		return nil
	}, logger); err != nil {
		logger.Error("Failed to update environment status after retries", zap.Error(err))
		// Don't fail the reconciliation for status update failures
	}

	return reconcile.Result{}, nil
}

// handleDeletion handles the deletion of an environment and its child resources
func (r *EnvironmentReconciler) handleDeletion(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	// Update status to show deletion in progress
	if environment.Status.State != environmentsv1.EnvironmentStateDeleting {
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateDeleting, "Deleting environment and cleaning up resources", logger); err != nil {
			logger.Error("Failed to update environment status to deleting after retries", zap.Error(err))
			// Continue with deletion even if status update fails
		}
	}

	// Clean up workspace environment connections referencing this environment
	if err := r.cleanupWorkspaceConnections(ctx, environment, logger); err != nil {
		logger.Error("Failed to cleanup workspace connections", zap.Error(err))
		// Continue with deletion even if cleanup fails
	}

	// Check if namespace exists
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: environment.Spec.TargetNamespace}, namespace)

	if err == nil {
		// Namespace exists, delete it
		logger.Info("Deleting namespace for environment",
			zap.String("namespace", environment.Spec.TargetNamespace))

		if err := r.Delete(ctx, namespace); err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete namespace", zap.Error(err))
				return reconcile.Result{RequeueAfter: 5 * time.Second}, err
			}
		}

		// Requeue to wait for namespace deletion to complete
		logger.Info("Waiting for namespace deletion to complete")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check namespace", zap.Error(err))
		return reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	// Namespace is deleted, remove finalizer
	logger.Info("Namespace deleted, removing finalizer from environment")

	if controllerutil.ContainsFinalizer(environment, environmentFinalizer) {
		controllerutil.RemoveFinalizer(environment, environmentFinalizer)
		if err := r.Update(ctx, environment); err != nil {
			logger.Error("Failed to remove finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
	}

	logger.Info("Environment cleanup completed successfully")
	return reconcile.Result{}, nil
}

// handleCloning handles cloning resources from a source environment including PVCs
func (r *EnvironmentReconciler) handleCloning(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	sourceName := environment.Spec.CloneFrom

	logger.Info("Starting environment cloning process",
		zap.String("target", environment.Name),
		zap.String("source", sourceName))

	// Initialize cloning status if not already set
	if environment.Status.CloningStatus == nil {
		now := metav1.Now()
		environment.Status.CloningStatus = &environmentsv1.CloningStatus{
			Phase:     environmentsv1.CloningPhasePending,
			Message:   "Initializing cloning process",
			StartTime: &now,
		}
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			return nil
		}, logger); err != nil {
			logger.Error("Failed to initialize cloning status", zap.Error(err))
		}
	}

	// Validate source environment exists and is accessible
	sourceEnv := &environmentsv1.Environment{}
	err := r.Get(ctx, client.ObjectKey{Name: sourceName}, sourceEnv)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error("Source environment not found", zap.String("source", sourceName))
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, "Source environment not found: "+sourceName, logger)
			if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError, "Source environment not found: "+sourceName, logger); err != nil {
				logger.Error("Failed to update environment status after retries", zap.Error(err))
			}
			return reconcile.Result{}, fmt.Errorf("source environment '%s' not found", sourceName)
		}
		logger.Error("Failed to get source environment", zap.Error(err))
		return reconcile.Result{}, fmt.Errorf("failed to access source environment '%s': %w", sourceName, err)
	}

	// Validate source environment state
	if sourceEnv.Status.State == environmentsv1.EnvironmentStateDeleting || sourceEnv.Status.State == environmentsv1.EnvironmentStateError {
		logger.Error("Source environment is not in a clonable state",
			zap.String("source", sourceName),
			zap.String("sourceState", string(sourceEnv.Status.State)))
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed,
			fmt.Sprintf("Source environment is in %s state", sourceEnv.Status.State), logger)
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError,
			fmt.Sprintf("Source environment '%s' is in %s state and cannot be cloned", sourceName, sourceEnv.Status.State), logger); err != nil {
			logger.Error("Failed to update environment status after retries", zap.Error(err))
		}
		return reconcile.Result{}, fmt.Errorf("source environment '%s' is in %s state", sourceName, sourceEnv.Status.State)
	}

	sourceNamespace := sourceEnv.Spec.TargetNamespace
	targetNamespace := environment.Spec.TargetNamespace

	logger.Info("Cloning environment resources",
		zap.String("source", sourceName),
		zap.String("sourceNamespace", sourceNamespace),
		zap.String("targetNamespace", targetNamespace))

	// Phase 1: Suspend source environment if it's active
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhasePending {
		if sourceEnv.Spec.Activated {
			logger.Info("Suspending source environment for safe cloning")
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseSuspending, "Suspending source environment", logger)

			// Set source cloning status
			now := metav1.Now()
			sourceEnv.Status.SourceCloningStatus = &environmentsv1.SourceCloningStatus{
				TargetEnvironmentName: environment.Name,
				Phase:                 environmentsv1.SourceCloningPhaseSuspended,
				Message:               fmt.Sprintf("Environment suspended for cloning to %s", environment.Name),
				StartTime:             &now,
			}
			if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, sourceEnv, func() error {
				return nil
			}, logger); err != nil {
				logger.Error("Failed to update source cloning status", zap.Error(err))
			}

			// Scale down source environment
			if err := r.suspendEnvironment(ctx, sourceEnv, logger); err != nil {
				logger.Error("Failed to suspend source environment", zap.Error(err))
				r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, "Failed to suspend source environment", logger)
				return reconcile.Result{RequeueAfter: 10 * time.Second}, err
			}
			logger.Info("Source environment suspended successfully")
		}

		// Move to next phase
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCloningResources, "Starting resource cloning", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Note: TargetNamespace is always set by the mutation webhook for the cloned environment.
	// Controller creates the actual Kubernetes namespace resource if it doesn't exist.

	// Create target namespace if it doesn't exist
	namespace := &corev1.Namespace{}
	err = r.Get(ctx, client.ObjectKey{Name: targetNamespace}, namespace)
	if apierrors.IsNotFound(err) {
		logger.Info("Creating namespace for cloned environment", zap.String("namespace", targetNamespace))

		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: targetNamespace,
				Labels: map[string]string{
					"kloudlite.io/environment": environment.Name,
				},
				Annotations: map[string]string{
					"kloudlite.io/environment-uid": string(environment.UID),
					"kloudlite.io/creation-reason": "auto-created-for-cloned-environment",
					"kloudlite.io/created-by":      environment.Spec.OwnedBy,
					"kloudlite.io/cloned-from":     sourceName,
				},
			},
		}

		// Apply labels and annotations using helper function
		r.applyLabelsAndAnnotations(namespace, environment)

		// Add cloning-specific annotations
		if namespace.Annotations == nil {
			namespace.Annotations = make(map[string]string)
		}
		namespace.Annotations["kloudlite.io/cloned-from"] = sourceName
		namespace.Annotations["kloudlite.io/creation-reason"] = "auto-created-for-cloned-environment"

		if err := r.Create(ctx, namespace); err != nil && !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create namespace for cloned environment", zap.Error(err))
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, fmt.Sprintf("Failed to create namespace: %v", err), logger)
			if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError,
				fmt.Sprintf("Failed to create namespace: %v", err), logger); err != nil {
				logger.Error("Failed to update environment status after retries", zap.Error(err))
			}
			return reconcile.Result{RequeueAfter: 30 * time.Second}, err
		}
		logger.Info("Successfully created namespace for cloned environment", zap.String("namespace", targetNamespace))
	}

	// Clone ConfigMaps with label "kloudlite.io/resource-type: environment-config"
	logger.Info("Cloning ConfigMaps from source environment")
	configMapList := &corev1.ConfigMapList{}
	err = r.List(ctx, configMapList,
		client.InNamespace(sourceNamespace),
		client.MatchingLabels{"kloudlite.io/resource-type": "environment-config"})
	if err != nil {
		logger.Error("Failed to list source configmaps", zap.Error(err))
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError,
			fmt.Sprintf("Failed to list ConfigMaps from source environment '%s': %v", sourceName, err), logger); err != nil {
			logger.Error("Failed to update environment status after retries", zap.Error(err))
		}
		return reconcile.Result{}, fmt.Errorf("failed to list source ConfigMaps: %w", err)
	}

	clonedConfigMaps := 0
	for _, srcCM := range configMapList.Items {
		newCM := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      srcCM.Name,
				Namespace: targetNamespace,
				Labels:    srcCM.Labels,
			},
			Data: srcCM.Data,
		}

		// Update the environment label
		if newCM.Labels == nil {
			newCM.Labels = make(map[string]string)
		}
		newCM.Labels["kloudlite.io/environment"] = environment.Name

		if err := r.Create(ctx, newCM); err != nil && !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to clone configmap",
				zap.String("name", srcCM.Name),
				zap.Error(err))
			// Continue with other resources instead of failing completely
			continue
		}
		clonedConfigMaps++
		logger.Debug("Cloned configmap", zap.String("name", srcCM.Name))
	}
	logger.Info("ConfigMap cloning completed", zap.Int("cloned", clonedConfigMaps), zap.Int("total", len(configMapList.Items)))

	// Clone Secrets with label "kloudlite.io/resource-type: environment-config"
	logger.Info("Cloning Secrets from source environment")
	secretList := &corev1.SecretList{}
	err = r.List(ctx, secretList,
		client.InNamespace(sourceNamespace),
		client.MatchingLabels{"kloudlite.io/resource-type": "environment-config"})
	if err != nil {
		logger.Error("Failed to list source secrets", zap.Error(err))
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError,
			fmt.Sprintf("Failed to list Secrets from source environment '%s': %v", sourceName, err), logger); err != nil {
			logger.Error("Failed to update environment status after retries", zap.Error(err))
		}
		return reconcile.Result{}, fmt.Errorf("failed to list source Secrets: %w", err)
	}

	clonedSecrets := 0
	for _, srcSecret := range secretList.Items {
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      srcSecret.Name,
				Namespace: targetNamespace,
				Labels:    srcSecret.Labels,
			},
			Type: srcSecret.Type,
			Data: srcSecret.Data,
		}

		// Update the environment label
		if newSecret.Labels == nil {
			newSecret.Labels = make(map[string]string)
		}
		newSecret.Labels["kloudlite.io/environment"] = environment.Name

		if err := r.Create(ctx, newSecret); err != nil && !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to clone secret",
				zap.String("name", srcSecret.Name),
				zap.Error(err))
			// Continue with other resources instead of failing completely
			continue
		}
		clonedSecrets++
		logger.Debug("Cloned secret", zap.String("name", srcSecret.Name))
	}
	logger.Info("Secret cloning completed", zap.Int("cloned", clonedSecrets), zap.Int("total", len(secretList.Items)))

	// Transition to PVC cloning phase after completing resource cloning
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseCloningResources {
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCloningPVCs, "Starting PVC cloning", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Phase 3: Clone PVCs (create empty PVCs in target namespace)
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseCloningPVCs {
		logger.Info("Creating PVCs in target namespace")

		// List PVCs from source namespace with kloudlite.io/managed label
		pvcList := &corev1.PersistentVolumeClaimList{}
		err = r.List(ctx, pvcList,
			client.InNamespace(sourceNamespace),
			client.MatchingLabels{"kloudlite.io/managed": "true"})
		if err != nil {
			logger.Error("Failed to list source PVCs", zap.Error(err))
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, fmt.Sprintf("Failed to list PVCs: %v", err), logger)
			return reconcile.Result{}, fmt.Errorf("failed to list source PVCs: %w", err)
		}

		// Update total PVC count
		environment.Status.CloningStatus.TotalPVCs = int32(len(pvcList.Items))
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update PVC count", zap.Error(err))
		}

		// Create empty PVCs in target namespace (data will be copied later)
		for _, srcPVC := range pvcList.Items {
			newPVC := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      srcPVC.Name,
					Namespace: targetNamespace,
					Labels:    srcPVC.Labels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes:      srcPVC.Spec.AccessModes,
					StorageClassName: srcPVC.Spec.StorageClassName,
					Resources:        srcPVC.Spec.Resources,
				},
			}

			// Update environment label
			if newPVC.Labels == nil {
				newPVC.Labels = make(map[string]string)
			}
			newPVC.Labels["kloudlite.io/environment"] = environment.Name

			if err := r.Create(ctx, newPVC); err != nil && !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to create PVC",
					zap.String("name", srcPVC.Name),
					zap.Error(err))
				// Track failed PVC
				environment.Status.CloningStatus.FailedPVCs = append(environment.Status.CloningStatus.FailedPVCs, srcPVC.Name)
				continue
			}
			logger.Debug("Created PVC", zap.String("name", srcPVC.Name))
		}

		// Move to creating copy jobs phase
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCreatingCopyJobs, "Creating copy jobs", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Phase 4: Create ALL copy jobs (sender + receiver)
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseCreatingCopyJobs {
		logger.Info("Creating all PVC copy jobs")

		// List PVCs from source namespace
		pvcList := &corev1.PersistentVolumeClaimList{}
		err = r.List(ctx, pvcList,
			client.InNamespace(sourceNamespace),
			client.MatchingLabels{"kloudlite.io/managed": "true"})
		if err != nil {
			logger.Error("Failed to list source PVCs", zap.Error(err))
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed, fmt.Sprintf("Failed to list PVCs: %v", err), logger)
			return reconcile.Result{}, fmt.Errorf("failed to list source PVCs: %w", err)
		}

		if len(pvcList.Items) == 0 {
			logger.Info("No PVCs to clone, moving to resuming phase")
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseResuming, "No PVCs to clone", logger)
			return reconcile.Result{Requeue: true}, nil
		}

		// Initialize copy jobs status tracking
		copyJobsStatus := make([]environmentsv1.PVCCopyJobStatus, 0, len(pvcList.Items))
		copier := NewPVCCopier(r.Client, sourceNamespace, targetNamespace)
		now := metav1.Now()

		// Create sender and receiver jobs for ALL PVCs
		for _, srcPVC := range pvcList.Items {
			jobStatus := environmentsv1.PVCCopyJobStatus{
				PVCName:         srcPVC.Name,
				Phase:           "Creating",
				SenderJobName:   fmt.Sprintf("pvc-copy-sender-%s", srcPVC.Name),
				ReceiverJobName: fmt.Sprintf("pvc-copy-receiver-%s", srcPVC.Name),
				StartTime:       &now,
			}

			logger.Info("Creating copy jobs for PVC", zap.String("pvc", srcPVC.Name))
			if err := copier.CopyPVC(ctx, srcPVC.Name, srcPVC.Name, environment); err != nil {
				logger.Error("Failed to create copy jobs", zap.String("pvc", srcPVC.Name), zap.Error(err))
				jobStatus.Phase = "Failed"
				jobStatus.ErrorMessage = err.Error()
				environment.Status.CloningStatus.FailedPVCs = append(environment.Status.CloningStatus.FailedPVCs, srcPVC.Name)
			} else {
				jobStatus.Phase = "Copying"
			}
			copyJobsStatus = append(copyJobsStatus, jobStatus)
		}

		// Update status with copy jobs info
		environment.Status.CloningStatus.CopyJobsStatus = copyJobsStatus
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update copy jobs status", zap.Error(err))
		}

		// Update source environment to Copying phase
		sourceEnv := &environmentsv1.Environment{}
		if err := r.Get(ctx, client.ObjectKey{Name: sourceName}, sourceEnv); err == nil {
			if sourceEnv.Status.SourceCloningStatus != nil {
				sourceEnv.Status.SourceCloningStatus.Phase = environmentsv1.SourceCloningPhaseCopying
				sourceEnv.Status.SourceCloningStatus.Message = "PVC data is being copied"
				if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, sourceEnv, func() error {
					return nil
				}, logger); err != nil {
					logger.Error("Failed to update source copying status", zap.Error(err))
				}
			}
		}

		logger.Info("All copy jobs created", zap.Int("total", len(copyJobsStatus)))
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseWaitingForCopyCompletion,
			fmt.Sprintf("Waiting for %d PVC copies to complete", len(copyJobsStatus)), logger)
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// Phase 5: Wait for ALL copy jobs to complete
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseWaitingForCopyCompletion {
		logger.Info("Checking status of all PVC copy jobs")

		copier := NewPVCCopier(r.Client, sourceNamespace, targetNamespace)
		allCompleted := true
		anyFailed := false
		completedCount := int32(0)
		now := metav1.Now()

		// Check status of ALL copy jobs
		for i := range environment.Status.CloningStatus.CopyJobsStatus {
			jobStatus := &environment.Status.CloningStatus.CopyJobsStatus[i]

			if jobStatus.Phase == "Completed" {
				completedCount++
				continue
			}

			if jobStatus.Phase == "Failed" {
				anyFailed = true
				continue
			}

			// Check receiver job status
			completed, failed, err := copier.GetCopyStatus(ctx, jobStatus.PVCName)
			if err != nil {
				logger.Warn("Failed to check copy status", zap.String("pvc", jobStatus.PVCName), zap.Error(err))
				allCompleted = false
				continue
			}

			if completed {
				jobStatus.Phase = "Completed"
				jobStatus.CompletionTime = &now
				completedCount++
				logger.Info("PVC copy completed", zap.String("pvc", jobStatus.PVCName))
			} else if failed {
				jobStatus.Phase = "Failed"
				jobStatus.ErrorMessage = "Copy job failed"
				anyFailed = true
				environment.Status.CloningStatus.FailedPVCs = append(environment.Status.CloningStatus.FailedPVCs, jobStatus.PVCName)
				logger.Error("PVC copy failed", zap.String("pvc", jobStatus.PVCName))
			} else {
				allCompleted = false
				jobStatus.Phase = "Copying"
			}
		}

		// Update cloned PVCs count
		environment.Status.CloningStatus.ClonedPVCs = completedCount

		// Update status
		if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
			return nil
		}, logger); err != nil {
			logger.Error("Failed to update copy status", zap.Error(err))
		}

		if anyFailed && allCompleted {
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed,
				fmt.Sprintf("Some PVC copies failed: %v", environment.Status.CloningStatus.FailedPVCs), logger)
			return reconcile.Result{}, fmt.Errorf("some PVC copies failed")
		}

		if !allCompleted {
			logger.Info("Waiting for PVC copies to complete",
				zap.Int32("completed", completedCount),
				zap.Int32("total", environment.Status.CloningStatus.TotalPVCs))
			return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
		}

		// ALL copies completed successfully!
		logger.Info("All PVC copies completed successfully",
			zap.Int32("total", completedCount))
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseVerifyingCopies, "Verifying copies", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Phase 6: Verify copies and cleanup
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseVerifyingCopies {
		logger.Info("Verifying all PVC copies")

		// Cleanup completed copy jobs
		copier := NewPVCCopier(r.Client, sourceNamespace, targetNamespace)
		for _, jobStatus := range environment.Status.CloningStatus.CopyJobsStatus {
			if jobStatus.Phase == "Completed" {
				if err := copier.CleanupCopyJobs(ctx, jobStatus.PVCName); err != nil {
					logger.Warn("Failed to cleanup copy jobs",
						zap.String("pvc", jobStatus.PVCName),
						zap.Error(err))
				}
			}
		}

		logger.Info("Verification complete, proceeding to clone Compositions")
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCloningCompositions, "Cloning Compositions to destination", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Phase 6.5: Clone Compositions (after PVC data is copied)
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseCloningCompositions {
		logger.Info("Cloning Compositions from source environment to destination (after PVC data is ready)")

		sourceName := environment.Spec.CloneFrom
		sourceNamespace := fmt.Sprintf("env-%s", sourceName)
		targetNamespace := environment.Spec.TargetNamespace

		// Clone Compositions
		compositionList := &environmentsv1.CompositionList{}
		err := r.List(ctx, compositionList, client.InNamespace(sourceNamespace))
		if err != nil {
			logger.Error("Failed to list source compositions", zap.Error(err))
			r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseFailed,
				fmt.Sprintf("Failed to list Compositions from source environment '%s': %v", sourceName, err), logger)
			return reconcile.Result{}, fmt.Errorf("failed to list source Compositions: %w", err)
		}

		clonedCompositions := 0
		for _, srcComp := range compositionList.Items {
			newComp := &environmentsv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Name:        srcComp.Name,
					Namespace:   targetNamespace,
					Labels:      srcComp.Labels,
					Annotations: srcComp.Annotations,
				},
				Spec: srcComp.Spec,
			}

			// Override NodeName with target environment's NodeName
			// This ensures composition pods run on the correct node where target PVCs are bound
			newComp.Spec.NodeName = environment.Spec.NodeName

			// Update the environment label
			if newComp.Labels == nil {
				newComp.Labels = make(map[string]string)
			}
			newComp.Labels["kloudlite.io/environment"] = environment.Name

			if err := r.Create(ctx, newComp); err != nil && !apierrors.IsAlreadyExists(err) {
				logger.Error("Failed to clone composition",
					zap.String("name", srcComp.Name),
					zap.Error(err))
				// Continue with other resources instead of failing completely
				continue
			}
			clonedCompositions++
			logger.Debug("Cloned composition", zap.String("name", srcComp.Name))
		}
		logger.Info("Composition cloning completed", zap.Int("cloned", clonedCompositions), zap.Int("total", len(compositionList.Items)))

		// Transition to Resuming phase
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseResuming, "Resuming source environment", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Phase 7: Resume source environment
	if environment.Status.CloningStatus.Phase == environmentsv1.CloningPhaseResuming {
		logger.Info("Resuming source environment")

		// Reload source environment
		sourceEnv := &environmentsv1.Environment{}
		if err := r.Get(ctx, client.ObjectKey{Name: sourceName}, sourceEnv); err != nil {
			logger.Error("Failed to get source environment for resuming", zap.Error(err))
		} else {
			// Update source cloning status to Resuming
			if sourceEnv.Status.SourceCloningStatus != nil {
				sourceEnv.Status.SourceCloningStatus.Phase = environmentsv1.SourceCloningPhaseResuming
				sourceEnv.Status.SourceCloningStatus.Message = "Resuming after cloning completed"
				if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, sourceEnv, func() error {
					return nil
				}, logger); err != nil {
					logger.Error("Failed to update source resuming status", zap.Error(err))
				}
			}

			// Scale up deployments back to original replica counts
			if err := r.resumeEnvironment(ctx, sourceEnv, logger); err != nil {
				logger.Error("Failed to resume source environment", zap.Error(err))
			} else {
				logger.Info("Source environment resumed successfully")

				// Clear source cloning status after successful resume
				if sourceEnv.Status.SourceCloningStatus != nil {
					sourceEnv.Status.SourceCloningStatus = nil
					if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, sourceEnv, func() error {
						return nil
					}, logger); err != nil {
						logger.Error("Failed to clear source cloning status", zap.Error(err))
					}
				}
			}
		}

		// Move to completed phase
		r.updateCloningStatus(ctx, environment, environmentsv1.CloningPhaseCompleted, "Cloning completed successfully", logger)
		return reconcile.Result{Requeue: true}, nil
	}

	// Only execute completion logic when phase is Completed
	if environment.Status.CloningStatus.Phase != environmentsv1.CloningPhaseCompleted {
		return reconcile.Result{}, nil
	}

	// Prepare cloning completion message with statistics
	sourceName = environment.Spec.CloneFrom
	successMessage := fmt.Sprintf("Successfully cloned environment from %s", sourceName)

	if environment.Status.CloningStatus.TotalPVCs > 0 {
		successMessage = fmt.Sprintf("Successfully cloned environment from %s (PVCs: %d/%d)",
			sourceName, environment.Status.CloningStatus.ClonedPVCs, environment.Status.CloningStatus.TotalPVCs)
	}

	// Update status to indicate cloning is complete
	desiredState := environmentsv1.EnvironmentStateInactive
	if environment.Spec.Activated {
		desiredState = environmentsv1.EnvironmentStateActive
	}

	if err := r.updateEnvironmentStatus(ctx, environment, desiredState, successMessage, logger); err != nil {
		logger.Error("Failed to update environment status after cloning, even after retries", zap.Error(err))
	}

	// Update status with condition for successful cloning
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		r.addOrUpdateCondition(environment, environmentsv1.EnvironmentConditionCloned, metav1.ConditionTrue, "CloningSuccessful", successMessage)
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update environment conditions after retries", zap.Error(err))
	}

	// Mark completion time
	now := metav1.Now()
	environment.Status.CloningStatus.CompletionTime = &now
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update completion time", zap.Error(err))
	}

	// Clear the CloneFrom field to mark cloning as complete
	environment.Spec.CloneFrom = ""
	if err := r.Update(ctx, environment); err != nil {
		logger.Error("Failed to clear CloneFrom field", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Environment cloning completed successfully",
		zap.String("source", sourceName),
		zap.Int32("clonedPVCs", environment.Status.CloningStatus.ClonedPVCs),
		zap.Int32("totalPVCs", environment.Status.CloningStatus.TotalPVCs))

	return reconcile.Result{Requeue: true}, nil
}

// cleanupWorkspaceConnections removes environment connections from all workspaces referencing this environment
func (r *EnvironmentReconciler) cleanupWorkspaceConnections(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Get Workspace type to list workspaces
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(ctx, workspaceList); err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	environmentName := environment.Name
	environmentNamespace := environment.Namespace
	if environmentNamespace == "" {
		environmentNamespace = "default"
	}

	cleanedCount := 0
	for i := range workspaceList.Items {
		workspace := &workspaceList.Items[i]

		// Check if this workspace references the environment being deleted
		if workspace.Spec.EnvironmentConnection == nil {
			continue
		}

		envRef := workspace.Spec.EnvironmentConnection.EnvironmentRef
		if envRef.Name == environmentName && envRef.Namespace == environmentNamespace {
			logger.Info("Removing environment connection from workspace",
				zap.String("workspace", workspace.Name),
				zap.String("environment", environmentName))

			// Remove the environment connection
			workspace.Spec.EnvironmentConnection = nil

			if err := r.Update(ctx, workspace); err != nil {
				logger.Error("Failed to remove environment connection from workspace",
					zap.String("workspace", workspace.Name),
					zap.Error(err))
				// Continue with other workspaces even if one fails
				continue
			}
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		logger.Info("Cleaned up workspace environment connections",
			zap.Int("count", cleanedCount))
	}

	return nil
}

// updateCloningStatus updates the cloning status phase and message
func (r *EnvironmentReconciler) updateCloningStatus(ctx context.Context, environment *environmentsv1.Environment, phase environmentsv1.CloningPhase, message string, logger *zap.Logger) {
	if environment.Status.CloningStatus == nil {
		now := metav1.Now()
		environment.Status.CloningStatus = &environmentsv1.CloningStatus{
			StartTime: &now,
		}
	}

	environment.Status.CloningStatus.Phase = phase
	environment.Status.CloningStatus.Message = message

	// Set completion time if phase is completed or failed
	if phase == environmentsv1.CloningPhaseCompleted || phase == environmentsv1.CloningPhaseFailed {
		now := metav1.Now()
		environment.Status.CloningStatus.CompletionTime = &now
	}

	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, environment, func() error {
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update cloning status", zap.Error(err))
	}
}

// suspendEnvironment scales down all deployments in the environment
// It stores the original replica count in annotations for later resumption
func (r *EnvironmentReconciler) suspendEnvironment(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	namespace := environment.Spec.TargetNamespace
	const originalReplicasAnnotation = "kloudlite.io/original-replicas"

	// Scale down deployments
	deployments := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployments, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, dep := range deployments.Items {
		if dep.Spec.Replicas != nil && *dep.Spec.Replicas > 0 {
			// Store original replica count in annotation
			if dep.Annotations == nil {
				dep.Annotations = make(map[string]string)
			}
			if _, exists := dep.Annotations[originalReplicasAnnotation]; !exists {
				dep.Annotations[originalReplicasAnnotation] = fmt.Sprintf("%d", *dep.Spec.Replicas)
			}

			zero := int32(0)
			dep.Spec.Replicas = &zero
			if err := r.Update(ctx, &dep); err != nil {
				logger.Error("Failed to scale down deployment", zap.String("deployment", dep.Name), zap.Error(err))
			} else {
				logger.Debug("Scaled down deployment", zap.String("deployment", dep.Name))
			}
		}
	}

	return nil
}

// resumeEnvironment scales up deployments to their original replica counts
func (r *EnvironmentReconciler) resumeEnvironment(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	namespace := environment.Spec.TargetNamespace
	const originalReplicasAnnotation = "kloudlite.io/original-replicas"

	// Scale up deployments
	deployments := &appsv1.DeploymentList{}
	if err := r.List(ctx, deployments, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	for _, dep := range deployments.Items {
		if dep.Annotations != nil {
			if originalReplicasStr, exists := dep.Annotations[originalReplicasAnnotation]; exists {
				var originalReplicas int32
				if _, err := fmt.Sscanf(originalReplicasStr, "%d", &originalReplicas); err == nil && originalReplicas > 0 {
					dep.Spec.Replicas = &originalReplicas
					// Remove the annotation after restoring
					delete(dep.Annotations, originalReplicasAnnotation)

					if err := r.Update(ctx, &dep); err != nil {
						logger.Error("Failed to scale up deployment", zap.String("deployment", dep.Name), zap.Error(err))
					} else {
						logger.Debug("Scaled up deployment", zap.String("deployment", dep.Name), zap.Int32("replicas", originalReplicas))
					}
				}
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.Environment{}).
		Owns(&corev1.Namespace{}). // Watch Namespaces owned by Environments
		Complete(r)
	// Note: We don't watch WorkMachine here because Environment references WorkMachine by name
	// The Environment controller will handle WorkMachine ownership during reconciliation
}
