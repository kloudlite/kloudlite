package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceInterceptWebhook struct {
	logger logger.Logger
	client client.Client
}

func NewServiceInterceptWebhook(logger logger.Logger, client client.Client) *ServiceInterceptWebhook {
	return &ServiceInterceptWebhook{
		logger: logger,
		client: client,
	}
}

// ValidateServiceIntercept handles validation webhook for ServiceIntercept CRD
func (w *ServiceInterceptWebhook) ValidateServiceIntercept(c *gin.Context) {
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

// MutateServiceIntercept handles mutation webhook for ServiceIntercept (adds labels)
func (w *ServiceInterceptWebhook) MutateServiceIntercept(c *gin.Context) {
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

func (w *ServiceInterceptWebhook) handleValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Handle DELETE operations - for DELETE, the object is in OldObject
	if req.Operation == admissionv1.Delete {
		var intercept interceptsv1.ServiceIntercept
		if err := json.Unmarshal(req.OldObject.Raw, &intercept); err != nil {
			w.logger.Error("Failed to unmarshal ServiceIntercept object for delete: " + err.Error())
			// Allow deletion to proceed even if we can't parse the object
			return &admissionv1.AdmissionResponse{
				Allowed: true,
			}
		}

		// Check if this ServiceIntercept is managed by a workspace
		workspaceName := intercept.Labels["workspaces.kloudlite.io/workspace-name"]

		if workspaceName != "" {
			// Check if the deletion is coming from the workspace controller (service account)
			// Service accounts have username format: system:serviceaccount:<namespace>:<name>
			isController := strings.HasPrefix(req.UserInfo.Username, "system:serviceaccount:")

			if !isController {
				// This is a manual deletion attempt - prevent it
				return &admissionv1.AdmissionResponse{
					Allowed: false,
					Result: &metav1.Status{
						Message: fmt.Sprintf("Cannot delete ServiceIntercept '%s' directly - it is managed by workspace '%s'. To remove this intercept, update the workspace's environmentConnection.intercepts field instead.",
							intercept.Name, workspaceName),
					},
				}
			}

			// Allow deletion by the controller
		}

		// Allow deletion of non-workspace-managed intercepts or controller-initiated deletions
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Parse the ServiceIntercept object for CREATE/UPDATE operations
	var intercept interceptsv1.ServiceIntercept
	if err := json.Unmarshal(req.Object.Raw, &intercept); err != nil {
		w.logger.Error("Failed to unmarshal ServiceIntercept object: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Failed to parse ServiceIntercept object: %v", err),
			},
		}
	}

	// Skip validation if the ServiceIntercept is being deleted (for finalizer updates)
	// This allows finalizer removal even if the service/workspace no longer exists
	if intercept.DeletionTimestamp != nil {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Validate the ServiceIntercept
	allowed := true
	var messages []string

	// Validate workspace reference
	if intercept.Spec.WorkspaceRef.Name == "" {
		allowed = false
		messages = append(messages, "WorkspaceRef.Name is required")
	} else {
		// Check if workspace exists and is running (cluster-scoped)
		workspace := &workspacesv1.Workspace{}
		err := w.client.Get(context.TODO(), client.ObjectKey{
			Name: intercept.Spec.WorkspaceRef.Name,
		}, workspace)

		if err != nil {
			allowed = false
			messages = append(messages, fmt.Sprintf("Workspace '%s' not found",
				intercept.Spec.WorkspaceRef.Name))
		} else if workspace.Status.Phase != "Running" {
			allowed = false
			messages = append(messages, fmt.Sprintf("Workspace '%s' is not running (current phase: %s)",
				intercept.Spec.WorkspaceRef.Name, workspace.Status.Phase))
		}
	}

	// Validate service reference
	if intercept.Spec.ServiceRef.Name == "" {
		allowed = false
		messages = append(messages, "ServiceRef.Name is required")
	} else if intercept.Spec.ServiceRef.Namespace == "" {
		allowed = false
		messages = append(messages, "ServiceRef.Namespace is required for cluster-scoped ServiceIntercept")
	} else {
		// Check if service exists
		service := &corev1.Service{}
		serviceNamespace := intercept.Spec.ServiceRef.Namespace

		err := w.client.Get(context.TODO(), client.ObjectKey{
			Name:      intercept.Spec.ServiceRef.Name,
			Namespace: serviceNamespace,
		}, service)

		if err != nil {
			allowed = false
			messages = append(messages, fmt.Sprintf("Service '%s' not found in namespace '%s'",
				intercept.Spec.ServiceRef.Name, serviceNamespace))
		} else {
			// Validate port mappings
			if len(intercept.Spec.PortMappings) == 0 {
				allowed = false
				messages = append(messages, "At least one port mapping is required")
			} else if len(service.Spec.Ports) > 0 {
				// For services with defined ports, validate that port mappings match service ports
				servicePortMap := make(map[int32]bool)
				for _, port := range service.Spec.Ports {
					servicePortMap[port.Port] = true
				}

				for _, mapping := range intercept.Spec.PortMappings {
					if !servicePortMap[mapping.ServicePort] {
						allowed = false
						messages = append(messages, fmt.Sprintf("Service port %d not found in service '%s'",
							mapping.ServicePort, service.Name))
					}
				}
			}
			// For portless services (headless services with no spec.ports), skip port validation
			// and trust the user-provided port mappings
		}

		// Check if another ServiceIntercept is already intercepting this service (cluster-scoped)
		existingIntercepts := &interceptsv1.ServiceInterceptList{}
		err = w.client.List(context.TODO(), existingIntercepts,
			client.MatchingLabels{
				"intercepts.kloudlite.io/service-name":      intercept.Spec.ServiceRef.Name,
				"intercepts.kloudlite.io/service-namespace": serviceNamespace,
			})

		if err != nil {
			w.logger.Error("Failed to check for existing intercepts: " + err.Error())
			allowed = false
			messages = append(messages, "Failed to validate service intercept uniqueness")
		} else {
			// Check if any existing intercept (other than this one) exists
			for _, existing := range existingIntercepts.Items {
				// Skip checking against itself (for UPDATE operations)
				if existing.Name == intercept.Name {
					continue
				}

				// Skip intercepts that are being deleted
				if existing.DeletionTimestamp != nil {
					continue
				}

				allowed = false
				messages = append(messages, fmt.Sprintf("Service '%s' is already being intercepted by workspace '%s' (intercept: %s)",
					intercept.Spec.ServiceRef.Name, existing.Spec.WorkspaceRef.Name, existing.Name))
				break
			}
		}
	}

	// Validate port mappings
	if len(intercept.Spec.PortMappings) == 0 {
		allowed = false
		messages = append(messages, "At least one port mapping is required")
	} else {
		for i, mapping := range intercept.Spec.PortMappings {
			if mapping.ServicePort < 1 || mapping.ServicePort > 65535 {
				allowed = false
				messages = append(messages, fmt.Sprintf("Invalid service port %d at index %d", mapping.ServicePort, i))
			}
			if mapping.WorkspacePort < 1 || mapping.WorkspacePort > 65535 {
				allowed = false
				messages = append(messages, fmt.Sprintf("Invalid workspace port %d at index %d", mapping.WorkspacePort, i))
			}
		}
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

func (w *ServiceInterceptWebhook) handleMutation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the ServiceIntercept object
	var intercept interceptsv1.ServiceIntercept
	if err := json.Unmarshal(req.Object.Raw, &intercept); err != nil {
		w.logger.Error("Failed to unmarshal ServiceIntercept object: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Failed to parse ServiceIntercept object: %v", err),
			},
		}
	}

	// Create patches for mutations
	var patches []patchOperation

	// Ensure labels map exists
	if intercept.ObjectMeta.Labels == nil {
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/metadata/labels",
			Value: make(map[string]string),
		})
	}

	// Add service and workspace labels for fast lookups
	// ServiceRef.Namespace is required and validated, no fallback needed
	labelsToAdd := map[string]string{
		"intercepts.kloudlite.io/service-name":      intercept.Spec.ServiceRef.Name,
		"intercepts.kloudlite.io/service-namespace": intercept.Spec.ServiceRef.Namespace,
		"intercepts.kloudlite.io/workspace-name":    intercept.Spec.WorkspaceRef.Name,
	}

	for key, value := range labelsToAdd {
		if intercept.ObjectMeta.Labels == nil {
			// Labels map doesn't exist, add all at once
			patches = append(patches, patchOperation{
				Op:    "add",
				Path:  "/metadata/labels",
				Value: labelsToAdd,
			})
			break
		} else {
			// Add or replace individual labels
			if existingValue, exists := intercept.ObjectMeta.Labels[key]; !exists || existingValue != value {
				op := "add"
				if exists {
					op = "replace"
				}
				// Escape forward slashes in label keys
				escapedKey := strings.ReplaceAll(key, "/", "~1")
				patches = append(patches, patchOperation{
					Op:    op,
					Path:  fmt.Sprintf("/metadata/labels/%s", escapedKey),
					Value: value,
				})
			}
		}
	}

	// Add finalizer to ensure cleanup (but not during deletion)
	if intercept.ObjectMeta.DeletionTimestamp == nil && !containsString(intercept.ObjectMeta.Finalizers, "intercepts.kloudlite.io/finalizer") {
		finalizers := intercept.ObjectMeta.Finalizers
		finalizers = append(finalizers, "intercepts.kloudlite.io/finalizer")
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/metadata/finalizers",
			Value: finalizers,
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
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
