package controllers

import (
	"context"
	"testing"

	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
	packagesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/packages/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/workspaces/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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
			ServerType: "ttyd",
			Packages:   []workspacesv1.PackageSpec{},
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
			ServerType: "ttyd",
			Status:     "active",
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
			ServerType: "ttyd",
			Status:     "active",
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

	// Status should be reset to Pending
	assert.Equal(t, "Pending", updatedPkgReq.Status.Phase)
	assert.Equal(t, "Package list updated, waiting for installation", updatedPkgReq.Status.Message)
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
			ServerType: "ttyd",
			Status:     "active",
			Packages:   []workspacesv1.PackageSpec{}, // No packages
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
			ServerType: "ttyd",
			Status:     "suspended", // Suspended status
			Packages:   []workspacesv1.PackageSpec{},
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
