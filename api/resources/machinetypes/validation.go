package machinetypes

import (
	"fmt"
	"regexp"

	workmachinev1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var validNamePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

var validCategories = map[string]struct{}{
	"general":           {},
	"compute-optimized": {},
	"memory-optimized":  {},
	"gpu":               {},
	"development":       {},
}

func ValidateDefinition(machineType workmachinev1.MachineType) error {
	if !validNamePattern.MatchString(machineType.Name) {
		return fmt.Errorf("invalid machine type name: must be lowercase alphanumeric with hyphens")
	}
	if machineType.Spec.DisplayName == "" {
		return fmt.Errorf("displayName is required")
	}
	if machineType.Spec.Category == "" {
		return fmt.Errorf("category is required")
	}
	if _, ok := validCategories[machineType.Spec.Category]; !ok {
		return fmt.Errorf("invalid category: must be one of %v", ValidCategories())
	}
	if err := ValidateResources(machineType.Spec.Resources); err != nil {
		return fmt.Errorf("invalid resources: %v", err)
	}
	return nil
}

func ValidateActiveDisplayNameUnique(candidate workmachinev1.MachineType, existing []workmachinev1.MachineType) error {
	if !candidate.Spec.Active {
		return nil
	}
	for _, item := range existing {
		if item.Spec.Active && item.Spec.DisplayName == candidate.Spec.DisplayName && item.Name != candidate.Name {
			return fmt.Errorf("active machine type with display name %s already exists", candidate.Spec.DisplayName)
		}
	}
	return nil
}

func DefaultName(items []workmachinev1.MachineType) (string, bool) {
	for _, item := range items {
		if item.Spec.IsDefault {
			return item.Name, true
		}
	}
	return "", false
}

func ValidateUsable(name string, machineType workmachinev1.MachineType) error {
	if machineType.Name != name {
		return fmt.Errorf("machine type %s not found", name)
	}
	if !machineType.Spec.Active {
		return fmt.Errorf("machine type %s is not active", name)
	}
	return nil
}

func ValidateResources(resources workmachinev1.MachineResources) error {
	if resources.CPU == "" {
		return fmt.Errorf("CPU is required")
	}
	if _, err := resource.ParseQuantity(resources.CPU); err != nil {
		return fmt.Errorf("invalid CPU quantity: %v", err)
	}
	if resources.Memory == "" {
		return fmt.Errorf("memory is required")
	}
	if _, err := resource.ParseQuantity(resources.Memory); err != nil {
		return fmt.Errorf("invalid memory quantity: %v", err)
	}
	if resources.GPU != "" {
		if _, err := resource.ParseQuantity(resources.GPU); err != nil {
			return fmt.Errorf("invalid GPU quantity: %v", err)
		}
	}
	return nil
}

func ValidCategories() []string {
	return []string{"general", "compute-optimized", "memory-optimized", "gpu", "development"}
}
