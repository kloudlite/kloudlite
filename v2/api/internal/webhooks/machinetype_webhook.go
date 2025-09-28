package webhooks

import (
	"context"
	"fmt"
	"regexp"

	machinesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/machines/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// MachineTypeWebhook handles validation and mutation for MachineType resources
type MachineTypeWebhook struct {
	k8sClient client.Client
}

// NewMachineTypeWebhook creates a new MachineType webhook
func NewMachineTypeWebhook(k8sClient client.Client) *MachineTypeWebhook {
	return &MachineTypeWebhook{
		k8sClient: k8sClient,
	}
}

// Default implements admission.Defaulter for mutation
func (w *MachineTypeWebhook) Default(ctx context.Context, obj runtime.Object) error {
	machineType := obj.(*machinesv1.MachineType)

	// Set default category if not specified
	if machineType.Spec.Category == "" {
		machineType.Spec.Category = "general"
	}

	// Set default labels
	if machineType.Labels == nil {
		machineType.Labels = make(map[string]string)
	}

	machineType.Labels["kloudlite.io/machine-type-category"] = machineType.Spec.Category
	if machineType.Spec.Active {
		machineType.Labels["kloudlite.io/machine-type-active"] = "true"
	} else {
		machineType.Labels["kloudlite.io/machine-type-active"] = "false"
	}

	// Set default priority if not specified
	if machineType.Spec.Priority == 0 {
		machineType.Spec.Priority = 100
	}

	// Set default storage if not specified
	if machineType.Spec.Resources.Storage == "" {
		machineType.Spec.Resources.Storage = "50Gi"
	}

	// Set default ephemeral storage if not specified
	if machineType.Spec.Resources.EphemeralStorage == "" {
		machineType.Spec.Resources.EphemeralStorage = "10Gi"
	}

	return nil
}

// ValidateCreate implements admission.Validator for create operations
func (w *MachineTypeWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	machineType := obj.(*machinesv1.MachineType)

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

	// Validate resources (includes validation of required CPU, Memory, Storage)
	if err := validateResources(&machineType.Spec.Resources); err != nil {
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

	return nil
}

// ValidateUpdate implements admission.Validator for update operations
func (w *MachineTypeWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	newMachineType := newObj.(*machinesv1.MachineType)

	// Validate DisplayName is provided (required field)
	if newMachineType.Spec.DisplayName == "" {
		return fmt.Errorf("displayName is required")
	}

	// Validate category is provided
	if newMachineType.Spec.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Validate category value
	validCategories := []string{"general", "compute-optimized", "memory-optimized", "gpu", "development"}
	isValidCategory := false
	for _, valid := range validCategories {
		if newMachineType.Spec.Category == valid {
			isValidCategory = true
			break
		}
	}
	if !isValidCategory {
		return fmt.Errorf("invalid category: must be one of %v", validCategories)
	}

	// Validate resources (includes validation of required CPU, Memory, Storage)
	if err := validateResources(&newMachineType.Spec.Resources); err != nil {
		return fmt.Errorf("invalid resources: %v", err)
	}

	// Check for duplicate display names if becoming active
	if newMachineType.Spec.Active {
		existingTypes := &machinesv1.MachineTypeList{}
		if err := w.k8sClient.List(ctx, existingTypes); err != nil {
			return fmt.Errorf("failed to list machine types: %v", err)
		}

		for _, existing := range existingTypes.Items {
			if existing.Spec.Active &&
			   existing.Spec.DisplayName == newMachineType.Spec.DisplayName &&
			   existing.Name != newMachineType.Name {
				return fmt.Errorf("active machine type with display name %s already exists", newMachineType.Spec.DisplayName)
			}
		}
	}

	return nil
}

// ValidateDelete implements admission.Validator for delete operations
func (w *MachineTypeWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	machineType := obj.(*machinesv1.MachineType)

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

// InjectDecoder injects the decoder
func (w *MachineTypeWebhook) InjectDecoder(d *admission.Decoder) error {
	return nil
}

// Helper functions

func isValidMachineTypeName(name string) bool {
	// Must be lowercase alphanumeric with hyphens, start and end with alphanumeric
	validName := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return validName.MatchString(name)
}

func validateResources(resources *machinesv1.MachineResources) error {
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

	// Validate Storage is provided (required field)
	if resources.Storage == "" {
		return fmt.Errorf("storage is required")
	}
	if _, err := resource.ParseQuantity(resources.Storage); err != nil {
		return fmt.Errorf("invalid storage quantity: %v", err)
	}

	// Validate GPU if specified (optional field)
	if resources.GPU != "" {
		if _, err := resource.ParseQuantity(resources.GPU); err != nil {
			return fmt.Errorf("invalid GPU quantity: %v", err)
		}
	}

	// Validate EphemeralStorage if specified (optional field)
	if resources.EphemeralStorage != "" {
		if _, err := resource.ParseQuantity(resources.EphemeralStorage); err != nil {
			return fmt.Errorf("invalid ephemeral storage quantity: %v", err)
		}
	}

	return nil
}