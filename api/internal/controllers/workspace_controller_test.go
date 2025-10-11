package controllers

import (
	"context"
	"testing"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/packages/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestWorkspaceReconciler_Reconcile_NotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacesv1.Workspace{}).
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

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Packages:    []workspacesv1.PackageSpec{},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacesv1.Workspace{}).
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
	updatedWorkspace := &workspacesv1.Workspace{}
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
			State: "Ready",
		},
	}

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-workspace",
			Namespace:  "test-namespace",
			Finalizers: []string{workspaceFinalizer},
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			Packages: []workspacesv1.PackageSpec{
				{Name: "git"},
				{Name: "curl"},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, workMachine).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacesv1.Workspace{}).
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

	// Multiple reconciles may be needed
	for i := 0; i < 3; i++ {
		_, _ = reconciler.Reconcile(context.Background(), req)
	}

	// Verify PackageRequest was created
	pkgReq := &packagesv1.PackageRequest{}
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

func TestWorkspaceReconciler_Reconcile_UpdatePackages(t *testing.T) {
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
			State: "Ready",
		},
	}

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-workspace",
			Namespace:  "test-namespace",
			Finalizers: []string{workspaceFinalizer},
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			Packages: []workspacesv1.PackageSpec{
				{Name: "git"},
				{Name: "vim"},
			},
		},
	}

	// Existing PackageRequest with different packages
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace-packages",
			Namespace: "test-namespace",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "workspaces.kloudlite.io/v1",
					Kind:       "Workspace",
					Name:       "test-workspace",
					UID:        workspace.UID,
				},
			},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "workspace-test-workspace-packages",
			Packages: []workspacesv1.PackageSpec{
				{Name: "git"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase:   "Ready",
			Message: "Packages installed",
			InstalledPackages: []workspacesv1.InstalledPackage{
				{Name: "git", BinPath: "/nix/var/nix/profiles/per-user/root/workspace-test-workspace-packages/bin"},
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, workMachine, pkgReq).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacesv1.Workspace{}).
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

	// Multiple reconciles may be needed
	for i := 0; i < 3; i++ {
		_, _ = reconciler.Reconcile(context.Background(), req)
	}

	// Verify PackageRequest was updated
	updatedPkgReq := &packagesv1.PackageRequest{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-workspace-packages",
		Namespace: "test-namespace",
	}, updatedPkgReq)
	assert.NoError(t, err)
	assert.Len(t, updatedPkgReq.Spec.Packages, 2)

	// Verify the packages were updated (the status remains as-is, the package reconciler will handle it)
	assert.Equal(t, "git", updatedPkgReq.Spec.Packages[0].Name)
	assert.Equal(t, "vim", updatedPkgReq.Spec.Packages[1].Name)
}

func TestWorkspaceReconciler_Reconcile_NoPackagesSkipsPackageRequest(t *testing.T) {
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
			State: "Ready",
		},
	}

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-workspace",
			Namespace:  "test-namespace",
			Finalizers: []string{workspaceFinalizer},
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "active",
			Packages:    []workspacesv1.PackageSpec{}, // No packages
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, workMachine).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacesv1.Workspace{}).
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

	// Multiple reconciles may be needed
	for i := 0; i < 3; i++ {
		_, _ = reconciler.Reconcile(context.Background(), req)
	}

	// Verify no PackageRequest was created (since there are no packages)
	pkgReq := &packagesv1.PackageRequest{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-workspace-packages",
		Namespace: "test-namespace",
	}, pkgReq)
	assert.Error(t, err, "PackageRequest should not exist when there are no packages")
}

func TestWorkspaceReconciler_Reconcile_SuspendedWorkspace(t *testing.T) {
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
			State: "Ready",
		},
	}

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-workspace",
			Namespace:  "test-namespace",
			Finalizers: []string{workspaceFinalizer},
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
			Owner:       "test@example.com",
			Status:      "suspended", // Suspended status
			Packages:    []workspacesv1.PackageSpec{},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, workMachine).
		WithStatusSubresource(&packagesv1.PackageRequest{}, &workspacesv1.Workspace{}).
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

	// Reconcile suspended workspace
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify workspace phase is stopped (suspended workspaces show as Stopped)
	updatedWorkspace := &workspacesv1.Workspace{}
	err = k8sClient.Get(context.Background(), req.NamespacedName, updatedWorkspace)
	assert.NoError(t, err)
	assert.Equal(t, "Stopped", updatedWorkspace.Status.Phase)
}

// Auto-Suspension Tests

