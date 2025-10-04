package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/v2/api/internal/repository"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EnvConfigMapName      = "env-config"
	EnvSecretName         = "env-secret"
	EnvFileConfigPrefix   = "env-file-"
	EnvResourceLabelKey   = "kloudlite.io/resource-type"
	EnvResourceLabelValue = "environment-config"
)

// EnvironmentConfigHandlers handles HTTP requests for environment configs, secrets, and files
type EnvironmentConfigHandlers struct {
	envRepo   repository.EnvironmentRepository
	k8sClient client.Client
	logger    *zap.Logger
}

// NewEnvironmentConfigHandlers creates a new EnvironmentConfigHandlers
func NewEnvironmentConfigHandlers(envRepo repository.EnvironmentRepository, k8sClient client.Client, logger *zap.Logger) *EnvironmentConfigHandlers {
	return &EnvironmentConfigHandlers{
		envRepo:   envRepo,
		k8sClient: k8sClient,
		logger:    logger,
	}
}

// SetConfig handles PUT /api/v1/environments/:name/config
func (h *EnvironmentConfigHandlers) SetConfig(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	var req struct {
		Data map[string]string `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse set config request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get environment to verify it exists and get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		h.logger.Error("Failed to get environment",
			zap.String("name", envName),
			zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace

	// Create or update ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvConfigMapName,
			Namespace: namespace,
			Labels: map[string]string{
				EnvResourceLabelKey: EnvResourceLabelValue,
				"kloudlite.io/environment": envName,
			},
		},
		Data: req.Data,
	}

	if err := h.createOrUpdateConfigMap(c.Request.Context(), configMap); err != nil {
		h.logger.Error("Failed to create/update config",
			zap.String("environment", envName),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to set config",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment config set successfully",
		zap.String("environment", envName),
		zap.String("namespace", namespace))

	c.JSON(http.StatusOK, gin.H{
		"message": "Config set successfully",
		"data":    req.Data,
	})
}

// GetConfig handles GET /api/v1/environments/:name/config
func (h *EnvironmentConfigHandlers) GetConfig(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get environment to get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace

	// Get ConfigMap
	configMap := &corev1.ConfigMap{}
	err = h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      EnvConfigMapName,
		Namespace: namespace,
	}, configMap)

	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusOK, gin.H{
				"message": "No config found",
				"data":    map[string]string{},
			})
			return
		}
		h.logger.Error("Failed to get config",
			zap.String("environment", envName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get config",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": configMap.Data,
	})
}

// DeleteConfig handles DELETE /api/v1/environments/:name/config
func (h *EnvironmentConfigHandlers) DeleteConfig(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get environment to get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace

	// Delete ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvConfigMapName,
			Namespace: namespace,
		},
	}

	if err := h.k8sClient.Delete(c.Request.Context(), configMap); err != nil {
		if !apierrors.IsNotFound(err) {
			h.logger.Error("Failed to delete config",
				zap.String("environment", envName),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete config",
				"details": err.Error(),
			})
			return
		}
	}

	h.logger.Info("Environment config deleted successfully",
		zap.String("environment", envName))

	c.JSON(http.StatusOK, gin.H{
		"message": "Config deleted successfully",
	})
}

// SetSecret handles PUT /api/v1/environments/:name/secret
func (h *EnvironmentConfigHandlers) SetSecret(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	var req struct {
		Data map[string]string `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse set secret request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get environment to verify it exists and get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace

	// Convert string data to byte data
	stringData := req.Data

	// Create or update Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvSecretName,
			Namespace: namespace,
			Labels: map[string]string{
				EnvResourceLabelKey: EnvResourceLabelValue,
				"kloudlite.io/environment": envName,
			},
		},
		StringData: stringData,
		Type:       corev1.SecretTypeOpaque,
	}

	if err := h.createOrUpdateSecret(c.Request.Context(), secret); err != nil {
		h.logger.Error("Failed to create/update secret",
			zap.String("environment", envName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to set secret",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment secret set successfully",
		zap.String("environment", envName))

	c.JSON(http.StatusOK, gin.H{
		"message": "Secret set successfully",
		"keys":    getKeys(req.Data),
	})
}

// GetSecret handles GET /api/v1/environments/:name/secret
func (h *EnvironmentConfigHandlers) GetSecret(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get environment to get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace

	// Get Secret
	secret := &corev1.Secret{}
	err = h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      EnvSecretName,
		Namespace: namespace,
	}, secret)

	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusOK, gin.H{
				"message": "No secret found",
				"keys":    []string{},
			})
			return
		}
		h.logger.Error("Failed to get secret",
			zap.String("environment", envName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get secret",
			"details": err.Error(),
		})
		return
	}

	// Return only keys, not values (for security)
	keys := make([]string, 0, len(secret.Data))
	for k := range secret.Data {
		keys = append(keys, k)
	}

	c.JSON(http.StatusOK, gin.H{
		"keys": keys,
	})
}

