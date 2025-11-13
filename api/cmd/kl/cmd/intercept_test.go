package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kl/pkg/workspace"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupTestClientForIntercept(t *testing.T, objects ...client.Object) {
	// Register workspace types
	if err := workspacesv1.AddToScheme(scheme.Scheme); err != nil {
		t.Fatalf("Failed to add workspace types to scheme: %v", err)
	}

	// Create fake client with objects
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(objects...).
		WithStatusSubresource(&workspacesv1.Workspace{}).
		Build()

	// Initialize the global WsClient for testing
	WsClient = &workspace.Client{
		K8sClient: fakeClient,
		Namespace: "test-namespace",
		Name:      "test-workspace",
	}
}

// TestWaitForInterceptSync_StartSuccess tests successful intercept activation
func TestWaitForInterceptSync_StartSuccess(t *testing.T) {
	// Create test workspace with active intercept in status
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			EnvironmentConnection: &workspacesv1.EnvironmentConnection{
				EnvironmentName: "test-env",
				Intercepts: []workspacesv1.InterceptSpec{
					{
						ServiceName: "api-server",
						PortMappings: []workspacesv1.PortMapping{
							{ServicePort: 8080, WorkspacePort: 8080, Protocol: "TCP"},
						},
					},
				},
			},
		},
		Status: workspacesv1.WorkspaceStatus{
			ActiveIntercepts: []workspacesv1.InterceptStatus{
				{
					ServiceName: "api-server",
					Phase:       "Active",
				},
			},
		},
	}

	setupTestClientForIntercept(t, workspace)

	// Test with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run waitForInterceptSync in a goroutine since it polls
	done := make(chan error, 1)
	go func() {
		done <- waitForInterceptSync("api-server", "env-test", "start")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Expected successful sync, got error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for intercept sync")
	}
}

// TestWaitForInterceptSync_StartFailure tests intercept activation failure
func TestWaitForInterceptSync_StartFailure(t *testing.T) {
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			EnvironmentConnection: &workspacesv1.EnvironmentConnection{
				EnvironmentName: "test-env",
				Intercepts: []workspacesv1.InterceptSpec{
					{
						ServiceName: "api-server",
						PortMappings: []workspacesv1.PortMapping{
							{ServicePort: 8080, WorkspacePort: 8080, Protocol: "TCP"},
						},
					},
				},
			},
		},
		Status: workspacesv1.WorkspaceStatus{
			ActiveIntercepts: []workspacesv1.InterceptStatus{
				{
					ServiceName: "api-server",
					Phase:       "Failed",
					Message:     "service not found",
				},
			},
		},
	}

	setupTestClientForIntercept(t, workspace)

	err := waitForInterceptSync("api-server", "env-test", "start")
	if err == nil {
		t.Fatal("Expected error for failed intercept, got nil")
	}

	expectedMsg := "service intercept failed: service not found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestWaitForInterceptSync_StopSuccess tests successful intercept deletion
func TestWaitForInterceptSync_StopSuccess(t *testing.T) {
	// Workspace with no active intercepts (simulating intercept was stopped)
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Status: workspacesv1.WorkspaceStatus{
			ActiveIntercepts: []workspacesv1.InterceptStatus{}, // No active intercepts
		},
	}

	setupTestClientForIntercept(t, workspace)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- waitForInterceptSync("api-server", "env-test", "stop")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Expected successful deletion sync, got error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for intercept deletion")
	}
}

// TestWaitForInterceptSync_Timeout tests timeout scenario
func TestWaitForInterceptSync_Timeout(t *testing.T) {
	t.Skip("Skipping timeout test as it takes 30+ seconds to run")

	// Workspace with intercept stuck in "Creating" phase
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Status: workspacesv1.WorkspaceStatus{
			ActiveIntercepts: []workspacesv1.InterceptStatus{
				{
					ServiceName: "api-server",
					Phase:       "Creating",
				},
			},
		},
	}

	setupTestClientForIntercept(t, workspace)

	err := waitForInterceptSync("api-server", "env-test", "start")
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if err.Error() != "timeout waiting for intercept sync after 30 seconds" {
		t.Errorf("Expected timeout error, got: %s", err.Error())
	}
}

// TestWaitForInterceptSync_FailureWithEmptyMessage tests failure with empty message
func TestWaitForInterceptSync_FailureWithEmptyMessage(t *testing.T) {
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Status: workspacesv1.WorkspaceStatus{
			ActiveIntercepts: []workspacesv1.InterceptStatus{
				{
					ServiceName: "api-server",
					Phase:       "Failed",
					Message:     "", // Empty message
				},
			},
		},
	}

	setupTestClientForIntercept(t, workspace)

	err := waitForInterceptSync("api-server", "env-test", "start")
	if err == nil {
		t.Fatal("Expected error for failed intercept, got nil")
	}

	expectedMsg := "service intercept failed: unknown error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestWaitForInterceptSync_NotFound tests when intercept doesn't exist yet (start action)
func TestWaitForInterceptSync_NotFoundThenCreated(t *testing.T) {
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Status: workspacesv1.WorkspaceStatus{
			ActiveIntercepts: []workspacesv1.InterceptStatus{}, // No intercepts initially
		},
	}

	setupTestClientForIntercept(t, workspace)

	// Update the workspace status after a delay to simulate controller creating it
	go func() {
		time.Sleep(500 * time.Millisecond)
		ctx := context.Background()
		ws := &workspacesv1.Workspace{}
		WsClient.K8sClient.Get(ctx, client.ObjectKey{Name: "test-workspace"}, ws)
		ws.Status.ActiveIntercepts = []workspacesv1.InterceptStatus{
			{
				ServiceName: "api-server",
				Phase:       "Active",
			},
		}
		WsClient.K8sClient.Status().Update(ctx, ws)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- waitForInterceptSync("api-server", "env-test", "start")
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Expected successful creation and activation, got error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("Test timed out waiting for intercept creation")
	}
}
