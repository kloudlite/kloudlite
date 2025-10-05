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

func TestEnvironmentReconciler_HandleDeletion(t *testing.T) {
	scheme := testutil.NewTestScheme()

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

	k8sClient := testutil.NewFakeClient(scheme, env, namespace).Build()

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
	scheme := testutil.NewTestScheme()

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

	k8sClient := testutil.NewFakeClient(scheme, env).Build()

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
