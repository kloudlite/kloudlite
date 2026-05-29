package environment

import (
	"context"
	"testing"

	"github.com/kloudlite/kloudlite/controllers/testutil"
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnvironmentReconciler_Reconcile_CreateNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
			UID:  types.UID("test-uid-123"),
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "test@example.com",
			Activated:       true,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env).
		WithStatusSubresource(&environmentsv1.Environment{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-env",
		},
	}

	// First reconcile - should add finalizer
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue)

	// Second reconcile - should create namespace
	result, err = reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify namespace was created
	namespace := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-namespace"}, namespace)
	assert.NoError(t, err)
	assert.Equal(t, "test-env", namespace.Labels["kloudlite.io/environment"])
	assert.Equal(t, "test@example.com", namespace.Annotations["kloudlite.io/created-by"])
}

func TestEnvironmentReconciler_ScopeAllowsCurrentWorkMachineEnvironment(t *testing.T) {
	scheme := testutil.NewTestScheme()
	env := scopedEnvironment("test-env", "wm-karthik-dev", "karthik-dev", "env-test-env")
	k8sClient := testutil.NewFakeClient(scheme, env).WithStatusSubresource(&environmentsv1.Environment{}).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop(), OwnNamespace: "wm-karthik-dev", WorkMachineName: "karthik-dev"}

	_, err := reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-env", Namespace: "wm-karthik-dev"}})
	assert.NoError(t, err)
	_, err = reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-env", Namespace: "wm-karthik-dev"}})
	assert.NoError(t, err)

	namespace := &corev1.Namespace{}
	err = reconciler.Get(context.Background(), types.NamespacedName{Name: "env-test-env"}, namespace)
	assert.NoError(t, err)
}

func TestEnvironmentReconciler_ScopeSkipsWrongNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()
	env := scopedEnvironment("test-env", "wm-other", "karthik-dev", "env-test-env")
	k8sClient := testutil.NewFakeClient(scheme, env).WithStatusSubresource(&environmentsv1.Environment{}).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop(), OwnNamespace: "wm-karthik-dev", WorkMachineName: "karthik-dev"}

	_, err := reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-env", Namespace: "wm-other"}})
	assert.NoError(t, err)

	namespace := &corev1.Namespace{}
	err = reconciler.Get(context.Background(), types.NamespacedName{Name: "env-test-env"}, namespace)
	assert.True(t, apierrors.IsNotFound(err), "expected wrong namespace environment to be skipped, got %v", err)
}

func TestEnvironmentReconciler_ScopeSkipsWrongWorkMachine(t *testing.T) {
	scheme := testutil.NewTestScheme()
	env := scopedEnvironment("test-env", "wm-karthik-dev", "other-machine", "env-test-env")
	k8sClient := testutil.NewFakeClient(scheme, env).WithStatusSubresource(&environmentsv1.Environment{}).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop(), OwnNamespace: "wm-karthik-dev", WorkMachineName: "karthik-dev"}

	_, err := reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-env", Namespace: "wm-karthik-dev"}})
	assert.NoError(t, err)

	namespace := &corev1.Namespace{}
	err = reconciler.Get(context.Background(), types.NamespacedName{Name: "env-test-env"}, namespace)
	assert.True(t, apierrors.IsNotFound(err), "expected wrong workmachine environment to be skipped, got %v", err)
}

func TestEnvironmentReconciler_ScopeFailsClosedWhenConfiguredIncomplete(t *testing.T) {
	scheme := testutil.NewTestScheme()
	env := scopedEnvironment("test-env", "wm-karthik-dev", "karthik-dev", "env-test-env")
	k8sClient := testutil.NewFakeClient(scheme, env).WithStatusSubresource(&environmentsv1.Environment{}).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop(), OwnNamespace: "wm-karthik-dev"}

	_, err := reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-env", Namespace: "wm-karthik-dev"}})
	assert.NoError(t, err)

	namespace := &corev1.Namespace{}
	err = reconciler.Get(context.Background(), types.NamespacedName{Name: "env-test-env"}, namespace)
	assert.True(t, apierrors.IsNotFound(err), "expected incomplete scope to skip reconciliation, got %v", err)
}

func scopedEnvironment(name, namespace, workMachineName, targetNamespace string) *environmentsv1.Environment {
	return &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, UID: types.UID(name + "-uid")},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: targetNamespace,
			WorkMachineName: workMachineName,
			OwnedBy:         "karthik",
			Activated:       true,
		},
	}
}

func TestEnvironmentReconciler_Reconcile_EnvironmentNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()

	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "nonexistent-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestEnvironmentReconciler_Reconcile_ExistingNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-env",
			UID:        types.UID("test-uid-existing"),
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "existing-namespace",
			OwnedBy:         "test@example.com",
			Activated:       false,
		},
	}

	existingNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-namespace",
		},
	}
	applyEnvironmentNamespaceOwnership(existingNamespace, env)

	k8sClient := testutil.NewFakeClient(scheme, env, existingNamespace).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify namespace was updated with labels
	namespace := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "existing-namespace"}, namespace)
	assert.NoError(t, err)
	assert.Equal(t, "test-env", namespace.Labels["kloudlite.io/environment"])
	assert.Equal(t, "test@example.com", namespace.Annotations["kloudlite.io/created-by"])
}

