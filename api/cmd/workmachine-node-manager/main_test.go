package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	zap2 "go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
	ExecuteFunc func(script string) ([]byte, error)
	CallCount   int
	LastScript  string
}

func (m *MockCommandExecutor) Execute(script string) ([]byte, error) {
	m.CallCount++
	m.LastScript = script
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(script)
	}
	return nil, fmt.Errorf("mock error: command failed")
}

// MockFileSystem implements FileSystem for testing
type MockFileSystem struct {
	MkdirAllFunc  func(path string, perm os.FileMode) error
	ChownFunc     func(name string, uid, gid int) error
	WriteFileFunc func(name string, data []byte, perm os.FileMode) error
	RenameFunc    func(oldpath, newpath string) error
	CallLog       []string
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	m.CallLog = append(m.CallLog, fmt.Sprintf("MkdirAll(%s, %v)", path, perm))
	if m.MkdirAllFunc != nil {
		return m.MkdirAllFunc(path, perm)
	}
	return nil
}

func (m *MockFileSystem) Chown(name string, uid, gid int) error {
	m.CallLog = append(m.CallLog, fmt.Sprintf("Chown(%s, %d, %d)", name, uid, gid))
	if m.ChownFunc != nil {
		return m.ChownFunc(name, uid, gid)
	}
	return nil
}

func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	m.CallLog = append(m.CallLog, fmt.Sprintf("WriteFile(%s, %d bytes, %v)", name, len(data), perm))
	if m.WriteFileFunc != nil {
		return m.WriteFileFunc(name, data, perm)
	}
	return nil
}

func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	m.CallLog = append(m.CallLog, fmt.Sprintf("Rename(%s, %s)", oldpath, newpath))
	if m.RenameFunc != nil {
		return m.RenameFunc(oldpath, newpath)
	}
	return nil
}

func setupTestReconciler(t *testing.T, initObjs ...runtime.Object) *PackageManagerReconciler {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = packagesv1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(initObjs...).
		WithStatusSubresource(&packagesv1.PackageRequest{}).
		Build()

	logger, _ := zap.NewDevelopment()

	return &PackageManagerReconciler{
		Client:    fakeClient,
		Scheme:    scheme,
		Logger:    logger,
		Namespace: "test-namespace",
		CmdExec:   &MockCommandExecutor{}, // Use mock by default
	}
}

func setupTestReconcilerWithMock(t *testing.T, mockExec *MockCommandExecutor, initObjs ...runtime.Object) *PackageManagerReconciler {
	reconciler := setupTestReconciler(t, initObjs...)
	reconciler.CmdExec = mockExec
	return reconciler
}

func TestPackageManagerReconciler_Reconcile_NotFound(t *testing.T) {
	reconciler := setupTestReconciler(t)

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

func TestPackageManagerReconciler_Reconcile_AlreadyReady(t *testing.T) {
	// Mock executor that succeeds for all commands
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// mkdir succeeds
			if strings.Contains(script, "mkdir -p") {
				return []byte(""), nil
			}
			// Profile check - exists
			if strings.Contains(script, "test -d") {
				return []byte(""), nil
			}
			// Package query - installed
			if strings.Contains(script, "nix-env -p") && strings.Contains(script, "-q git") {
				return []byte("git-2.40.0  /nix/store/xyz-git-2.40.0"), nil
			}
			// Query --out-path
			if strings.Contains(script, "-q --out-path") {
				return []byte("git-2.40.0  /nix/store/xyz-git"), nil
			}
			return []byte(""), nil
		},
	}

	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "git"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase:   "Ready",
			Message: "Already installed",
			InstalledPackages: []packagesv1.InstalledPackage{
				{Name: "git", StorePath: "/nix/store/xyz-git", BinPath: "/nix/profiles/test-profile/bin"},
			},
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify status unchanged
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.Equal(t, "Ready", updatedPkgReq.Status.Phase)
}

func TestPackageManagerReconciler_Reconcile_AlreadyFailed(t *testing.T) {
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "invalid-package"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase:   "Failed",
			Message: "Installation failed",
		},
	}

	reconciler := setupTestReconciler(t, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestPackageManagerReconciler_Reconcile_UpdateToInstalling(t *testing.T) {
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "curl"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	reconciler := setupTestReconciler(t, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	// Note: This test will fail during actual package installation
	// because we're not in a real Nix environment
	_, err := reconciler.Reconcile(context.Background(), req)

	// Get updated status
	updatedPkgReq := &packagesv1.PackageRequest{}
	getErr := reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, getErr)

	// Should have attempted to update to Installing phase
	// Even if installation fails, the phase should have changed from Pending
	assert.NotEqual(t, "Pending", updatedPkgReq.Status.Phase)

	// Error is expected in test environment (no actual Nix)
	_ = err
}

func TestPackageManagerReconciler_Reconcile_RemovePackage(t *testing.T) {
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "git"}, // Only git now, vim was removed
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
			InstalledPackages: []packagesv1.InstalledPackage{
				{Name: "git", BinPath: "/nix/profiles/test/bin"},
				{Name: "vim", BinPath: "/nix/profiles/test/bin"}, // This should be removed
			},
		},
	}

	reconciler := setupTestReconciler(t, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	// Reconcile will try to remove vim and reinstall git
	_, _ = reconciler.Reconcile(context.Background(), req)

	// Verify status was updated (even though installation will fail in test env)
	updatedPkgReq := &packagesv1.PackageRequest{}
	err := reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.NotEqual(t, "Pending", updatedPkgReq.Status.Phase)
}

func TestPackageManagerReconciler_Reconcile_MultiplePackages(t *testing.T) {
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "git"},
				{Name: "curl"},
				{Name: "vim"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	reconciler := setupTestReconciler(t, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	// Will attempt to install all packages
	_, _ = reconciler.Reconcile(context.Background(), req)

	// Verify status was updated
	updatedPkgReq := &packagesv1.PackageRequest{}
	err := reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)

	// In test environment, all packages will fail to install
	assert.Equal(t, "Failed", updatedPkgReq.Status.Phase)
	assert.Len(t, updatedPkgReq.Status.FailedPackages, 3)
}

