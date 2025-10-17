package composition

import (
	"context"
	"testing"

	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestCompositionReconciler_DeployComposition_Success(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Finalizers: []string{compositionFinalizer},
			Generation: 1,
		},
		Spec: compositionsv1.CompositionSpec{
			DisplayName: "Test Composition",
			ComposeContent: `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"`,
		},
		Status: compositionsv1.CompositionStatus{
			ObservedGeneration: 0,
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

	err := reconciler.deployComposition(context.Background(), composition, logger)
	assert.NoError(t, err)
	assert.NotNil(t, composition.Status.DeployedResources)
}

func TestCompositionReconciler_DeployComposition_ParseError(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
		Spec: compositionsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: `invalid yaml content [[[`,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme).Build()

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
	scheme := testutil.NewTestScheme()

	composition := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
		Spec: compositionsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: "",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme).Build()

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

func TestCompositionReconciler_DeployComposition_ConversionError(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-composition",
			Namespace: "test-namespace",
		},
		Spec: compositionsv1.CompositionSpec{
			DisplayName: "Test Composition",
			ComposeContent: `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "invalid-port-format"`, // This will cause conversion error
		},
	}

	k8sClient := testutil.NewFakeClient(scheme).Build()

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

func TestCompositionReconciler_DeployComposition_WithInactiveEnvironment(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Create inactive environment
	environment := &compositionsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: compositionsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "admin@example.com",
			Activated:       false, // Inactive
		},
	}

	composition := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Finalizers: []string{compositionFinalizer},
		},
		Spec: compositionsv1.CompositionSpec{
			DisplayName: "Test Composition",
			ComposeContent: `version: '3.8'
services:
  web:
    image: nginx:latest
    deploy:
      replicas: 3`,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition, environment).
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

	// Verify deployment was created with 0 replicas
	deploymentList := &appsv1.DeploymentList{}
	err = k8sClient.List(context.Background(), deploymentList, client.InNamespace("test-namespace"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deploymentList.Items))

	deployment := deploymentList.Items[0]
	assert.NotNil(t, deployment.Spec.Replicas)
	assert.Equal(t, int32(0), *deployment.Spec.Replicas)
	assert.Equal(t, "3", deployment.Annotations["kloudlite.io/original-replicas"])
}

func TestCompositionReconciler_DeployComposition_WithActiveEnvironment(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Create active environment
	environment := &compositionsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: compositionsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "admin@example.com",
			Activated:       true, // Active
		},
	}

	composition := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Finalizers: []string{compositionFinalizer},
		},
		Spec: compositionsv1.CompositionSpec{
			DisplayName: "Test Composition",
			ComposeContent: `version: '3.8'
services:
  web:
    image: nginx:latest
    deploy:
      replicas: 2`,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition, environment).
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

	// Verify deployment was created with original replicas
	deploymentList := &appsv1.DeploymentList{}
	err = k8sClient.List(context.Background(), deploymentList, client.InNamespace("test-namespace"))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deploymentList.Items))

	deployment := deploymentList.Items[0]
	assert.NotNil(t, deployment.Spec.Replicas)
	assert.Equal(t, int32(2), *deployment.Spec.Replicas)
	assert.NotContains(t, deployment.Annotations, "kloudlite.io/original-replicas")
}