// DeleteSecret handles DELETE /api/v1/environments/:name/secret
func (h *EnvironmentConfigHandlers) DeleteSecret(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get environment to get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace

	// Delete Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvSecretName,
			Namespace: namespace,
		},
	}

	if err := h.k8sClient.Delete(c.Request.Context(), secret); err != nil {
		if !apierrors.IsNotFound(err) {
			h.logger.Error("Failed to delete secret",
				zap.String("environment", envName),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete secret",
				"details": err.Error(),
			})
			return
		}
	}

	h.logger.Info("Environment secret deleted successfully",
		zap.String("environment", envName))

	c.JSON(http.StatusOK, gin.H{
		"message": "Secret deleted successfully",
	})
}

// SetFile handles PUT /api/v1/environments/:name/files/:filename
func (h *EnvironmentConfigHandlers) SetFile(c *gin.Context) {
	envName := c.Param("name")
	filename := c.Param("filename")

	if envName == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name and filename are required",
		})
		return
	}

	// Validate filename (no path traversal)
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid filename",
		})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse set file request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get environment to verify it exists and get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace
	configMapName := EnvFileConfigPrefix + filename

	// Create or update ConfigMap for file
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
			Labels: map[string]string{
				EnvResourceLabelKey: EnvResourceLabelValue,
				"kloudlite.io/environment": envName,
				"kloudlite.io/file-type":   "environment-file",
			},
		},
		Data: map[string]string{
			filename: req.Content,
		},
	}

	if err := h.createOrUpdateConfigMap(c.Request.Context(), configMap); err != nil {
		h.logger.Error("Failed to create/update file",
			zap.String("environment", envName),
			zap.String("filename", filename),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to set file",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Environment file set successfully",
		zap.String("environment", envName),
		zap.String("filename", filename))

	c.JSON(http.StatusOK, gin.H{
		"message":  "File set successfully",
		"filename": filename,
	})
}

// GetFile handles GET /api/v1/environments/:name/files/:filename
func (h *EnvironmentConfigHandlers) GetFile(c *gin.Context) {
	envName := c.Param("name")
	filename := c.Param("filename")

	if envName == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name and filename are required",
		})
		return
	}

	// Get environment to get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace
	configMapName := EnvFileConfigPrefix + filename

	// Get ConfigMap
	configMap := &corev1.ConfigMap{}
	err = h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      configMapName,
		Namespace: namespace,
	}, configMap)

	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}
		h.logger.Error("Failed to get file",
			zap.String("environment", envName),
			zap.String("filename", filename),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get file",
			"details": err.Error(),
		})
		return
	}

	content, ok := configMap.Data[filename]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File content not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename": filename,
		"content":  content,
	})
}

