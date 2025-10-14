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

type PodMutationWebhook struct {
	logger logger.Logger
	client client.Client
}

func NewPodMutationWebhook(logger logger.Logger, client client.Client) *PodMutationWebhook {
	return &PodMutationWebhook{
		logger: logger,
		client: client,
	}
}

// MutatePod handles mutation webhook for Pod resources to inject node selector for intercepted services
func (w *PodMutationWebhook) MutatePod(c *gin.Context) {
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

func (w *PodMutationWebhook) handleMutation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the Pod object
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		w.logger.Error("Failed to unmarshal Pod object: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("Failed to parse Pod object: %v", err),
			},
		}
	}

	// Skip workspace pods (don't intercept the interceptor!)
	if pod.Labels != nil {
		if _, isWorkspace := pod.Labels["workspaces.kloudlite.io/workspace-name"]; isWorkspace {
			return &admissionv1.AdmissionResponse{
				Allowed: true,
			}
		}
	}

	// List all active ServiceIntercepts in the pod's namespace
	interceptList := &interceptsv1.ServiceInterceptList{}
	err := w.client.List(context.TODO(), interceptList,
		client.InNamespace(req.Namespace))

	if err != nil {
		w.logger.Error("Failed to list ServiceIntercepts: " + err.Error())
		// Allow the pod to proceed without modification on error
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Check if this pod matches any intercepted service's original selector
	var matchedIntercept *interceptsv1.ServiceIntercept
	for i := range interceptList.Items {
		intercept := &interceptList.Items[i]

		// Only consider active intercepts that are not being deleted
		if intercept.DeletionTimestamp != nil || intercept.Spec.Status != "active" || intercept.Status.Phase != "Active" {
			continue
		}

		// Check if pod labels match the original service selector
		if intercept.Status.OriginalServiceSelector != nil {
			if podMatchesSelector(&pod, intercept.Status.OriginalServiceSelector) {
				matchedIntercept = intercept
				break
			}
		}
	}

	// If no intercept matches, allow pod as-is
	if matchedIntercept == nil {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Create patches to make pod unschedulable
	var patches []patchOperation

	// Add node selector to prevent scheduling
	if pod.Spec.NodeSelector == nil {
		patches = append(patches, patchOperation{
			Op:   "add",
			Path: "/spec/nodeSelector",
			Value: map[string]string{
				"kloudlite.io/intercept-hold": "non-existing",
			},
		})
	} else {
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/spec/nodeSelector/kloudlite.io~1intercept-hold",
			Value: "non-existing",
		})
	}

	// Add annotation to track which intercept caused this
	if pod.Annotations == nil {
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/metadata/annotations",
			Value: make(map[string]string),
		})
	}

	escapedKey := "intercepts.kloudlite.io~1held-by"
	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  fmt.Sprintf("/metadata/annotations/%s", escapedKey),
		Value: matchedIntercept.Name,
	})

	w.logger.Info(fmt.Sprintf("Holding pod '%s' in namespace '%s' due to service intercept '%s'",
		pod.Name, pod.Namespace, matchedIntercept.Name))

	// Create patch response
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		w.logger.Error("Failed to marshal patches: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: true, // Allow without modification on error
		}
	}

	patchType := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
	}
}

// podMatchesSelector checks if a pod's labels match a service selector
func podMatchesSelector(pod *corev1.Pod, selector map[string]string) bool {
	if pod.Labels == nil {
		return false
	}

	for key, value := range selector {
		if podValue, exists := pod.Labels[key]; !exists || podValue != value {
			return false
		}
	}

	return true
}
