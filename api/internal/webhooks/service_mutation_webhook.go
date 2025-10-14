package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	interceptsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/intercepts/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceMutationWebhook struct {
	logger logger.Logger
	client client.Client
}

func NewServiceMutationWebhook(logger logger.Logger, client client.Client) *ServiceMutationWebhook {
	return &ServiceMutationWebhook{
		logger: logger,
		client: client,
	}
}

// MutateService handles mutation webhook for Service resources to redirect traffic for intercepts
func (w *ServiceMutationWebhook) MutateService(c *gin.Context) {
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

func (w *ServiceMutationWebhook) handleMutation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the Service object
	var service corev1.Service
	if err := json.Unmarshal(req.Object.Raw, &service); err != nil {
		w.logger.Error("Failed to unmarshal Service object: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Failed to parse Service object: %v", err),
			},
		}
	}

	// Check if there's an active ServiceIntercept for this service
	interceptList := &interceptsv1.ServiceInterceptList{}
	err := w.client.List(context.TODO(), interceptList,
		client.InNamespace(req.Namespace),
		client.MatchingLabels{
			"intercepts.kloudlite.io/service-name": service.Name,
		})

	if err != nil {
		w.logger.Error("Failed to list ServiceIntercepts: " + err.Error())
		// Allow the service to proceed without interception on error
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Find an active intercept (skip ones being deleted)
	var activeIntercept *interceptsv1.ServiceIntercept
	for i := range interceptList.Items {
		intercept := &interceptList.Items[i]
		if intercept.DeletionTimestamp == nil && intercept.Spec.Status == "active" && intercept.Status.Phase == "Active" {
			activeIntercept = intercept
			break
		}
	}

	// If no active intercept, allow service as-is
	if activeIntercept == nil {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Create patches to redirect traffic to workspace pod
	var patches []patchOperation

	// Get the workspace pod to find its labels
	workspacePodName := activeIntercept.Status.WorkspacePodName
	if workspacePodName == "" {
		// If we don't have the pod name yet, allow without interception
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Modify the service selector to point to the workspace pod
	// We'll use a specific label that the workspace pod has
	newSelector := map[string]string{
		"workspaces.kloudlite.io/workspace-name": activeIntercept.Spec.WorkspaceRef.Name,
	}

	// Add patch to replace the selector
	patches = append(patches, patchOperation{
		Op:    "replace",
		Path:  "/spec/selector",
		Value: newSelector,
	})

	// Add annotation to track that this service is intercepted
	if service.Annotations == nil {
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/metadata/annotations",
			Value: make(map[string]string),
		})
	}

	escapedKey := "intercepts.kloudlite.io~1intercepted-by"
	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  fmt.Sprintf("/metadata/annotations/%s", escapedKey),
		Value: activeIntercept.Name,
	})

	// Update service ports to map to workspace ports
	if len(activeIntercept.Spec.PortMappings) > 0 {
		var newPorts []corev1.ServicePort
		for _, mapping := range activeIntercept.Spec.PortMappings {
			// Find the corresponding service port
			for _, svcPort := range service.Spec.Ports {
				if svcPort.Port == mapping.ServicePort {
					// Create a new port that targets the workspace port
					newPort := svcPort.DeepCopy()
					newPort.TargetPort.IntVal = mapping.WorkspacePort
					if mapping.Protocol != "" {
						newPort.Protocol = mapping.Protocol
					}
					newPorts = append(newPorts, *newPort)
				}
			}
		}

		if len(newPorts) > 0 {
			patches = append(patches, patchOperation{
				Op:    "replace",
				Path:  "/spec/ports",
				Value: newPorts,
			})
		}
	}

	w.logger.Info(fmt.Sprintf("Intercepting service '%s' in namespace '%s' with workspace '%s'",
		service.Name, service.Namespace, activeIntercept.Spec.WorkspaceRef.Name))

	// Create patch response
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		w.logger.Error("Failed to marshal patches: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: true, // Allow without interception on error
		}
	}

	patchType := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}
}