// ListFiles handles GET /api/v1/environments/:name/files
func (h *EnvironmentConfigHandlers) ListFiles(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get environment to get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace

	// List all file ConfigMaps
	configMapList := &corev1.ConfigMapList{}
	err = h.k8sClient.List(c.Request.Context(), configMapList,
		client.InNamespace(namespace),
		client.MatchingLabels{
			"kloudlite.io/file-type": "environment-file",
		},
	)

	if err != nil {
		h.logger.Error("Failed to list files",
			zap.String("environment", envName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list files",
			"details": err.Error(),
		})
		return
	}

	files := make([]gin.H, 0)
	for _, cm := range configMapList.Items {
		for filename := range cm.Data {
			files = append(files, gin.H{
				"name":         filename,
				"configMapName": cm.Name,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"count": len(files),
	})
}

// DeleteFile handles DELETE /api/v1/environments/:name/files/:filename
func (h *EnvironmentConfigHandlers) DeleteFile(c *gin.Context) {
	envName := c.Param("name")
	filename := c.Param("filename")

	if envName == "" || filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name and filename are required",
		})
		return
	}

	// Get environment to get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace
	configMapName := EnvFileConfigPrefix + filename

	// Delete ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
	}

	if err := h.k8sClient.Delete(c.Request.Context(), configMap); err != nil {
		if !apierrors.IsNotFound(err) {
			h.logger.Error("Failed to delete file",
				zap.String("environment", envName),
				zap.String("filename", filename),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete file",
				"details": err.Error(),
			})
			return
		}
	}

	h.logger.Info("Environment file deleted successfully",
		zap.String("environment", envName),
		zap.String("filename", filename))

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}

// Helper functions

func (h *EnvironmentConfigHandlers) createOrUpdateConfigMap(ctx context.Context, cm *corev1.ConfigMap) error {
	existing := &corev1.ConfigMap{}
	err := h.k8sClient.Get(ctx, client.ObjectKey{
		Name:      cm.Name,
		Namespace: cm.Namespace,
	}, existing)

	if err != nil {
		if apierrors.IsNotFound(err) {
			return h.k8sClient.Create(ctx, cm)
		}
		return err
	}

	// Update existing
	existing.Data = cm.Data
	existing.Labels = cm.Labels
	return h.k8sClient.Update(ctx, existing)
}

func (h *EnvironmentConfigHandlers) createOrUpdateSecret(ctx context.Context, secret *corev1.Secret) error {
	existing := &corev1.Secret{}
	err := h.k8sClient.Get(ctx, client.ObjectKey{
		Name:      secret.Name,
		Namespace: secret.Namespace,
	}, existing)

	if err != nil {
		if apierrors.IsNotFound(err) {
			return h.k8sClient.Create(ctx, secret)
		}
		return err
	}

	// Update existing
	existing.StringData = secret.StringData
	existing.Labels = secret.Labels
	return h.k8sClient.Update(ctx, existing)
}

func getKeys(data map[string]string) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

// EnvVar represents a unified environment variable (config or secret)
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"` // Empty for secrets (security)
	Type  string `json:"type"`  // "config" or "secret"
}

// GetEnvVars handles GET /api/v1/environments/:name/envvars
// Returns both configs and secrets in unified format
func (h *EnvironmentConfigHandlers) GetEnvVars(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	// Get environment to get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace
	envVars := make([]EnvVar, 0)

	// Get configs
	configMap := &corev1.ConfigMap{}
	err = h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      EnvConfigMapName,
		Namespace: namespace,
	}, configMap)

	if err == nil {
		for k, v := range configMap.Data {
			envVars = append(envVars, EnvVar{
				Key:   k,
				Value: v,
				Type:  "config",
			})
		}
	} else if !apierrors.IsNotFound(err) {
		h.logger.Error("Failed to get configs",
			zap.String("environment", envName),
			zap.Error(err))
	}

	// Get secrets (keys only)
	secret := &corev1.Secret{}
	err = h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      EnvSecretName,
		Namespace: namespace,
	}, secret)

	if err == nil {
		for k := range secret.Data {
			envVars = append(envVars, EnvVar{
				Key:   k,
				Value: "", // Don't expose secret values
				Type:  "secret",
			})
		}
	} else if !apierrors.IsNotFound(err) {
		h.logger.Error("Failed to get secrets",
			zap.String("environment", envName),
			zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{
		"envVars": envVars,
		"count":   len(envVars),
	})
}

