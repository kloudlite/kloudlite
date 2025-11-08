package composition

import (
	"context"
	"testing"

	compositionsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompositionReconciler_FindCompositionsForConfigMap(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition1 := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp1",
			Namespace: "test-namespace",
		},
	}

	composition2 := &compositionsv1.Composition{
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

	k8sClient := testutil.NewFakeClient(scheme, composition1, composition2, configMap).
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
	scheme := testutil.NewTestScheme()

	composition := &compositionsv1.Composition{
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

	k8sClient := testutil.NewFakeClient(scheme, composition, configMap).
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
	scheme := testutil.NewTestScheme()

	composition1 := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp1",
			Namespace: "test-namespace",
		},
	}

	composition2 := &compositionsv1.Composition{
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

	k8sClient := testutil.NewFakeClient(scheme, composition1, composition2, secret).
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
	scheme := testutil.NewTestScheme()

	composition := &compositionsv1.Composition{
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

	k8sClient := testutil.NewFakeClient(scheme, composition, secret).
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

func TestCompositionReconciler_FindCompositionsForEnvironment(t *testing.T) {
	scheme := testutil.NewTestScheme()

	composition1 := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp1",
			Namespace: "test-namespace",
		},
	}

	composition2 := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "comp2",
			Namespace: "test-namespace",
		},
	}

	environment := &compositionsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: compositionsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			OwnedBy:         "admin@example.com",
			Activated:       true,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, composition1, composition2, environment).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	requests := reconciler.findCompositionsForEnvironment(context.Background(), environment)

	// Should return reconcile requests for all compositions in the environment's namespace
	assert.Equal(t, 2, len(requests))

	names := []string{requests[0].Name, requests[1].Name}
	assert.Contains(t, names, "comp1")
	assert.Contains(t, names, "comp2")
}
