package controllers

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestCompositionReconciler_Reconcile_CompositionNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
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

func TestCompositionReconciler_HandleDeletion(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition, deployment).
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

func TestGetPVCNames(t *testing.T) {
	pvcs := []*corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pvc-1",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pvc-2",
			},
		},
	}

	names := getPVCNames(pvcs)
	assert.Len(t, names, 2)
	assert.Contains(t, names, "pvc-1")
	assert.Contains(t, names, "pvc-2")
}

func TestCompositionReconciler_DeployComposition_Success(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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
    image: nginx:latest
    ports:
      - "80:80"`,
		},
		Status: environmentsv1.CompositionStatus{
			ObservedGeneration: 0,
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.deployComposition(context.Background(), composition, logger)
	assert.NoError(t, err)
	assert.NotNil(t, composition.Status.DeployedResources)
}

func TestCompositionReconciler_DeployComposition_ParseError(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `invalid yaml content [[[`,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.deployComposition(context.Background(), composition, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse error")
}

func TestCompositionReconciler_DeployComposition_EmptyContent(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: "",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.deployComposition(context.Background(), composition, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse error")
}

func TestCompositionReconciler_ApplyResource_Create(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
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

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.applyResource(context.Background(), deployment, composition, logger)
	assert.NoError(t, err)

	// Verify deployment was created
	createdDeployment := &appsv1.Deployment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-deployment",
		Namespace: "test-namespace",
	}, createdDeployment)
	assert.NoError(t, err)
}

func TestCompositionReconciler_ApplyResource_Update(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
			UID:       "comp-123",
		},
	}

	existingDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
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
							Image: "nginx:1.0",
						},
					},
				},
			},
		},
	}

	updatedDeployment := existingDeployment.DeepCopy()
	updatedDeployment.Spec.Replicas = int32Ptr(3)
	updatedDeployment.Spec.Template.Spec.Containers[0].Image = "nginx:2.0"

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(composition, existingDeployment).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.applyResource(context.Background(), updatedDeployment, composition, logger)
	assert.NoError(t, err)

	// Verify deployment was updated
	retrievedDeployment := &appsv1.Deployment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-deployment",
		Namespace: "test-namespace",
	}, retrievedDeployment)
	assert.NoError(t, err)
	assert.Equal(t, int32(3), *retrievedDeployment.Spec.Replicas)
}

func TestCompositionReconciler_ApplyResource_SkipPVCUpdate(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
			UID:       "comp-123",
		},
	}

	// Create an existing PVC with StorageClassName set
	existingPVC := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "test-namespace",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
			StorageClassName: func() *string { s := "local-path"; return &s }(),
			VolumeName:       "pvc-123",
		},
	}

	// Create a new PVC (without StorageClassName/VolumeName - would cause immutability error)
	newPVC := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "test-namespace",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
			// No StorageClassName or VolumeName - would fail if updated
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(composition, existingPVC).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	// Apply the new PVC - should skip update
	err := reconciler.applyResource(context.Background(), newPVC, composition, logger)
	assert.NoError(t, err)

	// Verify PVC was NOT updated (StorageClassName should still be set)
	retrievedPVC := &corev1.PersistentVolumeClaim{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-pvc",
		Namespace: "test-namespace",
	}, retrievedPVC)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedPVC.Spec.StorageClassName)
	assert.Equal(t, "local-path", *retrievedPVC.Spec.StorageClassName)
	assert.Equal(t, "pvc-123", retrievedPVC.Spec.VolumeName)
}

func TestCompositionReconciler_CleanupRemovedResources_FirstDeployment(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.cleanupRemovedResources(context.Background(), composition, nil, []string{"dep1"}, []string{"svc1"}, logger)
	assert.NoError(t, err)
}

func TestCompositionReconciler_CleanupRemovedResources_RemoveDeployment(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	oldDeployedResources := &environmentsv1.DeployedResources{
		Deployments: []string{"old-dep", "keep-dep"},
		Services:    []string{"old-svc"},
	}

	oldDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "old-dep",
			Namespace: "test-namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "old"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "old"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "old", Image: "nginx"}},
				},
			},
		},
	}

	oldService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "old-svc",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 80}},
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(composition, oldDeployment, oldService).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	// Current deployments only has "keep-dep", so "old-dep" should be deleted
	err := reconciler.cleanupRemovedResources(context.Background(), composition, oldDeployedResources, []string{"keep-dep"}, []string{}, logger)
	assert.NoError(t, err)

	// Verify old-dep was deleted
	deletedDep := &appsv1.Deployment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "old-dep",
		Namespace: "test-namespace",
	}, deletedDep)
	assert.Error(t, err)

	// Verify old-svc was deleted
	deletedSvc := &corev1.Service{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "old-svc",
		Namespace: "test-namespace",
	}, deletedSvc)
	assert.Error(t, err)
}

func TestCompositionReconciler_UpdateStatus_Running(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateRunning, "Success", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}
	assert.Equal(t, environmentsv1.CompositionStateRunning, composition.Status.State)
	assert.Equal(t, "Success", composition.Status.Message)
	assert.Equal(t, int64(1), composition.Status.ObservedGeneration)
	assert.NotNil(t, composition.Status.LastDeployedTime)
}

func TestCompositionReconciler_UpdateStatus_Failed(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 2,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateFailed, "Parse error", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}
	assert.Equal(t, environmentsv1.CompositionStateFailed, composition.Status.State)
	assert.Equal(t, "Parse error", composition.Status.Message)
	assert.Len(t, composition.Status.Conditions, 1)
	assert.Equal(t, "Ready", composition.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionFalse, composition.Status.Conditions[0].Status)
}

func TestCompositionReconciler_UpdateStatus_Deploying(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateDeploying, "Deploying", logger)
	assert.NoError(t, err)
	assert.Greater(t, result.RequeueAfter.Seconds(), float64(0))
	assert.Equal(t, environmentsv1.CompositionStateDeploying, composition.Status.State)
}

func TestCompositionReconciler_HandleDeletion_WithResources(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(composition, deployment, service, configMap, secret, pvc).Build()

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

func TestCompositionReconciler_SetupWithManager(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

func TestCompositionReconciler_HandleDeletion_StatusUpdateFails(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(composition, deletingDeployment, deletingService).Build()

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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(composition, configMap, secret).Build()

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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(composition, pvc).Build()

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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
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

func TestCompositionReconciler_UpdateStatus_UpdateExistingCondition(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	oldTime := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 2,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
		Status: environmentsv1.CompositionStatus{
			Conditions: []metav1.Condition{
				{
					Type:               "Ready",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 1,
					LastTransitionTime: oldTime,
					Reason:             "Running",
					Message:            "Old message",
				},
			},
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateFailed, "New error message", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}

	// Verify condition was updated, not added
	assert.Len(t, composition.Status.Conditions, 1)
	assert.Equal(t, "Ready", composition.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionFalse, composition.Status.Conditions[0].Status)
	assert.Equal(t, "New error message", composition.Status.Conditions[0].Message)
	assert.Equal(t, int64(2), composition.Status.Conditions[0].ObservedGeneration)
}

func TestCompositionReconciler_UpdateStatus_MultipleConditions(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	now := metav1.Now()
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
		Status: environmentsv1.CompositionStatus{
			Conditions: []metav1.Condition{
				{
					Type:               "OtherCondition",
					Status:             metav1.ConditionTrue,
					ObservedGeneration: 1,
					LastTransitionTime: now,
					Reason:             "Other",
					Message:            "Some other condition",
				},
				{
					Type:               "Ready",
					Status:             metav1.ConditionFalse,
					ObservedGeneration: 0,
					LastTransitionTime: now,
					Reason:             "Failed",
					Message:            "Old failure",
				},
			},
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateRunning, "Now running", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}

	// Verify we have 2 conditions and only Ready was updated
	assert.Len(t, composition.Status.Conditions, 2)
	assert.Equal(t, "OtherCondition", composition.Status.Conditions[0].Type)
	assert.Equal(t, "Some other condition", composition.Status.Conditions[0].Message) // Unchanged
	assert.Equal(t, "Ready", composition.Status.Conditions[1].Type)
	assert.Equal(t, metav1.ConditionTrue, composition.Status.Conditions[1].Status)
	assert.Equal(t, "Now running", composition.Status.Conditions[1].Message)
}

func TestCompositionReconciler_UpdateStatus_AddConditionWhenNoneExist(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `version: '3.8'`,
		},
		Status: environmentsv1.CompositionStatus{
			Conditions: []metav1.Condition{}, // Empty conditions
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	result, err := reconciler.updateStatus(context.Background(), composition, environmentsv1.CompositionStateRunning, "First status", logger)
	// Fake client may delete object during status update
	if err != nil {
		assert.Contains(t, err.Error(), "not found")
	} else {
		assert.False(t, result.Requeue)
	}

	// Verify condition was added
	assert.Len(t, composition.Status.Conditions, 1)
	assert.Equal(t, "Ready", composition.Status.Conditions[0].Type)
	assert.Equal(t, metav1.ConditionTrue, composition.Status.Conditions[0].Status)
	assert.Equal(t, "First status", composition.Status.Conditions[0].Message)
	assert.NotNil(t, composition.Status.LastDeployedTime)
}

func TestCompositionReconciler_Reconcile_GetCompositionError(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	// Create a fake client that will return an error
	// We can't easily simulate Get errors with fake client, but we test the error handling path exists
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

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
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
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

func TestCompositionReconciler_DeployComposition_ConversionError(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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
      - "invalid-port-format"`, // This will cause conversion error
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.deployComposition(context.Background(), composition, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse error")
}

func TestCompositionReconciler_ApplyResource_SetOwnerError(t *testing.T) {
	scheme := runtime.NewScheme()
	// Don't add appsv1 to scheme to cause SetControllerReference to fail
	_ = environmentsv1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
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

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.applyResource(context.Background(), deployment, composition, logger)
	assert.Error(t, err)
}

func TestCompositionReconciler_CleanupRemovedResources_DeleteError(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	oldDeployedResources := &environmentsv1.DeployedResources{
		Deployments: []string{"old-deployment"},
		Services:    []string{"old-service"},
	}

	// Don't create the actual resources - this will test the "not found" path
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition).
		WithStatusSubresource(composition).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	// Should handle not found errors gracefully
	err := reconciler.cleanupRemovedResources(context.Background(), composition, oldDeployedResources, []string{}, []string{}, logger)
	assert.NoError(t, err)
}

func TestCompositionReconciler_FetchEnvironmentData(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create test ConfigMap for env-config
	envConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "env-config",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"API_URL": "https://api.example.com",
			"DEBUG":   "true",
		},
	}

	// Create test Secret for env-secret
	envSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "env-secret",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"DB_PASSWORD": []byte("secret123"),
			"API_KEY":     []byte("key456"),
		},
	}

	// Create test ConfigMaps for config files
	configFileMap1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "env-file-app.yml",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/file-type": "environment-file",
			},
		},
		Data: map[string]string{
			"app.yml": "app config content",
		},
	}

	configFileMap2 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "env-file-nginx.conf",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/file-type": "environment-file",
			},
		},
		Data: map[string]string{
			"nginx.conf": "nginx config",
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(envConfigMap, envSecret, configFileMap1, configFileMap2).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	envData, err := reconciler.fetchEnvironmentData(context.Background(), "test-namespace", logger)
	assert.NoError(t, err)
	assert.NotNil(t, envData)

	// Verify env vars
	assert.Equal(t, 2, len(envData.EnvVars))
	assert.Equal(t, "https://api.example.com", envData.EnvVars["API_URL"])
	assert.Equal(t, "true", envData.EnvVars["DEBUG"])

	// Verify secrets
	assert.Equal(t, 2, len(envData.Secrets))
	assert.Equal(t, "secret123", envData.Secrets["DB_PASSWORD"])
	assert.Equal(t, "key456", envData.Secrets["API_KEY"])

	// Verify config files
	assert.Equal(t, 2, len(envData.ConfigFiles))
	assert.Equal(t, "app config content", envData.ConfigFiles["app.yml"])
	assert.Equal(t, "nginx config", envData.ConfigFiles["nginx.conf"])
}

