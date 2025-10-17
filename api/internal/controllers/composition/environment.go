package composition

import (
	"context"
	"strings"

	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)


// getEnvironmentForNamespace finds the environment that owns the given namespace
func (r *CompositionReconciler) getEnvironmentForNamespace(ctx context.Context, namespace string, logger *zap.Logger) (*compositionsv1.Environment, error) {
	envList := &compositionsv1.EnvironmentList{}
	if err := r.List(ctx, envList); err != nil {
		logger.Error("Failed to list environments", zap.Error(err))
		return nil, err
	}

	for _, env := range envList.Items {
		if env.Spec.TargetNamespace == namespace {
			return &env, nil
		}
	}

	logger.Warn("No environment found for namespace", zap.String("namespace", namespace))
	return nil, nil
}

// fetchEnvironmentData fetches environment envvars and config files from ConfigMaps and Secrets
func (r *CompositionReconciler) fetchEnvironmentData(ctx context.Context, namespace string, logger *zap.Logger) (*EnvironmentData, error) {
	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	// Fetch environment envvars ConfigMap
	if err := r.fetchConfigMap(ctx, namespace, envConfigConfigMapName, envData.EnvVars, logger, "environment envvars"); err != nil {
		return envData, err // Return error for critical failures
	}

	// Fetch environment envvars Secret
	if err := r.fetchSecret(ctx, namespace, envSecretSecretName, envData.Secrets, logger, "environment envvars"); err != nil {
		return envData, err // Return error for critical failures
	}

	// Fetch environment config files from individual ConfigMaps
	if err := r.fetchConfigFiles(ctx, namespace, envData.ConfigFiles, logger); err != nil {
		return envData, err
	}

	return envData, nil
}

// fetchConfigMap fetches a ConfigMap and populates the target map with its data
func (r *CompositionReconciler) fetchConfigMap(ctx context.Context, namespace, name string, target map[string]string, logger *zap.Logger, description string) error {
	configMap := &corev1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, configMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ConfigMap not found, skipping", zap.String("name", name), zap.String("description", description))
			return nil
		}
		logger.Error("Failed to fetch ConfigMap", zap.String("name", name), zap.String("description", description), zap.Error(err))
		return err
	}

	if configMap.Data != nil {
		logger.Info("Loaded data from ConfigMap", zap.String("name", name), zap.String("description", description), zap.Int("count", len(configMap.Data)))
		for k, v := range configMap.Data {
			target[k] = v
		}
	}

	return nil
}

// fetchSecret fetches a Secret and populates the target map with its data
func (r *CompositionReconciler) fetchSecret(ctx context.Context, namespace, name string, target map[string]string, logger *zap.Logger, description string) error {
	secret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, secret)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Secret not found, skipping", zap.String("name", name), zap.String("description", description))
			return nil
		}
		logger.Error("Failed to fetch Secret", zap.String("name", name), zap.String("description", description), zap.Error(err))
		return err
	}

	if secret.Data != nil {
		logger.Info("Loaded data from Secret", zap.String("name", name), zap.String("description", description), zap.Int("count", len(secret.Data)))
		for k, v := range secret.Data {
			target[k] = string(v)
		}
	}

	return nil
}

// fetchConfigFiles fetches environment config files from ConfigMaps with the file-type label
func (r *CompositionReconciler) fetchConfigFiles(ctx context.Context, namespace string, target map[string]string, logger *zap.Logger) error {
	configMapList := &corev1.ConfigMapList{}
	err := r.List(ctx, configMapList, client.InNamespace(namespace), client.MatchingLabels{
		envFileTypeLabel: "environment-file",
	})
	if err != nil {
		logger.Error("Failed to list environment config file ConfigMaps", zap.Error(err))
		return err
	}

	logger.Info("Found environment config file ConfigMaps", zap.Int("count", len(configMapList.Items)))
	for _, cm := range configMapList.Items {
		// Extract filename from ConfigMap name (remove prefix)
		filename := strings.TrimPrefix(cm.Name, envFileConfigMapPrefix)

		// Get the file content (should be a single key in the ConfigMap data)
		if len(cm.Data) > 0 {
			// Use the first data entry
			for _, content := range cm.Data {
				target[filename] = content
				break // Only use the first data entry
			}
		}
	}

	return nil
}