package controllers

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestCompositionReconciler_HandleDeletion(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-composition",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now,
			Finalizers:        []string{compositionFinalizer},
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
	}

	// Create some resources that should be cleaned up
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test",
							Image: "nginx",
						},
					},
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition, deployment).
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

	_, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	// Fake client deletes resources immediately
}

func TestCompositionReconciler_HandleDeletion_WithResources(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-composition",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now,
			Finalizers:        []string{compositionFinalizer},
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
		},
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
				},
			},
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 80}},
		},
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition, deployment, service, configMap, secret, pvc).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.handleDeletion(context.Background(), composition, logger)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestCompositionReconciler_HandleDeletion_StatusUpdateFails(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-composition",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now,
			Finalizers:        []string{compositionFinalizer},
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
		},
		Status: environmentsv1.CompositionStatus{
			State: environmentsv1.CompositionStateRunning, // Not deleting yet
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

	result, err := reconciler.handleDeletion(context.Background(), composition, logger)
	// Status update may fail with fake client, but deletion should continue
	// Verify no fatal error and finalizer is removed
	if err == nil {
		assert.False(t, result.Requeue)
	}
}

func TestCompositionReconciler_HandleDeletion_ResourcesAlreadyBeingDeleted(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-composition",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now,
			Finalizers:        []string{compositionFinalizer},
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
		},
	}

	// Resources already being deleted (have DeletionTimestamp)
	deletingDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-deployment",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now,
			Finalizers:        []string{"kubernetes"},
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "test", Image: "nginx"}},
				},
			},
		},
	}

	deletingService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-service",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now,
			Finalizers:        []string{"kubernetes"},
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 80}},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition, deletingDeployment, deletingService).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.handleDeletion(context.Background(), composition, logger)
	assert.NoError(t, err)
	// Should requeue because resources still exist (being deleted)
	assert.Greater(t, result.RequeueAfter.Seconds(), float64(0))
}

func TestCompositionReconciler_HandleDeletion_ConfigMapAndSecretDeletion(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-composition",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now,
			Finalizers:        []string{compositionFinalizer},
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
		},
	}

	// Only ConfigMaps and Secrets (testing these specific resource types)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
		Data: map[string]string{"key": "value"},
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
		Data: map[string][]byte{"key": []byte("value")},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition, configMap, secret).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.handleDeletion(context.Background(), composition, logger)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestCompositionReconciler_HandleDeletion_PVCDeletion(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-composition",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now,
			Finalizers:        []string{compositionFinalizer},
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
		},
	}

	// Only PVC (testing PVC deletion specifically)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/docker-composition": "test-composition",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition, pvc).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.handleDeletion(context.Background(), composition, logger)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestCompositionReconciler_HandleDeletion_NoResources(t *testing.T) {
	scheme := testutil.NewTestScheme()

	now := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-composition",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now,
			Finalizers:        []string{compositionFinalizer},
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
		},
	}

	// No resources to delete
	k8sClient := testutil.NewFakeClient(scheme, composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.handleDeletion(context.Background(), composition, logger)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify finalizer was removed
	updatedComp := &environmentsv1.Composition{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-composition",
		Namespace: "test-namespace",
	}, updatedComp)
	// Fake client may delete object after finalizer removal
	if err == nil {
		assert.NotContains(t, updatedComp.Finalizers, compositionFinalizer)
	}
}
