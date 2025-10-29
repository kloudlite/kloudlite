package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	"gopkg.in/yaml.v3"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CompositionWebhook struct {
	logger    logger.Logger
	k8sClient client.Client
}

func NewCompositionWebhook(logger logger.Logger, k8sClient client.Client) *CompositionWebhook {
	return &CompositionWebhook{
		logger:    logger,
		k8sClient: k8sClient,
	}
}

// ValidateComposition handles validation webhook for Composition CRD
func (w *CompositionWebhook) ValidateComposition(c *gin.Context) {
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

	// Process the admission request
	response := w.handleValidation(admissionReview.Request)

	// Build the admission review response
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

// MutateComposition handles mutation webhook for Composition CRD
func (w *CompositionWebhook) MutateComposition(c *gin.Context) {
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

	// Process the admission request
	response := w.handleMutation(admissionReview.Request)

	// Build the admission review response
	admissionReview.Response = response
	admissionReview.Response.UID = admissionReview.Request.UID

	c.JSON(http.StatusOK, admissionReview)
}

func (w *CompositionWebhook) handleValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the composition object
	var composition environmentsv1.Composition
	if err := json.Unmarshal(req.Object.Raw, &composition); err != nil {
		w.logger.Error("Failed to unmarshal composition: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal composition object",
			},
		}
	}

	// Perform validation
	if err := w.validateComposition(&composition); err != nil {
		w.logger.Warn("Composition validation failed: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1.AdmissionResponse{
		Allowed: true,
	}
}

func (w *CompositionWebhook) handleMutation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the composition object
	var composition environmentsv1.Composition
	if err := json.Unmarshal(req.Object.Raw, &composition); err != nil {
		w.logger.Error("Failed to unmarshal composition: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal composition object",
			},
		}
	}

	// Create patches for mutations
	var patches []map[string]interface{}

	// Set default ComposeFormat if not provided
	if composition.Spec.ComposeFormat == "" {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/composeFormat",
			"value": "v3.8",
		})
	}

	// Set default AutoDeploy if not set
	// Note: We need to check if the field was explicitly set to false or just omitted
	// Since JSON unmarshaling will set bool to false by default, we can't distinguish
	// For now, we'll skip this mutation and let the controller handle it

	// If no patches needed, return allowed without patch
	if len(patches) == 0 {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Convert patches to JSON
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		w.logger.Error("Failed to marshal patches: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to create patches",
			},
		}
	}

	// Return response with patches
	patchType := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}
}

func (w *CompositionWebhook) validateComposition(composition *environmentsv1.Composition) error {
	ctx := context.Background()

	// Validate DisplayName length
	if len(composition.Spec.DisplayName) == 0 {
		return fmt.Errorf("displayName is required")
	}
	if len(composition.Spec.DisplayName) > 100 {
		return fmt.Errorf("displayName must be at most 100 characters")
	}

	// Validate Description length
	if len(composition.Spec.Description) > 500 {
		return fmt.Errorf("description must be at most 500 characters")
	}

	// Validate ComposeContent is not empty
	if composition.Spec.ComposeContent == "" {
		return fmt.Errorf("composeContent is required")
	}

	// Validate ComposeContent is valid YAML
	var composeData interface{}
	if err := yaml.Unmarshal([]byte(composition.Spec.ComposeContent), &composeData); err != nil {
		return fmt.Errorf("composeContent must be valid YAML: %v", err)
	}

	// Validate ComposeFormat is valid
	validFormats := []string{"v2", "v3", "v3.1", "v3.2", "v3.3", "v3.4", "v3.5", "v3.6", "v3.7", "v3.8", "v3.9"}
	if composition.Spec.ComposeFormat != "" {
		valid := false
		for _, format := range validFormats {
			if composition.Spec.ComposeFormat == format {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("composeFormat must be one of: %v", validFormats)
		}
	}

	// Validate EnvFrom references exist
	for _, envFrom := range composition.Spec.EnvFrom {
		if envFrom.Type != "ConfigMap" && envFrom.Type != "Secret" {
			return fmt.Errorf("envFrom type must be either ConfigMap or Secret")
		}
		if envFrom.Name == "" {
			return fmt.Errorf("envFrom name is required")
		}

		// Check if the referenced ConfigMap or Secret exists in the same namespace
		namespace := composition.Namespace
		if namespace == "" {
			namespace = "default"
		}

		if envFrom.Type == "ConfigMap" {
			var cm client.Object
			if err := w.k8sClient.Get(ctx, client.ObjectKey{
				Namespace: namespace,
				Name:      envFrom.Name,
			}, cm); err != nil {
				w.logger.Warn(fmt.Sprintf("ConfigMap %s not found in namespace %s: %v", envFrom.Name, namespace, err))
				// Note: We don't fail validation here since the ConfigMap might be created later
			}
		}
	}

	// Validate ResourceOverrides
	for serviceName, override := range composition.Spec.ResourceOverrides {
		if override.Replicas != nil {
			if *override.Replicas < 0 || *override.Replicas > 10 {
				return fmt.Errorf("replicas for service %s must be between 0 and 10", serviceName)
			}
		}
	}

	return nil
}
