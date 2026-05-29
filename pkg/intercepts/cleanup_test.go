package intercepts

import (
	"testing"

	environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestRemoveForWorkspaceKeepsOnlyInterceptsNotOwnedByWorkspace(t *testing.T) {
	intercepts := []environmentv1.ServiceInterceptConfig{
		{
			ServiceName: "api",
			WorkspaceRef: &corev1.ObjectReference{
				Name:      "ws-a",
				Namespace: "team-a",
			},
		},
		{
			ServiceName: "worker",
			WorkspaceRef: &corev1.ObjectReference{
				Name:      "ws-b",
				Namespace: "team-a",
			},
		},
		{
			ServiceName: "metrics",
			WorkspaceRef: &corev1.ObjectReference{
				Name:      "ws-a",
				Namespace: "team-b",
			},
		},
		{ServiceName: "legacy"},
	}

	kept, removed := RemoveForWorkspace(intercepts, "ws-a", "team-a")

	if len(removed) != 1 || removed[0] != "api" {
		t.Fatalf("expected only api removed, got %#v", removed)
	}
	if got, want := serviceNames(kept), []string{"worker", "metrics", "legacy"}; !sameStrings(got, want) {
		t.Fatalf("expected kept services %#v, got %#v", want, got)
	}
}

func TestRemoveForWorkspacePreservesNilInput(t *testing.T) {
	kept, removed := RemoveForWorkspace(nil, "ws-a", "team-a")

	if kept != nil {
		t.Fatalf("expected nil kept intercepts, got %#v", kept)
	}
	if len(removed) != 0 {
		t.Fatalf("expected no removed intercepts, got %#v", removed)
	}
}

func serviceNames(intercepts []environmentv1.ServiceInterceptConfig) []string {
	names := make([]string, 0, len(intercepts))
	for _, intercept := range intercepts {
		names = append(names, intercept.ServiceName)
	}
	return names
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