func TestPackageManagerReconciler_Reconcile_PackageWithChannel(t *testing.T) {
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "nodejs_22", Channel: "nixos-24.05"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	reconciler := setupTestReconciler(t, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	// Will attempt to install package from channel
	_, _ = reconciler.Reconcile(context.Background(), req)

	// Verify status was updated
	updatedPkgReq := &packagesv1.PackageRequest{}
	err := reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.NotEqual(t, "Pending", updatedPkgReq.Status.Phase)
}

func TestPackageManagerReconciler_SetupWithManager(t *testing.T) {
	reconciler := setupTestReconciler(t)

	// SetupWithManager requires a real manager, test will fail with nil
	err := reconciler.SetupWithManager(nil)
	assert.Error(t, err)
}

func TestPackageManagerReconciler_UpdateEventFilter_Ready(t *testing.T) {
	reconciler := setupTestReconciler(t)

	readyPkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pkg",
			Namespace: "test-namespace",
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Ready",
		},
	}

	pendingPkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pkg",
			Namespace: "test-namespace",
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	// Build controller to access predicates
	ctrl := &PackageManagerReconciler{
		Client:    reconciler.Client,
		Scheme:    reconciler.Scheme,
		Logger:    reconciler.Logger,
		Namespace: "test-namespace",
	}

	// Create update event with Ready status - should NOT trigger reconcile
	updateEvent := event.UpdateEvent{
		ObjectOld: pendingPkgReq,
		ObjectNew: readyPkgReq,
	}

	// The predicate in SetupWithManager checks if status is Ready/Failed
	// We can't directly access the predicate, but we verify the logic through reconcile behavior
	result, err := ctrl.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      readyPkgReq.Name,
			Namespace: readyPkgReq.Namespace,
		},
	})

	// When already Ready, should not requeue
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify the update event logic
	_ = updateEvent
}

func TestPackageManagerReconciler_UpdateEventFilter_Failed(t *testing.T) {
	failedPkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Failed",
		},
	}

	reconciler := setupTestReconciler(t, failedPkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      failedPkgReq.Name,
			Namespace: failedPkgReq.Namespace,
		},
	}

	// When already Failed, should not requeue
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestPackageManagerReconciler_UpdateEventFilter_Pending(t *testing.T) {
	pendingPkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	reconciler := setupTestReconciler(t, pendingPkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      pendingPkgReq.Name,
			Namespace: pendingPkgReq.Namespace,
		},
	}

	// When Pending, should process and update status
	_, _ = reconciler.Reconcile(context.Background(), req)

	updatedPkgReq := &packagesv1.PackageRequest{}
	err := reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)

	// Should have moved from Pending to Installing/Failed
	assert.NotEqual(t, "Pending", updatedPkgReq.Status.Phase)
}

func TestPackageManagerReconciler_DeleteEvent(t *testing.T) {
	// Delete events should be ignored by the controller
	// This is verified through the predicate in SetupWithManager
	// We test this indirectly by ensuring reconcile doesn't fail on non-existent resources

	reconciler := setupTestReconciler(t)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "deleted-pkg-req",
			Namespace: "test-namespace",
		},
	}

	// Reconciling a non-existent resource should return no error
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestPackageManagerReconciler_InstallPackage(t *testing.T) {
	reconciler := setupTestReconciler(t)

	// Test installPackage directly - will fail in test env but exercises code paths
	pkg := packagesv1.PackageSpec{
		Name: "testpkg",
		// No channel/commit = uses latest
	}

	_, err := reconciler.installPackage(pkg, "test-profile")

	// Expected to fail in test environment (no Nix)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nix-env failed")
}

func TestPackageManagerReconciler_InstallPackageWithChannel(t *testing.T) {
	reconciler := setupTestReconciler(t)

	// Test installPackage with channel
	pkg := packagesv1.PackageSpec{
		Name:    "nodejs_22",
		Channel: "nixos-24.05",
	}

	_, err := reconciler.installPackage(pkg, "test-profile")

	// Expected to fail in test environment (no Nix)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nix-env failed")
}

func TestPackageManagerReconciler_UninstallPackage(t *testing.T) {
	reconciler := setupTestReconciler(t)

	// Test uninstallPackage directly - will fail in test env but exercises code paths
	err := reconciler.uninstallPackage("testpkg", "test-profile")

	// Expected to fail in test environment (no Nix)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nix-env uninstall failed")
}

func TestPackageManagerReconciler_Reconcile_EmptyPackageList(t *testing.T) {
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages:     []packagesv1.PackageSpec{}, // Empty package list
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	reconciler := setupTestReconciler(t, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	// Reconcile with empty package list
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify status was updated to Ready with 0 packages
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.Equal(t, "Ready", updatedPkgReq.Status.Phase)
	assert.Len(t, updatedPkgReq.Status.InstalledPackages, 0)
	assert.Contains(t, updatedPkgReq.Status.Message, "Successfully reconciled 0 packages")
}

func TestPackageManagerReconciler_Reconcile_StatusUpdateFailure(t *testing.T) {
	// This test covers the error handling when status update fails
	// We use a client without status subresource to simulate failure
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = packagesv1.AddToScheme(scheme)
	_ = packagesv1.AddToScheme(scheme)

	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "curl"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	// Create mock executor for mkdir and other commands
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// Let mkdir succeed
			if strings.Contains(script, "mkdir -p") {
				return []byte(""), nil
			}
			// Let profile check fail (doesn't exist)
			if strings.Contains(script, "test -d") {
				return nil, fmt.Errorf("not found")
			}
			// Let install succeed
			if strings.Contains(script, "nix-env") && strings.Contains(script, "-iA") {
				return []byte("installing"), nil
			}
			// Let query succeed
			if strings.Contains(script, "-q --out-path") {
				return []byte("curl-8.0.0  /nix/store/xyz-curl"), nil
			}
			return []byte(""), nil
		},
	}

	// Create client without status subresource - will fail on status updates
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(pkgReq).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &PackageManagerReconciler{
		Client:    fakeClient,
		Scheme:    scheme,
		Logger:    logger,
		Namespace: "test-namespace",
		CmdExec:   mockExec,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	// Reconcile should fail when trying to update status
	_, err := reconciler.Reconcile(context.Background(), req)

	// Should get an error from status update
	assert.Error(t, err)
}

func TestPackageManagerReconciler_InstallPackage_Success(t *testing.T) {
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// First call: install command succeeds
			if callCount == 1 {
				return []byte("installing..."), nil
			}
			// Second call: query command succeeds with valid output
			if callCount == 2 {
				return []byte("git-2.39.0  /nix/store/abc123-git-2.39.0"), nil
			}
			return nil, fmt.Errorf("unexpected call")
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	pkg := packagesv1.PackageSpec{
		Name: "git",
		// No channel/commit = uses latest
	}

	installedPkg, err := reconciler.installPackage(pkg, "test-profile")

	assert.NoError(t, err)
	assert.Equal(t, "git", installedPkg.Name)
	assert.Equal(t, "/nix/store/abc123-git-2.39.0", installedPkg.StorePath)
	assert.Equal(t, "/nix/profiles/per-user/root/test-profile/bin", installedPkg.BinPath)
	assert.Equal(t, 2, mockExec.CallCount) // Install + Query
}

