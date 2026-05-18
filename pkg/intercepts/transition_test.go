package intercepts

import (
	"testing"

	environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestUpsertConfigReplacesExistingServiceIntercept(t *testing.T) {
	existing := []environmentv1.ServiceInterceptConfig{
		{ServiceName: "api", Enabled: false},
		{ServiceName: "worker", Enabled: true},
	}
	replacement := environmentv1.ServiceInterceptConfig{
		ServiceName: "api",
		Enabled:     true,
		WorkspaceRef: &corev1.ObjectReference{
			Name:      "ws-a",
			Namespace: "team-a",
		},
	}

	updated, replaced := UpsertConfig(existing, replacement)

	if !replaced {
		t.Fatalf("expected existing intercept to be replaced")
	}
	if len(updated) != 2 || updated[0].WorkspaceRef == nil || updated[0].WorkspaceRef.Name != "ws-a" {
		t.Fatalf("expected api intercept replaced in place, got %#v", updated)
	}
	if updated[1].ServiceName != "worker" {
		t.Fatalf("expected unrelated intercept order preserved, got %#v", updated)
	}
}

func TestRemoveServiceRemovesMatchingIntercept(t *testing.T) {
	existing := []environmentv1.ServiceInterceptConfig{
		{ServiceName: "api"},
		{ServiceName: "worker"},
	}

	updated, removed := RemoveService(existing, "api")

	if !removed {
		t.Fatalf("expected api intercept to be removed")
	}
	if got, want := serviceNames(updated), []string{"worker"}; !sameStrings(got, want) {
		t.Fatalf("expected remaining services %#v, got %#v", want, got)
	}
}
