package controllers

import (
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
