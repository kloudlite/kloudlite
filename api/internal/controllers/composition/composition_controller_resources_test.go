package composition

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestCompositionReconciler_ApplyResource_Create(t *testing.T) {
	scheme := testutil.NewTestScheme()

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

	k8sClient := testutil.NewFakeClient(scheme, composition).
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
	scheme := testutil.NewTestScheme()

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
			Replicas: testutil.Int32Ptr(1),
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
	updatedDeployment.Spec.Replicas = testutil.Int32Ptr(3)
	updatedDeployment.Spec.Template.Spec.Containers[0].Image = "nginx:2.0"

	k8sClient := testutil.NewFakeClient(scheme, composition, existingDeployment).Build()

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
	scheme := testutil.NewTestScheme()

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

	k8sClient := testutil.NewFakeClient(scheme, composition, existingPVC).Build()

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
	scheme := testutil.NewTestScheme()

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme).Build()

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
	scheme := testutil.NewTestScheme()

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

	k8sClient := testutil.NewFakeClient(scheme, composition, oldDeployment, oldService).Build()

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

func TestCompositionReconciler_CleanupRemovedResources_DeleteError(t *testing.T) {
	scheme := testutil.NewTestScheme()

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
	k8sClient := testutil.NewFakeClient(scheme, composition).
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
