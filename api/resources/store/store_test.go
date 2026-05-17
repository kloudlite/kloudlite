package store

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestStoreReadinessGatesReads(t *testing.T) {
	s := New()
	if s.Ready("configmaps", "default") {
		t.Fatal("expected scope to start not ready")
	}
	s.SetReady("configmaps", "default", true)
	if !s.Ready("configmaps", "default") {
		t.Fatal("expected scope to be ready")
	}
}

func TestStoreUpsertGetListAndDelete(t *testing.T) {
	s := New()
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default", Labels: map[string]string{"app": "demo", "kloudlite.io/hash": "abc"}}}

	s.Upsert("configmaps", cm)
	got, ok := s.Get("configmaps", "default", "app")
	if !ok || got.GetName() != "app" {
		t.Fatalf("expected to get stored configmap, ok=%v got=%#v", ok, got)
	}

	items := s.List("configmaps", "default", Selector{Labels: map[string]string{"app": "demo"}})
	if len(items) != 1 {
		t.Fatalf("expected one item by label, got %d", len(items))
	}

	items = s.List("configmaps", "default", Selector{Hash: "abc"})
	if len(items) != 1 {
		t.Fatalf("expected one item by hash, got %d", len(items))
	}

	s.Delete("configmaps", "default", "app")
	if _, ok := s.Get("configmaps", "default", "app"); ok {
		t.Fatal("expected deleted object to be missing")
	}
}

func TestStoreGetReturnsCopy(t *testing.T) {
	s := New()
	s.Upsert("configmaps", &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default", Labels: map[string]string{"app": "demo"}}})

	got, ok := s.Get("configmaps", "default", "app")
	if !ok {
		t.Fatal("expected configmap to exist")
	}
	got.SetLabels(map[string]string{"app": "mutated"})

	got, ok = s.Get("configmaps", "default", "app")
	if !ok {
		t.Fatal("expected configmap to exist")
	}
	if got.GetLabels()["app"] != "demo" {
		t.Fatalf("expected cached labels to be unchanged, got %#v", got.GetLabels())
	}
}

func TestStoreUpsertCopiesOriginalObject(t *testing.T) {
	s := New()
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default", Labels: map[string]string{"app": "demo"}}}

	s.Upsert("configmaps", cm)
	cm.SetLabels(map[string]string{"app": "mutated"})

	got, ok := s.Get("configmaps", "default", "app")
	if !ok {
		t.Fatal("expected configmap to exist")
	}
	if got.GetLabels()["app"] != "demo" {
		t.Fatalf("expected cached labels to be unchanged, got %#v", got.GetLabels())
	}
}

func TestStoreReplaceScopeReplacesObjectsAndMarksReady(t *testing.T) {
	s := New()
	s.Upsert("configmaps", &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "old", Namespace: "default"}})

	s.ReplaceScope("configmaps", "default", []client.Object{
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "new", Namespace: "default"}},
	})

	if !s.Ready("configmaps", "default") {
		t.Fatal("expected replaced scope to be ready")
	}
	if _, ok := s.Get("configmaps", "default", "old"); ok {
		t.Fatal("expected old object to be replaced")
	}
	if _, ok := s.Get("configmaps", "default", "new"); !ok {
		t.Fatal("expected new object to exist")
	}
}

func TestStoreReplaceScopeCopiesOriginalObjects(t *testing.T) {
	s := New()
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "default", Labels: map[string]string{"app": "demo"}}}

	s.ReplaceScope("configmaps", "default", []client.Object{cm})
	cm.SetLabels(map[string]string{"app": "mutated"})

	got, ok := s.Get("configmaps", "default", "app")
	if !ok {
		t.Fatal("expected configmap to exist")
	}
	if got.GetLabels()["app"] != "demo" {
		t.Fatalf("expected cached labels to be unchanged, got %#v", got.GetLabels())
	}
}
