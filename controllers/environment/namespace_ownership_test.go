package environment

import (
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestApplyEnvironmentNamespaceOwnership(t *testing.T) {
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "dev", Namespace: "wm-alice", UID: types.UID("env-uid")},
	}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "env-dev"}}

	applyEnvironmentNamespaceOwnership(ns, env)

	if ns.Labels[environmentManagedLabel] != "true" {
		t.Fatalf("managed label = %q, want true", ns.Labels[environmentManagedLabel])
	}
	if ns.Labels[environmentNameLabel] != "dev" {
		t.Fatalf("environment label = %q, want dev", ns.Labels[environmentNameLabel])
	}
	if ns.Labels[environmentNamespaceLabel] != "wm-alice" {
		t.Fatalf("environment namespace label = %q, want wm-alice", ns.Labels[environmentNamespaceLabel])
	}
	if ns.Annotations[environmentUIDAnnotation] != "env-uid" {
		t.Fatalf("environment uid annotation = %q, want env-uid", ns.Annotations[environmentUIDAnnotation])
	}
}

func TestEnvironmentNamespaceOwnershipMatches(t *testing.T) {
	env := &environmentsv1.Environment{ObjectMeta: metav1.ObjectMeta{Name: "dev", Namespace: "wm-alice", UID: types.UID("env-uid")}}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "env-dev"}}
	applyEnvironmentNamespaceOwnership(ns, env)

	if err := validateEnvironmentNamespaceOwnership(ns, env); err != nil {
		t.Fatalf("validateEnvironmentNamespaceOwnership returned error: %v", err)
	}
}

func TestEnvironmentNamespaceOwnershipRejectsMissingMarkers(t *testing.T) {
	env := &environmentsv1.Environment{ObjectMeta: metav1.ObjectMeta{Name: "dev", Namespace: "wm-alice", UID: types.UID("env-uid")}}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "env-dev"}}

	if err := validateEnvironmentNamespaceOwnership(ns, env); err == nil {
		t.Fatalf("validateEnvironmentNamespaceOwnership returned nil, want mismatch error")
	}
}

func TestEnvironmentNamespaceOwnershipRejectsDifferentUID(t *testing.T) {
	env := &environmentsv1.Environment{ObjectMeta: metav1.ObjectMeta{Name: "dev", Namespace: "wm-alice", UID: types.UID("env-uid")}}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "env-dev"}}
	applyEnvironmentNamespaceOwnership(ns, env)
	ns.Annotations[environmentUIDAnnotation] = "other-uid"

	if err := validateEnvironmentNamespaceOwnership(ns, env); err == nil {
		t.Fatalf("validateEnvironmentNamespaceOwnership returned nil, want uid mismatch error")
	}
}

func TestEnvironmentNamespaceOwnershipMarkersOverrideCustomMetadata(t *testing.T) {
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "dev", Namespace: "wm-alice", UID: types.UID("env-uid")},
		Spec: environmentsv1.EnvironmentSpec{
			OwnedBy: "alice@example.com",
			Labels: map[string]string{
				environmentManagedLabel:   "false",
				environmentNameLabel:      "other-env",
				environmentNamespaceLabel: "wm-other",
				environmentUIDAnnotation:  "label-uid",
				"custom":                  "value",
			},
			Annotations: map[string]string{
				environmentUIDAnnotation:       "other-uid",
				environmentCreatedByAnnotation: "mallory@example.com",
				"custom-annotation":            "annotation-value",
			},
		},
	}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "env-dev"}}
	reconciler := &EnvironmentReconciler{}

	reconciler.applyLabelsAndAnnotations(ns, env)

	if ns.Labels[environmentManagedLabel] != "true" {
		t.Fatalf("managed label = %q, want true", ns.Labels[environmentManagedLabel])
	}
	if ns.Labels[environmentNameLabel] != "dev" {
		t.Fatalf("environment label = %q, want dev", ns.Labels[environmentNameLabel])
	}
	if ns.Labels[environmentNamespaceLabel] != "wm-alice" {
		t.Fatalf("environment namespace label = %q, want wm-alice", ns.Labels[environmentNamespaceLabel])
	}
	if ns.Annotations[environmentUIDAnnotation] != "env-uid" {
		t.Fatalf("environment uid annotation = %q, want env-uid", ns.Annotations[environmentUIDAnnotation])
	}
	if ns.Annotations[environmentCreatedByAnnotation] != "alice@example.com" {
		t.Fatalf("created by annotation = %q, want alice@example.com", ns.Annotations[environmentCreatedByAnnotation])
	}
	if ns.Labels["custom"] != "value" {
		t.Fatalf("custom label = %q, want value", ns.Labels["custom"])
	}
	if ns.Annotations["custom-annotation"] != "annotation-value" {
		t.Fatalf("custom annotation = %q, want annotation-value", ns.Annotations["custom-annotation"])
	}
}
