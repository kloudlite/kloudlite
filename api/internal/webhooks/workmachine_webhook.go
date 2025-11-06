package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/config"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
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

	// Set ownedBy to "system" if not provided
	ctx := context.Background()
	if machine.Spec.OwnedBy == "" {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/ownedBy",
			"value": "system",
		})
		machine.Spec.OwnedBy = "system"
	}

	// Set displayName if not provided
	if machine.Spec.DisplayName == "" {
		displayName := fmt.Sprintf("WorkMachine for %s", machine.Spec.OwnedBy)
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/displayName",
			"value": displayName,
		})
	}

	// Set targetNamespace if not provided - generate as wm-{machine-name}
	if machine.Spec.TargetNamespace == "" {
		targetNS := fmt.Sprintf("wm-%s", machine.Name)
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/targetNamespace",
			"value": targetNS,
		})
		machine.Spec.TargetNamespace = targetNS // Update for label generation below
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
		"path":  "/metadata/labels/" + fn.LabelKeyEncoder("kloudlite.io/owned-by"),
		"value": userID,
	}
	patches = append(patches, ownerLabelPatch)

	// Add owner-email label (base64 encoded)
	if userEmail != "" {
		encodedEmail := fn.LabelValueEncoder(userEmail)
		emailLabelPatch := map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + fn.LabelKeyEncoder("kloudlite.io/owner-email"),
			"value": encodedEmail,
		}
		patches = append(patches, emailLabelPatch)
	}

	// Set default machine type if not provided
	if machine.Spec.MachineType == "" {
		// Find default machine type
		machineTypeList := &machinesv1.MachineTypeList{}
		if err := w.k8sClient.List(ctx, machineTypeList); err != nil {
			w.logger.Error("Failed to list machine types: " + err.Error())
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: "Failed to find default machine type",
				},
			}
		}

		var defaultMachineType string
		for _, mt := range machineTypeList.Items {
			if mt.Spec.IsDefault {
				defaultMachineType = mt.Name
				break
			}
		}

		if defaultMachineType == "" {
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: "No machine type specified and no default machine type found",
				},
			}
		}

		// Add patch to set the default machine type
		machineTypePatch := map[string]interface{}{
			"op":    "add",
			"path":  "/spec/machineType",
			"value": defaultMachineType,
		}
		patches = append(patches, machineTypePatch)
		machine.Spec.MachineType = defaultMachineType // Update for label generation below

		w.logger.Info(fmt.Sprintf("Auto-assigned default machine type '%s' to WorkMachine", defaultMachineType))
	}

	// Add machine-type label
	machineTypeLabelPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/" + fn.LabelKeyEncoder("kloudlite.io/machine-type"),
		"value": machine.Spec.MachineType,
	}
	patches = append(patches, machineTypeLabelPatch)

	// Add target-namespace label for easy validation and lookup
	if machine.Spec.TargetNamespace != "" {
		targetNamespaceLabelPatch := map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + fn.LabelKeyEncoder("kloudlite.io/target-namespace"),
			"value": machine.Spec.TargetNamespace,
		}
		patches = append(patches, targetNamespaceLabelPatch)
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

func (w *WorkMachineWebhook) validateWorkMachine(
	machine *machinesv1.WorkMachine, operation admissionv1.Operation,
) error {
	ctx := context.Background()

	// Note: ownedBy validation removed - users may not exist yet
	// If ownedBy is empty, it will be set to "system" in mutation webhook

	// Validate machine type exists and is active (only if machineType is specified)
	if machine.Spec.MachineType != "" {
		machineType := &machinesv1.MachineType{}
		if err := w.k8sClient.Get(ctx, client.ObjectKey{Name: machine.Spec.MachineType}, machineType); err != nil {
			return fmt.Errorf("machine type %s not found", machine.Spec.MachineType)
		}

		if !machineType.Spec.Active {
			return fmt.Errorf("machine type %s is not active", machine.Spec.MachineType)
		}
	}

	// Validate targetNamespace is unique across WorkMachines and not used by Environments
	if machine.Spec.TargetNamespace != "" && (operation == admissionv1.Create || operation == admissionv1.Update) {
		// Check if any other WorkMachine is using this targetNamespace (using label selector)
		workMachineList := &machinesv1.WorkMachineList{}
		if err := w.k8sClient.List(ctx, workMachineList, client.MatchingLabels{
			"kloudlite.io/target-namespace": machine.Spec.TargetNamespace,
		}); err != nil {
			return fmt.Errorf("failed to list workmachines: %v", err)
		}

		for _, wm := range workMachineList.Items {
			// Skip the current machine being created/updated
			if wm.Name == machine.Name {
				continue
			}

			return fmt.Errorf("targetNamespace '%s' is already used by WorkMachine '%s'. Each WorkMachine must have a unique targetNamespace",
				machine.Spec.TargetNamespace, wm.Name)
		}

		// Check if any Environment is using this namespace (using label selector)
		environmentList := &environmentsv1.EnvironmentList{}
		if err := w.k8sClient.List(ctx, environmentList, client.MatchingLabels{
			"kloudlite.io/target-namespace": machine.Spec.TargetNamespace,
		}); err != nil {
			return fmt.Errorf("failed to list environments: %v", err)
		}

		if len(environmentList.Items) > 0 {
			return fmt.Errorf("targetNamespace '%s' is already used by Environment '%s'. WorkMachine cannot use a namespace owned by an Environment",
				machine.Spec.TargetNamespace, environmentList.Items[0].Name)
		}
	}

	// On DELETE, check if machine is running
	if operation == admissionv1.Delete {
		if machine.Status.State == machinesv1.MachineStateRunning {
			return fmt.Errorf("cannot delete a running machine, please stop it first")
		}
	}

	return nil
}
