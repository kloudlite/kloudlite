package workspace

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func TestWorkspaceReconciler_handleDeletion(t *testing.T) {
	scheme := testutil.NewTestScheme()
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name                string
		workspace           *workspacev1.Workspace
		existingPod         *corev1.Pod
		expectFinalizerGone bool
		expectError         bool
	}{
		{
			name: "Delete workspace with pod",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-workspace",
					Namespace:  "default",
					Finalizers: []string{workspaceFinalizer},
				},
				Spec: workspacev1.WorkspaceSpec{
					OwnedBy:     "test-user",
					DisplayName: "Test Workspace",
					Status:      "active",
				},
			},
			existingPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "workspace-test-workspace",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "workspace", Image: "test:latest"},
					},
				},
			},
			expectFinalizerGone: true,
			expectError:         false,
		},
		{
			name: "Delete workspace without pod",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-workspace-2",
					Namespace:  "default",
					Finalizers: []string{workspaceFinalizer},
				},
				Spec: workspacev1.WorkspaceSpec{
					OwnedBy:     "test-user",
					DisplayName: "Test Workspace 2",
					Status:      "active",
				},
			},
			existingPod:         nil,
			expectFinalizerGone: true,
			expectError:         false,
		},
		{
			name: "Workspace without finalizer",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "test-workspace-3",
					Namespace:  "default",
					Finalizers: []string{}, // No finalizer
				},
				Spec: workspacev1.WorkspaceSpec{
					OwnedBy:     "test-user",
					DisplayName: "Test Workspace 3",
					Status:      "active",
				},
			},
			existingPod:         nil,
			expectFinalizerGone: true, // Already gone
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build list of objects to initialize
			objects := []client.Object{tt.workspace}
			if tt.existingPod != nil {
				objects = append(objects, tt.existingPod)
			}

			k8sClient := testutil.NewFakeClient(scheme, objects...).Build()

			reconciler := &WorkspaceReconciler{
				Client: k8sClient,
				Scheme: scheme,
				Logger: logger,
			}

			// Call handleDeletion
			_, err := reconciler.handleDeletion(context.Background(), tt.workspace, logger)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify finalizer was removed
			updatedWorkspace := &workspacev1.Workspace{}
			err = k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      tt.workspace.Name,
				Namespace: tt.workspace.Namespace,
			}, updatedWorkspace)

			if tt.expectFinalizerGone {
				assert.NoError(t, err)
				assert.False(t, controllerutil.ContainsFinalizer(updatedWorkspace, workspaceFinalizer),
					"Finalizer should be removed")
			}

			// Verify pod was deleted if it existed
			if tt.existingPod != nil {
				pod := &corev1.Pod{}
				err = k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      tt.existingPod.Name,
					Namespace: tt.existingPod.Namespace,
				}, pod)
				// Pod should be deleted (NotFound) or have deletion timestamp
				assert.True(t, apierrors.IsNotFound(err) || pod.DeletionTimestamp != nil,
					"Pod should be deleted or have deletion timestamp")
			}
		})
	}
}

