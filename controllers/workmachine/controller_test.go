package workmachine

import (
	"context"
	"errors"
	"testing"

	workmachinev1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
)

func TestSkipCloudPermissionValidation(t *testing.T) {
	t.Setenv("KLOUDLITE_SKIP_CLOUD_PERMISSION_VALIDATION", "true")

	if !skipCloudPermissionValidation() {
		t.Fatal("expected cloud permission validation to be skipped")
	}
}

func TestSkipCloudPermissionValidationDefaultsToFalse(t *testing.T) {
	if skipCloudPermissionValidation() {
		t.Fatal("expected cloud permission validation to run by default")
	}
}

func TestSetupCloudProviderRejectsUnsupportedProvider(t *testing.T) {
	_, err := setupCloudProvider(context.Background(), Env{CloudProvider: workmachinev1.CloudProvider("unknown")})
	if err == nil {
		t.Fatal("expected unsupported cloud provider error")
	}
}

func TestValidateProviderPermissionsHonorsSkipToggle(t *testing.T) {
	t.Setenv("KLOUDLITE_SKIP_CLOUD_PERMISSION_VALIDATION", "true")
	provider := &recordingProvider{validateErr: errors.New("should be skipped")}

	if err := validateProviderPermissions(context.Background(), provider); err != nil {
		t.Fatalf("expected validation to be skipped, got %v", err)
	}
	if provider.validateCalls != 0 {
		t.Fatalf("expected no validation calls, got %d", provider.validateCalls)
	}
}

func TestWorkMachineLifecycleStepNamesPreserveOrder(t *testing.T) {
	r := &WorkMachineReconciler{}
	steps := r.lifecycleSteps()

	got := make([]string, 0, len(steps))
	for _, step := range steps {
		got = append(got, step.Name)
	}

	want := []string{
		"setup-namespace",
		"ensure-network-policy",
		"sync-wildcard-cert-secret",
		"setup-host-manager-RBAC",
		"ensure-ssh-host-keys",
		"ensure-sshd-config",
		"when-running/ensure-wm-ingress-controller",
		"when-stopped/cleanup-wm-ingress-controller",
		"when-running/ensure-host-manager",
		"when-stopped/cleanup-host-manager",
		"when-running/ensure-buildkit",
		"when-stopped/cleanup-buildkit",
		"handle-machine-type-change",
		"handle-node-reboot-request",
		"when-running/ensure-tunnel-server",
		"when-stopped/cleanup-tunnel-server",
		"when-running/ensure-code-analyzer",
		"when-stopped/cleanup-code-analyzer",
		"check-auto-shutdown",
		"setup cloud machine",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d steps, got %d: %#v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestWorkMachineAddOnPlacementPolicyTargetsWorkMachineNode(t *testing.T) {
	placement := workMachineAddOnPlacement("wm-1")

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

type recordingProvider struct {
	validateCalls int
	validateErr   error
}

func (p *recordingProvider) ValidatePermissions(context.Context) error {
	p.validateCalls++
	return p.validateErr
}

func (p *recordingProvider) CreateMachine(context.Context, *workmachinev1.WorkMachine) (*workmachinev1.MachineInfo, error) {
	return nil, nil
}

func (p *recordingProvider) GetMachineStatus(context.Context, string) (*workmachinev1.MachineInfo, error) {
	return nil, nil
}

func (p *recordingProvider) StartMachine(context.Context, string) error { return nil }

func (p *recordingProvider) StopMachine(context.Context, string) error { return nil }

func (p *recordingProvider) RebootMachine(context.Context, string) error { return nil }

func (p *recordingProvider) IncreaseVolumeSize(context.Context, string, int32) error { return nil }

func (p *recordingProvider) ChangeMachine(context.Context, string, string) error { return nil }

func (p *recordingProvider) DeleteMachine(context.Context, string) error { return nil }
