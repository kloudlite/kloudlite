package webhooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	platformv1alpha1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/platform/v1alpha1"
	"github.com/kloudlite/kloudlite/v2/api/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UserWebhook struct {
	logger logger.Logger
}

func NewUserWebhook(logger logger.Logger) *UserWebhook {
	return &UserWebhook{
		logger: logger,
	}
}

// ValidateUser handles validation webhook for User CRD
func (w *UserWebhook) ValidateUser(c *gin.Context) {
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

// MutateUser handles mutation webhook for User CRD (adds/updates labels)
func (w *UserWebhook) MutateUser(c *gin.Context) {
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

func (w *UserWebhook) handleValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the User object
	var user platformv1alpha1.User
	if err := json.Unmarshal(req.Object.Raw, &user); err != nil {
		w.logger.Error("Failed to unmarshal user object: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Failed to parse User object: %v", err),
			},
		}
	}

	// Validate the User
	allowed := true
	var messages []string

	// Validate email
	if user.Spec.Email == "" {
		allowed = false
		messages = append(messages, "Email is required")
	} else if !isValidEmail(user.Spec.Email) {
		allowed = false
		messages = append(messages, "Invalid email format")
	}

	// Validate name (if provided)
	if user.Name != "" && !isValidKubernetesName(user.Name) {
		allowed = false
		messages = append(messages, "Invalid name format (must be lowercase alphanumeric or '-')")
	}

	// Validate roles
	if len(user.Spec.Roles) == 0 {
		allowed = false
		messages = append(messages, "At least one role is required")
	}

	// Build response
	response := &admissionv1.AdmissionResponse{
		Allowed: allowed,
	}

	if !allowed {
		response.Result = &metav1.Status{
			Message: strings.Join(messages, "; "),
		}
	}

	return response
}

func (w *UserWebhook) handleMutation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the User object
	var user platformv1alpha1.User
	if err := json.Unmarshal(req.Object.Raw, &user); err != nil {
		w.logger.Error("Failed to unmarshal user object: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Failed to parse User object: %v", err),
			},
		}
	}

	// Create patches for mutations
	var patches []patchOperation

	// Ensure labels map exists
	if user.ObjectMeta.Labels == nil {
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/metadata/labels",
			Value: make(map[string]string),
		})
	}

	// Add email label for efficient lookups
	if user.Spec.Email != "" {
		emailLabel := sanitizeEmailForLabel(user.Spec.Email)
		labelPath := "/metadata/labels/platform.kloudlite.io~1user-email"

		// Check if we need to add or replace
		if user.ObjectMeta.Labels == nil {
			patches = append(patches, patchOperation{
				Op:   "add",
				Path: "/metadata/labels",
				Value: map[string]string{
					"platform.kloudlite.io/user-email": emailLabel,
				},
			})
		} else if existingLabel, exists := user.ObjectMeta.Labels["platform.kloudlite.io/user-email"]; !exists || existingLabel != emailLabel {
			op := "add"
			if exists {
				op = "replace"
			}
			patches = append(patches, patchOperation{
				Op:    op,
				Path:  labelPath,
				Value: emailLabel,
			})
		}
	}

	// Add default metadata if missing
	if user.Spec.Metadata == nil {
		patches = append(patches, patchOperation{
			Op:   "add",
			Path: "/spec/metadata",
			Value: map[string]string{
				"name":  user.Spec.DisplayName,
				"email": user.Spec.Email,
			},
		})
	}

	// Create patch response
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		w.logger.Error("Failed to marshal patches: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Failed to create patches: %v", err),
			},
		}
	}

	patchType := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: &patchType,
	}
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// Helper functions
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func isValidKubernetesName(name string) bool {
	if len(name) == 0 || len(name) > 253 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

func sanitizeEmailForLabel(email string) string {
	// Replace special characters with hyphens for label value
	sanitized := strings.ReplaceAll(email, "@", "-at-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-dot-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ToLower(sanitized)

	// Ensure it starts and ends with alphanumeric
	sanitized = strings.Trim(sanitized, "-")

	// Limit length to 63 characters (Kubernetes label value limit)
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
	}

	return sanitized
}