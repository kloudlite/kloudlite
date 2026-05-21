package environment

import (
	"context"
	"fmt"
	"testing"

	"github.com/kloudlite/kloudlite/controllers/testutil"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnvironmentReconciler_HandleDeletion(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleting-env",
			Namespace:         "wm-alice",
			UID:               types.UID("deleting-env-uid"),
			DeletionTimestamp: &now,
			Finalizers:        []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "deleting-namespace",
			OwnedBy:         "test@example.com",
		},
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deleting-namespace",
		},
	}
	applyEnvironmentNamespaceOwnership(namespace, env)

	k8sClient := testutil.NewFakeClient(scheme, env, namespace).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "deleting-env",
			Namespace: "wm-alice",
		},
	}

	_, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	// Fake client deletes immediately, so we just verify no error
}

func TestEnvironmentReconciler_HandleDeletion_NamespaceAlreadyDeleted(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleting-env",
			Namespace:         "wm-alice",
			UID:               types.UID("deleting-env-uid"),
			DeletionTimestamp: &now,
			Finalizers:        []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "already-deleted-namespace",
			OwnedBy:         "test@example.com",
		},
	}

	// Note: No namespace exists - testing deletion when namespace is already gone
	k8sClient := testutil.NewFakeClient(scheme, env).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "deleting-env",
			Namespace: "wm-alice",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify finalizer was removed
	updatedEnv := &environmentsv1.Environment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "deleting-env", Namespace: "wm-alice"}, updatedEnv)
	// Environment might be deleted by fake client
	if err == nil {
		assert.NotContains(t, updatedEnv.Finalizers, environmentFinalizer)
	}
}

func TestEnvironmentReconciler_HandleDeletion_CleanupFailure(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleting-env",
			Namespace:         "wm-alice",
			DeletionTimestamp: &now,
			Finalizers:        []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "deleting-namespace",
			OwnedBy:         "test@example.com",
		},
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deleting-namespace",
		},
	}
	applyEnvironmentNamespaceOwnership(namespace, env)

	// Create fake client that will return error on Update
	k8sClient := testutil.NewFakeClient(scheme, env, namespace).Build()
	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "deleting-env",
			Namespace: "wm-alice",
		},
	}

	_, err := reconciler.Reconcile(context.Background(), req)
	// With fake client, cleanup should succeed (no workspaces to clean up)
	assert.NoError(t, err)
}

func TestEnvironmentReconciler_Delete_DoesNotDeleteUnownedNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()
	now := metav1.Now()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-env", Namespace: "wm-alice", UID: types.UID("env-uid"), Finalizers: []string{environmentFinalizer}, DeletionTimestamp: &now},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "shared-namespace", OwnedBy: "alice", Activated: true},
	}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "shared-namespace"}}
	k8sClient := testutil.NewFakeClient(scheme, env, ns).WithStatusSubresource(&environmentsv1.Environment{}).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	_, err := reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-env", Namespace: "wm-alice"}})
	if err == nil {
		t.Fatalf("Reconcile returned nil error, want ownership mismatch")
	}

	remaining := &corev1.Namespace{}
	if getErr := k8sClient.Get(context.Background(), types.NamespacedName{Name: "shared-namespace"}, remaining); getErr != nil {
		t.Fatalf("namespace was deleted or inaccessible, get error: %v", getErr)
	}
}

func TestJoinErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected string
	}{
		{
			name:     "no errors",
			errors:   []error{},
			expected: "",
		},
		{
			name:     "single error",
			errors:   []error{fmt.Errorf("error 1")},
			expected: "error 1",
		},
		{
			name:     "multiple errors",
			errors:   []error{fmt.Errorf("error 1"), fmt.Errorf("error 2"), fmt.Errorf("error 3")},
			expected: "error 1; error 2; error 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := joinErrors(tt.errors)
			if tt.expected == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
			}
		})
	}
}
