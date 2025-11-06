package workspace

import (
	"context"
	"testing"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestHasActiveConnections_PodNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	hasConnections, count, err := reconciler.hasActiveConnections(context.Background(), workspace)
	assert.Error(t, err)
	assert.False(t, hasConnections)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "failed to get pod")
}

func TestHasActiveConnections_PodNoPodIP(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Pod without PodIP
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "test-namespace",
		},
		Status: corev1.PodStatus{
			PodIP: "", // No IP assigned yet
			Phase: corev1.PodPending,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, pod).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	hasConnections, count, err := reconciler.hasActiveConnections(context.Background(), workspace)
	assert.NoError(t, err)
	assert.False(t, hasConnections)
	assert.Equal(t, 0, count)
}

func TestHasActiveConnections_PodNotRunning(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Pod in pending state
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "test-namespace",
		},
		Status: corev1.PodStatus{
			PodIP: "10.0.0.1",
			Phase: corev1.PodPending, // Not running yet
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, pod).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	hasConnections, count, err := reconciler.hasActiveConnections(context.Background(), workspace)
	assert.NoError(t, err)
	assert.True(t, hasConnections) // Consider as active while starting
	assert.Equal(t, 0, count)
}

func TestHasActiveConnections_PodJustStarted(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Pod started 1 minute ago (within 2-minute grace period)
	startTime := metav1.NewTime(time.Now().Add(-1 * time.Minute))
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "workspace"},
			},
		},
		Status: corev1.PodStatus{
			PodIP:     "10.0.0.1",
			Phase:     corev1.PodRunning,
			StartTime: &startTime,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, pod).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	hasConnections, count, err := reconciler.hasActiveConnections(context.Background(), workspace)
	assert.NoError(t, err)
	assert.True(t, hasConnections) // Grace period - consider as having connections
	assert.Equal(t, 0, count)
}

func TestHasActiveConnections_NoContainers(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Pod started long ago but has no containers
	startTime := metav1.NewTime(time.Now().Add(-10 * time.Minute))
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "test-namespace",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{}, // No containers
		},
		Status: corev1.PodStatus{
			PodIP:     "10.0.0.1",
			Phase:     corev1.PodRunning,
			StartTime: &startTime,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, pod).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	hasConnections, count, err := reconciler.hasActiveConnections(context.Background(), workspace)
	assert.NoError(t, err)
	assert.False(t, hasConnections)
	assert.Equal(t, 0, count)
}

func TestIsWorkspaceIdle_WithActiveConnections(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Pod not running (which counts as active during startup)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "test-namespace",
		},
		Status: corev1.PodStatus{
			PodIP: "10.0.0.1",
			Phase: corev1.PodPending, // Not running = considered active
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, pod).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	isIdle, count, err := reconciler.isWorkspaceIdle(context.Background(), workspace)
	assert.NoError(t, err)
	assert.False(t, isIdle) // Should not be idle when pod is starting
	assert.Equal(t, 0, count)
}

func TestIsWorkspaceIdle_NoConnections(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
	}

	// Pod with no IP (no connections possible)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "test-namespace",
		},
		Status: corev1.PodStatus{
			PodIP: "", // No IP = no connections
			Phase: corev1.PodPending,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, pod).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	isIdle, count, err := reconciler.isWorkspaceIdle(context.Background(), workspace)
	assert.NoError(t, err)
	assert.True(t, isIdle) // No IP = idle
	assert.Equal(t, 0, count)
}

func TestCheckAndSuspendIdleWorkspace_AutoStopNotEnabled(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			Status: "active",
			Settings: &workspacev1.WorkspaceSettings{
				AutoStop: false, // Auto-stop not enabled
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).
		WithStatusSubresource(&workspacev1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.checkAndSuspendIdleWorkspace(context.Background(), workspace, logger)
	assert.NoError(t, err)

	// Verify workspace was not suspended
	updatedWorkspace := &workspacev1.Workspace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      workspace.Name,
		Namespace: workspace.Namespace,
	}, updatedWorkspace)
	assert.NoError(t, err)
	assert.Equal(t, "active", updatedWorkspace.Spec.Status)
}

