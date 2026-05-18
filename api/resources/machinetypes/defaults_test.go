package machinetypes

import (
	"testing"

	workmachinev1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
)

func TestDefaultsForAWS(t *testing.T) {
	items, err := Defaults("aws")
	if err != nil {
		t.Fatalf("Defaults(aws) error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected aws defaults to include machine types")
	}

	defaults := 0
	for _, item := range items {
		if item.Name == "" {
			t.Fatalf("machine type has empty name: %#v", item)
		}
		if item.Spec.DisplayName == "" {
			t.Fatalf("machine type %q has empty display name", item.Name)
		}
		if item.Spec.Resources.CPU == "" || item.Spec.Resources.Memory == "" {
			t.Fatalf("machine type %q has incomplete resources: %#v", item.Name, item.Spec.Resources)
		}
		if item.Spec.IsDefault {
			defaults++
		}
	}
	if defaults != 1 {
		t.Fatalf("expected exactly one default machine type, got %d", defaults)
	}
}

func TestDefaultsRejectUnknownProvider(t *testing.T) {
	if _, err := Defaults("digitalocean"); err == nil {
		t.Fatal("expected unknown provider to fail")
	}
}

func TestDefaultsForAllProvidersHaveStableDefaultNames(t *testing.T) {
	tests := []struct {
		provider    string
		defaultName string
	}{
		{provider: "aws", defaultName: "aws-t3-medium"},
		{provider: "gcp", defaultName: "gcp-e2-standard-2"},
		{provider: "azure", defaultName: "azure-standard-b2s"},
		{provider: "oci", defaultName: "oci-vm-standard-e4-flex-small"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			items, err := Defaults(tt.provider)
			if err != nil {
				t.Fatalf("Defaults(%s) error = %v", tt.provider, err)
			}
			defaults := []string{}
			for _, item := range items {
				if item.Spec.IsDefault {
					defaults = append(defaults, item.Name)
				}
			}
			if len(defaults) != 1 || defaults[0] != tt.defaultName {
				t.Fatalf("expected default %q, got %#v", tt.defaultName, defaults)
			}
		})
	}
}

func TestValidateDefinitionRejectsInvalidCategoryAndResources(t *testing.T) {
	item := machineType("custom", "Custom", "", "invalid", "not-cpu", "", "", false, 10)

	err := ValidateDefinition(item)

	if err == nil {
		t.Fatal("expected invalid machine type definition")
	}
}

func TestValidateActiveDisplayNameUniqueRejectsDuplicateActiveName(t *testing.T) {
	candidate := machineType("custom-a", "Shared", "", "general", "1", "1Gi", "", false, 10)
	candidate.Spec.Active = true
	existing := machineType("custom-b", "Shared", "", "general", "1", "1Gi", "", false, 10)
	existing.Spec.Active = true

	if err := ValidateActiveDisplayNameUnique(candidate, []workmachinev1.MachineType{existing}); err == nil {
		t.Fatal("expected duplicate active display name to fail")
	}
}

func TestDefaultNameFindsExactlyOneDefault(t *testing.T) {
	items, err := Defaults("aws")
	if err != nil {
		t.Fatal(err)
	}
	name, ok := DefaultName(items)
	if !ok || name != "aws-t3-medium" {
		t.Fatalf("expected aws-t3-medium default, got %q ok=%v", name, ok)
	}
}

func TestValidateUsableRejectsInactiveOrWrongMachineType(t *testing.T) {
	item := machineType("small", "Small", "", "general", "1", "1Gi", "", false, 10)
	item.Spec.Active = false

	if err := ValidateUsable("small", item); err == nil {
		t.Fatal("expected inactive machine type to fail")
	}
	item.Spec.Active = true
	if err := ValidateUsable("large", item); err == nil {
		t.Fatal("expected mismatched machine type name to fail")
	}
}
