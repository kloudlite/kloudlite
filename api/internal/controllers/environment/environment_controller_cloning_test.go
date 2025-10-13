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

func TestEnvironmentReconciler_HandleCloning_Success(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Create source environment
	sourceEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "source-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "source-namespace",
			CreatedBy:       "admin@example.com",
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
			CreatedBy:       "admin@example.com",
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

	k8sClient := testutil.NewFakeClient(scheme, sourceEnv, targetEnv, sourceConfigMap, sourceSecret, sourceComposition).
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

	// Verify target namespace was created
	targetNamespace := &corev1.Namespace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "target-namespace"}, targetNamespace)
	assert.NoError(t, err)
	assert.Equal(t, "source-env", targetNamespace.Annotations["kloudlite.io/cloned-from"])

	// Verify ConfigMap was cloned
	clonedConfigMap := &corev1.ConfigMap{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "env-config", Namespace: "target-namespace"}, clonedConfigMap)
	assert.NoError(t, err)
	assert.Equal(t, "https://api.example.com", clonedConfigMap.Data["API_URL"])
	assert.Equal(t, "target-env", clonedConfigMap.Labels["kloudlite.io/environment"])

	// Verify Secret was cloned
	clonedSecret := &corev1.Secret{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "env-secret", Namespace: "target-namespace"}, clonedSecret)
	assert.NoError(t, err)
	assert.Equal(t, []byte("secret123"), clonedSecret.Data["DB_PASSWORD"])
	assert.Equal(t, "target-env", clonedSecret.Labels["kloudlite.io/environment"])

	// Verify Composition was cloned
	clonedComposition := &environmentsv1.Composition{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{Name: "web-app", Namespace: "target-namespace"}, clonedComposition)
	assert.NoError(t, err)
	assert.Equal(t, "Web App", clonedComposition.Spec.DisplayName)
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
			CreatedBy:       "admin@example.com",
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