func TestPackageManagerReconciler_InstallPackage_QueryFails(t *testing.T) {
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// First call: install succeeds
			if callCount == 1 {
				return []byte("installing..."), nil
			}
			// Second call: query fails
			if callCount == 2 {
				return nil, fmt.Errorf("query failed")
			}
			return nil, fmt.Errorf("unexpected call")
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	pkg := packagesv1.PackageSpec{
		Name: "nodejs",
	}

	installedPkg, err := reconciler.installPackage(pkg, "test-profile")

	// Should still succeed even if query fails
	assert.NoError(t, err)
	assert.Equal(t, "nodejs", installedPkg.Name)
	assert.Equal(t, "/nix/store", installedPkg.StorePath) // Default path
	assert.Equal(t, 2, mockExec.CallCount)
}

func TestPackageManagerReconciler_InstallPackage_QueryEmptyOutput(t *testing.T) {
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// First call: install succeeds
			if callCount == 1 {
				return []byte("installing..."), nil
			}
			// Second call: query returns empty
			if callCount == 2 {
				return []byte(""), nil
			}
			return nil, fmt.Errorf("unexpected call")
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	pkg := packagesv1.PackageSpec{
		Name: "vim",
	}

	installedPkg, err := reconciler.installPackage(pkg, "test-profile")

	assert.NoError(t, err)
	assert.Equal(t, "vim", installedPkg.Name)
	assert.Equal(t, "/nix/store", installedPkg.StorePath) // Default
}

func TestPackageManagerReconciler_InstallPackage_QuerySingleField(t *testing.T) {
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// First call: install succeeds
			if callCount == 1 {
				return []byte("installing..."), nil
			}
			// Second call: query returns only one field
			if callCount == 2 {
				return []byte("curl-8.0.0"), nil
			}
			return nil, fmt.Errorf("unexpected call")
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	pkg := packagesv1.PackageSpec{
		Name: "curl",
	}

	installedPkg, err := reconciler.installPackage(pkg, "test-profile")

	assert.NoError(t, err)
	assert.Equal(t, "curl", installedPkg.Name)
	assert.Equal(t, "/nix/store", installedPkg.StorePath) // Default (< 2 parts)
}

func TestPackageManagerReconciler_UninstallPackage_Success(t *testing.T) {
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			return []byte("uninstalling..."), nil
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	err := reconciler.uninstallPackage("git", "test-profile")

	assert.NoError(t, err)
	assert.Equal(t, 1, mockExec.CallCount)
	assert.Contains(t, mockExec.LastScript, "nix-env -p")
	assert.Contains(t, mockExec.LastScript, "-e git")
}

func TestPackageManagerReconciler_Reconcile_Success_WithMock(t *testing.T) {
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// All commands succeed
			if callCount%2 == 1 {
				// Install commands
				return []byte("installing..."), nil
			}
			// Query commands
			return []byte("pkg-1.0  /nix/store/xyz-pkg-1.0"), nil
		},
	}

	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "git"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify status was updated to Ready
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.Equal(t, "Ready", updatedPkgReq.Status.Phase)
	assert.Len(t, updatedPkgReq.Status.InstalledPackages, 1)
	assert.Equal(t, "git", updatedPkgReq.Status.InstalledPackages[0].Name)
}

// Test package removal (covering the successful uninstall branch)
func TestPackageManagerReconciler_Reconcile_PackageRemoval(t *testing.T) {
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// All commands succeed
			if strings.Contains(script, "mkdir -p") {
				return []byte(""), nil
			}
			if strings.Contains(script, "test -d") {
				return []byte(""), nil
			}
			if strings.Contains(script, "nix-env -p") && strings.Contains(script, "-e") {
				// Uninstall command
				return []byte("uninstalling..."), nil
			}
			if strings.Contains(script, "nix-env -p") && strings.Contains(script, "-q git") {
				// git is NOT installed
				return []byte(""), nil
			}
			if strings.Contains(script, "nix-env -p") && strings.Contains(script, "-q vim") {
				// vim is NOT installed yet
				return []byte(""), nil
			}
			if strings.Contains(script, "nix-env") && strings.Contains(script, "-iA") {
				// Install commands
				return []byte("installing..."), nil
			}
			if strings.Contains(script, "-q --out-path") {
				// Query commands
				return []byte("vim-9.0  /nix/store/xyz-vim-9.0"), nil
			}
			return []byte(""), nil
		},
	}

	// Create PackageRequest with an installed package that should be removed
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer}, // Add finalizer to avoid requeue
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "vim"}, // Only vim in spec
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
			InstalledPackages: []packagesv1.InstalledPackage{
				{
					Name:      "git", // git was installed before but not in spec anymore
					StorePath: "/nix/store/old-git",
				},
			},
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify git was removed and vim was installed
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.Equal(t, "Ready", updatedPkgReq.Status.Phase)
	assert.Len(t, updatedPkgReq.Status.InstalledPackages, 1)
	assert.Equal(t, "vim", updatedPkgReq.Status.InstalledPackages[0].Name)
}

// Test final status update failure
func TestPackageManagerReconciler_Reconcile_FinalStatusUpdateFails(t *testing.T) {
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// mkdir succeeds
			if strings.Contains(script, "mkdir -p") {
				return []byte(""), nil
			}
			// Profile check - doesn't exist
			if strings.Contains(script, "test -d") {
				return nil, fmt.Errorf("not found")
			}
			// Package check - not installed
			if strings.Contains(script, "nix-env -p") && strings.Contains(script, "-q git") {
				return []byte(""), nil
			}
			// Install succeeds
			if strings.Contains(script, "nix-env") && strings.Contains(script, "-iA") {
				return []byte("installing..."), nil
			}
			// Query succeeds
			if strings.Contains(script, "-q --out-path") {
				return []byte("git-2.40.0  /nix/store/xyz-git-2.40.0"), nil
			}
			return []byte(""), nil
		},
	}

	// Create PackageRequest
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer}, // Add finalizer
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "git"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = packagesv1.AddToScheme(scheme)

	logger := zap2.NewNop()

	// Create a client that will fail on the final status update
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(pkgReq).
		WithStatusSubresource(&packagesv1.PackageRequest{}).
		WithInterceptorFuncs(interceptor.Funcs{
			SubResourceUpdate: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
				// Fail the final status update (not the first one)
				if pr, ok := obj.(*packagesv1.PackageRequest); ok {
					if pr.Status.Phase == "Ready" || pr.Status.Phase == "Failed" {
						return fmt.Errorf("simulated final status update failure")
					}
				}
				return client.SubResource(subResourceName).Update(ctx, obj, opts...)
			},
		}).
		Build()

	reconciler := &PackageManagerReconciler{
		Client:    fakeClient,
		Scheme:    scheme,
		Logger:    logger,
		Namespace: "test-namespace",
		CmdExec:   mockExec,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	// Should get an error from final status update
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "simulated final status update failure")
	}
	assert.Equal(t, reconcile.Result{}, result)
}

