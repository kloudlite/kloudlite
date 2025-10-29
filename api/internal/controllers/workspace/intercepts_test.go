package workspace

import (
	"context"
	"testing"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestPortMappingsEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []interceptsv1.PortMapping
		b        []interceptsv1.PortMapping
		expected bool
	}{
		{
			name: "equal mappings",
			a: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
				{ServicePort: 443, WorkspacePort: 8443, Protocol: "TCP"},
			},
			b: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
				{ServicePort: 443, WorkspacePort: 8443, Protocol: "TCP"},
			},
			expected: true,
		},
		{
			name: "different order same content",
			a: []interceptsv1.PortMapping{
				{ServicePort: 443, WorkspacePort: 8443, Protocol: "TCP"},
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
			},
			b: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
				{ServicePort: 443, WorkspacePort: 8443, Protocol: "TCP"},
			},
			expected: true,
		},
		{
			name: "different lengths",
			a: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
			},
			b: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
				{ServicePort: 443, WorkspacePort: 8443, Protocol: "TCP"},
			},
			expected: false,
		},
		{
			name: "different workspace ports",
			a: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
			},
			b: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 9090, Protocol: "TCP"},
			},
			expected: false,
		},
		{
			name: "different protocols",
			a: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
			},
			b: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "UDP"},
			},
			expected: false,
		},
		{
			name:     "both empty",
			a:        []interceptsv1.PortMapping{},
			b:        []interceptsv1.PortMapping{},
			expected: true,
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := portMappingsEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetServiceInterceptsForWorkspace(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
	}

	// Create intercepts with matching labels
	intercept1 := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-test-workspace",
			Namespace: "env-ns",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "test-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			ServiceRef: corev1.ObjectReference{Name: "web"},
		},
	}

	intercept2 := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-test-workspace",
			Namespace: "env-ns",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "test-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			ServiceRef: corev1.ObjectReference{Name: "api"},
		},
	}

	// Intercept in different namespace (should not be returned)
	intercept3 := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-test-workspace",
			Namespace: "other-ns",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "test-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			ServiceRef: corev1.ObjectReference{Name: "db"},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, intercept1, intercept2, intercept3).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	intercepts, err := reconciler.getServiceInterceptsForWorkspace(context.Background(), workspace, "env-ns")
	require.NoError(t, err)
	assert.Len(t, intercepts, 2)

	// Verify the intercepts are the right ones
	names := make(map[string]bool)
	for _, i := range intercepts {
		names[i.Name] = true
	}
	assert.True(t, names["web-test-workspace"])
	assert.True(t, names["api-test-workspace"])
	assert.False(t, names["db-test-workspace"])
}

func TestCreateServiceIntercept(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
	}

	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-env",
			Namespace: "workspace-ns",
		},
		Spec: environmentv1.EnvironmentSpec{
			TargetNamespace: "env-target-ns",
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "env-target-ns",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, env, service).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	interceptSpec := workspacev1.InterceptSpec{
		ServiceName: "web",
		PortMappings: []interceptsv1.PortMapping{
			{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
		},
	}

	err := reconciler.createServiceIntercept(context.Background(), workspace, env, interceptSpec, logger)
	require.NoError(t, err)

	// Verify intercept was created
	createdIntercept := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "web-test-workspace",
		Namespace: "env-target-ns",
	}, createdIntercept)
	require.NoError(t, err)

	// Verify labels
	assert.Equal(t, "test-workspace", createdIntercept.Labels["workspaces.kloudlite.io/workspace-name"])
	assert.Equal(t, "workspace-ns", createdIntercept.Labels["workspaces.kloudlite.io/workspace-namespace"])
	assert.Equal(t, "web", createdIntercept.Labels["intercepts.kloudlite.io/service-name"])

	// Verify spec
	assert.Equal(t, "test-workspace", createdIntercept.Spec.WorkspaceRef.Name)
	assert.Equal(t, "workspace-ns", createdIntercept.Spec.WorkspaceRef.Namespace)
	assert.Equal(t, "web", createdIntercept.Spec.ServiceRef.Name)
	assert.Equal(t, "env-target-ns", createdIntercept.Spec.ServiceRef.Namespace)
	assert.Len(t, createdIntercept.Spec.PortMappings, 1)
	assert.Equal(t, int32(80), createdIntercept.Spec.PortMappings[0].ServicePort)
	assert.Equal(t, int32(8080), createdIntercept.Spec.PortMappings[0].WorkspacePort)
}

