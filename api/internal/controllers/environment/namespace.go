package environment

import (
	"context"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/pkg/utils"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

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

// createNamespace creates the namespace for the environment
func (r *EnvironmentReconciler) createNamespace(ctx context.Context, environment *environmentsv1.Environment, logger *zap.Logger) error {
	// Create namespace for the Environment
	logger.Info("Creating namespace for environment",
		zap.String("namespace", environment.Spec.TargetNamespace))

	namespace := &corev1.Namespace{
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

		// Apply labels and annotations using helper function
		r.applyLabelsAndAnnotations(namespace, environment)

		// Update the namespace
		if err := r.Update(ctx, namespace); err != nil {
			logger.Warn("Failed to update namespace labels/annotations", zap.Error(err))
		}

		return true, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check existing namespace", zap.Error(err))
		return false, err
	}

	// Namespace doesn't exist, create it
	if err := r.createNamespace(ctx, environment, logger); err != nil {
		return false, err
	}

	return false, nil
}

// createNamespaceForCloning creates namespace for a cloned environment
func (r *EnvironmentReconciler) createNamespaceForCloning(ctx context.Context, environment *environmentsv1.Environment, sourceName string, logger *zap.Logger) error {
	targetNamespace := environment.Spec.TargetNamespace

	logger.Info("Creating namespace for cloned environment", zap.String("namespace", targetNamespace))

	namespace := &corev1.Namespace{
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
		return err
	}
	logger.Info("Successfully created namespace for cloned environment", zap.String("namespace", targetNamespace))

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