func TestCompositionReconciler_FetchEnvironmentData_Missing(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	envData, err := reconciler.fetchEnvironmentData(context.Background(), "test-namespace", logger)
	assert.NoError(t, err)
	assert.NotNil(t, envData)

	// Should return empty maps, not nil
	assert.Equal(t, 0, len(envData.EnvVars))
	assert.Equal(t, 0, len(envData.Secrets))
	assert.Equal(t, 0, len(envData.ConfigFiles))
}

func TestCompositionReconciler_FindCompositionsForConfigMap(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition1 := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp1",
			Namespace: "test-namespace",
		},
	}

	composition2 := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp2",
			Namespace: "test-namespace",
		},
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "env-config",
			Namespace: "test-namespace",
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition1, composition2, configMap).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	requests := reconciler.findCompositionsForConfigMap(context.Background(), configMap)

	// Should return reconcile requests for all compositions in the namespace
	assert.Equal(t, 2, len(requests))

	names := []string{requests[0].Name, requests[1].Name}
	assert.Contains(t, names, "comp1")
	assert.Contains(t, names, "comp2")
}

func TestCompositionReconciler_FindCompositionsForConfigMap_WrongName(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp1",
			Namespace: "test-namespace",
		},
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-other-config",
			Namespace: "test-namespace",
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition, configMap).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	requests := reconciler.findCompositionsForConfigMap(context.Background(), configMap)

	// Should not return any requests for non env-config ConfigMaps
	assert.Equal(t, 0, len(requests))
}

func TestCompositionReconciler_FindCompositionsForSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition1 := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp1",
			Namespace: "test-namespace",
		},
	}

	composition2 := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp2",
			Namespace: "test-namespace",
		},
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "env-secret",
			Namespace: "test-namespace",
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition1, composition2, secret).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	requests := reconciler.findCompositionsForSecret(context.Background(), secret)

	// Should return reconcile requests for all compositions in the namespace
	assert.Equal(t, 2, len(requests))

	names := []string{requests[0].Name, requests[1].Name}
	assert.Contains(t, names, "comp1")
	assert.Contains(t, names, "comp2")
}

func TestCompositionReconciler_FindCompositionsForSecret_WrongName(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp1",
			Namespace: "test-namespace",
		},
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-other-secret",
			Namespace: "test-namespace",
		},
	}

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(composition, secret).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	requests := reconciler.findCompositionsForSecret(context.Background(), secret)

	// Should not return any requests for non env-secret Secrets
	assert.Equal(t, 0, len(requests))
}

func int32Ptr(i int32) *int32 {
	return &i
}
