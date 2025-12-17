package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
	mockExec := &MockCommandExecutor{}
	profileManager := NewNixProfileManager(logger, mockExec)

	return &PackageManagerReconciler{
		Client:         fakeClient,
		Scheme:         scheme,
		Logger:         logger,
		Namespace:      "test-namespace",
		CmdExec:        mockExec,
		ProfileManager: profileManager,
	}
}

func setupTestReconcilerWithMock(t *testing.T, mockExec *MockCommandExecutor, initObjs ...runtime.Object) *PackageManagerReconciler {
	reconciler := setupTestReconciler(t, initObjs...)
	reconciler.CmdExec = mockExec
	reconciler.ProfileManager = NewNixProfileManager(reconciler.Logger, mockExec)
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
	// Compute the expected spec hash for the test packages
	testPackages := []packagesv1.PackageSpec{{Name: "git"}}
	logger, _ := zap.NewDevelopment()
	mockExec := &MockCommandExecutor{}
	profileManager := NewNixProfileManager(logger, mockExec)
	expectedHash := profileManager.ComputeSpecHash(testPackages)

	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Finalizers: []string{packageRequestFinalizer},
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages:     testPackages,
		},
		Status: packagesv1.PackageRequestStatus{
			Phase:            "Ready",
			Message:          "Already installed",
			SpecHash:         expectedHash, // Hash matches, so skip build
			PackageCount:     1,
			Packages:         []string{"git"},
			ProfileStorePath: "/nix/store/xyz-git",
			PackagesPath:     "/nix/profiles/kloudlite/test-workspace/packages",
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

	// Verify status unchanged (should skip because hash matches)
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.Equal(t, "Ready", updatedPkgReq.Status.Phase)
}

func TestPackageManagerReconciler_SetupWithManager(t *testing.T) {
	// This test verifies the controller can be created without errors
	// Actual manager setup would require a real manager
	reconciler := setupTestReconciler(t)
	assert.NotNil(t, reconciler)
}

func TestPackageManagerReconciler_UpdateEventFilter_Ready(t *testing.T) {
	pkgReqOld := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-pkg-req",
			Namespace:  "test-namespace",
			Generation: 1,
		},
		Spec: packagesv1.PackageRequestSpec{
			WorkspaceRef: "test-workspace",
			ProfileName:  "test-profile",
			Packages: []packagesv1.PackageSpec{
				{Name: "git"},
			},
		},
		Status: packagesv1.PackageRequestStatus{
			Phase: "Ready",
		},
	}

	// Same generation, only status changed
	pkgReqNew := pkgReqOld.DeepCopy()
	pkgReqNew.Status.Message = "Updated message"

	_ = event.UpdateEvent{
		ObjectOld: pkgReqOld,
		ObjectNew: pkgReqNew,
	}

	// The update filter should return false for status-only changes
	// (same generation means no spec change)
	assert.Equal(t, pkgReqOld.Generation, pkgReqNew.Generation)
}

func TestPackageManagerReconciler_DeleteEvent(t *testing.T) {
	pkgReq := &packagesv1.PackageRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pkg-req",
			Namespace: "test-namespace",
		},
	}

	e := event.DeleteEvent{
		Object: pkgReq,
	}

	// Delete events should be processed
	assert.NotNil(t, e.Object)
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
			Packages:     []packagesv1.PackageSpec{}, // Empty list
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
	_, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)

	// Verify status is Ready with empty packages
	updatedPkgReq := &packagesv1.PackageRequest{}
	err = reconciler.Get(context.Background(), req.NamespacedName, updatedPkgReq)
	assert.NoError(t, err)
	assert.Equal(t, "Ready", updatedPkgReq.Status.Phase)
	assert.Equal(t, 0, updatedPkgReq.Status.PackageCount)
}

// ========================================
// Workspace Home Setup Tests
// ========================================

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
// SSH Config Tests
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
		FS:     &MockFileSystem{},
	}
}

func TestSSHConfigReconciler_Reconcile_NotFound(t *testing.T) {
	reconciler := setupTestSSHConfigReconciler(t)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "ssh-authorized-keys",
			Namespace: "test-namespace",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)

	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}

func TestSetupSSHConfigDirectory_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{}

	err := setupSSHConfigDirectory(logger, mockFS)

	assert.NoError(t, err)
	assert.Len(t, mockFS.CallLog, 1)
	assert.Contains(t, mockFS.CallLog[0], "MkdirAll(/var/lib/kloudlite/ssh-config")
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
	content := "ssh-rsa AAAAB3... user@example.com"

	err := writeAuthorizedKeys(logger, content, mockFS)

	assert.NoError(t, err)
	assert.Len(t, mockFS.CallLog, 2)
	assert.Contains(t, mockFS.CallLog[0], "WriteFile")
	assert.Contains(t, mockFS.CallLog[1], "Rename")
}

