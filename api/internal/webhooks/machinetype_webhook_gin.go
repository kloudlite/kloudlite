package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MachineTypeGinWebhook struct {
	logger    logger.Logger
	k8sClient client.Client
}

func NewMachineTypeGinWebhook(logger logger.Logger, k8sClient client.Client) *MachineTypeGinWebhook {
	return &MachineTypeGinWebhook{
		logger:    logger,
		k8sClient: k8sClient,
	}
}

// ValidateMachineType handles validation webhook for MachineType CRD
func (w *MachineTypeGinWebhook) ValidateMachineType(c *gin.Context) {
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

// MutateMachineType handles mutation webhook for MachineType CRD
func (w *MachineTypeGinWebhook) MutateMachineType(c *gin.Context) {
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

func (w *MachineTypeGinWebhook) handleValidation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	var machineType machinesv1.MachineType
	var err error

	// For DELETE operations, the object is in OldObject, not Object
	if req.Operation == admissionv1.Delete {
		if err := json.Unmarshal(req.OldObject.Raw, &machineType); err != nil {
			w.logger.Error("Failed to unmarshal machine type from OldObject: " + err.Error())
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: "Failed to unmarshal machine type object",
				},
			}
		}
		err = w.validateDelete(&machineType)
	} else {
		// For CREATE and UPDATE, parse from Object
		if err := json.Unmarshal(req.Object.Raw, &machineType); err != nil {
			w.logger.Error("Failed to unmarshal machine type: " + err.Error())
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: "Failed to unmarshal machine type object",
				},
			}
		}

		// Perform validation based on operation
		switch req.Operation {
		case admissionv1.Create:
			err = w.validateCreate(&machineType)
		case admissionv1.Update:
			err = w.validateUpdate(&machineType)
		}
	}

	if err != nil {
		w.logger.Warn("MachineType validation failed: " + err.Error())
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

func (w *MachineTypeGinWebhook) handleMutation(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Parse the machine type object
	var machineType machinesv1.MachineType
	if err := json.Unmarshal(req.Object.Raw, &machineType); err != nil {
		w.logger.Error("Failed to unmarshal machine type: " + err.Error())
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: "Failed to unmarshal machine type object",
			},
		}
	}

	// Create patches for mutations
	var patches []map[string]interface{}

	// Set default category if not specified
	if machineType.Spec.Category == "" {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/category",
			"value": "general",
		})
		machineType.Spec.Category = "general" // Update for label generation
	}

	// Ensure labels map exists
	if machineType.Labels == nil {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels",
			"value": map[string]string{},
		})
	}

	// Add category label
	categoryLabelPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/kloudlite.io~1machine-type-category",
		"value": machineType.Spec.Category,
	}
	patches = append(patches, categoryLabelPatch)

	// Add machine-type-name label
	machineTypeNameLabelPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/kloudlite.io~1machine-type-name",
		"value": machineType.Name,
	}
	patches = append(patches, machineTypeNameLabelPatch)

	// Add active label
	activeValue := "false"
	if machineType.Spec.Active {
		activeValue = "true"
	}
	activeLabelPatch := map[string]interface{}{
		"op":    "add",
		"path":  "/metadata/labels/kloudlite.io~1machine-type-active",
		"value": activeValue,
	}
	patches = append(patches, activeLabelPatch)

	// Set default priority if not specified (0 is valid, so check for unset)
	if machineType.Spec.Priority == 0 {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/spec/priority",
			"value": 100,
		})
	}

	// Check if there are any default machine types
	// If no default exists and this is not explicitly set to false, make it default
	// Skip this logic if being updated by webhook to prevent race condition
	ctx := context.Background()
	if !machineType.Spec.IsDefault && (machineType.Annotations == nil || machineType.Annotations["kloudlite.io/webhook-updating"] != "true") {
		existingTypes := &machinesv1.MachineTypeList{}
		if err := w.k8sClient.List(ctx, existingTypes); err == nil {
			hasDefault := false
			for _, mt := range existingTypes.Items {
				// Skip the current machine type being created
				if mt.Name == machineType.Name {
					continue
				}
				// Check both spec.isDefault and the label to avoid race conditions
				if mt.Spec.IsDefault || (mt.Labels != nil && mt.Labels["kloudlite.io/machinetype.default"] == "true") {
					hasDefault = true
					break
				}
			}

			// If no default exists, make this one default
			if !hasDefault {
				w.logger.Info(fmt.Sprintf("No default machine type found, setting '%s' as default", machineType.Name))
				patches = append(patches, map[string]interface{}{
					"op":    "add",
					"path":  "/spec/isDefault",
					"value": true,
				})
				machineType.Spec.IsDefault = true // Update for label generation below
			}
		}
	}

	// Handle default machine type - ensure only one is default at a time
	if machineType.Spec.IsDefault {
		// Add default label to this machine type
		patches = append(patches, map[string]any{
			"op":    "add",
			"path":  "/metadata/labels/kloudlite.io~1machinetype.default",
			"value": "true",
		})

		// Remove default flag from all other machine types
		ctx := context.Background()
		if err := w.removeDefaultFromOthers(ctx, machineType.Name); err != nil {
			w.logger.Error("Failed to remove default flag from other machine types: " + err.Error())
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: fmt.Sprintf("Failed to update other machine types: %v", err),
				},
			}
		}
	} else if _, ok := machineType.Labels["kloudlite.io/machinetype.default"]; ok {
		// Remove default label if isDefault is false but label exists
		patches = append(patches, map[string]any{
			"op":   "remove",
			"path": "/metadata/labels/kloudlite.io~1machinetype.default",
		})
	}

	// Remove webhook-updating annotation if it exists
	if machineType.Annotations != nil && machineType.Annotations["kloudlite.io/webhook-updating"] == "true" {
		patches = append(patches, map[string]any{
			"op":   "remove",
			"path": "/metadata/annotations/kloudlite.io~1webhook-updating",
		})
	}

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

