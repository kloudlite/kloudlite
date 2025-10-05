package controllers

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnvironmentReconciler_Reconcile_ActiveEnvironment(t *testing.T) {
	scheme := testutil.NewTestScheme()

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

	k8sClient := testutil.NewFakeClient(scheme, env).Build()

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
	scheme := testutil.NewTestScheme()

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

	k8sClient := testutil.NewFakeClient(scheme, env).Build()

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

func TestEnvironmentReconciler_ActivationStatusUpdate(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "admin@example.com",
			Activated:       true,
		},
		Status: environmentsv1.EnvironmentStatus{
			State: environmentsv1.EnvironmentStateInactive,
		},
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, namespace).
		WithStatusSubresource(env).
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

	// Verify status was updated to active
	updatedEnv := &environmentsv1.Environment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-env"}, updatedEnv)
	assert.NoError(t, err)
	assert.Equal(t, environmentsv1.EnvironmentStateActive, updatedEnv.Status.State)
	assert.Equal(t, "Environment is active", updatedEnv.Status.Message)
	assert.NotNil(t, updatedEnv.Status.LastActivatedTime)
}

func TestEnvironmentReconciler_DeactivationStatusUpdate(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "admin@example.com",
			Activated:       false,
		},
		Status: environmentsv1.EnvironmentStatus{
			State: environmentsv1.EnvironmentStateActive,
		},
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, namespace).
		WithStatusSubresource(env).
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

	// Verify status was updated to inactive
	updatedEnv := &environmentsv1.Environment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "test-env"}, updatedEnv)
	assert.NoError(t, err)
	assert.Equal(t, environmentsv1.EnvironmentStateInactive, updatedEnv.Status.State)
	assert.Equal(t, "Environment is inactive", updatedEnv.Status.Message)
	assert.NotNil(t, updatedEnv.Status.LastDeactivatedTime)
}
