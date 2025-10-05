package controllers

import (
	"context"
	"time"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
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

// Reconcile handles Environment events and ensures namespace exists
func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("environment", req.Name),
	)

	logger.Info("Reconciling Environment")

	// Fetch the Environment instance
	environment := &environmentsv1.Environment{}
	err := r.Get(ctx, req.NamespacedName, environment)
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

	// Check if cloning is requested
	if environment.Spec.CloneFrom != "" {
		logger.Info("Cloning requested from source environment",
			zap.String("source", environment.Spec.CloneFrom))
		return r.handleCloning(ctx, environment, logger)
	}

	// Check if namespace already exists
	namespace := &corev1.Namespace{}
	err = r.Get(ctx, client.ObjectKey{Name: environment.Spec.TargetNamespace}, namespace)

	if err == nil {
		// Namespace already exists
		logger.Info("Namespace already exists for environment",
			zap.String("namespace", environment.Spec.TargetNamespace))

		// Update namespace labels if needed
		if namespace.Labels == nil {
			namespace.Labels = make(map[string]string)
		}

		// Add environment labels
		namespace.Labels["kloudlite.io/environment"] = environment.Name

		// Update namespace annotations if needed
		if namespace.Annotations == nil {
			namespace.Annotations = make(map[string]string)
		}

		// Add createdBy to annotations (emails contain invalid label characters)
		namespace.Annotations["kloudlite.io/created-by"] = environment.Spec.CreatedBy

		// Add any custom labels from the environment spec to annotations
		// (they may contain invalid label characters like @, =, etc.)
		if environment.Spec.Labels != nil {
			for k, v := range environment.Spec.Labels {
				namespace.Annotations[k] = v
			}
		}

		// Add any custom annotations from the environment spec
		if environment.Spec.Annotations != nil {
			for k, v := range environment.Spec.Annotations {
				namespace.Annotations[k] = v
			}
		}

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

		if environment.Status.State != desiredState {
			environment.Status.State = desiredState
			if desiredState == environmentsv1.EnvironmentStateActive {
				environment.Status.Message = "Environment is active"
				now := metav1.Now()
				environment.Status.LastActivatedTime = &now
			} else {
				environment.Status.Message = "Environment is inactive"
				now := metav1.Now()
				environment.Status.LastDeactivatedTime = &now
			}

			if err := r.Status().Update(ctx, environment); err != nil {
				logger.Warn("Failed to update environment status", zap.Error(err))
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

	// Add custom labels from environment spec (skip invalid label values)
	if environment.Spec.Labels != nil {
		for k, v := range environment.Spec.Labels {
			// Only add valid label values (no @, =, etc.)
			// Labels with emails or special chars should go to annotations
			namespace.Annotations[k] = v
		}
	}

	// Add custom annotations from environment spec
	if environment.Spec.Annotations != nil {
		for k, v := range environment.Spec.Annotations {
			namespace.Annotations[k] = v
		}
	}

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
	environment.Status.State = environmentsv1.EnvironmentStateInactive
	if environment.Spec.Activated {
		environment.Status.State = environmentsv1.EnvironmentStateActive
	}
	environment.Status.Message = "Namespace created successfully"

	// Add condition for namespace creation
	condition := environmentsv1.EnvironmentCondition{
		Type:               environmentsv1.EnvironmentConditionNamespaceCreated,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: &metav1.Time{Time: time.Now()},
		Reason:             "NamespaceCreated",
		Message:            "Namespace has been created successfully",
	}

	// Initialize conditions if nil
	if environment.Status.Conditions == nil {
		environment.Status.Conditions = []environmentsv1.EnvironmentCondition{}
	}

	// Add or update the condition
	found := false
	for i, c := range environment.Status.Conditions {
		if c.Type == environmentsv1.EnvironmentConditionNamespaceCreated {
			environment.Status.Conditions[i] = condition
			found = true
			break
		}
	}
	if !found {
		environment.Status.Conditions = append(environment.Status.Conditions, condition)
	}

	// Update the environment status
	if err := r.Status().Update(ctx, environment); err != nil {
		logger.Warn("Failed to update environment status", zap.Error(err))
		// Don't fail the reconciliation for status update failures
	}

	return reconcile.Result{}, nil
}

// handleDeletion handles the deletion of an environment and its child resources
func (r *EnvironmentReconciler) handleDeletion(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (reconcile.Result, error) {
	// Update status to show deletion in progress
	if environment.Status.State != environmentsv1.EnvironmentStateDeleting {
		environment.Status.State = environmentsv1.EnvironmentStateDeleting
		environment.Status.Message = "Deleting environment and cleaning up resources"

		if err := r.Status().Update(ctx, environment); err != nil {
			logger.Warn("Failed to update environment status to deleting", zap.Error(err))
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

	// Fetch the source environment
	sourceEnv := &environmentsv1.Environment{}
	err := r.Get(ctx, client.ObjectKey{Name: sourceName}, sourceEnv)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error("Source environment not found", zap.String("source", sourceName))
			environment.Status.State = environmentsv1.EnvironmentStateError
			environment.Status.Message = "Source environment not found: " + sourceName
			if err := r.Status().Update(ctx, environment); err != nil {
				logger.Warn("Failed to update environment status", zap.Error(err))
			}
			return reconcile.Result{}, err
		}
		logger.Error("Failed to get source environment", zap.Error(err))
		return reconcile.Result{}, err
	}

	sourceNamespace := sourceEnv.Spec.TargetNamespace
	targetNamespace := environment.Spec.TargetNamespace

	logger.Info("Cloning environment resources",
		zap.String("source", sourceName),
		zap.String("sourceNamespace", sourceNamespace),
		zap.String("targetNamespace", targetNamespace))

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

		if err := r.Create(ctx, namespace); err != nil && !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create namespace", zap.Error(err))
			return reconcile.Result{RequeueAfter: 30 * time.Second}, err
		}
	}

	// Clone ConfigMaps with label "kloudlite.io/resource-type: environment-config"
	configMapList := &corev1.ConfigMapList{}
	err = r.List(ctx, configMapList,
		client.InNamespace(sourceNamespace),
		client.MatchingLabels{"kloudlite.io/resource-type": "environment-config"})
	if err != nil {
		logger.Error("Failed to list source configmaps", zap.Error(err))
		return reconcile.Result{}, err
	}

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
			return reconcile.Result{}, err
		}
		logger.Info("Cloned configmap", zap.String("name", srcCM.Name))
	}

	// Clone Secrets with label "kloudlite.io/resource-type: environment-config"
	secretList := &corev1.SecretList{}
	err = r.List(ctx, secretList,
		client.InNamespace(sourceNamespace),
		client.MatchingLabels{"kloudlite.io/resource-type": "environment-config"})
	if err != nil {
		logger.Error("Failed to list source secrets", zap.Error(err))
		return reconcile.Result{}, err
	}

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
			return reconcile.Result{}, err
		}
		logger.Info("Cloned secret", zap.String("name", srcSecret.Name))
	}

	// Clone Compositions
	compositionList := &environmentsv1.CompositionList{}
	err = r.List(ctx, compositionList, client.InNamespace(sourceNamespace))
	if err != nil {
		logger.Error("Failed to list source compositions", zap.Error(err))
		return reconcile.Result{}, err
	}

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

		if err := r.Create(ctx, newComp); err != nil && !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to clone composition",
				zap.String("name", srcComp.Name),
				zap.Error(err))
			return reconcile.Result{}, err
		}
		logger.Info("Cloned composition", zap.String("name", srcComp.Name))
	}

	// Update status to indicate cloning is complete
	environment.Status.State = environmentsv1.EnvironmentStateInactive
	if environment.Spec.Activated {
		environment.Status.State = environmentsv1.EnvironmentStateActive
	}
	environment.Status.Message = "Successfully cloned from " + sourceName

	// Add condition for successful cloning
	condition := environmentsv1.EnvironmentCondition{
		Type:               environmentsv1.EnvironmentConditionCloned,
		Status:             metav1.ConditionTrue,
		LastTransitionTime: &metav1.Time{Time: time.Now()},
		Reason:             "CloningSuccessful",
		Message:            "Resources successfully cloned from " + sourceName,
	}

	// Initialize conditions if nil
	if environment.Status.Conditions == nil {
		environment.Status.Conditions = []environmentsv1.EnvironmentCondition{}
	}

	// Add the condition
	environment.Status.Conditions = append(environment.Status.Conditions, condition)

	// Update the environment status
	if err := r.Status().Update(ctx, environment); err != nil {
		logger.Warn("Failed to update environment status", zap.Error(err))
	}

	// Clear the CloneFrom field to mark cloning as complete
	environment.Spec.CloneFrom = ""
	if err := r.Update(ctx, environment); err != nil {
		logger.Error("Failed to clear CloneFrom field", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Environment cloning completed successfully")
	return reconcile.Result{Requeue: true}, nil
}


// SetupWithManager sets up the controller with the Manager
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&environmentsv1.Environment{}).
		Owns(&corev1.Namespace{}). // Watch Namespaces owned by Environments
		Complete(r)
}
