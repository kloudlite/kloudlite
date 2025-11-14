package environment

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
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestEnvironmentReconciler_HandleCloning_Success(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Create source environment
	sourceEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "source-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "source-namespace",
			OwnedBy:         "admin@example.com",
			Activated:       true,
		},
	}

	// Create target environment with cloneFrom
	targetEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "target-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "target-namespace",
			OwnedBy:         "admin@example.com",
			CloneFrom:       "source-env",
			Activated:       false,
		},
	}

	// Create source ConfigMap
	sourceConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "env-config",
			Namespace: "source-namespace",
			Labels: map[string]string{
				"kloudlite.io/resource-type": "environment-config",
			},
		},
		Data: map[string]string{
			"API_URL": "https://api.example.com",
		},
	}

	// Create source Secret
	sourceSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "env-secret",
			Namespace: "source-namespace",
			Labels: map[string]string{
				"kloudlite.io/resource-type": "environment-config",
			},
		},
		Data: map[string][]byte{
			"DB_PASSWORD": []byte("secret123"),
		},
	}

	// Create source Composition
	sourceComposition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-app",
			Namespace: "source-namespace",
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "Web App",
			ComposeContent: "version: '3.8'",
		},
	}

	// Create namespaces
	sourceNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "source-namespace",
		},
	}
	targetNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "target-namespace",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, sourceEnv, targetEnv, sourceNamespace, targetNamespace, sourceConfigMap, sourceSecret, sourceComposition).
		WithStatusSubresource(targetEnv).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "target-env",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue)

	// Verify cloning status was initialized
	updatedEnv := &environmentsv1.Environment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "target-env"}, updatedEnv)
	assert.NoError(t, err)
	assert.NotNil(t, updatedEnv.Status.CloningStatus)
	assert.Equal(t, environmentsv1.CloningPhaseCloningResources, updatedEnv.Status.CloningStatus.Phase)

	// Verify target namespace was created
	retrievedNamespace := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "target-namespace"}, retrievedNamespace)
	assert.NoError(t, err)
}

func TestEnvironmentReconciler_HandleCloning_SourceNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Create target environment with cloneFrom pointing to nonexistent source
	targetEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "target-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "target-namespace",
			OwnedBy:         "admin@example.com",
			CloneFrom:       "nonexistent-env",
			Activated:       false,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, targetEnv).
		WithStatusSubresource(targetEnv).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "target-env",
		},
	}

	_, err := reconciler.Reconcile(context.Background(), req)
	assert.Error(t, err)
}

func TestEnvironmentReconciler_HandleCloning_WithPVCs(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Create source environment
	sourceEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "source-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "source-namespace",
			OwnedBy:         "admin@example.com",
			Activated:       true,
		},
	}

	// Create target environment with cloneFrom
	targetEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "target-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "target-namespace",
			OwnedBy:         "admin@example.com",
			CloneFrom:       "source-env",
			Activated:       false,
		},
	}

	// Create source PVC with kloudlite.io/managed label
	storageQuantity := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse("1Gi"),
	}
	sourcePVC := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "data-pvc",
			Namespace: "source-namespace",
			Labels: map[string]string{
				"kloudlite.io/managed": "true",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: storageQuantity,
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, sourceEnv, targetEnv, sourcePVC).
		WithStatusSubresource(targetEnv).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "target-env",
		},
	}

	// First reconcile - should initialize cloning status and move to CloningResources
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue)

	// Check that cloning status was initialized and progressed
	updatedEnv := &environmentsv1.Environment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "target-env"}, updatedEnv)
	assert.NoError(t, err)
	assert.NotNil(t, updatedEnv.Status.CloningStatus)
	// The reconciler moves directly from initialization to CloningResources
	assert.Equal(t, environmentsv1.CloningPhaseCloningResources, updatedEnv.Status.CloningStatus.Phase)
	// Verify cloning process started
	assert.NotNil(t, updatedEnv.Status.CloningStatus.StartTime)
}

func TestEnvironmentReconciler_SuspendEnvironment(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "admin@example.com",
			Activated:       true,
		},
	}

	// Create a deployment with 3 replicas
	replicas := int32(3)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
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
							Image: "test:latest",
						},
					},
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, deployment).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	// Suspend the environment
	err := reconciler.suspendEnvironment(context.Background(), env, logger)
	assert.NoError(t, err)

	// Verify deployment was scaled to 0
	updatedDeployment := &appsv1.Deployment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-deployment",
		Namespace: "test-namespace",
	}, updatedDeployment)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), *updatedDeployment.Spec.Replicas)

	// Verify original replica count was stored in annotation
	assert.Equal(t, "3", updatedDeployment.Annotations["kloudlite.io/original-replicas"])
}