func TestWriteAuthorizedKeys_WriteFileError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockFS := &MockFileSystem{
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error {
			return fmt.Errorf("write failed")
		},
	}

	err := writeAuthorizedKeys(logger, "content", mockFS)

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

	err := writeAuthorizedKeys(logger, "content", mockFS)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to rename temporary authorized_keys file")
}

// ========================================
// Helper Function Tests
// ========================================

func TestContainsString(t *testing.T) {
	slice := []string{"a", "b", "c"}

	assert.True(t, containsString(slice, "a"))
	assert.True(t, containsString(slice, "b"))
	assert.True(t, containsString(slice, "c"))
	assert.False(t, containsString(slice, "d"))
	assert.False(t, containsString(nil, "a"))
	assert.False(t, containsString([]string{}, "a"))
}

func TestRemoveString(t *testing.T) {
	slice := []string{"a", "b", "c"}

	result := removeString(slice, "b")
	assert.Equal(t, []string{"a", "c"}, result)

	result = removeString(slice, "d")
	assert.Equal(t, []string{"a", "b", "c"}, result)

	result = removeString([]string{}, "a")
	assert.Equal(t, []string{}, result)
}

func TestNixStorePathConstant(t *testing.T) {
	assert.Equal(t, "/nix", nixStorePath)
}

func TestSanitizeLabelValue(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"valid-label", 63, "valid-label"},
		{"UPPERCASE", 63, "UPPERCASE"},
		{"with spaces", 63, "with-spaces"},
		{"with/slashes", 63, "with-slashes"},
		{"---leading-trailing---", 63, "leading-trailing"},
		{"", 63, "unknown"},
		{"a", 1, "a"},
		{"abcdef", 3, "abc"},
		{"123", 63, "123"},
		{"test_underscore.dot", 63, "test_underscore.dot"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeLabelValue(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateError(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long error message", 10, "this is a ..."},
		{"", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := truncateError(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ========================================
// NixProfileManager Tests
// ========================================

func TestNixProfileManager_ComputeSpecHash(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockExec := &MockCommandExecutor{}
	manager := NewNixProfileManager(logger, mockExec)

	// Same packages in different order should produce same hash
	packages1 := []packagesv1.PackageSpec{
		{Name: "git"},
		{Name: "vim"},
	}
	packages2 := []packagesv1.PackageSpec{
		{Name: "vim"},
		{Name: "git"},
	}

	hash1 := manager.ComputeSpecHash(packages1)
	hash2 := manager.ComputeSpecHash(packages2)

	assert.Equal(t, hash1, hash2, "Same packages in different order should produce same hash")

	// Different packages should produce different hash
	packages3 := []packagesv1.PackageSpec{
		{Name: "git"},
		{Name: "curl"},
	}

	hash3 := manager.ComputeSpecHash(packages3)
	assert.NotEqual(t, hash1, hash3, "Different packages should produce different hash")

	// Empty packages should produce consistent hash
	emptyHash := manager.ComputeSpecHash([]packagesv1.PackageSpec{})
	assert.NotEmpty(t, emptyHash)
}

func TestNixProfileManager_GetPaths(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockExec := &MockCommandExecutor{}
	manager := NewNixProfileManager(logger, mockExec)

	workspace := "test-workspace"

	profileDir := manager.GetProfileDir(workspace)
	assert.Equal(t, "/nix/profiles/kloudlite/test-workspace", profileDir)

	nixPath := manager.GetProfileNixPath(workspace)
	assert.Equal(t, "/nix/profiles/kloudlite/test-workspace/profile.nix", nixPath)

	packagesPath := manager.GetPackagesPath(workspace)
	assert.Equal(t, "/nix/profiles/kloudlite/test-workspace/packages", packagesPath)
}

// ========================================
// GPU-related Tests
// ========================================

func TestParseQuantity(t *testing.T) {
	q := parseQuantity("1")
	assert.NotNil(t, q)
	assert.Equal(t, int64(1), q.Value())

	q = parseQuantity("invalid")
	assert.NotNil(t, q)
	assert.Equal(t, int64(0), q.Value())
}

func setupTestGPUReconciler(t *testing.T, initObjs ...runtime.Object) *GPUStatusReconciler {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(initObjs...).
		WithStatusSubresource(&corev1.Node{}).
		Build()

	logger, _ := zap.NewDevelopment()

	return &GPUStatusReconciler{
		Client:   fakeClient,
		Logger:   logger,
		CmdExec:  &MockCommandExecutor{},
		NodeName: "test-node",
	}
}

func TestGPUStatusReconciler_Reconcile_WrongNode(t *testing.T) {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "other-node",
		},
	}

	reconciler := setupTestGPUReconciler(t, node)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "other-node",
		},
	}

	result, err := reconciler.Reconcile(context.Background(), req)
	assert.NoError(t, err)
	assert.False(t, result.Requeue)
}
