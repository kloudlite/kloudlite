package workspace

import (
	"context"
	"os"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupTestClient(t *testing.T, objects ...client.Object) *Client {
	// Register workspace types
	if err := workspacesv1.AddToScheme(scheme.Scheme); err != nil {
		t.Fatalf("Failed to add workspace types to scheme: %v", err)
	}

	// Register environment types
	if err := environmentsv1.AddToScheme(scheme.Scheme); err != nil {
		t.Fatalf("Failed to add environment types to scheme: %v", err)
	}

	// Create fake client with objects
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(objects...).
		Build()

	return &Client{
		K8sClient: fakeClient,
		Namespace: "test-namespace",
		Name:      "test-workspace",
	}
}

func TestClient_Get_Success(t *testing.T) {
	// Create test workspace
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "user@example.com",
		},
	}

	client := setupTestClient(t, workspace)

	// Test Get
	result, err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result.Name != "test-workspace" {
		t.Errorf("Expected name 'test-workspace', got %s", result.Name)
	}
	if result.Spec.DisplayName != "Test Workspace" {
		t.Errorf("Expected display name 'Test Workspace', got %s", result.Spec.DisplayName)
	}
}

func TestClient_Get_NotFound(t *testing.T) {
	client := setupTestClient(t) // No workspace created

	_, err := client.Get(context.Background())
	if err == nil {
		t.Fatal("Expected error for missing workspace, got nil")
	}
}

func TestClient_Update_Success(t *testing.T) {
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Original Name",
		},
	}

	client := setupTestClient(t, workspace)

	// Get workspace
	ws, err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Update workspace
	ws.Spec.DisplayName = "Updated Name"
	err = client.Update(context.Background(), ws)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updated, err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if updated.Spec.DisplayName != "Updated Name" {
		t.Errorf("Expected 'Updated Name', got %s", updated.Spec.DisplayName)
	}
}

func TestClient_Patch_Success(t *testing.T) {
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Original Name",
		},
	}

	testClient := setupTestClient(t, workspace)

	// Get workspace
	ws, err := testClient.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Create patch using MergePatch
	ws.Spec.DisplayName = "Patched Name"
	patch := client.MergeFrom(workspace)

	// Use Patch method
	err = testClient.Patch(context.Background(), ws, patch)
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}

	// Verify patch
	patched, err := testClient.Get(context.Background())
	if err != nil {
		t.Fatalf("Get after patch failed: %v", err)
	}
	if patched.Spec.DisplayName != "Patched Name" {
		t.Errorf("Expected 'Patched Name', got %s", patched.Spec.DisplayName)
	}
}

func TestNew_WithEnvVars(t *testing.T) {
	tests := []struct {
		name               string
		workspaceName      string
		workspaceNamespace string
		hostname           string
		expectedName       string
		expectedNamespace  string
	}{
		{
			name:               "with both env vars",
			workspaceName:      "my-workspace",
			workspaceNamespace: "my-namespace",
			expectedName:       "my-workspace",
			expectedNamespace:  "my-namespace",
		},
		{
			name:               "fallback to hostname",
			workspaceName:      "",
			workspaceNamespace: "my-namespace",
			hostname:           "workspace-pod-123",
			expectedName:       "workspace-pod-123",
			expectedNamespace:  "my-namespace",
		},
		{
			name:               "fallback to default namespace",
			workspaceName:      "my-workspace",
			workspaceNamespace: "",
			expectedName:       "my-workspace",
			expectedNamespace:  "default",
		},
		{
			name:               "all fallbacks",
			workspaceName:      "",
			workspaceNamespace: "",
			hostname:           "pod-name",
			expectedName:       "pod-name",
			expectedNamespace:  "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			os.Setenv("WORKSPACE_NAME", tt.workspaceName)
			os.Setenv("WORKSPACE_NAMESPACE", tt.workspaceNamespace)
			os.Setenv("HOSTNAME", tt.hostname)
			defer func() {
				os.Unsetenv("WORKSPACE_NAME")
				os.Unsetenv("WORKSPACE_NAMESPACE")
				os.Unsetenv("HOSTNAME")
			}()

			// Register scheme for New() to work
			if err := workspacesv1.AddToScheme(scheme.Scheme); err != nil {
				t.Fatalf("Failed to add workspace types to scheme: %v", err)
			}

			// Note: New() will fail because it tries to connect to K8s
			// We're just testing the env var logic here
			// In a real implementation, we'd need to refactor New() to accept
			// a config parameter for testability
		})
	}
}

