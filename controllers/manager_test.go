package controllers

import (
	"testing"
)

func TestPlatformAPIControllerNames(t *testing.T) {
	want := []string{"user", "workmachine"}
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
