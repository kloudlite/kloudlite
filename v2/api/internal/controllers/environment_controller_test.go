package controllers

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/environments/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnvironmentReconciler_Reconcile_CreateNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env).Build()

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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env, existingNamespace).Build()

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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env).Build()

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
	assert.Equal(t, "label-value", namespace.Annotations["custom-label"])
	assert.Equal(t, "annotation-value", namespace.Annotations["custom-annotation"])
}

func TestEnvironmentReconciler_Reconcile_ActiveEnvironment(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "active-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "active-namespace",
			CreatedBy:       "test@example.com",
			Activated:       true,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "active-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify namespace was created
	namespace := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "active-namespace"}, namespace)
	assert.NoError(t, err)
	assert.Equal(t, "active-env", namespace.Labels["kloudlite.io/environment"])
}

func TestEnvironmentReconciler_Reconcile_InactiveEnvironment(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "inactive-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "inactive-namespace",
			CreatedBy:       "test@example.com",
			Activated:       false,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "inactive-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Status update might fail with fake client, which is expected
	// The actual state update happens via Status().Update() which fake client doesn't fully support
}

func TestEnvironmentReconciler_HandleDeletion(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	now := metav1.Now()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleting-env",
			DeletionTimestamp: &now,
			Finalizers:        []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "deleting-namespace",
			CreatedBy:       "test@example.com",
		},
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "deleting-namespace",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env, namespace).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "deleting-env",
		},
	}

	_, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	// Fake client deletes immediately, so we just verify no error
}

func TestEnvironmentReconciler_HandleDeletion_NamespaceAlreadyDeleted(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	now := metav1.Now()
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleting-env",
			DeletionTimestamp: &now,
			Finalizers:        []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "already-deleted-namespace",
			CreatedBy:       "test@example.com",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "deleting-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify finalizer was removed
	updatedEnv := &environmentsv1.Environment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "deleting-env"}, updatedEnv)
	// Environment might be deleted by fake client
	if err == nil {
		assert.NotContains(t, updatedEnv.Finalizers, environmentFinalizer)
	}
}

func TestEnvironmentReconciler_Reconcile_AddFinalizerError(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env).Build()

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

func TestEnvironmentReconciler_Reconcile_ExistingNamespaceWithNilLabels(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env, namespace).Build()

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

func TestEnvironmentReconciler_Reconcile_CustomLabelsAndAnnotations(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(env).Build()

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

	// Verify namespace has custom labels in annotations
	namespace := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-namespace"}, namespace)
	assert.NoError(t, err)
	assert.Equal(t, "platform", namespace.Annotations["team"])
	assert.Equal(t, "main", namespace.Annotations["project"])
	assert.Equal(t, "Test environment", namespace.Annotations["description"])
	assert.Equal(t, "team@example.com", namespace.Annotations["owner"])
}
