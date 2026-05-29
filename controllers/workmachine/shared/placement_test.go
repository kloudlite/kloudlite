package shared

import "testing"

func TestWorkMachineAddOnPlacementTargetsWorkMachineNode(t *testing.T) {
	placement := WorkMachineAddOnPlacement("wm-1")

	if placement.NodeSelector["kloudlite.io/workmachine"] != "wm-1" {
		t.Fatalf("unexpected node selector: %#v", placement.NodeSelector)
	}
	if len(placement.Tolerations) != 3 {
		t.Fatalf("expected three tolerations, got %#v", placement.Tolerations)
	}
	if placement.Tolerations[0].Key != "kloudlite.io/workmachine" {
		t.Fatalf("expected workmachine taint toleration, got %#v", placement.Tolerations[0])
	}
}
