package composition

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompositionReconciler_GetEnvironmentForNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()

	environment := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "admin@example.com",
			Activated:       true,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, environment).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	env, err := reconciler.getEnvironmentForNamespace(context.Background(), "test-namespace", logger)
	assert.NoError(t, err)
	assert.NotNil(t, env)
	assert.Equal(t, "test-env", env.Name)
	assert.True(t, env.Spec.Activated)
}

func TestCompositionReconciler_GetEnvironmentForNamespace_NotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()

	k8sClient := testutil.NewFakeClient(scheme).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &CompositionReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	env, err := reconciler.getEnvironmentForNamespace(context.Background(), "nonexistent-namespace", logger)
	assert.NoError(t, err)
	assert.Nil(t, env)
}

func TestCompositionReconciler_FetchEnvironmentData(t *testing.T) {
	scheme := testutil.NewTestScheme()

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

	k8sClient := testutil.NewFakeClient(scheme, envConfigMap, envSecret, configFileMap1, configFileMap2).
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
	scheme := testutil.NewTestScheme()

	k8sClient := testutil.NewFakeClient(scheme).Build()

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
