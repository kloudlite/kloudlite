package azure

import "testing"

func TestWorkMachineAzureResourceNamesFromMachineID(t *testing.T) {
	names := workMachineAzureResourceNames("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/kl-workmachine-wm")

	if names.vm != "kl-workmachine-wm" {
		t.Fatalf("vm name = %q", names.vm)
	}
	if names.nic != "kl-workmachine-wm-nic" {
		t.Fatalf("nic name = %q", names.nic)
	}
	if names.publicIP != "kl-workmachine-wm-pip" {
		t.Fatalf("public ip name = %q", names.publicIP)
	}
}
