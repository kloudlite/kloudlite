package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DomainRequestWebhook struct {
	logger    logger.Logger
	k8sClient client.Client
}

func NewDomainRequestWebhook(logger logger.Logger, k8sClient client.Client) *DomainRequestWebhook {
	return &DomainRequestWebhook{
		logger:    logger,
		k8sClient: k8sClient,
	}
}

// ValidateDomainRequest handles validation webhook for DomainRequest CRD
func (w *DomainRequestWebhook) ValidateDomainRequest(c *gin.Context) {
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

func (w *DomainRequestWebhook) handleValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the DomainRequest object
	var domainRequest domainrequestv1.DomainRequest
	if err := json.Unmarshal(req.Object.Raw, &domainRequest); err != nil {
		w.logger.Error("Failed to unmarshal DomainRequest: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal DomainRequest object",
			},
		}
	}

	// Perform validation
	if err := w.validateDomainRequest(&domainRequest, req); err != nil {
		w.logger.Warn("DomainRequest validation failed: " + err.Error())
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

func (w *DomainRequestWebhook) validateDomainRequest(dr *domainrequestv1.DomainRequest, req *admissionv1.AdmissionRequest) error {
	ctx := context.Background()

	// Validate unique name across all namespaces
	// List all DomainRequests across all namespaces
	drList := &domainrequestv1.DomainRequestList{}
	if err := w.k8sClient.List(ctx, drList); err != nil {
		return fmt.Errorf("failed to list DomainRequests: %w", err)
	}

	// Check for name conflicts
	for _, existingDR := range drList.Items {
		// Skip the current resource if this is an update
		if req.Operation == admissionv1.Update {
			// For updates, skip if it's the same resource (same namespace + name)
			if existingDR.Namespace == req.Namespace && existingDR.Name == req.Name {
				continue
			}
		}

		// Check if another DomainRequest with the same name exists
		if existingDR.Name == dr.Name {
			return fmt.Errorf("DomainRequest with name '%s' already exists in namespace '%s'. DomainRequest names must be unique across all namespaces",
				dr.Name, existingDR.Namespace)
		}
	}

	return nil
}