func TestEnvironmentReconciler_ResumeEnvironment(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "admin@example.com",
			Activated:       false,
		},
	}

	// Create a deployment scaled to 0 with original replica annotation
	zero := int32(0)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
			Annotations: map[string]string{
				"kloudlite.io/original-replicas": "5",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &zero,
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
							Image: "test:latest",
						},
					},
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, deployment).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	// Resume the environment
	err := reconciler.resumeEnvironment(context.Background(), env, logger)
	assert.NoError(t, err)

	// Verify deployment was scaled back to original replicas
	updatedDeployment := &appsv1.Deployment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-deployment",
		Namespace: "test-namespace",
	}, updatedDeployment)
	assert.NoError(t, err)
	assert.Equal(t, int32(5), *updatedDeployment.Spec.Replicas)

	// Verify annotation was removed after restoration
	_, exists := updatedDeployment.Annotations["kloudlite.io/original-replicas"]
	assert.False(t, exists)
}

func TestPVCCopier_CreateJobs(t *testing.T) {
	scheme := testutil.NewTestScheme()

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
			UID:  "test-uid-123",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "target-namespace",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env).Build()

	copier := NewPVCCopier(k8sClient, "source-namespace")

	// Test sender job creation
	senderJob := copier.createSenderJob("source-pvc", "target-pvc", env)
	assert.NotNil(t, senderJob)
	assert.Equal(t, "pvc-copy-sender-target-pvc", senderJob.Name)
	assert.Equal(t, "source-namespace", senderJob.Namespace)
	assert.Equal(t, "sender", senderJob.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "alpine:latest", senderJob.Spec.Template.Spec.Containers[0].Image)

	// Verify sender job has source PVC mounted as ReadOnly
	volumeMounts := senderJob.Spec.Template.Spec.Containers[0].VolumeMounts
	assert.Len(t, volumeMounts, 1)
	assert.Equal(t, "source-volume", volumeMounts[0].Name)
	assert.True(t, volumeMounts[0].ReadOnly)

	// Verify sender job exposes port 8080
	ports := senderJob.Spec.Template.Spec.Containers[0].Ports
	assert.Len(t, ports, 1)
	assert.Equal(t, int32(8080), ports[0].ContainerPort)

	// Test receiver job creation
	receiverJob := copier.createReceiverJob("source-pvc", "target-pvc", "10.0.0.1", env)
	assert.NotNil(t, receiverJob)
	assert.Equal(t, "pvc-copy-receiver-target-pvc", receiverJob.Name)
	assert.Equal(t, "source-namespace", receiverJob.Namespace)
	assert.Equal(t, "receiver", receiverJob.Spec.Template.Spec.Containers[0].Name)

	// Verify receiver job has target PVC mounted as ReadWrite
	receiverVolumeMounts := receiverJob.Spec.Template.Spec.Containers[0].VolumeMounts
	assert.Len(t, receiverVolumeMounts, 1)
	assert.Equal(t, "target-volume", receiverVolumeMounts[0].Name)
	assert.False(t, receiverVolumeMounts[0].ReadOnly)
}

func TestEnvironmentReconciler_CloningStatusProgress(t *testing.T) {
	scheme := testutil.NewTestScheme()

	sourceEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "source-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "source-namespace",
			OwnedBy:         "admin@example.com",
			Activated:       false, // Already deactivated
		},
	}

	targetEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "target-env",
			Finalizers: []string{environmentFinalizer},
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "target-namespace",
			OwnedBy:         "admin@example.com",
			CloneFrom:       "source-env",
			Activated:       false,
		},
		Status: environmentsv1.EnvironmentStatus{
			CloningStatus: &environmentsv1.CloningStatus{
				Phase:   environmentsv1.CloningPhasePending,
				Message: "Initializing",
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, sourceEnv, targetEnv).
		WithStatusSubresource(targetEnv).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &EnvironmentReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "target-env",
		},
	}

	// Reconcile - should move to CloningResources phase
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue)

	// Verify cloning status progressed
	updatedEnv := &environmentsv1.Environment{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "target-env"}, updatedEnv)
	assert.NoError(t, err)
	assert.NotNil(t, updatedEnv.Status.CloningStatus)
	assert.Equal(t, environmentsv1.CloningPhaseCloningResources, updatedEnv.Status.CloningStatus.Phase)
}