func TestHasActiveConnections_PodNotFound(t *testing.T) {
	scheme := testutil.NewTestScheme()
	k8sClient := testutil.NewFakeClient(scheme).Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	workspace := &workspacesv1.Workspace{
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

	workspace := &workspacesv1.Workspace{
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

	workspace := &workspacesv1.Workspace{
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

	workspace := &workspacesv1.Workspace{
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

	workspace := &workspacesv1.Workspace{
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

	workspace := &workspacesv1.Workspace{
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

	workspace := &workspacesv1.Workspace{
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

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			Status: "active",
			Settings: &workspacesv1.WorkspaceSettings{
				AutoStop: false, // Auto-stop not enabled
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).
		WithStatusSubresource(&workspacesv1.Workspace{}).
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
	updatedWorkspace := &workspacesv1.Workspace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      workspace.Name,
		Namespace: workspace.Namespace,
	}, updatedWorkspace)
	assert.NoError(t, err)
	assert.Equal(t, "active", updatedWorkspace.Spec.Status)
}

func TestCheckAndSuspendIdleWorkspace_WorkspaceNotActive(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			Status: "suspended", // Already suspended
			Settings: &workspacesv1.WorkspaceSettings{
				AutoStop: true,
			},
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace).
		WithStatusSubresource(&workspacesv1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.checkAndSuspendIdleWorkspace(context.Background(), workspace, logger)
	assert.NoError(t, err)
}

func TestCheckAndSuspendIdleWorkspace_WithActiveConnections(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			Status: "active",
			Settings: &workspacesv1.WorkspaceSettings{
				AutoStop:    true,
				IdleTimeout: 30,
			},
		},
	}

	// Pod with active connections (not running = considered active)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-test-workspace",
			Namespace: "test-namespace",
		},
		Status: corev1.PodStatus{
			PodIP: "10.0.0.1",
			Phase: corev1.PodPending, // Not running = active
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, workspace, pod).
		WithStatusSubresource(&workspacesv1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.checkAndSuspendIdleWorkspace(context.Background(), workspace, logger)
	assert.NoError(t, err)

	// Verify workspace was NOT suspended (has active connections)
	updatedWorkspace := &workspacesv1.Workspace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      workspace.Name,
		Namespace: workspace.Namespace,
	}, updatedWorkspace)
	assert.NoError(t, err)
	assert.Equal(t, "active", updatedWorkspace.Spec.Status)
	assert.NotNil(t, updatedWorkspace.Status.LastActivityTime)
}

func TestCheckAndSuspendIdleWorkspace_IdleButNoTimeout(t *testing.T) {
	scheme := testutil.NewTestScheme()

	// Set LastActivityTime to 20 minutes ago (less than 30 min timeout)
	lastActivityTime := metav1.NewTime(time.Now().Add(-20 * time.Minute))

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			Status: "active",
			Settings: &workspacesv1.WorkspaceSettings{
				AutoStop:    true,
				IdleTimeout: 30, // 30 minutes
			},
		},
		Status: workspacesv1.WorkspaceStatus{
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
		WithStatusSubresource(&workspacesv1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.checkAndSuspendIdleWorkspace(context.Background(), workspace, logger)
	assert.NoError(t, err)

	// Verify workspace was NOT suspended (idle time not exceeded)
	updatedWorkspace := &workspacesv1.Workspace{}
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

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			Status: "active",
			Settings: &workspacesv1.WorkspaceSettings{
				AutoStop:    true,
				IdleTimeout: 30, // 30 minutes
			},
		},
		Status: workspacesv1.WorkspaceStatus{
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
		WithStatusSubresource(&workspacesv1.Workspace{}).
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
	updatedWorkspace := &workspacesv1.Workspace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      workspace.Name,
		Namespace: workspace.Namespace,
	}, updatedWorkspace)
	assert.NoError(t, err)
	assert.Equal(t, "suspended", updatedWorkspace.Spec.Status)
}

func TestCheckAndSuspendIdleWorkspace_NoLastActivityTime(t *testing.T) {
	scheme := testutil.NewTestScheme()

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			Status: "active",
			Settings: &workspacesv1.WorkspaceSettings{
				AutoStop:    true,
				IdleTimeout: 30,
			},
		},
		Status: workspacesv1.WorkspaceStatus{
			LastActivityTime: nil, // No activity time set yet
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
		WithStatusSubresource(&workspacesv1.Workspace{}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &WorkspaceReconciler{
		Client: k8sClient,
		Scheme: scheme,
		Logger: logger,
	}

	err := reconciler.checkAndSuspendIdleWorkspace(context.Background(), workspace, logger)
	assert.NoError(t, err)

	// Verify LastActivityTime was initialized
	updatedWorkspace := &workspacesv1.Workspace{}
	err = k8sClient.Get(context.Background(), types.NamespacedName{
		Name:      workspace.Name,
		Namespace: workspace.Namespace,
	}, updatedWorkspace)
	assert.NoError(t, err)
	assert.NotNil(t, updatedWorkspace.Status.LastActivityTime)
	assert.Equal(t, "active", updatedWorkspace.Spec.Status) // Not suspended yet
}
