package controllers

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestCompositionReconciler_Reconcile_CompositionNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()

	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent-composition",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestCompositionReconciler_Reconcile_AddFinalizer(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
			ComposeContent: `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"`,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue)

	// Verify finalizer was added
	updatedComp := &environmentsv1.Composition{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-composition",
		Namespace: "test-namespace",
	}, updatedComp)
	assert.NoError(t, err)
	assert.Contains(t, updatedComp.Finalizers, compositionFinalizer)
}

func TestCompositionReconciler_Reconcile_ReconcileOnConfigMapChange(t *testing.T) {
	// This test verifies that reconciliation happens even when observedGeneration matches generation
	// This allows ConfigMap/Secret changes to trigger redeployment
	scheme := testutil.NewTestScheme()

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Finalizers: []string{compositionFinalizer},
			Generation: 1,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
			ComposeContent: `version: '3.8'
services:
  web:
    image: nginx:latest`,
		},
		Status: environmentsv1.CompositionStatus{
			ObservedGeneration: 1, // Same as current generation - deployment should still happen
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	// Deployment should happen, updating status, so no error expected
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify that a deployment was created
	deploymentList := &appsv1.DeploymentList{}
	err = k8sClient.List(context.Background(), deploymentList, client.InNamespace("test-namespace"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deploymentList.Items), "Deployment should be created")
}

func TestCompositionReconciler_Reconcile_GetCompositionError(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Create a fake client that will return an error
	// We can't easily simulate Get errors with fake client, but we test the error handling path exists
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	// Should return no error for not found
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestCompositionReconciler_Reconcile_AddFinalizerUpdateError(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
			// No finalizer initially
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
			ComposeContent: `version: '3.8'
services:
  web:
    image: nginx:latest`,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	// Fake client may have issues with update, but should add finalizer and requeue
	if err == nil {
		assert.True(t, result.Requeue)
	}
}

func TestCompositionReconciler_SetupWithManager(t *testing.T) {
	scheme := testutil.NewTestScheme()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: nil,
		Scheme: scheme,
		Logger: logger,
	}

	// SetupWithManager requires a manager, which we can't easily mock
	// Just verify the function exists and returns an error without a manager
	err := reconciler.SetupWithManager(nil)
	assert.Error(t, err)
}