func TestWorkspaceReconciler_validateHostPath(t *testing.T) {
	tests := []struct {
		name          string
		hostPath      string
		workspaceName string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Valid path matching workspace name",
			hostPath:      "/home/kl/workspaces/my-workspace",
			workspaceName: "my-workspace",
			expectError:   false,
		},
		{
			name:          "Path traversal attempt with ..",
			hostPath:      "/home/kl/workspaces/../../../etc/passwd",
			workspaceName: "my-workspace",
			expectError:   true,
			errorContains: "must end with workspace name",
		},
		{
			name:          "Path not ending with workspace name",
			hostPath:      "/home/kl/workspaces/wrong-name",
			workspaceName: "my-workspace",
			expectError:   true,
			errorContains: "must end with workspace name",
		},
		{
			name:          "Path outside workspaces directory",
			hostPath:      "/etc/passwd",
			workspaceName: "my-workspace",
			expectError:   true,
			errorContains: "must be within",
		},
		{
			name:          "Empty path",
			hostPath:      "",
			workspaceName: "my-workspace",
			expectError:   true,
			errorContains: "cannot be empty",
		},
		{
			name:          "Path with shell metacharacters",
			hostPath:      "/home/kl/workspaces/my-workspace; rm -rf /",
			workspaceName: "my-workspace",
			expectError:   true,
			errorContains: "must end with workspace name",
		},
		{
			name:          "Valid path with hyphens",
			hostPath:      "/home/kl/workspaces/my-test-workspace-123",
			workspaceName: "my-test-workspace-123",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := testutil.NewTestScheme()
			logger, _ := zap.NewDevelopment()
			k8sClient := testutil.NewFakeClient(scheme).Build()

			reconciler := &WorkspaceReconciler{
				Client: k8sClient,
				Scheme: scheme,
				Logger: logger,
			}

			err := reconciler.validateHostPath(tt.hostPath, tt.workspaceName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWorkspaceReconciler_deleteHostDirectory(t *testing.T) {
	scheme := testutil.NewTestScheme()
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name          string
		hostPath      string
		workspaceName string
		expectError   bool
		expectPod     bool
	}{
		{
			name:          "Create cleanup pod for valid path",
			hostPath:      "/home/kl/workspaces/test-workspace",
			workspaceName: "test-workspace",
			expectError:   false,
			expectPod:     true,
		},
		{
			name:          "Reject unsafe path with traversal",
			hostPath:      "/home/kl/workspaces/../../../etc/passwd",
			workspaceName: "test-workspace",
			expectError:   true,
			expectPod:     false,
		},
		{
			name:          "Reject path outside workspaces",
			hostPath:      "/etc/passwd",
			workspaceName: "test-workspace",
			expectError:   true,
			expectPod:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := testutil.NewFakeClient(scheme).Build()

			reconciler := &WorkspaceReconciler{
				Client: k8sClient,
				Scheme: scheme,
				Logger: logger,
			}

			err := reconciler.deleteHostDirectory(context.Background(), tt.hostPath, tt.workspaceName, logger)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check if cleanup pod was created
			if tt.expectPod {
				// The pod name is generated from the path, so we list all pods and check
				podList := &corev1.PodList{}
				err := k8sClient.List(context.Background(), podList,
					client.InNamespace("default"),
					client.MatchingLabels{"app": "workspace-cleanup"})

				assert.NoError(t, err)
				assert.Greater(t, len(podList.Items), 0, "Cleanup pod should be created")

				// Verify pod has correct labels
				pod := podList.Items[0]
				assert.Equal(t, "workspace-cleanup", pod.Labels["app"])
				assert.Equal(t, "temporary", pod.Labels["type"])

				// Verify pod uses correct image and command
				assert.Equal(t, "alpine:latest", pod.Spec.Containers[0].Image)
				assert.Equal(t, []string{"rm", "-rf", tt.hostPath}, pod.Spec.Containers[0].Command)

				// Verify host path volume mount
				assert.Len(t, pod.Spec.Volumes, 1)
				assert.Equal(t, "host-home", pod.Spec.Volumes[0].Name)
				assert.NotNil(t, pod.Spec.Volumes[0].HostPath)
				assert.Equal(t, "/home/kl", pod.Spec.Volumes[0].HostPath.Path)

				// Verify ActiveDeadlineSeconds is set
				assert.NotNil(t, pod.Spec.ActiveDeadlineSeconds)
				assert.Equal(t, int64(300), *pod.Spec.ActiveDeadlineSeconds)
			}
		})
	}
}

func TestWorkspaceReconciler_handleSuspendedWorkspace(t *testing.T) {
	scheme := testutil.NewTestScheme()
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name              string
		workspace         *workspacev1.Workspace
		existingPod       *corev1.Pod
		expectedPhase     string
		expectedMessage   string
		expectPodDeletion bool
	}{
		{
			name: "Suspend workspace with running pod",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workspace",
					Namespace: "default",
				},
				Spec: workspacev1.WorkspaceSpec{
					OwnedBy:     "test-user",
					DisplayName: "Test Workspace",
					Status:      "suspended",
				},
			},
			existingPod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "workspace-test-workspace",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "workspace", Image: "test:latest"},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			expectedPhase:     "Stopping",
			expectedMessage:   "Workspace is being stopped",
			expectPodDeletion: true,
		},
		{
			name: "Suspend workspace without pod (already stopped)",
			workspace: &workspacev1.Workspace{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workspace-2",
					Namespace: "default",
				},
				Spec: workspacev1.WorkspaceSpec{
					OwnedBy:     "test-user",
					DisplayName: "Test Workspace 2",
					Status:      "suspended",
				},
			},
			existingPod:       nil,
			expectedPhase:     "Stopped",
			expectedMessage:   "Workspace is stopped",
			expectPodDeletion: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build list of objects to initialize
			objects := []client.Object{tt.workspace}
			if tt.existingPod != nil {
				objects = append(objects, tt.existingPod)
			}

			k8sClient := testutil.NewFakeClient(scheme, objects...).
				WithStatusSubresource(&workspacev1.Workspace{}).
				Build()

			reconciler := &WorkspaceReconciler{
				Client: k8sClient,
				Scheme: scheme,
				Logger: logger,
			}

			// Call handleSuspendedWorkspace
			result, err := reconciler.handleSuspendedWorkspace(context.Background(), tt.workspace, logger)

			assert.NoError(t, err)

			// Verify result
			if tt.existingPod != nil {
				// Should requeue when deleting pod
				assert.True(t, result.RequeueAfter > 0)
			} else {
				// Should not requeue when already stopped
				assert.Equal(t, time.Duration(0), result.RequeueAfter)
			}

			// Verify workspace status was updated
			updatedWorkspace := &workspacev1.Workspace{}
			err = k8sClient.Get(context.Background(), types.NamespacedName{
				Name:      tt.workspace.Name,
				Namespace: tt.workspace.Namespace,
			}, updatedWorkspace)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPhase, updatedWorkspace.Status.Phase)
			assert.Equal(t, tt.expectedMessage, updatedWorkspace.Status.Message)

			// Verify pod deletion
			if tt.expectPodDeletion {
				pod := &corev1.Pod{}
				err = k8sClient.Get(context.Background(), types.NamespacedName{
					Name:      fmt.Sprintf("workspace-%s", tt.workspace.Name),
					Namespace: tt.workspace.Namespace,
				}, pod)
				// Pod should have deletion timestamp
				assert.True(t, apierrors.IsNotFound(err) || pod.DeletionTimestamp != nil)
			}

			// Verify status fields are cleared when stopped
			if tt.expectedPhase == "Stopped" {
				assert.Empty(t, updatedWorkspace.Status.PodName)
				assert.Empty(t, updatedWorkspace.Status.PodIP)
				assert.Empty(t, updatedWorkspace.Status.NodeName)
				assert.NotNil(t, updatedWorkspace.Status.StopTime)
			}
		})
	}
}