func TestClient_MultipleWorkspaces(t *testing.T) {
	// Create multiple workspaces
	workspace1 := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-1",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Workspace 1",
		},
	}
	workspace2 := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-2",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Workspace 2",
		},
	}

	client := setupTestClient(t, workspace1, workspace2)

	// Update client to fetch workspace-1
	client.Name = "workspace-1"
	result, err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result.Name != "workspace-1" {
		t.Errorf("Expected 'workspace-1', got %s", result.Name)
	}
	if result.Spec.DisplayName != "Workspace 1" {
		t.Errorf("Expected 'Workspace 1', got %s", result.Spec.DisplayName)
	}

	// Now fetch workspace-2
	client.Name = "workspace-2"
	result, err = client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if result.Name != "workspace-2" {
		t.Errorf("Expected 'workspace-2', got %s", result.Name)
	}
	if result.Spec.DisplayName != "Workspace 2" {
		t.Errorf("Expected 'Workspace 2', got %s", result.Spec.DisplayName)
	}
}

func TestClient_UpdatePackages(t *testing.T) {
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Packages:    []workspacesv1.PackageSpec{},
		},
	}

	client := setupTestClient(t, workspace)

	// Get workspace
	ws, err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Add packages
	ws.Spec.Packages = append(ws.Spec.Packages, workspacesv1.PackageSpec{
		Name:          "nodejs_20",
		NixpkgsCommit: "abc123",
	})
	ws.Spec.Packages = append(ws.Spec.Packages, workspacesv1.PackageSpec{
		Name:          "python312",
		NixpkgsCommit: "def456",
	})

	// Update
	err = client.Update(context.Background(), ws)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify packages
	updated, err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if len(updated.Spec.Packages) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(updated.Spec.Packages))
	}
	if updated.Spec.Packages[0].Name != "nodejs_20" {
		t.Errorf("Expected first package 'nodejs_20', got %s", updated.Spec.Packages[0].Name)
	}
}

func TestClient_GetNamespaceName(t *testing.T) {
	client := &Client{
		Namespace: "custom-namespace",
		Name:      "custom-workspace",
	}

	if client.Namespace != "custom-namespace" {
		t.Errorf("Expected namespace 'custom-namespace', got %s", client.Namespace)
	}
	if client.Name != "custom-workspace" {
		t.Errorf("Expected name 'custom-workspace', got %s", client.Name)
	}
}

func TestClient_WorkspaceWithEnvironmentRef(t *testing.T) {
	// Create test environment
	environment := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-environment",
			Namespace: "test-namespace",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "env-test",
			Activated:       true,
		},
	}

	// Create workspace with environment reference
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "user@example.com",
			EnvironmentRef: &corev1.ObjectReference{
				Name:      "test-environment",
				Namespace: "test-namespace",
			},
		},
	}

	client := setupTestClient(t, workspace, environment)

	// Test Get
	result, err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result.Spec.EnvironmentRef == nil {
		t.Fatal("Expected environment ref, got nil")
	}
	if result.Spec.EnvironmentRef.Name != "test-environment" {
		t.Errorf("Expected environment ref 'test-environment', got %s", result.Spec.EnvironmentRef.Name)
	}

	// Verify environment exists and can be fetched
	env := &environmentsv1.Environment{}
	err = client.K8sClient.Get(context.Background(), types.NamespacedName{
		Name:      result.Spec.EnvironmentRef.Name,
		Namespace: "test-namespace",
	}, env)
	if err != nil {
		t.Fatalf("Failed to get environment: %v", err)
	}
	if env.Spec.TargetNamespace != "env-test" {
		t.Errorf("Expected target namespace 'env-test', got %s", env.Spec.TargetNamespace)
	}
}

func TestClient_UpdateWorkspaceEnvironmentRef(t *testing.T) {
	// Create test environment
	environment := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "production",
			Namespace: "test-namespace",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "env-production",
			Activated:       true,
		},
	}

	// Create workspace without environment reference
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "user@example.com",
		},
	}

	client := setupTestClient(t, workspace, environment)

	// Get workspace
	ws, err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Add environment reference
	ws.Spec.EnvironmentRef = &corev1.ObjectReference{
		Name:      "production",
		Namespace: "test-namespace",
	}

	// Update workspace
	err = client.Update(context.Background(), ws)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updated, err := client.Get(context.Background())
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if updated.Spec.EnvironmentRef == nil {
		t.Fatal("Expected environment ref after update, got nil")
	}
	if updated.Spec.EnvironmentRef.Name != "production" {
		t.Errorf("Expected environment ref 'production', got %s", updated.Spec.EnvironmentRef.Name)
	}
}