func TestCreateServiceIntercept_ServiceNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
	}

	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-env",
			Namespace: "workspace-ns",
		},
		Spec: environmentv1.EnvironmentSpec{
			TargetNamespace: "env-target-ns",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, env).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	interceptSpec := workspacev1.InterceptSpec{
		ServiceName: "nonexistent",
		PortMappings: []interceptsv1.PortMapping{
			{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
		},
	}

	err := reconciler.createServiceIntercept(context.Background(), workspace, env, interceptSpec, logger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCleanupServiceIntercepts(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
	}

	// Create intercepts across different namespaces
	intercept1 := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-test-workspace",
			Namespace: "env-ns-1",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "test-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
	}

	intercept2 := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-test-workspace",
			Namespace: "env-ns-2",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "test-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
	}

	// Intercept from different workspace (should not be deleted)
	intercept3 := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "db-other-workspace",
			Namespace: "env-ns-1",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "other-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, intercept1, intercept2, intercept3).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.cleanupServiceIntercepts(context.Background(), workspace, logger)
	require.NoError(t, err)

	// Verify intercepts for test-workspace are deleted
	interceptList := &interceptsv1.ServiceInterceptList{}
	err = k8sClient.List(context.Background(), interceptList, client.MatchingLabels{
		"workspaces.kloudlite.io/workspace-name":      "test-workspace",
		"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
	})
	require.NoError(t, err)
	assert.Len(t, interceptList.Items, 0)

	// Verify intercept from other workspace still exists
	otherIntercept := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "db-other-workspace",
		Namespace: "env-ns-1",
	}, otherIntercept)
	require.NoError(t, err)
}

func TestReconcileServiceIntercepts_CreateNew(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
		Spec: workspacev1.WorkspaceSpec{
			EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
				EnvironmentRef: corev1.ObjectReference{Name: "test-env"},
				Intercepts: []workspacev1.InterceptSpec{
					{
						ServiceName: "web",
						PortMappings: []interceptsv1.PortMapping{
							{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
						},
					},
				},
			},
		},
	}

	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-env",
			Namespace: "workspace-ns",
		},
		Spec: environmentv1.EnvironmentSpec{
			TargetNamespace: "env-target-ns",
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "env-target-ns",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, env, service).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.reconcileServiceIntercepts(context.Background(), workspace, env, logger)
	require.NoError(t, err)

	// Verify intercept was created
	createdIntercept := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "web-test-workspace",
		Namespace: "env-target-ns",
	}, createdIntercept)
	require.NoError(t, err)
	assert.Equal(t, "web", createdIntercept.Spec.ServiceRef.Name)
}

