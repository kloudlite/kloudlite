package intercepts

import (
	"testing"

	environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestActiveStatusesFiltersByWorkspace(t *testing.T) {
	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "dev"},
		Status: environmentv1.EnvironmentStatus{
			ComposeStatus: &environmentv1.CompositionStatus{
				ActiveIntercepts: []environmentv1.InterceptStatus{
					{ServiceName: "api", WorkspaceName: "ws-a", Phase: "active", Message: "ready"},
					{ServiceName: "worker", WorkspaceName: "ws-b", Phase: "creating"},
				},
			},
		},
	}

	statuses := ActiveStatuses(env, "ws-a")

	if len(statuses) != 1 {
		t.Fatalf("expected one status, got %#v", statuses)
	}
	if statuses[0] != (ActiveStatus{EnvironmentName: "dev", ServiceName: "api", Phase: "active", Message: "ready"}) {
		t.Fatalf("unexpected status %#v", statuses[0])
	}
}

func TestContextForWorkspaceCombinesActiveStatusWithSpecPortMappings(t *testing.T) {
	env := &environmentv1.Environment{
		Spec: environmentv1.EnvironmentSpec{
			Compose: &environmentv1.CompositionSpec{
				Intercepts: []environmentv1.ServiceInterceptConfig{
					{
						ServiceName: "api",
						PortMappings: []environmentv1.PortMapping{
							{ServicePort: 80, WorkspacePort: 8080, Protocol: corev1.ProtocolTCP},
						},
					},
					{
						ServiceName: "worker",
						PortMappings: []environmentv1.PortMapping{
							{ServicePort: 9000, WorkspacePort: 9001, Protocol: corev1.ProtocolTCP},
						},
					},
				},
			},
		},
		Status: environmentv1.EnvironmentStatus{
			ComposeStatus: &environmentv1.CompositionStatus{
				ActiveIntercepts: []environmentv1.InterceptStatus{
					{ServiceName: "api", WorkspaceName: "ws-a"},
					{ServiceName: "worker", WorkspaceName: "ws-b"},
				},
			},
		},
	}

	entries := ContextForWorkspace(env, "ws-a")

	if len(entries) != 1 {
		t.Fatalf("expected one context entry, got %#v", entries)
	}
	if entries[0].ServiceName != "api" {
		t.Fatalf("expected api context entry, got %#v", entries[0])
	}
	if len(entries[0].PortMappings) != 1 || entries[0].PortMappings[0].WorkspacePort != 8080 {
		t.Fatalf("expected api port mapping, got %#v", entries[0].PortMappings)
	}
}
