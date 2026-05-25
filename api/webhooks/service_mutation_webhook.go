package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/pkg/logger"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
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

	// Check if there's an active intercept for this service in any Environment whose
	// embedded compose resources run in this service's namespace.
	environmentList := &environmentsv1.EnvironmentList{}
	err := w.client.List(context.TODO(), environmentList)
	if err != nil {
		// Handle TLS/certificate errors gracefully for development environments
		if isTLSError(err) {
			w.logger.Warn("tls error when checking environments (development mode), allowing service to proceed: " + err.Error())
			return &admissionv1.AdmissionResponse{
				Allowed: true,
			}
		}
		w.logger.Error("failed to list environments: " + err.Error())
		// Allow the service to proceed without interception on error
		return &admissionv1.AdmissionResponse{
			Allowed: true,
		}
	}

	// Find an active intercept for this service
	var hasActiveIntercept bool
	for _, environment := range environmentList.Items {
		if environment.DeletionTimestamp != nil || environment.Spec.TargetNamespace != req.Namespace {
			continue
		}

		if environment.Status.ComposeStatus == nil {
			continue
		}

		for _, activeIntercept := range environment.Status.ComposeStatus.ActiveIntercepts {
			if activeIntercept.ServiceName == service.Name {
				hasActiveIntercept = true
				break
			}
		}

		if hasActiveIntercept {
			break
		}
	}

	// If no active intercept, allow service as-is
	if !hasActiveIntercept {
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

// isTLSError checks if the error is related to TLS certificate verification
func isTLSError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	tlsErrorStrings := []string{
		"tls: failed to verify certificate",
		"x509: certificate signed by unknown authority",
		"certificate not trusted",
		"certificate has expired",
		"certificate is not yet valid",
		"tls handshake error",
		"certificate authority",
	}

	for _, tlsStr := range tlsErrorStrings {
		if strings.Contains(errStr, tlsStr) {
			return true
		}
	}

	return false
}
