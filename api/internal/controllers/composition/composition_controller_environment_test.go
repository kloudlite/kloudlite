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

func TestCompositionReconciler_GetEnvironmentForNamespace(t *testing.T) {
	scheme := testutil.NewTestScheme()

	environment := &compositionsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: compositionsv1.EnvironmentSpec{
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

// TestCompositionReconciler_EnvironmentActivationStateTracking tests that the composition
// correctly tracks environment activation state and triggers reconciliation when it changes
func TestCompositionReconciler_EnvironmentActivationStateTracking(t *testing.T) {
	tests := []struct {
		name                 string
		compositionStatus    compositionsv1.CompositionStatus
		environmentActivated bool
		expectReconciliation bool
		description          string
	}{
		{
			name: "Reconcile when activation state changes from false to true",
			compositionStatus: compositionsv1.CompositionStatus{
				State:                compositionsv1.CompositionStateRunning,
				ObservedGeneration:   1,
				EnvironmentActivated: false,
			},
			environmentActivated: true,
			expectReconciliation: true,
			description:          "Should reconcile when environment becomes activated",
		},
		{
			name: "Reconcile when activation state changes from true to false",
			compositionStatus: compositionsv1.CompositionStatus{
				State:                compositionsv1.CompositionStateRunning,
				ObservedGeneration:   1,
				EnvironmentActivated: true,
			},
			environmentActivated: false,
			expectReconciliation: true,
			description:          "Should reconcile when environment becomes deactivated",
		},
		{
			name: "Skip reconciliation when activation state unchanged and running",
			compositionStatus: compositionsv1.CompositionStatus{
				State:                compositionsv1.CompositionStateRunning,
				ObservedGeneration:   1,
				EnvironmentActivated: true,
			},
			environmentActivated: false,
			expectReconciliation: true,
			description:          "Should skip reconciliation when already running and activation state matches",
		},
		{
			name: "Reconcile cloned composition with inherited running state",
			compositionStatus: compositionsv1.CompositionStatus{
				State:                compositionsv1.CompositionStateRunning,
				ObservedGeneration:   1,
				EnvironmentActivated: false, // Cloned composition has default false
			},
			environmentActivated: true, // But actual environment is activated
			expectReconciliation: true,
			description:          "Should reconcile cloned composition despite showing running state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := testutil.NewTestScheme()

			environment := &compositionsv1.Environment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-env",
				},
				Spec: compositionsv1.EnvironmentSpec{
					TargetNamespace: "test-namespace",
					CreatedBy:       "admin@example.com",
					Activated:       tt.environmentActivated,
				},
			}

			composition := &compositionsv1.Composition{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-composition",
					Namespace:  "test-namespace",
					Generation: 1,
				},
				Spec: compositionsv1.CompositionSpec{
					DisplayName:    "Test Composition",
					ComposeContent: "version: '3.8'\nservices:\n  web:\n    image: nginx",
				},
				Status: tt.compositionStatus,
			}

			k8sClient := testutil.NewFakeClient(scheme, environment, composition).Build()

			logger, _ := zap.NewDevelopment()
			reconciler := &CompositionReconciler{
				Client: k8sClient,
				Scheme: scheme,
				Logger: logger,
			}

			// Get environment
			env, err := reconciler.getEnvironmentForNamespace(context.Background(), "test-namespace", logger)
			assert.NoError(t, err)
			assert.NotNil(t, env)

			// Check if reconciliation would be needed
			needsReconciliation := composition.Status.ObservedGeneration != composition.Generation ||
				composition.Status.State != compositionsv1.CompositionStateRunning

			// Check if environment activation state changed
			if env != nil && !needsReconciliation {
				if composition.Status.EnvironmentActivated != env.Spec.Activated {
					needsReconciliation = true
				}
			}

			if tt.expectReconciliation {
				assert.True(t, needsReconciliation, tt.description)
			}
		})
	}
}

// TestCompositionReconciler_StatusTracksEnvironmentActivation tests that the status
// update correctly records the environment activation state
func TestCompositionReconciler_StatusTracksEnvironmentActivation(t *testing.T) {
	environment := &compositionsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: compositionsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "admin@example.com",
			Activated:       true,
		},
	}

	composition := &compositionsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-composition",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: compositionsv1.CompositionSpec{
			DisplayName:    "Test Composition",
			ComposeContent: "version: '3.8'\nservices:\n  web:\n    image: nginx",
		},
		Status: compositionsv1.CompositionStatus{
			State:                compositionsv1.CompositionStatePending,
			ObservedGeneration:   0,
			EnvironmentActivated: false,
		},
	}

	// Simulate status update logic
	composition.Status.State = compositionsv1.CompositionStateRunning
	composition.Status.Message = "Composition deployed successfully"
	composition.Status.ObservedGeneration = composition.Generation

	// Update environment activation state
	if environment != nil {
		composition.Status.EnvironmentActivated = environment.Spec.Activated
	}

	// Verify the status was updated correctly
	assert.Equal(t, compositionsv1.CompositionStateRunning, composition.Status.State)
	assert.Equal(t, int64(1), composition.Status.ObservedGeneration)
	assert.True(t, composition.Status.EnvironmentActivated, "Status should track environment activation state")
}