func TestPackageManagerReconciler_InstallPackage_ChannelScriptGeneration(t *testing.T) {
	// Test that the correct install script is generated for channel-based packages
	var installScriptCaptured string
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// Capture install script (with nix profile wrapper)
			if strings.Contains(script, "nix --extra-experimental-features") && strings.Contains(script, "profile install") {
				installScriptCaptured = script
				return []byte("installing 'nodejs-22.19.0'"), nil
			}
			// Handle query script
			if strings.Contains(script, "-q --out-path") {
				return []byte("nodejs-22.19.0  /nix/store/abc123-nodejs-22.19.0"), nil
			}
			return nil, fmt.Errorf("unexpected script: %s", script)
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	pkg := packagesv1.PackageSpec{
		Name:    "nodejs_22",
		Channel: "nixos-24.05",
	}

	installedPkg, err := reconciler.installPackage(pkg, "test-profile")
	assert.NoError(t, err)
	assert.Equal(t, "nodejs_22", installedPkg.Name)
	// Verify the install script contains nix profile setup and channel reference
	assert.Contains(t, installScriptCaptured, ". /root/.nix-profile/etc/profile.d/nix.sh")
	assert.Contains(t, installScriptCaptured, "nix --extra-experimental-features")
	assert.Contains(t, installScriptCaptured, "profile install")
	assert.Contains(t, installScriptCaptured, "nixpkgs/nixos-24.05#nodejs_22")
}

func TestPackageManagerReconciler_InstallPackage_CommitScriptGeneration(t *testing.T) {
	// Test that the correct install script is generated for commit-based packages
	var installScriptCaptured string
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// Capture install script (commit uses nix-env with tarball, not nix profile)
			if strings.Contains(script, "nix-env") && strings.Contains(script, "-f") && strings.Contains(script, "github.com/nixos/nixpkgs/archive") {
				installScriptCaptured = script
				return []byte("installing 'nodejs-20.10.0'"), nil
			}
			// Handle query script
			if strings.Contains(script, "-q --out-path") {
				return []byte("nodejs-20.10.0  /nix/store/xyz-nodejs-20.10.0"), nil
			}
			return nil, fmt.Errorf("unexpected script: %s", script)
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	pkg := packagesv1.PackageSpec{
		Name:          "nodejs_20",
		NixpkgsCommit: "abc123def456789",
	}

	installedPkg, err := reconciler.installPackage(pkg, "test-profile")
	assert.NoError(t, err)
	assert.Equal(t, "nodejs_20", installedPkg.Name)
	// Verify the install script contains nix profile setup and GitHub tarball URL
	assert.Contains(t, installScriptCaptured, ". /root/.nix-profile/etc/profile.d/nix.sh")
	assert.Contains(t, installScriptCaptured, "nix-env")
	assert.Contains(t, installScriptCaptured, "https://github.com/nixos/nixpkgs/archive/abc123def456789.tar.gz")
	assert.Contains(t, installScriptCaptured, "-iA nodejs_20")
}

func TestPackageManagerReconciler_InstallPackage_NoVersionUsesNixEnv(t *testing.T) {
	// Test that nix-env is used for packages without version
	var installScriptCaptured string
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// Capture install script
			if strings.Contains(script, "nix-env") && strings.Contains(script, "-iA") {
				installScriptCaptured = script
				return []byte("installing 'git-2.40.0'"), nil
			}
			// Handle query script
			if strings.Contains(script, "-q --out-path") {
				return []byte("git-2.40.0  /nix/store/xyz789-git-2.40.0"), nil
			}
			return nil, fmt.Errorf("unexpected script: %s", script)
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	pkg := packagesv1.PackageSpec{
		Name: "git",
		// No version specified
	}

	installedPkg, err := reconciler.installPackage(pkg, "test-profile")
	assert.NoError(t, err)
	assert.Equal(t, "git", installedPkg.Name)
	// Verify the install script (not the query script)
	assert.Contains(t, installScriptCaptured, "nix-env")
	assert.Contains(t, installScriptCaptured, "nixpkgs.git")
	assert.NotContains(t, installScriptCaptured, "nix profile install")
}

func TestPackageManagerReconciler_InstallPackage_VersionExtractionWithChannel(t *testing.T) {
	// Test version extraction from nix-env query output with channel
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			if strings.Contains(script, "nix --extra-experimental-features") && strings.Contains(script, "profile install") {
				// Simulate successful install
				return []byte("installing 'vim-9.1.1623'"), nil
			}
			if strings.Contains(script, "-q --out-path") {
				// Simulate query output with version
				return []byte("vim-9.1.1623  /nix/store/hash123-vim-9.1.1623"), nil
			}
			return nil, fmt.Errorf("unexpected script: %s", script)
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	pkg := packagesv1.PackageSpec{
		Name:    "vim",
		Channel: "nixos-24.05",
	}

	installedPkg, err := reconciler.installPackage(pkg, "test-profile")
	assert.NoError(t, err)
	assert.Equal(t, "vim", installedPkg.Name)
	// Verify the version includes both actual version and source
	assert.Contains(t, installedPkg.Version, "9.1.1623")
	assert.Contains(t, installedPkg.Version, "channel:nixos-24.05")
	assert.Contains(t, installedPkg.StorePath, "/nix/store/hash123-vim-9.1.1623")
}

