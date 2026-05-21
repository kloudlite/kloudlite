package environment

import (
	"context"
	"fmt"

	"github.com/kloudlite/kloudlite/pkg/utils"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// applyLabelsAndAnnotations applies labels and annotations from environment spec to namespace
func (r *EnvironmentReconciler) applyLabelsAndAnnotations(namespace *corev1.Namespace, environment *environmentsv1.Environment) {
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string)
	}
	if namespace.Annotations == nil {
		namespace.Annotations = make(map[string]string)
	}

	// Add workmachine label for network policy targeting
	if environment.Spec.WorkMachineName != "" {
		namespace.Labels["kloudlite.io/workmachine-name"] = environment.Spec.WorkMachineName
	}

	// Add owned-by label for network policy (sanitized for label value)
	// This enables cross-namespace communication between resources owned by the same user
	if environment.Spec.OwnedBy != "" {
		namespace.Labels["kloudlite.io/owned-by"] = utils.SanitizeForLabel(environment.Spec.OwnedBy)
	}

	// Add custom labels from environment spec (move to annotations for invalid characters)
	if environment.Spec.Labels != nil {
		for k, v := range environment.Spec.Labels {
			if isEnvironmentNamespaceManagedMetadataKey(k) {
				continue
			}
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
			if isEnvironmentNamespaceManagedMetadataKey(k) {
				continue
			}
			namespace.Annotations[k] = v
		}
	}

	applyEnvironmentNamespaceOwnership(namespace, environment)
	namespace.Annotations[environmentCreatedByAnnotation] = environment.Spec.OwnedBy
}

// createNamespace creates the namespace for the environment
func (r *EnvironmentReconciler) createNamespace(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Create namespace for the Environment
	logger.Info("Creating namespace for environment",
		zap.String("namespace", environment.Spec.TargetNamespace))

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: environment.Spec.TargetNamespace,
			Annotations: map[string]string{
				"kloudlite.io/creation-reason": "auto-created-for-environment",
			},
		},
	}

	// Apply labels and annotations using helper function
	r.applyLabelsAndAnnotations(namespace, environment)

	// Create the namespace
	if err := r.Create(ctx, namespace); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Another reconciliation might have created it
			logger.Info("Namespace already exists (race condition)")
			existingNamespace := &corev1.Namespace{}
			if getErr := r.Get(ctx, client.ObjectKey{Name: environment.Spec.TargetNamespace}, existingNamespace); getErr != nil {
				return getErr
			}
			if ownershipErr := validateEnvironmentNamespaceOwnership(existingNamespace, environment); ownershipErr != nil {
				logger.Error("Target namespace from create race is not owned by Environment", zap.Error(ownershipErr))
				return ownershipErr
			}
			return nil
		}
		logger.Error("Failed to create namespace", zap.Error(err))
		// Retry after a delay
		return err
	}

	logger.Info("Successfully created namespace for environment",
		zap.String("namespace", environment.Spec.TargetNamespace))

	return nil
}

// ensureNamespaceExists checks if namespace exists and creates it if needed
func (r *EnvironmentReconciler) ensureNamespaceExists(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (bool, error) {
	// Check if namespace already exists
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: environment.Spec.TargetNamespace}, namespace)

	if err == nil {
		// Namespace already exists
		logger.Info("Namespace already exists for environment",
			zap.String("namespace", environment.Spec.TargetNamespace))

		if ownershipErr := validateEnvironmentNamespaceOwnership(namespace, environment); ownershipErr != nil {
			logger.Error("Target namespace is not owned by Environment", zap.Error(ownershipErr))
			return false, ownershipErr
		}

		// Apply labels and annotations using helper function
		r.applyLabelsAndAnnotations(namespace, environment)

		// Update the namespace
		if err := r.Update(ctx, namespace); err != nil {
			logger.Warn("Failed to update namespace labels/annotations", zap.Error(err))
			return false, fmt.Errorf("failed to update namespace labels/annotations: %w", err)
		}

		return true, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check existing namespace", zap.Error(err))
		return false, err
	}

	// Namespace doesn't exist, create it
	// Update status to show namespace is being created
	if environment.Status.State != environmentsv1.EnvironmentStateInactive {
		// Only update if we're not already in an appropriate state
		if err := r.updateEnvironmentStatus(ctx, environment, environmentsv1.EnvironmentStateInactive, "Creating namespace for environment", logger); err != nil {
			logger.Warn("Failed to update status to creating", zap.Error(err))
			// Continue with namespace creation even if status update fails
		}
	}

	if err := r.createNamespace(ctx, environment, logger); err != nil {
		return false, err
	}

	// Return true since namespace now exists and reconciliation should continue
	return true, nil
}

// createNamespaceForForking creates namespace for a forked environment
func (r *EnvironmentReconciler) createNamespaceForForking(ctx context.Context, environment *environmentsv1.Environment, sourceName string, logger *zap.Logger) error {
	targetNamespace := environment.Spec.TargetNamespace

	logger.Info("Creating namespace for forked environment", zap.String("namespace", targetNamespace))

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: targetNamespace,
			Annotations: map[string]string{
				"kloudlite.io/creation-reason": "auto-created-for-forked-environment",
				"kloudlite.io/forked-from":     sourceName,
			},
		},
	}

	// Apply labels and annotations using helper function
	r.applyLabelsAndAnnotations(namespace, environment)

	// Add forking-specific annotations
	if namespace.Annotations == nil {
		namespace.Annotations = make(map[string]string)
	}
	namespace.Annotations["kloudlite.io/forked-from"] = sourceName
	namespace.Annotations["kloudlite.io/creation-reason"] = "auto-created-for-forked-environment"

	if err := r.Create(ctx, namespace); err != nil {
		if apierrors.IsAlreadyExists(err) {
			existingNamespace := &corev1.Namespace{}
			if getErr := r.Get(ctx, client.ObjectKey{Name: targetNamespace}, existingNamespace); getErr != nil {
				return getErr
			}
			if ownershipErr := validateEnvironmentNamespaceOwnership(existingNamespace, environment); ownershipErr != nil {
				logger.Error("Target namespace from fork create race is not owned by Environment", zap.Error(ownershipErr))
				return ownershipErr
			}
		} else {
			logger.Error("Failed to create namespace for forked environment", zap.Error(err))
			return err
		}
	}
	logger.Info("Successfully created namespace for forked environment", zap.String("namespace", targetNamespace))

	return nil
}

// deleteNamespace deletes the namespace for the environment
func (r *EnvironmentReconciler) deleteNamespace(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) (bool, error) {
	// Check if namespace exists
	namespace := &corev1.Namespace{}
	err := r.Get(ctx, client.ObjectKey{Name: environment.Spec.TargetNamespace}, namespace)

	if err == nil {
		// Namespace exists, delete it
		logger.Info("Deleting namespace for environment",
			zap.String("namespace", environment.Spec.TargetNamespace))
		if ownershipErr := validateEnvironmentNamespaceOwnership(namespace, environment); ownershipErr != nil {
			logger.Error("Refusing to delete namespace not owned by Environment", zap.Error(ownershipErr))
			return false, ownershipErr
		}

		if err := r.Delete(ctx, namespace); err != nil {
			if !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete namespace", zap.Error(err))
				return false, err
			}
		}

		// Requeue to wait for namespace deletion to complete
		logger.Info("Waiting for namespace deletion to complete")
		return false, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check namespace", zap.Error(err))
		return false, err
	}

	// Namespace is deleted
	logger.Info("Namespace deleted successfully")
	return true, nil
}
