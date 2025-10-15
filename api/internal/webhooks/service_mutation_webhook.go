package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
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
		if intercept.DeletionTimestamp == nil && intercept.Status.Phase == "Active" {
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

	// With SOCAT-based interception, the SOCAT pod already has the original service selector labels
	// copied from the service. The service doesn't need modification - it will naturally route
	// to the SOCAT pod using its existing selector.
	//
	// We DO NOT modify the service selector because:
	// 1. SOCAT pod has the original service selector labels (copied in controller)
	// 2. Service continues to work with its original selector
	// 3. Traffic flows: Service -> SOCAT pod (via original labels) -> Workspace (via headless service)
	w.logger.Info(fmt.Sprintf("Service '%s' is being intercepted via SOCAT pod with matching labels, no modification needed",
		service.Name))

	return &admissionv1.AdmissionResponse{
		Allowed: true,
	}
}
