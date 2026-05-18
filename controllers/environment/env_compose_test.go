package environment

import (
	"context"
	"testing"

	"github.com/kloudlite/kloudlite/controllers/composition"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// TestMakeStringSet tests the makeStringSet helper
func TestMakeStringSet(t *testing.T) {
	tests := []struct {
		name     string
		items    []string
		expected map[string]bool
	}{
		{
			name:     "empty slice",
			items:    []string{},
			expected: map[string]bool{},
		},
		{
			name:     "single item",
			items:    []string{"item1"},
			expected: map[string]bool{"item1": true},
		},
		{
			name:     "multiple items",
			items:    []string{"item1", "item2", "item3"},
			expected: map[string]bool{"item1": true, "item2": true, "item3": true},
		},
		{
			name:     "duplicate items",
			items:    []string{"item1", "item1", "item2"},
			expected: map[string]bool{"item1": true, "item2": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeStringSet(tt.items)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected %s: %v, got %v", k, v, result[k])
				}
			}
		})
	}
}

func TestFindEnvironmentForComposeResourceUsesOwnershipLabels(t *testing.T) {
	reconciler := &EnvironmentReconciler{}
	resource := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-config",
			Namespace: "env-dev",
			Labels: map[string]string{
				composition.DockerCompositionLabel:    "dev",
				composition.EnvironmentNamespaceLabel: "wm-alice",
			},
		},
	}

	requests := reconciler.findEnvironmentForComposeResource(context.Background(), resource)

	if len(requests) != 1 {
		t.Fatalf("expected one reconcile request, got %#v", requests)
	}
	if requests[0] != (reconcile.Request{NamespacedName: types.NamespacedName{Name: "dev", Namespace: "wm-alice"}}) {
		t.Fatalf("expected request for owning environment, got %#v", requests[0])
	}
}

func TestFindEnvironmentForComposeResourceIgnoresUnownedResources(t *testing.T) {
	reconciler := &EnvironmentReconciler{}

	tests := []struct {
		name     string
		resource *corev1.ConfigMap
	}{
		{
			name: "no labels",
			resource: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "api-config",
				Namespace: "env-dev",
			}},
		},
		{
			name: "missing environment namespace label",
			resource: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name:      "api-config",
				Namespace: "env-dev",
				Labels: map[string]string{
					composition.DockerCompositionLabel: "dev",
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := reconciler.findEnvironmentForComposeResource(context.Background(), tt.resource)
			if len(requests) != 0 {
				t.Fatalf("expected no reconcile requests, got %#v", requests)
			}
		})
	}
}