func TestEnvironmentReconciler_Reconcile_WithCustomLabelsAndAnnotations(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "test@example.com",
			Labels: map[string]string{
				"custom-label": "label-value",
			},
			Annotations: map[string]string{
				"custom-annotation": "annotation-value",
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env).
		WithStatusSubresource(&environmentsv1.Environment{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify namespace has custom labels and annotations
	namespace := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-namespace"}, namespace)
	assert.NoError(t, err)
	assert.Equal(t, "label-value", namespace.Labels["custom-label"])
	assert.Equal(t, "annotation-value", namespace.Annotations["custom-annotation"])
}

func TestEnvironmentReconciler_Reconcile_CustomLabelsAndAnnotations(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-env",
			Finalizers: []string{environmentFinalizer},
			UID:        types.UID("test-uid-456"),
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "admin@example.com",
			Labels: map[string]string{
				"team":    "platform",
				"project": "main",
			},
			Annotations: map[string]string{
				"description": "Test environment",
				"owner":       "team@example.com",
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env).
		WithStatusSubresource(&environmentsv1.Environment{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify namespace has custom labels and annotations
	namespace := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-namespace"}, namespace)
	assert.NoError(t, err)
	assert.Equal(t, "platform", namespace.Labels["team"])
	assert.Equal(t, "main", namespace.Labels["project"])
	assert.Equal(t, "Test environment", namespace.Annotations["description"])
	assert.Equal(t, "team@example.com", namespace.Annotations["owner"])
}

func TestEnvironmentReconciler_Reconcile_ExistingNamespaceWithNilLabels(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "test@example.com",
		},
	}

	// Existing namespace with nil labels and annotations
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-namespace",
			Labels:      nil,
			Annotations: nil,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, namespace).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	// Existing unmarked namespaces must not be adopted.
	assert.Error(t, err)
	assert.False(t, result.Requeue)

	// Verify labels and annotations were not added.
	updatedNs := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-namespace"}, updatedNs)
	assert.NoError(t, err)
	assert.Nil(t, updatedNs.Labels)
	assert.Nil(t, updatedNs.Annotations)
}

func TestEnvironmentReconciler_Reconcile_RejectsUnownedExistingNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-env", Namespace: "wm-alice", UID: types.UID("env-uid"), Finalizers: []string{environmentFinalizer}},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "existing-namespace", OwnedBy: "alice", Activated: true},
	}
	existingNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "existing-namespace"}}
	k8sClient := testutil.NewFakeClient(scheme, env, existingNamespace).WithStatusSubresource(&environmentsv1.Environment{}).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	_, err := reconciler.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "test-env", Namespace: "wm-alice"}})
	if err == nil {
		t.Fatalf("Reconcile returned nil error, want ownership mismatch")
	}

	ns := &corev1.Namespace{}
	if getErr := k8sClient.Get(context.Background(), types.NamespacedName{Name: "existing-namespace"}, ns); getErr != nil {
		t.Fatalf("get namespace: %v", getErr)
	}
	if ns.Labels[environmentNameLabel] == "test-env" {
		t.Fatalf("namespace was adopted despite missing ownership markers: %#v", ns.Labels)
	}
}

func TestEnvironmentReconciler_CreateNamespaceRejectsUnownedAlreadyExists(t *testing.T) {
	scheme := testutil.NewTestScheme()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-env", Namespace: "wm-alice", UID: types.UID("env-uid")},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "existing-namespace", OwnedBy: "alice"},
	}
	existingNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "existing-namespace"}}
	k8sClient := testutil.NewFakeClient(scheme, existingNamespace).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	err := reconciler.createNamespace(context.Background(), env, zap.NewNop())
	if err == nil {
		t.Fatalf("createNamespace returned nil error, want ownership mismatch")
	}
}

func TestEnvironmentReconciler_CreateNamespaceForForkingRejectsUnownedAlreadyExists(t *testing.T) {
	scheme := testutil.NewTestScheme()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "fork-env", Namespace: "wm-alice", UID: types.UID("fork-env-uid")},
		Spec:       environmentsv1.EnvironmentSpec{TargetNamespace: "fork-namespace", OwnedBy: "alice"},
	}
	existingNamespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fork-namespace"}}
	k8sClient := testutil.NewFakeClient(scheme, existingNamespace).Build()
	reconciler := &EnvironmentReconciler{Client: k8sClient, Scheme: scheme, Logger: zap.NewNop()}

	err := reconciler.createNamespaceForForking(context.Background(), env, "source-env", zap.NewNop())
	if err == nil {
		t.Fatalf("createNamespaceForForking returned nil error, want ownership mismatch")
	}
}

func TestEnvironmentReconciler_Reconcile_AddFinalizerError(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
			// No finalizer
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "test@example.com",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env).
		WithStatusSubresource(&environmentsv1.Environment{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	// Fake client should successfully add finalizer
	if err == nil {
		assert.True(t, result.Requeue)
	}
}
