package controllers

import (
	"strings"
	"testing"
)

func TestPlatformAPIControllerNames(t *testing.T) {
	want := []string{"user", "workmachine-platform-scoped"}
	got := PlatformAPIControllerNames()

	if len(got) != len(want) {
		t.Fatalf("PlatformAPIControllerNames() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("PlatformAPIControllerNames() = %v, want %v", got, want)
		}
	}
}

func TestMachineScopedControllerNames(t *testing.T) {
	want := []string{"workmachine-manager", "environment", "workspace"}
	got := MachineScopedControllerNames()

	if len(got) != len(want) {
		t.Fatalf("MachineScopedControllerNames() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("MachineScopedControllerNames() = %v, want %v", got, want)
		}
	}
}

func TestMachineScopedRuntimeScopeUsesPodNamespaceAndWorkMachineName(t *testing.T) {
	t.Setenv("POD_NAMESPACE", "wm-alice")
	t.Setenv("NAMESPACE", "")
	t.Setenv("WORKMACHINE_NAME", "alice-dev")

	namespace, workMachineName, err := machineScopedRuntimeScope()
	if err != nil {
		t.Fatalf("machineScopedRuntimeScope() error = %v", err)
	}
	if namespace != "wm-alice" {
		t.Fatalf("namespace = %q, want %q", namespace, "wm-alice")
	}
	if workMachineName != "alice-dev" {
		t.Fatalf("workMachineName = %q, want %q", workMachineName, "alice-dev")
	}
}

func TestMachineScopedRuntimeScopeFallsBackToNamespace(t *testing.T) {
	t.Setenv("POD_NAMESPACE", "")
	t.Setenv("NAMESPACE", "wm-bob")
	t.Setenv("WORKMACHINE_NAME", "bob-dev")

	namespace, workMachineName, err := machineScopedRuntimeScope()
	if err != nil {
		t.Fatalf("machineScopedRuntimeScope() error = %v", err)
	}
	if namespace != "wm-bob" {
		t.Fatalf("namespace = %q, want %q", namespace, "wm-bob")
	}
	if workMachineName != "bob-dev" {
		t.Fatalf("workMachineName = %q, want %q", workMachineName, "bob-dev")
	}
}

func TestMachineScopedRuntimeScopeErrorsWhenNamespaceMissing(t *testing.T) {
	t.Setenv("POD_NAMESPACE", "")
	t.Setenv("NAMESPACE", "")
	t.Setenv("WORKMACHINE_NAME", "alice-dev")

	_, _, err := machineScopedRuntimeScope()
	if err == nil {
		t.Fatal("machineScopedRuntimeScope() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "namespace") {
		t.Fatalf("machineScopedRuntimeScope() error = %q, want namespace error", err)
	}
}

func TestMachineScopedRuntimeScopeErrorsWhenWorkMachineNameMissing(t *testing.T) {
	t.Setenv("POD_NAMESPACE", "wm-alice")
	t.Setenv("NAMESPACE", "")
	t.Setenv("WORKMACHINE_NAME", "")

	_, _, err := machineScopedRuntimeScope()
	if err == nil {
		t.Fatal("machineScopedRuntimeScope() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "WORKMACHINE_NAME") {
		t.Fatalf("machineScopedRuntimeScope() error = %q, want WORKMACHINE_NAME error", err)
	}
}