// CreateEnvVar handles POST /api/v1/environments/:name/envvars
// Creates a new environment variable (config or secret)
func (h *EnvironmentConfigHandlers) CreateEnvVar(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
		Type  string `json:"type" binding:"required,oneof=config secret"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create envvar request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get environment to verify it exists and get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace

	// Check if key already exists in ConfigMap
	configMap := &corev1.ConfigMap{}
	err = h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      EnvConfigMapName,
		Namespace: namespace,
	}, configMap)
	if err == nil && configMap.Data != nil {
		if _, exists := configMap.Data[req.Key]; exists {
			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("Key '%s' already exists as a config. Please use a different key or update the existing one.", req.Key),
			})
			return
		}
	}

	// Check if key already exists in Secret
	secret := &corev1.Secret{}
	err = h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      EnvSecretName,
		Namespace: namespace,
	}, secret)
	if err == nil && secret.Data != nil {
		if _, exists := secret.Data[req.Key]; exists {
			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("Key '%s' already exists as a secret. Please use a different key or update the existing one.", req.Key),
			})
			return
		}
	}

	h.setEnvVarInternal(c, envName, namespace, req.Key, req.Value, req.Type)
}

// SetEnvVar handles PUT /api/v1/environments/:name/envvars
// Updates an existing environment variable (config or secret)
func (h *EnvironmentConfigHandlers) SetEnvVar(c *gin.Context) {
	envName := c.Param("name")
	if envName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name is required",
		})
		return
	}

	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
		Type  string `json:"type" binding:"required,oneof=config secret"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update envvar request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get environment to verify it exists and get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace

	h.setEnvVarInternal(c, envName, namespace, req.Key, req.Value, req.Type)
}

// setEnvVarInternal is the internal implementation for creating/updating envvars
func (h *EnvironmentConfigHandlers) setEnvVarInternal(c *gin.Context, envName, namespace, key, value, envType string) {
	if envType == "config" {
		// Get existing config or create new
		configMap := &corev1.ConfigMap{}
		err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
			Name:      EnvConfigMapName,
			Namespace: namespace,
		}, configMap)

		if apierrors.IsNotFound(err) {
			configMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      EnvConfigMapName,
					Namespace: namespace,
					Labels: map[string]string{
						EnvResourceLabelKey:         EnvResourceLabelValue,
						"kloudlite.io/environment": envName,
						"kloudlite.io/config-type":  "envvars",
					},
				},
				Data: map[string]string{key: value},
			}
			if err := h.k8sClient.Create(c.Request.Context(), configMap); err != nil {
				h.logger.Error("Failed to create config",
					zap.String("environment", envName),
					zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to set config",
					"details": err.Error(),
				})
				return
			}
		} else if err != nil {
			h.logger.Error("Failed to get config",
				zap.String("environment", envName),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to set config",
				"details": err.Error(),
			})
			return
		} else {
			// Update existing
			configMap.Data[key] = value
			if err := h.k8sClient.Update(c.Request.Context(), configMap); err != nil {
				h.logger.Error("Failed to update config",
					zap.String("environment", envName),
					zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to set config",
					"details": err.Error(),
				})
				return
			}
		}
	} else {
		// Secret
		secret := &corev1.Secret{}
		err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
			Name:      EnvSecretName,
			Namespace: namespace,
		}, secret)

		if apierrors.IsNotFound(err) {
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      EnvSecretName,
					Namespace: namespace,
					Labels: map[string]string{
						EnvResourceLabelKey:         EnvResourceLabelValue,
						"kloudlite.io/environment": envName,
						"kloudlite.io/config-type":  "envvars",
					},
				},
				StringData: map[string]string{key: value},
				Type:       corev1.SecretTypeOpaque,
			}
			if err := h.k8sClient.Create(c.Request.Context(), secret); err != nil {
				h.logger.Error("Failed to create secret",
					zap.String("environment", envName),
					zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to set secret",
					"details": err.Error(),
				})
				return
			}
		} else if err != nil {
			h.logger.Error("Failed to get secret",
				zap.String("environment", envName),
				zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to set secret",
				"details": err.Error(),
			})
			return
		} else {
			// Update existing
			secret.StringData = map[string]string{key: value}
			if err := h.k8sClient.Update(c.Request.Context(), secret); err != nil {
				h.logger.Error("Failed to update secret",
					zap.String("environment", envName),
					zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to set secret",
					"details": err.Error(),
				})
				return
			}
		}
	}

	h.logger.Info("Environment variable set successfully",
		zap.String("environment", envName),
		zap.String("key", key),
		zap.String("type", envType))

	c.JSON(http.StatusOK, gin.H{
		"message": "Environment variable set successfully",
		"key":     key,
		"type":    envType,
	})
}