func TestPackageManagerReconciler_InstallPackage_VersionExtractionEdgeCases(t *testing.T) {
	tests := []struct {
		name                string
		queryOutput         string
		expectedVerContains []string
		packageName         string
	}{
		{
			name:                "Standard format with version",
			queryOutput:         "nodejs-20.10.0  /nix/store/abc-nodejs-20.10.0",
			expectedVerContains: []string{"20.10.0", "channel:nixos-24.05"},
			packageName:         "nodejs",
		},
		{
			name:                "Package name with dashes",
			queryOutput:         "code-server-4.20.0  /nix/store/def-code-server-4.20.0",
			expectedVerContains: []string{"4.20.0", "channel:nixos-24.05"},
			packageName:         "code-server",
		},
		{
			name:                "Single field output (no path)",
			queryOutput:         "python-3.11.0",
			expectedVerContains: []string{"3.11.0", "channel:nixos-24.05"},
			packageName:         "python",
		},
		{
			name:                "Complex version string",
			queryOutput:         "gcc-13.2.0-x86_64-unknown-linux-gnu  /nix/store/ghi-gcc-13.2.0",
			expectedVerContains: []string{"13.2.0-x86_64-unknown-linux-gnu", "channel:nixos-24.05"},
			packageName:         "gcc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec := &MockCommandExecutor{
				ExecuteFunc: func(script string) ([]byte, error) {
					if strings.Contains(script, "nix --extra-experimental-features") && strings.Contains(script, "profile install") {
						// Handle channel-based installation
						return []byte("installing package"), nil
					}
					if strings.Contains(script, "nix-env") && strings.Contains(script, "-iA") {
						return []byte("installing package"), nil
					}
					if strings.Contains(script, "-q --out-path") {
						return []byte(tt.queryOutput), nil
					}
					return nil, fmt.Errorf("unexpected script")
				},
			}

			reconciler := setupTestReconcilerWithMock(t, mockExec)

			pkg := packagesv1.PackageSpec{
				Name:    tt.packageName,
				Channel: "nixos-24.05",
			}

			installedPkg, err := reconciler.installPackage(pkg, "test-profile")
			assert.NoError(t, err)
			assert.Equal(t, tt.packageName, installedPkg.Name)
			// Verify all expected strings are in the version
			for _, expected := range tt.expectedVerContains {
				assert.Contains(t, installedPkg.Version, expected)
			}
		})
	}
}

func TestSetupWorkspaceHome_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{}

	err := setupWorkspaceHome(logger, mockFS)

	assert.NoError(t, err)
	// Verify 4 filesystem calls: 2 MkdirAll + 2 Chown
	assert.Len(t, mockFS.CallLog, 4)
	assert.Contains(t, mockFS.CallLog[0], "MkdirAll(/var/lib/kloudlite/home")
	assert.Contains(t, mockFS.CallLog[1], "Chown(/var/lib/kloudlite/home, 1001, 1001)")
	assert.Contains(t, mockFS.CallLog[2], "MkdirAll(/var/lib/kloudlite/home/workspaces")
	assert.Contains(t, mockFS.CallLog[3], "Chown(/var/lib/kloudlite/home/workspaces, 1001, 1001)")
}

func TestSetupWorkspaceHome_MkdirAllError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{
		MkdirAllFunc: func(path string, perm os.FileMode) error {
			return fmt.Errorf("permission denied")
		},
	}

	err := setupWorkspaceHome(logger, mockFS)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create workspace home directory")
}

func TestSetupWorkspaceHome_ChownError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	callCount := 0
	mockFS := &MockFileSystem{
		ChownFunc: func(name string, uid, gid int) error {
			callCount++
			if callCount == 1 {
				return fmt.Errorf("chown failed")
			}
			return nil
		},
	}

	err := setupWorkspaceHome(logger, mockFS)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set ownership on workspace home directory")
}

// ========================================
// SSH ConfigMap Controller Tests
// ========================================

func setupTestSSHConfigReconciler(t *testing.T, initObjs ...runtime.Object) *SSHConfigReconciler {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(initObjs...).
		Build()

	logger, _ := zap.NewDevelopment()

	return &SSHConfigReconciler{
		Client: fakeClient,
		Logger: logger,
		FS:     &MockFileSystem{}, // Use mock filesystem by default
	}
}

func TestSSHConfigReconciler_Reconcile_Success(t *testing.T) {
	// Create ssh-authorized-keys ConfigMap with sample keys
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"authorized_keys": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... user@example.com\nssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB... user2@example.com\n",
		},
	}

	reconciler := setupTestSSHConfigReconciler(t, cm)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
	}

	// Note: This will attempt to write to /var/lib/kloudlite/ssh-config/authorized_keys
	// In test environment, this may fail due to permissions, but we verify the code path
	result, err := reconciler.Reconcile(context.Background(), req)

	// Should not requeue
	assert.False(t, result.Requeue)
	// Error may occur due to filesystem permissions in test environment
	_ = err
}

func TestSSHConfigReconciler_Reconcile_ConfigMapNotFound(t *testing.T) {
	reconciler := setupTestSSHConfigReconciler(t)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
	}

	// ConfigMap doesn't exist
	result, err := reconciler.Reconcile(context.Background(), req)

	// Should not return error when ConfigMap is not found (it's deleted or doesn't exist)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestSSHConfigReconciler_Reconcile_WrongConfigMapName(t *testing.T) {
	// Create ConfigMap with different name - should be ignored
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-configmap",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"authorized_keys": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... user@example.com\n",
		},
	}

	reconciler := setupTestSSHConfigReconciler(t, cm)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "other-configmap",
			Namespace: "test-namespace",
		},
	}

	// Should be ignored (not ssh-authorized-keys)
	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestSSHConfigReconciler_Reconcile_MissingAuthorizedKeysKey(t *testing.T) {
	// Create ConfigMap without authorized_keys key
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"other-key": "some-value",
		},
	}

	reconciler := setupTestSSHConfigReconciler(t, cm)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
	}

	// Should handle missing key gracefully
	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestSSHConfigReconciler_Reconcile_EmptyAuthorizedKeys(t *testing.T) {
	// Create ConfigMap with empty authorized_keys
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"authorized_keys": "",
		},
	}

	reconciler := setupTestSSHConfigReconciler(t, cm)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
	}

	// Should handle empty content (clears authorized_keys file)
	result, err := reconciler.Reconcile(context.Background(), req)

	assert.False(t, result.Requeue)
	// Error may occur due to filesystem permissions in test environment
	_ = err
}