func TestCheckAndSuspendIdleWorkspace_IdleTimeoutExceeded(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Set LastActivityTime to 31 minutes ago (exceeds 30 min timeout)
	lastActivityTime := metav1.NewTime(time.Now().Add(-31 * time.Minute))

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			Status: "active",
			Settings: &workspacev1.WorkspaceSettings{
				AutoStop:    true,
				IdleTimeout: 30, // 30 minutes
			},
		},
		Status: workspacev1.WorkspaceStatus{
			LastActivityTime: &lastActivityTime,
		},
	}

	// Pod with no connections (idle)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "test-namespace",
		},
		Status: corev1.PodStatus{
			PodIP: "", // No IP = no connections = idle
			Phase: corev1.PodPending,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, pod).
		WithStatusSubresource(&workspacev1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.checkAndSuspendIdleWorkspace(context.Background(), workspace, logger)
	assert.NoError(t, err)

	// Verify workspace was suspended
	updatedWorkspace := &workspacev1.Workspace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      workspace.Name,
		Namespace: workspace.Namespace,
	}, updatedWorkspace)
	assert.NoError(t, err)
	assert.Equal(t, "suspended", updatedWorkspace.Spec.Status)
}

func TestWorkspaceReconciler_CreateWorkspacePod_NixVolumeMount(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			Packages:    []workspacev1.PackageSpec{},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacev1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	// Create workspace pod
	pod, err := reconciler.createWorkspacePod(workspace)
	assert.NoError(t, err)
	assert.NotNil(t, pod)

	// Find the workspace container
	var workspaceContainer *corev1.Container
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == "workspace" {
			workspaceContainer = &pod.Spec.Containers[i]
			break
		}
	}
	assert.NotNil(t, workspaceContainer, "workspace container not found")

	// Verify nix-store volume mount is at /nix (single mount, not three subPath mounts)
	var nixStoreMount *corev1.VolumeMount
	for i := range workspaceContainer.VolumeMounts {
		if workspaceContainer.VolumeMounts[i].Name == "nix-store" {
			nixStoreMount = &workspaceContainer.VolumeMounts[i]
			break
		}
	}
	assert.NotNil(t, nixStoreMount, "nix-store volume mount not found")
	assert.Equal(t, "/nix", nixStoreMount.MountPath, "nix-store should be mounted at /nix")
	assert.Empty(t, nixStoreMount.SubPath, "nix-store mount should not use subPath")

	// Verify there are no other nix-store mounts with subPaths
	nixStoreMountCount := 0
	for i := range workspaceContainer.VolumeMounts {
		if workspaceContainer.VolumeMounts[i].Name == "nix-store" {
			nixStoreMountCount++
		}
	}
	assert.Equal(t, 1, nixStoreMountCount, "should only have one nix-store mount")

	// Verify nix-store volume is defined
	var nixStoreVolume *corev1.Volume
	for i := range pod.Spec.Volumes {
		if pod.Spec.Volumes[i].Name == "nix-store" {
			nixStoreVolume = &pod.Spec.Volumes[i]
			break
		}
	}
	assert.NotNil(t, nixStoreVolume, "nix-store volume not found")
	assert.NotNil(t, nixStoreVolume.HostPath, "nix-store should be a hostPath volume")
	assert.Equal(t, "/var/lib/kloudlite/nix-store", nixStoreVolume.HostPath.Path, "nix-store hostPath should be /var/lib/kloudlite/nix-store")
}