func (w *MachineTypeGinWebhook) validateCreate(machineType *machinesv1.MachineType) error {
	ctx := context.Background()

	// Validate name format
	if !isValidMachineTypeName(machineType.Name) {
		return fmt.Errorf("invalid machine type name: must be lowercase alphanumeric with hyphens")
	}

	// Validate DisplayName is provided (required field)
	if machineType.Spec.DisplayName == "" {
		return fmt.Errorf("displayName is required")
	}

	// Validate category is provided
	if machineType.Spec.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Validate category value
	validCategories := []string{"general", "compute-optimized", "memory-optimized", "gpu", "development"}
	isValidCategory := false
	for _, valid := range validCategories {
		if machineType.Spec.Category == valid {
			isValidCategory = true
			break
		}
	}
	if !isValidCategory {
		return fmt.Errorf("invalid category: must be one of %v", validCategories)
	}

	// Validate resources
	if err := w.validateResources(&machineType.Spec.Resources); err != nil {
		return fmt.Errorf("invalid resources: %v", err)
	}

	// Check for duplicate display names among active types
	if machineType.Spec.Active {
		existingTypes := &machinesv1.MachineTypeList{}
		if err := w.k8sClient.List(ctx, existingTypes); err != nil {
			return fmt.Errorf("failed to list machine types: %v", err)
		}

		for _, existing := range existingTypes.Items {
			if existing.Spec.Active &&
				existing.Spec.DisplayName == machineType.Spec.DisplayName &&
				existing.Name != machineType.Name {
				return fmt.Errorf("active machine type with display name %s already exists", machineType.Spec.DisplayName)
			}
		}
	}

	// Note: We don't validate single default here because the mutation webhook
	// will automatically remove the default flag from other machine types

	return nil
}

func (w *MachineTypeGinWebhook) validateUpdate(machineType *machinesv1.MachineType) error {
	ctx := context.Background()

	// Validate DisplayName is provided (required field)
	if machineType.Spec.DisplayName == "" {
		return fmt.Errorf("displayName is required")
	}

	// Validate category is provided
	if machineType.Spec.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Validate category value
	validCategories := []string{"general", "compute-optimized", "memory-optimized", "gpu", "development"}
	isValidCategory := false
	for _, valid := range validCategories {
		if machineType.Spec.Category == valid {
			isValidCategory = true
			break
		}
	}
	if !isValidCategory {
		return fmt.Errorf("invalid category: must be one of %v", validCategories)
	}

	// Validate resources
	if err := w.validateResources(&machineType.Spec.Resources); err != nil {
		return fmt.Errorf("invalid resources: %v", err)
	}

	// Check for duplicate display names if becoming active
	if machineType.Spec.Active {
		existingTypes := &machinesv1.MachineTypeList{}
		if err := w.k8sClient.List(ctx, existingTypes); err != nil {
			return fmt.Errorf("failed to list machine types: %v", err)
		}

		for _, existing := range existingTypes.Items {
			if existing.Spec.Active &&
				existing.Spec.DisplayName == machineType.Spec.DisplayName &&
				existing.Name != machineType.Name {
				return fmt.Errorf("active machine type with display name %s already exists", machineType.Spec.DisplayName)
			}
		}
	}

	// Note: We don't validate single default here because the mutation webhook
	// will automatically remove the default flag from other machine types

	return nil
}