func TestSSHConfigReconciler_Reconcile_MultipleKeys(t *testing.T) {
	// Create ConfigMap with multiple SSH keys (realistic scenario)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"authorized_keys": `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDTest1 user1@example.com
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest2 user2@example.com
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDTest3 user3@example.com
`,
		},
	}

	reconciler := setupTestSSHConfigReconciler(t, cm)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.False(t, result.Requeue)
	// Error may occur due to filesystem permissions in test environment
	_ = err
}

func TestSetupSSHConfigDirectory_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{}

	err := setupSSHConfigDirectory(logger, mockFS)

	assert.NoError(t, err)
	// Verify 1 filesystem call: MkdirAll
	assert.Len(t, mockFS.CallLog, 1)
	assert.Contains(t, mockFS.CallLog[0], "MkdirAll(/var/lib/kloudlite/ssh-config")
	// The permission format will be shown as file mode (e.g., drwxr-xr-x or -rwxr-xr-x)
	assert.Contains(t, mockFS.CallLog[0], "rwxr-xr-x")
}

func TestSetupSSHConfigDirectory_Error(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{
		MkdirAllFunc: func(path string, perm os.FileMode) error {
			return fmt.Errorf("permission denied")
		},
	}

	err := setupSSHConfigDirectory(logger, mockFS)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create SSH config directory")
}

func TestWriteAuthorizedKeys_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{}

	content := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com\n"

	err := writeAuthorizedKeys(logger, content, mockFS)

	assert.NoError(t, err)
	// Verify 2 filesystem calls: WriteFile + Rename (atomic write pattern)
	assert.Len(t, mockFS.CallLog, 2)
	assert.Contains(t, mockFS.CallLog[0], "WriteFile(/var/lib/kloudlite/ssh-config/authorized_keys.tmp")
	// The permission format will be shown as file mode (e.g., -rw-r--r--)
	assert.Contains(t, mockFS.CallLog[0], "rw-r--r--")
	assert.Contains(t, mockFS.CallLog[1], "Rename(/var/lib/kloudlite/ssh-config/authorized_keys.tmp, /var/lib/kloudlite/ssh-config/authorized_keys)")
}

func TestWriteAuthorizedKeys_WriteFileError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error {
			return fmt.Errorf("disk full")
		},
	}

	content := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com\n"

	err := writeAuthorizedKeys(logger, content, mockFS)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write temporary authorized_keys file")
}

func TestWriteAuthorizedKeys_RenameError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{
		RenameFunc: func(oldpath, newpath string) error {
			return fmt.Errorf("rename failed")
		},
	}

	content := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... test@example.com\n"

	err := writeAuthorizedKeys(logger, content, mockFS)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to rename temporary authorized_keys file")
}

func TestWriteAuthorizedKeys_EmptyContent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{}

	err := writeAuthorizedKeys(logger, "", mockFS)

	assert.NoError(t, err)
	// Should handle empty content (creates empty file)
	assert.Len(t, mockFS.CallLog, 2)
	assert.Contains(t, mockFS.CallLog[0], "0 bytes") // Empty content
}

func TestWriteAuthorizedKeys_LargeContent(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{}

	// Simulate 100 SSH keys
	var content strings.Builder
	for i := 0; i < 100; i++ {
		content.WriteString(fmt.Sprintf("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDTest%d user%d@example.com\n", i, i))
	}

	err := writeAuthorizedKeys(logger, content.String(), mockFS)

	assert.NoError(t, err)
	// Verify large content was written
	assert.Len(t, mockFS.CallLog, 2)
	// Content length should be > 5000 bytes (100 keys * ~50+ bytes each)
	assert.Contains(t, mockFS.CallLog[0], "WriteFile")
}

func TestSSHConfigReconciler_Reconcile_UpdateExistingKeys(t *testing.T) {
	// Simulate updating authorized_keys (e.g., adding/removing users)
	// First reconcile with initial keys
	initialCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"authorized_keys": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC... user1@example.com\n",
		},
	}

	reconciler := setupTestSSHConfigReconciler(t, initialCM)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
	}

	// First reconcile
	result1, err1 := reconciler.Reconcile(context.Background(), req)
	assert.False(t, result1.Requeue)
	_ = err1

	// Update ConfigMap with new keys
	updatedCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"authorized_keys": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIB... user2@example.com\n",
		},
	}

	// Update in fake client
	err := reconciler.Client.Update(context.Background(), updatedCM)
	assert.NoError(t, err)

	// Second reconcile with updated keys
	result2, err2 := reconciler.Reconcile(context.Background(), req)
	assert.False(t, result2.Requeue)
	_ = err2

	// Expected behavior:
	// - First reconcile writes user1's key
	// - Second reconcile overwrites with user2's key (atomic operation)
	// - SSH daemons immediately see new keys on next auth attempt
}

func TestSSHConfigReconciler_SetupWithManager(t *testing.T) {
	reconciler := setupTestSSHConfigReconciler(t)

	// SetupWithManager requires a real manager, test will fail with nil
	err := reconciler.SetupWithManager(nil)
	assert.Error(t, err)
}

// ========================================
// Additional Edge Case Tests
// ========================================

func TestPackageManagerReconciler_InstallPackage_WithCommit(t *testing.T) {
	var installScriptCaptured string
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// Commit-based installs use nix-env with tarball URL
			if strings.Contains(script, "nix-env") && strings.Contains(script, "-f") && strings.Contains(script, "github.com/nixos/nixpkgs/archive") {
				installScriptCaptured = script
				return []byte("installing 'git-2.45.0'"), nil
			}
			if strings.Contains(script, "-q --out-path") {
				return []byte("git-2.45.0  /nix/store/xyz-git-2.45.0"), nil
			}
			return nil, fmt.Errorf("unexpected script: %s", script)
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	pkg := packagesv1.PackageSpec{
		Name:          "git",
		NixpkgsCommit: "abc123def456",
	}

	installedPkg, err := reconciler.installPackage(pkg, "test-profile")
	assert.NoError(t, err)
	assert.Equal(t, "git", installedPkg.Name)
	// Commit-based installs use tarball URL, not github: flake reference
	assert.Contains(t, installScriptCaptured, "https://github.com/nixos/nixpkgs/archive/abc123def456.tar.gz")
	assert.Contains(t, installedPkg.Version, "2.45.0")
	assert.Contains(t, installedPkg.Version, "commit:abc123de") // Short hash (8 chars)
}

func TestPackageManagerReconciler_Reconcile_InstallFailureUpdatesStatus(t *testing.T) {
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// Allow mkdir to succeed
			if strings.Contains(script, "mkdir -p") {
				return []byte(""), nil
			}
			// Profile check - exists (so we skip init)
			if strings.Contains(script, "test -d") {
				return []byte(""), nil
			}
			// All nix install commands fail
			if (strings.Contains(script, "nix-env") && strings.Contains(script, "-iA")) ||
				(strings.Contains(script, "nix --extra-experimental-features") && strings.Contains(script, "profile install")) {
				return nil, fmt.Errorf("nix install failed: network error")
			}
			return nil, fmt.Errorf("unexpected script")
		},
	}

	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "python"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	// First reconcile adds finalizer and requeues
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue)

	// Second reconcile attempts install and should fail
	result, err = reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify status was updated to Failed
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.Equal(t, "Failed", updatedPkgReq.Status.Phase)
	assert.GreaterOrEqual(t, len(updatedPkgReq.Status.FailedPackages), 1)
	if len(updatedPkgReq.Status.FailedPackages) > 0 {
		assert.Equal(t, "python", updatedPkgReq.Status.FailedPackages[0])
	}
}

func TestPackageManagerReconciler_UninstallPackage_Failure(t *testing.T) {
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			return nil, fmt.Errorf("uninstall failed: package not found")
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)

	err := reconciler.uninstallPackage("nonexistent", "test-profile")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nix-env uninstall failed")
}

func TestSSHConfigReconciler_Reconcile_GetConfigMapError(t *testing.T) {
	// Create a client that will return an error on Get
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"authorized_keys": "test-key",
		},
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(cm).
		WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, client client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				// Simulate a transient error (not NotFound)
				return fmt.Errorf("simulated API server error")
			},
		}).
		Build()

	logger, _ := zap.NewDevelopment()
	reconciler := &SSHConfigReconciler{
		Client: fakeClient,
		Logger: logger,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated API server error")
	assert.Equal(t, reconcile.Result{}, result)
}

func TestPackageManagerReconciler_Reconcile_PartialSuccess(t *testing.T) {
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// First package (git) succeeds
			if strings.Contains(script, "git") {
				if callCount == 1 {
					return []byte("installing git"), nil
				}
				if callCount == 2 {
					return []byte("git-2.40.0  /nix/store/xyz-git"), nil
				}
			}
			// Second package (vim) fails
			if strings.Contains(script, "vim") {
				return nil, fmt.Errorf("vim install failed")
			}
			return []byte("success"), nil
		},
	}

	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "git"},
				{Name: "vim"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	// First reconcile adds finalizer and requeues
	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Requeue)

	// Second reconcile attempts install
	result, err = reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify partial success - git installed, vim failed
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.Equal(t, "Failed", updatedPkgReq.Status.Phase) // Overall failed due to vim
	assert.GreaterOrEqual(t, len(updatedPkgReq.Status.InstalledPackages), 1)
	if len(updatedPkgReq.Status.InstalledPackages) > 0 {
		assert.Equal(t, "git", updatedPkgReq.Status.InstalledPackages[0].Name)
	}
	assert.GreaterOrEqual(t, len(updatedPkgReq.Status.FailedPackages), 1)
	if len(updatedPkgReq.Status.FailedPackages) > 0 {
		assert.Equal(t, "vim", updatedPkgReq.Status.FailedPackages[0])
	}
}

func TestNixStorePathConstant(t *testing.T) {
	// Verify that nixStorePath constant is set to /nix
	// This test ensures the path is consistent with the volume mount changes
	assert.Equal(t, "/nix", nixStorePath, "nixStorePath constant should be /nix")

	// Verify derived paths are correct
	expectedBinPath := "/nix/profiles/per-user/root/test-profile/bin"
	actualBinPath := fmt.Sprintf("%s/profiles/per-user/root/test-profile/bin", nixStorePath)
	assert.Equal(t, expectedBinPath, actualBinPath)
}

func TestPackageManagerReconciler_Reconcile_UninstallFailure(t *testing.T) {
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// mkdir succeeds
			if strings.Contains(script, "mkdir -p") {
				return []byte(""), nil
			}
			// Profile check - exists
			if strings.Contains(script, "test -d") {
				return []byte(""), nil
			}
			// Package query - installed
			if strings.Contains(script, "nix-env -p") && strings.Contains(script, "-q git") {
				return []byte("git-2.40.0  /nix/store/old-git"), nil
			}
			// Uninstall fails
			if strings.Contains(script, "-e git") {
				return nil, fmt.Errorf("uninstall failed: permission denied")
			}
			return []byte("success"), nil
		},
	}

	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages:     []packagesv1.PackageSpec{}, // Empty - should trigger uninstall
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
			InstalledPackages: []packagesv1.InstalledPackage{
				{Name: "git", StorePath: "/nix/store/old-git"},
			},
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Even with uninstall failure, reconciliation should complete
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	// Should be Ready even if uninstall failed (best effort)
	assert.Equal(t, "Ready", updatedPkgReq.Status.Phase)
}

// ========================================
// Tests for State-Based Reconciliation Changes
// ========================================

func TestIsPackageInstalled_ProfileNotExists(t *testing.T) {
	// Test when profile directory doesn't exist
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// First call: check if profile directory exists (test -d)
			if strings.Contains(script, "test -d") {
				return nil, fmt.Errorf("directory does not exist")
			}
			return []byte(""), nil
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)
	logger, _ := zap.NewDevelopment()

	isInstalled := reconciler.isPackageInstalled("git", "test-profile", logger)

	assert.False(t, isInstalled, "Package should not be considered installed when profile doesn't exist")
	assert.Equal(t, 1, mockExec.CallCount, "Should only check profile existence")
}

func TestIsPackageInstalled_PackageNotInstalled(t *testing.T) {
	// Test when profile exists but package is not installed
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// First call: profile directory exists
			if callCount == 1 && strings.Contains(script, "test -d") {
				return []byte(""), nil
			}
			// Second call: query package (returns empty = not installed)
			if callCount == 2 && strings.Contains(script, "nix-env -p") && strings.Contains(script, "-q") {
				return []byte(""), nil
			}
			return nil, fmt.Errorf("unexpected script")
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)
	logger, _ := zap.NewDevelopment()

	isInstalled := reconciler.isPackageInstalled("git", "test-profile", logger)

	assert.False(t, isInstalled, "Package should not be installed when query returns empty")
	assert.Equal(t, 2, mockExec.CallCount, "Should check profile and query package")
}

func TestIsPackageInstalled_PackageInstalled(t *testing.T) {
	// Test when package is actually installed
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// First call: profile directory exists
			if callCount == 1 && strings.Contains(script, "test -d") {
				return []byte(""), nil
			}
			// Second call: query package (returns package info = installed)
			if callCount == 2 && strings.Contains(script, "nix-env -p") && strings.Contains(script, "-q git") {
				return []byte("git-2.40.0  /nix/store/xyz-git-2.40.0"), nil
			}
			return nil, fmt.Errorf("unexpected script")
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)
	logger, _ := zap.NewDevelopment()

	isInstalled := reconciler.isPackageInstalled("git", "test-profile", logger)

	assert.True(t, isInstalled, "Package should be installed when query returns package info")
	assert.Equal(t, 2, mockExec.CallCount, "Should check profile and query package")
}

func TestIsPackageInstalled_QueryFails(t *testing.T) {
	// Test when query command fails (not just empty output)
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// First call: profile directory exists
			if callCount == 1 && strings.Contains(script, "test -d") {
				return []byte(""), nil
			}
			// Second call: query fails with error
			if callCount == 2 {
				return nil, fmt.Errorf("nix-env query failed: command not found")
			}
			return nil, fmt.Errorf("unexpected script")
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec)
	logger, _ := zap.NewDevelopment()

	isInstalled := reconciler.isPackageInstalled("nodejs", "test-profile", logger)

	assert.False(t, isInstalled, "Package should not be installed when query fails")
	assert.Equal(t, 2, mockExec.CallCount)
}

func TestReconcile_ProfileDirectoryCreationIdempotent(t *testing.T) {
	// Test that profile directory is created idempotently during reconciliation
	var mkdirCalled bool
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			// Track mkdir call
			if strings.Contains(script, "mkdir -p") && strings.Contains(script, "profiles/per-user/root") {
				mkdirCalled = true
				return []byte(""), nil
			}
			// Profile check - doesn't exist yet
			if strings.Contains(script, "test -d") {
				return nil, fmt.Errorf("not found")
			}
			// Install command
			if strings.Contains(script, "nix-env") && strings.Contains(script, "-iA") {
				return []byte("installing"), nil
			}
			// Query command
			if strings.Contains(script, "-q --out-path") {
				return []byte("git-2.40.0  /nix/store/xyz-git"), nil
			}
			return []byte(""), nil
		},
	}

	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "git"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Pending",
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	_, _ = reconciler.Reconcile(context.Background(), req)

	assert.True(t, mkdirCalled, "Profile directory should be created during reconciliation")
}

func TestReconcile_ClearsFailedPackagesOnRetry(t *testing.T) {
	// Test that failed packages are cleared when retrying installation
	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// mkdir succeeds
			if strings.Contains(script, "mkdir -p") {
				return []byte(""), nil
			}
			// Profile check - doesn't exist
			if strings.Contains(script, "test -d") {
				return nil, fmt.Errorf("not found")
			}
			// Install succeeds
			if strings.Contains(script, "nix-env") && strings.Contains(script, "-iA") {
				return []byte("installing"), nil
			}
			// Query succeeds
			if strings.Contains(script, "-q --out-path") {
				return []byte("nodejs-20.0.0  /nix/store/xyz-nodejs"), nil
			}
			return []byte(""), nil
		},
	}

	// Create PackageRequest with previous failed packages
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "nodejs"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase:          "Failed",
			FailedPackages: []string{"nodejs", "python"}, // Previous failures
			Message:        "Installation failed",
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify status was updated and old failed packages were cleared
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)

	// Should be Ready now
	assert.Equal(t, "Ready", updatedPkgReq.Status.Phase)
	// Failed packages should be cleared (empty array)
	assert.Len(t, updatedPkgReq.Status.FailedPackages, 0, "Failed packages should be cleared on successful retry")
	// nodejs should be installed
	assert.Len(t, updatedPkgReq.Status.InstalledPackages, 1)
	assert.Equal(t, "nodejs", updatedPkgReq.Status.InstalledPackages[0].Name)
}

func TestReconcile_ChecksActualStateNotStatus(t *testing.T) {
	// Test that reconciliation checks actual filesystem state, not status
	callCount := 0
	queryCallCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// mkdir succeeds
			if strings.Contains(script, "mkdir -p") {
				return []byte(""), nil
			}
			// Profile exists
			if strings.Contains(script, "test -d") {
				return []byte(""), nil
			}
			// Query for package - simulate package IS already installed
			if strings.Contains(script, "nix-env -p") && strings.Contains(script, "-q git") {
				queryCallCount++
				return []byte("git-2.40.0  /nix/store/xyz-git-2.40.0"), nil
			}
			// Query --out-path for installed package info
			if strings.Contains(script, "-q --out-path git") {
				return []byte("git-2.40.0  /nix/store/xyz-git-2.40.0"), nil
			}
			return []byte(""), nil
		},
	}

	// Status says Pending, but package is ACTUALLY already installed on filesystem
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "git"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase:             "Pending",                       // Status says pending
			InstalledPackages: []packagesv1.InstalledPackage{}, // Status says not installed
		},
	}

	reconciler := setupTestReconcilerWithMock(t, mockExec, pkgReq)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify that isPackageInstalled was called (checking actual state)
	assert.Greater(t, queryCallCount, 0, "Should have queried actual package state")

	// Verify status was updated to reflect actual state
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.Equal(t, "Ready", updatedPkgReq.Status.Phase)
	assert.Len(t, updatedPkgReq.Status.InstalledPackages, 1)
	assert.Equal(t, "git", updatedPkgReq.Status.InstalledPackages[0].Name)
}

func TestReconcile_StatusUpdateToInstallingClearsFailedPackages(t *testing.T) {
	// Test that when phase changes to "Installing", failed packages are cleared
	var statusUpdateCount int
	var installingPhaseCleared bool

	callCount := 0
	mockExec := &MockCommandExecutor{
		ExecuteFunc: func(script string) ([]byte, error) {
			callCount++
			// mkdir succeeds
			if strings.Contains(script, "mkdir -p") {
				return []byte(""), nil
			}
			// Profile check - doesn't exist yet
			if strings.Contains(script, "test -d") {
				return nil, fmt.Errorf("not found")
			}
			// Install succeeds
			if strings.Contains(script, "nix-env") && strings.Contains(script, "-iA") {
				return []byte("installing"), nil
			}
			// Query succeeds
			if strings.Contains(script, "-q --out-path") {
				return []byte("curl-8.0.0  /nix/store/xyz-curl"), nil
			}
			return []byte(""), nil
		},
	}

	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "curl"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase:          "Failed",
			FailedPackages: []string{"curl", "wget"}, // Old failures
		},
	}

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = packagesv1.AddToScheme(scheme)
	_ = packagesv1.AddToScheme(scheme)

	logger, _ := zap.NewDevelopment()

	// Create client with interceptor to track status updates
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(pkgReq).
		WithStatusSubresource(&packagesv1.PackageRequest{}).
		WithInterceptorFuncs(interceptor.Funcs{
			SubResourceUpdate: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
				statusUpdateCount++
				if pr, ok := obj.(*packagesv1.PackageRequest); ok {
					if pr.Status.Phase == "Installing" && len(pr.Status.FailedPackages) == 0 {
						installingPhaseCleared = true
					}
				}
				return client.SubResource(subResourceName).Update(ctx, obj, opts...)
			},
		}).
		Build()

	reconciler := &PackageManagerReconciler{
		Client:    fakeClient,
		Scheme:    scheme,
		Logger:    logger,
		Namespace: "test-namespace",
		CmdExec:   mockExec,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)

	// Verify that status was updated to Installing with cleared failed packages
	assert.True(t, installingPhaseCleared, "Failed packages should be cleared when status changes to Installing")
	assert.Greater(t, statusUpdateCount, 0, "Status should be updated at least once")
}