func TestWorkspaceReconciler_CreateWorkspacePod_KloudliteBinMount(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			Packages:    []workspacev1.PackageSpec{},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacev1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	// Create workspace pod
	pod, err := reconciler.createWorkspacePod(workspace)
	assert.NoError(t, err)
	assert.NotNil(t, pod)

	// Find the workspace container
	var workspaceContainer *corev1.Container
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == "workspace" {
			workspaceContainer = &pod.Spec.Containers[i]
			break
		}
	}
	assert.NotNil(t, workspaceContainer, "workspace container not found")

	// Verify kloudlite-bin volume mount is at /kloudlite/bin (NOT /usr/local/bin/kl with SubPath)
	var klBinMount *corev1.VolumeMount
	for i := range workspaceContainer.VolumeMounts {
		if workspaceContainer.VolumeMounts[i].Name == "kloudlite-bin" {
			klBinMount = &workspaceContainer.VolumeMounts[i]
			break
		}
	}
	assert.NotNil(t, klBinMount, "kloudlite-bin volume mount not found")
	assert.Equal(t, "/kloudlite/bin", klBinMount.MountPath, "kloudlite-bin should be mounted at /kloudlite/bin")
	assert.Empty(t, klBinMount.SubPath, "kloudlite-bin mount should not use subPath")
	assert.True(t, klBinMount.ReadOnly, "kloudlite-bin should be read-only")

	// Verify there's only one kloudlite-bin mount
	klBinMountCount := 0
	for i := range workspaceContainer.VolumeMounts {
		if workspaceContainer.VolumeMounts[i].Name == "kloudlite-bin" {
			klBinMountCount++
		}
	}
	assert.Equal(t, 1, klBinMountCount, "should only have one kloudlite-bin mount")

	// Verify kloudlite-bin volume is defined as HostPath
	var klBinVolume *corev1.Volume
	for i := range pod.Spec.Volumes {
		if pod.Spec.Volumes[i].Name == "kloudlite-bin" {
			klBinVolume = &pod.Spec.Volumes[i]
			break
		}
	}
	assert.NotNil(t, klBinVolume, "kloudlite-bin volume not found")
	assert.NotNil(t, klBinVolume.HostPath, "kloudlite-bin should be a hostPath volume")
	assert.Equal(t, "/kloudlite/bin", klBinVolume.HostPath.Path, "kloudlite-bin hostPath should be /kloudlite/bin")

	// Verify PATH environment variable includes /kloudlite/bin
	var pathEnv *corev1.EnvVar
	for i := range workspaceContainer.Env {
		if workspaceContainer.Env[i].Name == "PATH" {
			pathEnv = &workspaceContainer.Env[i]
			break
		}
	}
	assert.NotNil(t, pathEnv, "PATH environment variable not found")
	assert.Contains(t, pathEnv.Value, "/kloudlite/bin", "PATH should include /kloudlite/bin")
	assert.True(t,
		len(pathEnv.Value) > len("/kloudlite/bin"),
		"PATH should contain more than just /kloudlite/bin",
	)

	// Verify /kloudlite/bin is at the start of PATH (highest priority)
	assert.True(t,
		len(pathEnv.Value) >= len("/kloudlite/bin") && pathEnv.Value[:len("/kloudlite/bin")] == "/kloudlite/bin",
		"PATH should start with /kloudlite/bin for highest priority",
	)
}

func TestWorkspaceReconciler_CreateWorkspacePod_PathInEnvironmentFile(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			Packages:    []workspacev1.PackageSpec{},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacev1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	// Create workspace pod
	pod, err := reconciler.createWorkspacePod(workspace)
	assert.NoError(t, err)
	assert.NotNil(t, pod)

	// Find the init-workspace-dir init container
	var initContainer *corev1.Container
	for i := range pod.Spec.InitContainers {
		if pod.Spec.InitContainers[i].Name == "init-workspace-dir" {
			initContainer = &pod.Spec.InitContainers[i]
			break
		}
	}
	assert.NotNil(t, initContainer, "init-workspace-dir container not found")

	// Verify the init container command includes PATH configuration in /etc/environment
	commandStr := ""
	if len(initContainer.Command) > 0 {
		commandStr = initContainer.Command[len(initContainer.Command)-1]
	}

	assert.Contains(t, commandStr, "/etc/environment", "init container should create /etc/environment")
	assert.Contains(t, commandStr, "PATH=/kloudlite/bin", "init container should set PATH in /etc/environment with /kloudlite/bin")
	assert.Contains(t, commandStr, "/nix/profiles/per-user/root/workspace-", "PATH should include nix profiles path")
}