func TestReconcileServiceIntercepts_DeleteObsolete(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
		Spec: workspacev1.WorkspaceSpec{
			EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
				EnvironmentRef: corev1.ObjectReference{Name: "test-env"},
				Intercepts: []workspacev1.InterceptSpec{
					// Only keep web intercept
					{
						ServiceName: "web",
						PortMappings: []interceptsv1.PortMapping{
							{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
						},
					},
				},
			},
		},
	}

	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-env",
			Namespace: "workspace-ns",
		},
		Spec: environmentv1.EnvironmentSpec{
			TargetNamespace: "env-target-ns",
		},
	}

	// Existing intercepts
	webIntercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-test-workspace",
			Namespace: "env-target-ns",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "test-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			ServiceRef: corev1.ObjectReference{Name: "web"},
			PortMappings: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"},
			},
		},
	}

	apiIntercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-test-workspace",
			Namespace: "env-target-ns",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "test-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			ServiceRef: corev1.ObjectReference{Name: "api"},
			PortMappings: []interceptsv1.PortMapping{
				{ServicePort: 3000, WorkspacePort: 3000, Protocol: "TCP"},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, env, webIntercept, apiIntercept).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.reconcileServiceIntercepts(context.Background(), workspace, env, logger)
	require.NoError(t, err)

	// Verify web intercept still exists
	webCheck := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "web-test-workspace",
		Namespace: "env-target-ns",
	}, webCheck)
	require.NoError(t, err)

	// Verify api intercept was deleted
	apiCheck := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "api-test-workspace",
		Namespace: "env-target-ns",
	}, apiCheck)
	require.Error(t, err)
	assert.True(t, client.IgnoreNotFound(err) == nil)
}

func TestReconcileServiceIntercepts_UpdatePortMappings(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
		Spec: workspacev1.WorkspaceSpec{
			EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
				EnvironmentRef: corev1.ObjectReference{Name: "test-env"},
				Intercepts: []workspacev1.InterceptSpec{
					{
						ServiceName: "web",
						PortMappings: []interceptsv1.PortMapping{
							{ServicePort: 80, WorkspacePort: 9090, Protocol: "TCP"},  // Changed port
							{ServicePort: 443, WorkspacePort: 9443, Protocol: "TCP"}, // New port
						},
					},
				},
			},
		},
	}

	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-env",
			Namespace: "workspace-ns",
		},
		Spec: environmentv1.EnvironmentSpec{
			TargetNamespace: "env-target-ns",
		},
	}

	// Existing intercept with old port mappings
	webIntercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-test-workspace",
			Namespace: "env-target-ns",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "test-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			ServiceRef: corev1.ObjectReference{Name: "web"},
			PortMappings: []interceptsv1.PortMapping{
				{ServicePort: 80, WorkspacePort: 8080, Protocol: "TCP"}, // Old port
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, env, webIntercept).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.reconcileServiceIntercepts(context.Background(), workspace, env, logger)
	require.NoError(t, err)

	// Verify intercept was updated with new port mappings
	updatedIntercept := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "web-test-workspace",
		Namespace: "env-target-ns",
	}, updatedIntercept)
	require.NoError(t, err)
	assert.Len(t, updatedIntercept.Spec.PortMappings, 2)

	// Check the updated ports
	portMap := make(map[int32]int32)
	for _, pm := range updatedIntercept.Spec.PortMappings {
		portMap[pm.ServicePort] = pm.WorkspacePort
	}
	assert.Equal(t, int32(9090), portMap[80])
	assert.Equal(t, int32(9443), portMap[443])
}

func TestReconcileServiceIntercepts_NoEnvironmentConnection(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
		Spec: workspacev1.WorkspaceSpec{
			EnvironmentConnection: nil, // No environment connection
		},
	}

	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-env",
			Namespace: "workspace-ns",
		},
		Spec: environmentv1.EnvironmentSpec{
			TargetNamespace: "env-target-ns",
		},
	}

	// Existing intercept that should be deleted
	existingIntercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web-test-workspace",
			Namespace: "env-target-ns",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name":      "test-workspace",
				"workspaces.kloudlite.io/workspace-namespace": "workspace-ns",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			ServiceRef: corev1.ObjectReference{Name: "web"},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, env, existingIntercept).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.reconcileServiceIntercepts(context.Background(), workspace, env, logger)
	require.NoError(t, err)

	// Verify existing intercept was deleted
	interceptCheck := &interceptsv1.ServiceIntercept{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "web-test-workspace",
		Namespace: "env-target-ns",
	}, interceptCheck)
	require.Error(t, err)
	assert.True(t, client.IgnoreNotFound(err) == nil)
}
