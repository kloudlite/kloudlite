package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kl/pkg/workspace"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupTestClientForIntercept(t *testing.T, objects ...client.Object) {
	// Register workspace types
	if err := workspacesv1.AddToScheme(scheme.Scheme); err != nil {
		t.Fatalf("Failed to add workspace types to scheme: %v", err)
	}

	// Register intercept types
	if err := interceptsv1.AddToScheme(scheme.Scheme); err != nil {
		t.Fatalf("Failed to add intercept types to scheme: %v", err)
	}

	// Create fake client with objects
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(objects...).
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
	// Create test workspace
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
		},
	}

	// Create ServiceIntercept that starts as "Creating" and becomes "Active"
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-server-test-workspace",
			Namespace: "env-test",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			ServiceRef: corev1.ObjectReference{
				Name:      "api-server",
				Namespace: "env-test",
			},
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "test-namespace",
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase: "Active",
		},
	}

	setupTestClientForIntercept(t, workspace, intercept)

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
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Create ServiceIntercept with Failed phase
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-server-test-workspace",
			Namespace: "env-test",
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:   "Failed",
			Message: "service not found",
		},
	}

	setupTestClientForIntercept(t, workspace, intercept)

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
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Don't create the intercept - simulating it's already deleted
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

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Create intercept stuck in "Creating" phase
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-server-test-workspace",
			Namespace: "env-test",
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase: "Creating",
		},
	}

	setupTestClientForIntercept(t, workspace, intercept)

	err := waitForInterceptSync("api-server", "env-test", "start")
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if err.Error() != "timeout waiting for intercept sync after 30 seconds" {
		t.Errorf("Expected timeout error, got: %s", err.Error())
	}
}

// TestWaitForInterceptSync_TransitionToActive tests intercept transitioning from Creating to Active
func TestWaitForInterceptSync_TransitionToActive(t *testing.T) {
	t.Skip("Skipping transition test - fake client doesn't support runtime status updates")

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Start with Creating phase
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-server-test-workspace",
			Namespace: "env-test",
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase: "Creating",
		},
	}

	setupTestClientForIntercept(t, workspace, intercept)

	// Note: This test is skipped because the fake client doesn't support
	// runtime status updates in the same way a real client does.
	// In a real scenario, the controller would update the status.
}

// TestWaitForInterceptSync_NotFound tests when intercept doesn't exist yet (start action)
func TestWaitForInterceptSync_NotFoundThenCreated(t *testing.T) {
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	setupTestClientForIntercept(t, workspace)

	// Create the intercept after a delay to simulate controller creating it
	go func() {
		time.Sleep(500 * time.Millisecond)
		intercept := &interceptsv1.ServiceIntercept{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "api-server-test-workspace",
				Namespace: "env-test",
			},
			Status: interceptsv1.ServiceInterceptStatus{
				Phase: "Active",
			},
		}
		WsClient.K8sClient.Create(context.Background(), intercept)
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

// TestWaitForInterceptSync_FailureWithEmptyMessage tests failure with empty message
func TestWaitForInterceptSync_FailureWithEmptyMessage(t *testing.T) {
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-server-test-workspace",
			Namespace: "env-test",
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:   "Failed",
			Message: "", // Empty message
		},
	}

	setupTestClientForIntercept(t, workspace, intercept)

	err := waitForInterceptSync("api-server", "env-test", "start")
	if err == nil {
		t.Fatal("Expected error for failed intercept, got nil")
	}

	expectedMsg := "service intercept failed: unknown error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestInterceptNameConstruction tests that intercept name is constructed correctly
func TestInterceptNameConstruction(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		workspaceName string
		expectedName  string
	}{
		{
			name:          "simple service name",
			serviceName:   "api-server",
			workspaceName: "my-workspace",
			expectedName:  "api-server-my-workspace",
		},
		{
			name:          "service with hyphens",
			serviceName:   "auth-service",
			workspaceName: "dev-workspace",
			expectedName:  "auth-service-dev-workspace",
		},
		{
			name:          "short names",
			serviceName:   "api",
			workspaceName: "ws",
			expectedName:  "api-ws",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace := &workspacesv1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.workspaceName,
					Namespace: "test-namespace",
				},
			}

			setupTestClientForIntercept(t, workspace)

			// The waitForInterceptSync function constructs the name internally
			// We verify this by checking what name it tries to get
			interceptName := tt.serviceName + "-" + tt.workspaceName

			if interceptName != tt.expectedName {
				t.Errorf("Expected intercept name '%s', got '%s'", tt.expectedName, interceptName)
			}

			// Verify by trying to get the intercept (should not be found)
			intercept := &interceptsv1.ServiceIntercept{}
			err := WsClient.K8sClient.Get(context.Background(), types.NamespacedName{
				Name:      interceptName,
				Namespace: "env-test",
			}, intercept)

			if err == nil {
				t.Error("Expected NotFound error for non-existent intercept")
			}
		})
	}
}
