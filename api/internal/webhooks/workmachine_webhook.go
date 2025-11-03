package webhooks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/config"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkMachineWebhook struct {
	logger    logger.Logger
	k8sClient client.Client
	config    *config.Config
}

func NewWorkMachineWebhook(
	logger logger.Logger, k8sClient client.Client, cfg *config.Config,
) *WorkMachineWebhook {
	return &WorkMachineWebhook{
		logger:    logger,
		k8sClient: k8sClient,
		config:    cfg,
	}
}

// ValidateWorkMachine handles validation webhook for WorkMachine CRD
func (w *WorkMachineWebhook) ValidateWorkMachine(c *gin.Context) {
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

// MutateWorkMachine handles mutation webhook for WorkMachine CRD
func (w *WorkMachineWebhook) MutateWorkMachine(c *gin.Context) {
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

func (w *WorkMachineWebhook) handleValidation(
	req *admissionv1.AdmissionRequest,
) *admissionv1.AdmissionResponse {
	// Parse the work machine object
	var machine machinesv1.WorkMachine
	if err := json.Unmarshal(req.Object.Raw, &machine); err != nil {
		w.logger.Error("Failed to unmarshal work machine: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal work machine object",
			},
		}
	}

	// Perform validation
	if err := w.validateWorkMachine(&machine, req.Operation); err != nil {
		w.logger.Warn("WorkMachine validation failed: " + err.Error())
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

func (w *WorkMachineWebhook) handleMutation(
	req *admissionv1.AdmissionRequest,
) *admissionv1.AdmissionResponse {
	// Parse the work machine object
	var machine machinesv1.WorkMachine
	if err := json.Unmarshal(req.Object.Raw, &machine); err != nil {
		w.logger.Error("Failed to unmarshal work machine: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal work machine object",
			},
		}
	}

	// Create patches for mutations
	var patches []map[string]interface{}

	// Generate name if not provided (one machine per user)
	if machine.Name == "" {
		owner := machine.Spec.OwnedBy
		// Create a deterministic name based on owner
		sanitizedOwner := strings.ReplaceAll(owner, "@", "-at-")
		sanitizedOwner = strings.ReplaceAll(sanitizedOwner, ".", "-")
		generatedName := fmt.Sprintf("wm-%s", sanitizedOwner)

		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/name",
			"value": generatedName,
		})
		machine.Name = generatedName // Update for label generation below
	}

	// Ensure labels map exists
	if machine.Labels == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels",
			"value": map[string]string{},
		})
	}

	// Find user by username or email to get the actual user ID
	ctx := context.Background()
	var userID, userEmail string
	ownedBy := machine.Spec.OwnedBy

	if strings.Contains(ownedBy, "@") {
		// OwnedBy is an email, lookup user
		userList := &platformv1alpha1.UserList{}
		if err := w.k8sClient.List(ctx, userList); err == nil {
			for _, user := range userList.Items {
				if user.Spec.Email == ownedBy {
					userID = user.Name
					userEmail = user.Spec.Email
					break
				}
			}
		}

		if userID == "" {
			// If user not found, use sanitized email as userID
			userID = strings.ReplaceAll(strings.Split(ownedBy, "@")[0], ".", "-")
			userEmail = ownedBy
		}
	} else {
		// OwnedBy is a username, lookup user
		user := &platformv1alpha1.User{}
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: ownedBy}, user); err == nil {
			userID = user.Name
			userEmail = user.Spec.Email
		} else {
			// User not found, use the ownedBy value as userID
			userID = ownedBy
		}
	}

	// Add owned-by label
	ownerLabelPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/kloudlite.io~1owned-by",
		"value": userID,
	}
	patches = append(patches, ownerLabelPatch)

	// Add owner-email label (base64 encoded)
	if userEmail != "" {
		encodedEmail := base64.URLEncoding.EncodeToString([]byte(userEmail))
		emailLabelPatch := map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/kloudlite.io~1owner-email",
			"value": encodedEmail,
		}
		patches = append(patches, emailLabelPatch)
	}

	// Add machine-type label
	machineTypeLabelPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/kloudlite.io~1machine-type",
		"value": machine.Spec.MachineType,
	}
	patches = append(patches, machineTypeLabelPatch)

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

func (w *WorkMachineWebhook) validateWorkMachine(
	machine *machinesv1.WorkMachine, operation admissionv1.Operation,
) error {
	ctx := context.Background()

	// Validate owner exists
	ownedBy := machine.Spec.OwnedBy
	if ownedBy == "" {
		return fmt.Errorf("ownedBy field is required")
	}

	var foundUser *platformv1alpha1.User

	if strings.Contains(ownedBy, "@") {
		// Check by email
		userList := &platformv1alpha1.UserList{}
		if err := w.k8sClient.List(ctx, userList); err != nil {
			return fmt.Errorf("failed to list users: %v", err)
		}

		for _, user := range userList.Items {
			if user.Spec.Email == ownedBy {
				foundUser = &user
				break
			}
		}
	} else {
		// Check by username
		user := &platformv1alpha1.User{}
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: ownedBy}, user); err == nil {
			foundUser = user
		}
	}

	if foundUser == nil {
		return fmt.Errorf("owner %s not found", ownedBy)
	}

	// Note: The "one workmachine per user" constraint is an application-level
	// business rule enforced in handlers, not a resource validation concern.
	// Webhooks validate resource fields; handlers enforce business logic.

	// Validate machine type exists and is active
	machineType := &machinesv1.MachineType{}
	if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: machine.Spec.MachineType}, machineType); err != nil {
		return fmt.Errorf("machine type %s not found", machine.Spec.MachineType)
	}

	if !machineType.Spec.Active {
		return fmt.Errorf("machine type %s is not active", machine.Spec.MachineType)
	}

	// On UPDATE, prevent changing owner
	if operation == admissionv1.Update {
		// We would need the old object to validate this properly
		// This is handled by comparing oldObj vs newObj in the admission request
		// For now, we'll just validate the new state is correct
	}

	// On DELETE, check if machine is running
	if operation == admissionv1.Delete {
		if machine.Status.State == machinesv1.MachineStateRunning {
			return fmt.Errorf("cannot delete a running machine, please stop it first")
		}
	}

	return nil
}
