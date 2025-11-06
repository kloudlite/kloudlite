package environment

import (
	"context"
	"fmt"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	serviceinterceptv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	workmachinevl "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"github.com/kloudlite/kloudlite/api/pkg/utils"
	"go.uber.org/zap"
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
	namespace.Annotations["kloudlite.io/created-by"] = environment.Spec.CreatedBy

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
	deletedIntercepts := 0

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

	// 2. Find and delete all service intercepts targeting services in this environment's namespace
	// ServiceIntercepts are cluster-scoped, so list all and filter by serviceRef.namespace
	logger.Info("Finding service intercepts targeting services in environment namespace",
		zap.String("targetNamespace", environment.Spec.TargetNamespace))

	serviceInterceptList := &serviceinterceptv1.ServiceInterceptList{}
	if err := r.List(ctx, serviceInterceptList); err != nil {
		logger.Error("Failed to list service intercepts", zap.Error(err))
		return fmt.Errorf("failed to list service intercepts: %w", err)
	}

	for i := range serviceInterceptList.Items {
		intercept := &serviceInterceptList.Items[i]

		// Only delete intercepts targeting services in this environment's namespace
		if intercept.Spec.ServiceRef.Namespace != environment.Spec.TargetNamespace {
			continue
		}

		logger.Info("Deleting service intercept",
			zap.String("intercept", intercept.Name),
			zap.String("service", intercept.Spec.ServiceRef.Name),
			zap.String("serviceNamespace", intercept.Spec.ServiceRef.Namespace))

		if err := r.Delete(ctx, intercept); err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete service intercept",
					zap.String("intercept", intercept.Name),
					zap.Error(err))
				// Continue with other intercepts instead of failing
				continue
			}
		}

		deletedIntercepts++
		logger.Info("Successfully deleted service intercept",
			zap.String("intercept", intercept.Name))
	}

	logger.Info("Environment deactivation cleanup completed",
		zap.String("environment", environment.Name),
		zap.Int("disconnectedWorkspaces", disconnectedWorkspaces),
		zap.Int("deletedServiceIntercepts", deletedIntercepts))

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
				"kloudlite.io/created-by":      environment.Spec.CreatedBy, // Move email to annotations
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

// handleCloning handles cloning resources from a source environment
func (r *EnvironmentReconciler) handleCloning(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	sourceName := environment.Spec.CloneFrom

	logger.Info("Starting environment cloning process",
		zap.String("target", environment.Name),
		zap.String("source", sourceName))

	// Validate source environment exists and is accessible
	sourceEnv := &environmentsv1.Environment{}
	err := r.Get(ctx, client.ObjectKey{Name: sourceName}, sourceEnv)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error("Source environment not found", zap.String("source", sourceName))
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
					"kloudlite.io/created-by":      environment.Spec.CreatedBy,
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

	// Clone Compositions
	logger.Info("Cloning Compositions from source environment")
	compositionList := &environmentsv1.CompositionList{}
	err = r.List(ctx, compositionList, client.InNamespace(sourceNamespace))
	if err != nil {
		logger.Error("Failed to list source compositions", zap.Error(err))
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateError,
			fmt.Sprintf("Failed to list Compositions from source environment '%s': %v", sourceName, err), logger); err != nil {
			logger.Error("Failed to update environment status after retries", zap.Error(err))
		}
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

	// Prepare cloning completion message with statistics
	totalResources := len(configMapList.Items) + len(secretList.Items) + len(compositionList.Items)
	clonedResources := clonedConfigMaps + clonedSecrets + clonedCompositions

	successMessage := fmt.Sprintf("Successfully cloned %d/%d resources from %s (ConfigMaps: %d, Secrets: %d, Compositions: %d)",
		clonedResources, totalResources, sourceName, clonedConfigMaps, clonedSecrets, clonedCompositions)

	if clonedResources == 0 {
		successMessage = fmt.Sprintf("No clonable resources found in %s", sourceName)
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

	// Clear the CloneFrom field to mark cloning as complete
	environment.Spec.CloneFrom = ""
	if err := r.Update(ctx, environment); err != nil {
		logger.Error("Failed to clear CloneFrom field", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Environment cloning completed successfully",
		zap.String("source", sourceName),
		zap.Int("clonedResources", clonedResources),
		zap.Int("totalResources", totalResources))

	return reconcile.Result{Requeue: true}, nil
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
