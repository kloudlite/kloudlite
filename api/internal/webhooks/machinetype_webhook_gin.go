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

	// Perform validation based on operation
	var err error
	switch req.Operation {
	case admissionv1.Create:
		err = w.validateCreate(&machineType)
	case admissionv1.Update:
		err = w.validateUpdate(&machineType)
	case admissionv1.Delete:
		err = w.validateDelete(&machineType)
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

	// Validate only one machine type is marked as default
	if machineType.Spec.IsDefault {
		if err := w.validateSingleDefault(ctx, machineType.Name); err != nil {
			return err
		}
	}

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

	// Validate only one machine type is marked as default
	if machineType.Spec.IsDefault {
		if err := w.validateSingleDefault(ctx, machineType.Name); err != nil {
			return err
		}
	}

	return nil
}

func (w *MachineTypeGinWebhook) validateDelete(machineType *machinesv1.MachineType) error {
	ctx := context.Background()

	// Check if any WorkMachines are using this type
	workMachines := &machinesv1.WorkMachineList{}
	if err := w.k8sClient.List(ctx, workMachines); err != nil {
		return fmt.Errorf("failed to list work machines: %v", err)
	}

	inUseCount := 0
	for _, machine := range workMachines.Items {
		if machine.Spec.MachineType == machineType.Name {
			inUseCount++
		}
	}

	if inUseCount > 0 {
		return fmt.Errorf("cannot delete machine type %s: %d work machines are using it", machineType.Name, inUseCount)
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
