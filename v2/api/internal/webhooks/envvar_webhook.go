package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/v2/api/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EnvConfigMapName = "env-config"
	EnvSecretName    = "env-secret"
)

// Regular expression for valid environment variable keys
// Must start with letter or underscore, followed by letters, digits, or underscores
var envKeyRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type EnvVarWebhook struct {
	logger    logger.Logger
	k8sClient client.Client
}

func NewEnvVarWebhook(logger logger.Logger, k8sClient client.Client) *EnvVarWebhook {
	return &EnvVarWebhook{
		logger:    logger,
		k8sClient: k8sClient,
	}
}

// ValidateConfigMap handles validation webhook for ConfigMap
func (w *EnvVarWebhook) ValidateConfigMap(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.logger.Error("Failed to read request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		w.logger.Error("Failed to unmarshal admission review: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal admission review"})
		return
	}

	response := w.handleConfigMapValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

// ValidateSecret handles validation webhook for Secret
func (w *EnvVarWebhook) ValidateSecret(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		w.logger.Error("Failed to read request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		w.logger.Error("Failed to unmarshal admission review: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to unmarshal admission review"})
		return
	}

	response := w.handleSecretValidation(admissionReview.Request)
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

func (w *EnvVarWebhook) handleConfigMapValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	var configMap corev1.ConfigMap
	if err := json.Unmarshal(req.Object.Raw, &configMap); err != nil {
		w.logger.Error("Failed to unmarshal configmap: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal configmap object",
			},
		}
	}

	// Check if this ConfigMap has the envvars label
	if configMap.Labels != nil {
		if configType, exists := configMap.Labels["kloudlite.io/config-type"]; exists && configType == "envvars" {
			// Only env-config ConfigMap can have this label
			if configMap.Name != EnvConfigMapName {
				return &admissionv1.AdmissionResponse{
					Allowed: false,
					Result: &metav1.Status{
						Message: fmt.Sprintf("Only ConfigMap named '%s' can have label kloudlite.io/config-type=envvars", EnvConfigMapName),
					},
				}
			}
		}
	}

	// Only validate env-config ConfigMap for additional validations
	if configMap.Name != EnvConfigMapName {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	// Validate keys and values
	if err := w.validateConfigMapData(configMap.Data); err != nil {
		w.logger.Warn("ConfigMap validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	// Check for duplicate keys in Secret
	if err := w.validateNoDuplicateKeys(req.Namespace, configMap.Data, "secret"); err != nil {
		w.logger.Warn("ConfigMap validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

func (w *EnvVarWebhook) handleSecretValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	var secret corev1.Secret
	if err := json.Unmarshal(req.Object.Raw, &secret); err != nil {
		w.logger.Error("Failed to unmarshal secret: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal secret object",
			},
		}
	}

	// Check if this Secret has the envvars label
	if secret.Labels != nil {
		if configType, exists := secret.Labels["kloudlite.io/config-type"]; exists && configType == "envvars" {
			// Only env-secret Secret can have this label
			if secret.Name != EnvSecretName {
				return &admissionv1.AdmissionResponse{
					Allowed: false,
					Result: &metav1.Status{
						Message: fmt.Sprintf("Only Secret named '%s' can have label kloudlite.io/config-type=envvars", EnvSecretName),
					},
				}
			}
		}
	}

	// Only validate env-secret Secret for duplicate keys
	if secret.Name != EnvSecretName {
		return &admissionv1.AdmissionResponse{Allowed: true}
	}

	// Validate secret keys
	if err := w.validateSecretKeys(secret.Data); err != nil {
		w.logger.Warn("Secret validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	// Convert secret.Data ([]byte) to map[string]string for validation
	secretData := make(map[string]string)
	for k := range secret.Data {
		secretData[k] = "" // We only care about keys, not values
	}

	// Check for duplicate keys in ConfigMap
	if err := w.validateNoDuplicateKeys(req.Namespace, secretData, "config"); err != nil {
		w.logger.Warn("Secret validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{Allowed: true}
}

// validateConfigMapData validates ConfigMap keys and values
func (w *EnvVarWebhook) validateConfigMapData(data map[string]string) error {
	if len(data) == 0 {
		return fmt.Errorf("ConfigMap data cannot be empty")
	}

	for key, value := range data {
		// Validate key format
		if err := validateEnvVarKey(key); err != nil {
			return err
		}

		// Validate value is not empty for configs
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("config value for key '%s' cannot be empty", key)
		}
	}

	return nil
}

// validateSecretKeys validates Secret keys
func (w *EnvVarWebhook) validateSecretKeys(data map[string][]byte) error {
	if len(data) == 0 {
		return fmt.Errorf("Secret data cannot be empty")
	}

	for key := range data {
		// Validate key format
		if err := validateEnvVarKey(key); err != nil {
			return err
		}
	}

	return nil
}

// validateEnvVarKey validates environment variable key format
func validateEnvVarKey(key string) error {
	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}

	if !envKeyRegex.MatchString(key) {
		return fmt.Errorf("invalid environment variable key '%s': must start with a letter or underscore, followed by letters, digits, or underscores", key)
	}

	return nil
}

// validateNoDuplicateKeys checks that keys in the current resource don't exist in the opposite type
func (w *EnvVarWebhook) validateNoDuplicateKeys(namespace string, keys map[string]string, checkType string) error {
	ctx := context.Background()

	if checkType == "secret" {
		// Check if any keys exist in Secret
		secret := &corev1.Secret{}
		err := w.k8sClient.Get(ctx, client.ObjectKey{
			Name:      EnvSecretName,
			Namespace: namespace,
		}, secret)
		if err == nil {
			// Secret exists, check for duplicate keys
			for key := range keys {
				if _, exists := secret.Data[key]; exists {
					return fmt.Errorf("key '%s' already exists as a secret. Please use a different key or delete the existing secret first", key)
				}
			}
		}
	} else if checkType == "config" {
		// Check if any keys exist in ConfigMap
		configMap := &corev1.ConfigMap{}
		err := w.k8sClient.Get(ctx, client.ObjectKey{
			Name:      EnvConfigMapName,
			Namespace: namespace,
		}, configMap)
		if err == nil {
			// ConfigMap exists, check for duplicate keys
			for key := range keys {
				if _, exists := configMap.Data[key]; exists {
					return fmt.Errorf("key '%s' already exists as a config. Please use a different key or delete the existing config first", key)
				}
			}
		}
	}

	return nil
}