func (w *MachineTypeGinWebhook) validateDelete(machineType *machinesv1.MachineType) error {
	ctx := context.Background()

	// Check if any WorkMachines are using this type
	workMachines := &machinesv1.WorkMachineList{}
	if err := w.k8sClient.List(ctx, workMachines); err != nil {
		return fmt.Errorf("failed to list work machines: %v", err)
	}

	var workMachinesUsing []string
	for _, machine := range workMachines.Items {
		if machine.Spec.MachineType == machineType.Name {
			workMachinesUsing = append(workMachinesUsing, machine.Name)
		}
	}

	if len(workMachinesUsing) > 0 {
		if len(workMachinesUsing) <= 5 {
			// Show specific WorkMachine names if there are 5 or fewer
			return fmt.Errorf("cannot delete machine type '%s': the following %d work machine(s) are using it: %v. Please delete or update these work machines first",
				machineType.Name, len(workMachinesUsing), workMachinesUsing)
		} else {
			// If there are more than 5, just show the count and first few
			return fmt.Errorf("cannot delete machine type '%s': %d work machines are using it (including: %v...). Please delete or update these work machines first",
				machineType.Name, len(workMachinesUsing), workMachinesUsing[:5])
		}
	}

	return nil
}

// Helper functions

func isValidMachineTypeName(name string) bool {
	// Must be lowercase alphanumeric with hyphens, start and end with alphanumeric
	validName := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return validName.MatchString(name)
}

func (w *MachineTypeGinWebhook) validateResources(resources *machinesv1.MachineResources) error {
	// Validate CPU is provided (required field)
	if resources.CPU == "" {
		return fmt.Errorf("CPU is required")
	}
	if _, err := resource.ParseQuantity(resources.CPU); err != nil {
		return fmt.Errorf("invalid CPU quantity: %v", err)
	}

	// Validate Memory is provided (required field)
	if resources.Memory == "" {
		return fmt.Errorf("memory is required")
	}
	if _, err := resource.ParseQuantity(resources.Memory); err != nil {
		return fmt.Errorf("invalid memory quantity: %v", err)
	}

	// Validate GPU if specified (optional field)
	if resources.GPU != "" {
		if _, err := resource.ParseQuantity(resources.GPU); err != nil {
			return fmt.Errorf("invalid GPU quantity: %v", err)
		}
	}

	return nil
}

// validateSingleDefault ensures only one machine type is marked as default
func (w *MachineTypeGinWebhook) validateSingleDefault(ctx context.Context, currentName string) error {
	existingTypes := &machinesv1.MachineTypeList{}
	if err := w.k8sClient.List(ctx, existingTypes); err != nil {
		return fmt.Errorf("failed to list machine types: %v", err)
	}

	for _, existing := range existingTypes.Items {
		if existing.Spec.IsDefault && existing.Name != currentName {
			return fmt.Errorf("another machine type '%s' is already marked as default; only one default machine type is allowed", existing.Name)
		}
	}

	return nil
}

// removeDefaultFromOthers removes the default flag from all machine types except the specified one
func (w *MachineTypeGinWebhook) removeDefaultFromOthers(ctx context.Context, currentName string) error {
	existingTypes := &machinesv1.MachineTypeList{}
	if err := w.k8sClient.List(ctx, existingTypes); err != nil {
		return fmt.Errorf("failed to list machine types: %v", err)
	}

	for i := range existingTypes.Items {
		machineType := &existingTypes.Items[i]

		// Skip the current machine type being created/updated
		if machineType.Name == currentName {
			continue
		}

		// If this machine type is marked as default, update it
		if machineType.Spec.IsDefault {
			w.logger.Info(fmt.Sprintf("Removing default flag from machine type: %s", machineType.Name))

			// Update the spec
			machineType.Spec.IsDefault = false

			// Remove the default label if it exists
			if machineType.Labels != nil {
				delete(machineType.Labels, "kloudlite.io/machinetype.default")
			}

			// Add annotation to skip auto-default logic in mutation webhook
			if machineType.Annotations == nil {
				machineType.Annotations = make(map[string]string)
			}
			machineType.Annotations["kloudlite.io/webhook-updating"] = "true"

			// Update the machine type
			if err := w.k8sClient.Update(ctx, machineType); err != nil {
				return fmt.Errorf("failed to update machine type '%s': %v", machineType.Name, err)
			}

			w.logger.Info(fmt.Sprintf("Successfully removed default flag from machine type: %s", machineType.Name))
		}
	}

	return nil
}