func TestWorkspaceReconciler_CreateWorkspacePod_CustomResourceQuota(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			ResourceQuota: &workspacev1.ResourceQuota{
				CPU:    "2",
				Memory: "4Gi",
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	pod, err := reconciler.createWorkspacePod(workspace)
	assert.NoError(t, err)
	assert.NotNil(t, pod)

	// Find workspace container
	var workspaceContainer *corev1.Container
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == "workspace" {
			workspaceContainer = &pod.Spec.Containers[i]
			break
		}
	}
	assert.NotNil(t, workspaceContainer)

	// Verify custom resource limits applied
	assert.Equal(t, "2", workspaceContainer.Resources.Limits.Cpu().String())
	assert.Equal(t, "4Gi", workspaceContainer.Resources.Limits.Memory().String())
}

func TestWorkspaceReconciler_CreateWorkspacePod_CustomEnvironmentVariables(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			Settings: &workspacev1.WorkspaceSettings{
				EnvironmentVariables: map[string]string{
					"CUSTOM_VAR": "custom-value",
					"API_KEY":    "secret-key",
				},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	pod, err := reconciler.createWorkspacePod(workspace)
	assert.NoError(t, err)
	assert.NotNil(t, pod)

	// Find workspace container
	var workspaceContainer *corev1.Container
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == "workspace" {
			workspaceContainer = &pod.Spec.Containers[i]
			break
		}
	}
	assert.NotNil(t, workspaceContainer)

	// Verify custom environment variables are set
	envVars := make(map[string]string)
	for _, env := range workspaceContainer.Env {
		envVars[env.Name] = env.Value
	}

	assert.Equal(t, "custom-value", envVars["CUSTOM_VAR"])
	assert.Equal(t, "secret-key", envVars["API_KEY"])
}

func TestWorkspaceReconciler_CreateWorkspacePod_StartupScript(t *testing.T) {
	scheme := testutil.NewTestScheme()

	startupScript := "#!/bin/bash\necho 'Starting workspace'"
	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			Settings: &workspacev1.WorkspaceSettings{
				StartupScript: startupScript,
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	pod, err := reconciler.createWorkspacePod(workspace)
	assert.NoError(t, err)
	assert.NotNil(t, pod)

	// Find workspace container
	var workspaceContainer *corev1.Container
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == "workspace" {
			workspaceContainer = &pod.Spec.Containers[i]
			break
		}
	}
	assert.NotNil(t, workspaceContainer)

	// Verify startup script environment variable is set
	var startupScriptEnv *corev1.EnvVar
	for i := range workspaceContainer.Env {
		if workspaceContainer.Env[i].Name == "STARTUP_SCRIPT" {
			startupScriptEnv = &workspaceContainer.Env[i]
			break
		}
	}

	assert.NotNil(t, startupScriptEnv)
	assert.Equal(t, startupScript, startupScriptEnv.Value)
}

func TestWorkspaceReconciler_CreateWorkspacePod_SSHHostKeys(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	pod, err := reconciler.createWorkspacePod(workspace)
	assert.NoError(t, err)
	assert.NotNil(t, pod)

	// Find workspace container
	var workspaceContainer *corev1.Container
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == "workspace" {
			workspaceContainer = &pod.Spec.Containers[i]
			break
		}
	}
	assert.NotNil(t, workspaceContainer)

	// Verify SSH host key mounts
	sshKeyMounts := []string{}
	for _, mount := range workspaceContainer.VolumeMounts {
		if mount.Name == "ssh-host-keys" {
			sshKeyMounts = append(sshKeyMounts, mount.MountPath)
		}
	}

	// Should have RSA keys mounted
	assert.Contains(t, sshKeyMounts, "/etc/ssh/ssh_host_rsa_key")
	assert.Contains(t, sshKeyMounts, "/etc/ssh/ssh_host_rsa_key.pub")
}

func TestWorkspaceReconciler_CreateWorkspacePod_DNSConfiguration(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacev1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacev1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	pod, err := reconciler.createWorkspacePod(workspace)
	assert.NoError(t, err)
	assert.NotNil(t, pod)

	// Verify DNS policy is set to None (manual management)
	assert.Equal(t, corev1.DNSNone, pod.Spec.DNSPolicy)

	// Verify minimal DNS config is present
	assert.NotNil(t, pod.Spec.DNSConfig)
	assert.Contains(t, pod.Spec.DNSConfig.Nameservers, "10.43.0.10")
}
