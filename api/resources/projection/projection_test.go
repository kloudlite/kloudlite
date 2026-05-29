package projection

import (
	"testing"

	"github.com/kloudlite/kloudlite/api/resources/events"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSelectorMatchesLabelsAndHash(t *testing.T) {
	selector := Selector{Labels: map[string]string{"app": "demo"}, Hash: "abc"}
	labels := map[string]string{"app": "demo", HashLabel: "abc"}

	if !MatchesLabels(labels, selector) {
		t.Fatal("expected labels to match selector")
	}
	if MatchesLabels(map[string]string{"app": "demo", HashLabel: "other"}, selector) {
		t.Fatal("expected hash mismatch")
	}
	if MatchesLabels(map[string]string{"app": "other", HashLabel: "abc"}, selector) {
		t.Fatal("expected label mismatch")
	}
}

func TestDirtyTransitionCreatesScopedDirtyObjectWithClonedLabels(t *testing.T) {
	labels := map[string]string{"app": "demo"}
	dirty := NewDirtyObject("configmaps", "default", "app", events.DirtyReasonPendingPatch, labels)
	labels["app"] = "changed"

	if dirty.Resource != "configmaps" || dirty.Namespace != "default" || dirty.Name != "app" || dirty.Reason != events.DirtyReasonPendingPatch {
		t.Fatalf("unexpected dirty object: %#v", dirty)
	}
	if dirty.Labels["app"] != "demo" {
		t.Fatalf("expected labels to be cloned, got %#v", dirty.Labels)
	}
}

func TestScopedNamespaceClearsClusterScopedNamespace(t *testing.T) {
	if got := ScopedNamespace(true, "default"); got != "" {
		t.Fatalf("expected cluster scoped namespace to be empty, got %q", got)
	}
	if got := ScopedNamespace(false, "default"); got != "default" {
		t.Fatalf("expected namespaced namespace to be preserved, got %q", got)
	}
}

func TestObjectInitialStateIncludesObjectAndDirtyMetadata(t *testing.T) {
	object := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default"}}
	dirty := events.DirtyObject{Resource: "configmaps", Namespace: "default", Name: "app", Reason: events.DirtyReasonPendingPatch}

	state := ObjectInitialState(object, &dirty)

	if state.Object.GetName() != "app" {
		t.Fatalf("unexpected object: %#v", state.Object)
	}
	if state.Dirty == nil || state.Dirty.Name != "app" {
		t.Fatalf("expected dirty metadata, got %#v", state.Dirty)
	}
}
