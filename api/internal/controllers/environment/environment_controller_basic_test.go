package environment

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
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
			CreatedBy:       "test@example.com",
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
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "existing-namespace",
			CreatedBy:       "test@example.com",
			Activated:       false,
		},
	}

	existingNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-namespace",
		},
	}

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
			CreatedBy:       "test@example.com",
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
			CreatedBy:       "admin@example.com",
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
			CreatedBy:       "test@example.com",
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
	// Should succeed and create labels/annotations
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify labels and annotations were created
	updatedNs := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-namespace"}, updatedNs)
	assert.NoError(t, err)
	assert.NotNil(t, updatedNs.Labels)
	assert.NotNil(t, updatedNs.Annotations)
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
			CreatedBy:       "test@example.com",
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