// DeleteEnvVar handles DELETE /api/v1/environments/:name/envvars/:key
// Deletes an environment variable (checks both config and secret)
func (h *EnvironmentConfigHandlers) DeleteEnvVar(c *gin.Context) {
	envName := c.Param("name")
	key := c.Param("key")

	if envName == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Environment name and key are required",
		})
		return
	}

	// Get environment to get namespace
	env, err := h.envRepo.Get(c.Request.Context(), envName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Environment not found",
			"details": err.Error(),
		})
		return
	}

	namespace := env.Spec.TargetNamespace
	deleted := false

	// Try deleting from config
	configMap := &corev1.ConfigMap{}
	err = h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      EnvConfigMapName,
		Namespace: namespace,
	}, configMap)

	if err == nil && configMap.Data != nil {
		if _, exists := configMap.Data[key]; exists {
			delete(configMap.Data, key)
			if len(configMap.Data) == 0 {
				// Delete entire ConfigMap if empty
				if err := h.k8sClient.Delete(c.Request.Context(), configMap); err != nil && !apierrors.IsNotFound(err) {
					h.logger.Error("Failed to delete config",
						zap.String("environment", envName),
						zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to delete environment variable",
						"details": err.Error(),
					})
					return
				}
			} else {
				if err := h.k8sClient.Update(c.Request.Context(), configMap); err != nil {
					h.logger.Error("Failed to update config",
						zap.String("environment", envName),
						zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to delete environment variable",
						"details": err.Error(),
					})
					return
				}
			}
			deleted = true
		}
	}

	// Try deleting from secret if not found in config
	if !deleted {
		secret := &corev1.Secret{}
		err = h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
			Name:      EnvSecretName,
			Namespace: namespace,
		}, secret)

		if err == nil && secret.Data != nil {
			if _, exists := secret.Data[key]; exists {
				delete(secret.Data, key)
				if len(secret.Data) == 0 {
					// Delete entire Secret if empty
					if err := h.k8sClient.Delete(c.Request.Context(), secret); err != nil && !apierrors.IsNotFound(err) {
						h.logger.Error("Failed to delete secret",
							zap.String("environment", envName),
							zap.Error(err))
						c.JSON(http.StatusInternalServerError, gin.H{
							"error":   "Failed to delete environment variable",
							"details": err.Error(),
						})
						return
					}
				} else {
					if err := h.k8sClient.Update(c.Request.Context(), secret); err != nil {
						h.logger.Error("Failed to update secret",
							zap.String("environment", envName),
							zap.Error(err))
						c.JSON(http.StatusInternalServerError, gin.H{
							"error":   "Failed to delete environment variable",
							"details": err.Error(),
						})
						return
					}
				}
				deleted = true
			}
		}
	}

	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Environment variable not found",
		})
		return
	}

	h.logger.Info("Environment variable deleted successfully",
		zap.String("environment", envName),
		zap.String("key", key))

	c.JSON(http.StatusOK, gin.H{
		"message": "Environment variable deleted successfully",
	})
}
