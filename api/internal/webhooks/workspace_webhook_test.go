package webhooks

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupWorkspaceWebhookTest(t *testing.T, objects ...client.Object) *WorkspaceWebhook {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = workspacesv1.AddToScheme(scheme)
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = environmentv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()

	zapLogger, _ := zap.NewDevelopment()
	return NewWorkspaceWebhook(logger.NewZapLogger(zapLogger), k8sClient)
}

// Test DNS-1123 Label Validation
func TestValidateDNS1123Label_Valid(t *testing.T) {
	tests := []struct {
		name  string
		label string
	}{
		{"simple", "workspace"},
		{"with-dash", "my-workspace"},
		{"with-number", "workspace123"},
		{"start-with-number", "1workspace"},
		{"max-length", "a23456789b23456789c23456789d23456789e23456789f23456789g2345678"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDNS1123Label(tt.label)
			assert.NoError(t, err)
		})
	}
}

func TestValidateDNS1123Label_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		label string
		error string
	}{
		{"empty", "", "name cannot be empty"},
		{"too-long", "a23456789b23456789c23456789d23456789e23456789f23456789g234567890", "no more than 63 characters"},
		{"uppercase", "MyWorkspace", "lower case alphanumeric"},
		{"underscore", "my_workspace", "lower case alphanumeric"},
		{"dot", "my.workspace", "lower case alphanumeric"},
		{"end-with-dash", "workspace-", "start and end with an alphanumeric"},
		{"start-with-dash", "-workspace", "start and end with an alphanumeric"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDNS1123Label(tt.label)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
		})
	}
}

// Test Resource Quota Validation
func TestValidateResourceQuota_Valid(t *testing.T) {
	tests := []struct {
		name  string
		quota *workspacesv1.ResourceQuota
	}{
		{"cpu-cores", &workspacesv1.ResourceQuota{CPU: "2"}},
		{"cpu-millicores", &workspacesv1.ResourceQuota{CPU: "500m"}},
		{"cpu-decimal", &workspacesv1.ResourceQuota{CPU: "1.5"}},
		{"memory-gi", &workspacesv1.ResourceQuota{Memory: "4Gi"}},
		{"memory-mi", &workspacesv1.ResourceQuota{Memory: "512Mi"}},
		{"storage-ti", &workspacesv1.ResourceQuota{Storage: "1Ti"}},
		{"gpus-valid", &workspacesv1.ResourceQuota{GPUs: 2}},
		{"gpus-zero", &workspacesv1.ResourceQuota{GPUs: 0}},
		{"gpus-max", &workspacesv1.ResourceQuota{GPUs: 8}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceQuota(tt.quota)
			assert.NoError(t, err)
		})
	}
}

func TestValidateResourceQuota_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		quota *workspacesv1.ResourceQuota
		error string
	}{
		{"invalid-cpu", &workspacesv1.ResourceQuota{CPU: "abc"}, "invalid CPU format"},
		{"invalid-memory", &workspacesv1.ResourceQuota{Memory: "4GB"}, "invalid memory format"},
		{"invalid-storage", &workspacesv1.ResourceQuota{Storage: "10TB"}, "invalid storage format"},
		{"gpus-negative", &workspacesv1.ResourceQuota{GPUs: -1}, "GPUs must be between 0 and 8"},
		{"gpus-too-high", &workspacesv1.ResourceQuota{GPUs: 10}, "GPUs must be between 0 and 8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceQuota(tt.quota)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
		})
	}
}

// Test Workspace Settings Validation
func TestValidateWorkspaceSettings_Valid(t *testing.T) {
	tests := []struct {
		name     string
		settings *workspacesv1.WorkspaceSettings
	}{
		{"idle-timeout-zero", &workspacesv1.WorkspaceSettings{IdleTimeout: 0}},
		{"idle-timeout-max", &workspacesv1.WorkspaceSettings{IdleTimeout: 10080}},
		{"max-runtime-zero", &workspacesv1.WorkspaceSettings{MaxRuntime: 0}},
		{"max-runtime-max", &workspacesv1.WorkspaceSettings{MaxRuntime: 43200}},
		{"dotfiles-https", &workspacesv1.WorkspaceSettings{DotfilesRepo: "https://github.com/user/dotfiles"}},
		{"dotfiles-http", &workspacesv1.WorkspaceSettings{DotfilesRepo: "http://example.com/dotfiles.git"}},
		{"dotfiles-git", &workspacesv1.WorkspaceSettings{DotfilesRepo: "git@github.com:user/dotfiles.git"}},
		{"git-config-valid", &workspacesv1.WorkspaceSettings{GitConfig: &workspacesv1.GitConfig{UserEmail: "user@example.com"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWorkspaceSettings(tt.settings)
			assert.NoError(t, err)
		})
	}
}

func TestValidateWorkspaceSettings_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		settings *workspacesv1.WorkspaceSettings
		error    string
	}{
		{"idle-timeout-negative", &workspacesv1.WorkspaceSettings{IdleTimeout: -1}, "idleTimeout must be between 0 and 10080"},
		{"idle-timeout-too-high", &workspacesv1.WorkspaceSettings{IdleTimeout: 10081}, "idleTimeout must be between 0 and 10080"},
		{"max-runtime-negative", &workspacesv1.WorkspaceSettings{MaxRuntime: -1}, "maxRuntime must be between 0 and 43200"},
		{"max-runtime-too-high", &workspacesv1.WorkspaceSettings{MaxRuntime: 43201}, "maxRuntime must be between 0 and 43200"},
		{"dotfiles-invalid-url", &workspacesv1.WorkspaceSettings{DotfilesRepo: "ftp://example.com"}, "valid git repository URL"},
		{"git-config-invalid-email", &workspacesv1.WorkspaceSettings{GitConfig: &workspacesv1.GitConfig{UserEmail: "notanemail"}}, "valid email address"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWorkspaceSettings(tt.settings)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.error)
		})
	}
}

// Test Validation - Valid Workspace
func TestWorkspaceWebhook_ValidateWorkspace_Success(t *testing.T) {
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testuser",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testuser-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy: "testuser",
		},
	}

	webhook := setupWorkspaceWebhookTest(t, user, workMachine)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-ns",
		},
		Spec: workspacesv1.WorkspaceSpec{
			OwnedBy:     "testuser",
			DisplayName: "Test Workspace",
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: workspaceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateWorkspace)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
}

// Test Validation - Missing Owner
func TestWorkspaceWebhook_ValidateWorkspace_MissingOwner(t *testing.T) {
	webhook := setupWorkspaceWebhookTest(t)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: workspaceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateWorkspace)
	router.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "owner is required")
}

// Test Validation - Owner Not Found
func TestWorkspaceWebhook_ValidateWorkspace_OwnerNotFound(t *testing.T) {
	webhook := setupWorkspaceWebhookTest(t)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			OwnedBy:     "nonexistent",
			DisplayName: "Test Workspace",
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: workspaceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateWorkspace)
	router.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "does not exist")
}

// Test Validation - No WorkMachine
func TestWorkspaceWebhook_ValidateWorkspace_NoWorkMachine(t *testing.T) {
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testuser",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
		},
	}

	webhook := setupWorkspaceWebhookTest(t, user)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Spec: workspacesv1.WorkspaceSpec{
			OwnedBy:     "testuser",
			DisplayName: "Test Workspace",
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: workspaceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateWorkspace)
	router.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "does not have a WorkMachine")
}

// Test Validation - DisplayName Required
func TestWorkspaceWebhook_ValidateWorkspace_DisplayNameRequired(t *testing.T) {
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser"},
		Spec:       platformv1alpha1.UserSpec{Email: "test@example.com"},
	}
	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser-machine"},
		Spec:       machinesv1.WorkMachineSpec{OwnedBy: "testuser"},
	}

	webhook := setupWorkspaceWebhookTest(t, user, workMachine)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-workspace"},
		Spec:       workspacesv1.WorkspaceSpec{OwnedBy: "testuser"},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: workspaceBytes},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateWorkspace)
	router.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "displayName is required")
}

// Test Validation - DisplayName Too Long
func TestWorkspaceWebhook_ValidateWorkspace_DisplayNameTooLong(t *testing.T) {
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser"},
		Spec:       platformv1alpha1.UserSpec{Email: "test@example.com"},
	}
	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser-machine"},
		Spec:       machinesv1.WorkMachineSpec{OwnedBy: "testuser"},
	}

	webhook := setupWorkspaceWebhookTest(t, user, workMachine)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-workspace"},
		Spec: workspacesv1.WorkspaceSpec{
			OwnedBy:     "testuser",
			DisplayName: "a123456789b123456789c123456789d123456789e123456789f12345678901234567890123456789012345678901234567890",
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: workspaceBytes},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateWorkspace)
	router.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "no more than 100 characters")
}

// Test Validation - Invalid Status
func TestWorkspaceWebhook_ValidateWorkspace_InvalidStatus(t *testing.T) {
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser"},
		Spec:       platformv1alpha1.UserSpec{Email: "test@example.com"},
	}
	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser-machine"},
		Spec:       machinesv1.WorkMachineSpec{OwnedBy: "testuser"},
	}

	webhook := setupWorkspaceWebhookTest(t, user, workMachine)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-workspace"},
		Spec: workspacesv1.WorkspaceSpec{
			OwnedBy:     "testuser",
			DisplayName: "Test",
			Status:      "invalid",
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: workspaceBytes},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateWorkspace)
	router.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "invalid status")
}

// Test Validation - Service Intercepts Conflict
func TestWorkspaceWebhook_ValidateWorkspace_InterceptConflict(t *testing.T) {
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser"},
		Spec:       platformv1alpha1.UserSpec{Email: "test@example.com"},
	}
	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser-machine"},
		Spec:       machinesv1.WorkMachineSpec{OwnedBy: "testuser"},
	}
	env := &environmentv1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-env", Namespace: "test-ns"},
		Spec:       environmentv1.EnvironmentSpec{TargetNamespace: "target-ns"},
	}
	existingIntercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{Name: "existing-intercept", Namespace: "test-ns"},
		Spec: interceptsv1.ServiceInterceptSpec{
			WorkspaceRef: corev1.ObjectReference{Name: "other-workspace"},
			ServiceRef:   corev1.ObjectReference{Name: "my-service", Namespace: "target-ns"},
		},
	}

	webhook := setupWorkspaceWebhookTest(t, user, workMachine, env, existingIntercept)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-workspace", Namespace: "test-ns"},
		Spec: workspacesv1.WorkspaceSpec{
			OwnedBy:     "testuser",
			DisplayName: "Test",
			EnvironmentConnection: &workspacesv1.EnvironmentConnectionSpec{
				EnvironmentRef: corev1.ObjectReference{Name: "test-env", Namespace: "test-ns"},
				Intercepts: []workspacesv1.InterceptSpec{
					{ServiceName: "my-service"},
				},
			},
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: workspaceBytes},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateWorkspace)
	router.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "already being intercepted")
}

// Test Mutation - Default Values
func TestWorkspaceWebhook_MutateWorkspace_DefaultValues(t *testing.T) {
	webhook := setupWorkspaceWebhookTest(t)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-workspace"},
		Spec: workspacesv1.WorkspaceSpec{
			OwnedBy:     "testuser",
			DisplayName: "Test",
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: workspaceBytes},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateWorkspace)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.NotNil(t, response.Response.Patch)

	var patches []map[string]interface{}
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)

	// Check for default values
	foundStatus := false
	foundStorageSize := false
	foundWorkspacePath := false
	foundVSCodeVersion := false

	for _, patch := range patches {
		switch patch["path"] {
		case "/spec/status":
			foundStatus = true
			assert.Equal(t, "active", patch["value"])
		case "/spec/storageSize":
			foundStorageSize = true
			assert.Equal(t, "10Gi", patch["value"])
		case "/spec/workspacePath":
			foundWorkspacePath = true
			assert.Equal(t, "/workspace", patch["value"])
		case "/spec/vscodeVersion":
			foundVSCodeVersion = true
			assert.Equal(t, "latest", patch["value"])
		}
	}

	assert.True(t, foundStatus)
	assert.True(t, foundStorageSize)
	assert.True(t, foundWorkspacePath)
	assert.True(t, foundVSCodeVersion)
}

// Test Mutation - Labels and Finalizer
func TestWorkspaceWebhook_MutateWorkspace_LabelsAndFinalizer(t *testing.T) {
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser"},
		Spec:       platformv1alpha1.UserSpec{Email: "test@example.com"},
	}

	webhook := setupWorkspaceWebhookTest(t, user)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-workspace"},
		Spec: workspacesv1.WorkspaceSpec{
			OwnedBy:     "testuser",
			DisplayName: "Test",
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: workspaceBytes},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateWorkspace)
	router.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response.Response.Allowed)

	var patches []map[string]interface{}
	json.Unmarshal(response.Response.Patch, &patches)

	// Check for workspace-name label
	foundWorkspaceNameLabel := false
	foundOwnedByLabel := false
	foundOwnerEmailLabel := false
	foundFinalizer := false

	for _, patch := range patches {
		path, _ := patch["path"].(string)
		if path == "/metadata/labels/kloudlite.io~1workspace-name" {
			foundWorkspaceNameLabel = true
			assert.Equal(t, "test-workspace", patch["value"])
		}
		if path == "/metadata/labels/kloudlite.io~1owned-by" {
			foundOwnedByLabel = true
			assert.Equal(t, "testuser", patch["value"])
		}
		if path == "/metadata/labels/kloudlite.io~1owner-email" {
			foundOwnerEmailLabel = true
			expectedEmail := base64.URLEncoding.EncodeToString([]byte("test@example.com"))
			assert.Equal(t, expectedEmail, patch["value"])
		}
		if path == "/metadata/finalizers" {
			finalizers, ok := patch["value"].([]interface{})
			if ok && len(finalizers) > 0 {
				foundFinalizer = true
				assert.Contains(t, finalizers, "workspaces.kloudlite.io/finalizer")
			}
		}
	}

	assert.True(t, foundWorkspaceNameLabel)
	assert.True(t, foundOwnedByLabel)
	assert.True(t, foundOwnerEmailLabel)
	assert.True(t, foundFinalizer)
}

// Test Mutation - Owner by Email
func TestWorkspaceWebhook_MutateWorkspace_OwnerByEmail(t *testing.T) {
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{Name: "testuser"},
		Spec:       platformv1alpha1.UserSpec{Email: "test@example.com"},
	}

	webhook := setupWorkspaceWebhookTest(t, user)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-workspace"},
		Spec: workspacesv1.WorkspaceSpec{
			OwnedBy:     "test@example.com",
			DisplayName: "Test",
		},
	}

	workspaceBytes, _ := json.Marshal(workspace)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object:    runtime.RawExtension{Raw: workspaceBytes},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateWorkspace)
	router.ServeHTTP(w, req)

	var response admissionv1.AdmissionReview
	json.Unmarshal(w.Body.Bytes(), &response)

	var patches []map[string]interface{}
	json.Unmarshal(response.Response.Patch, &patches)

	// Should resolve email to username
	foundOwnedByLabel := false
	for _, patch := range patches {
		path, _ := patch["path"].(string)
		if path == "/metadata/labels/kloudlite.io~1owned-by" {
			foundOwnedByLabel = true
			assert.Equal(t, "testuser", patch["value"])
		}
	}
	assert.True(t, foundOwnedByLabel)
}

// Test Invalid JSON
func TestWorkspaceWebhook_ValidateWorkspace_InvalidJSON(t *testing.T) {
	webhook := setupWorkspaceWebhookTest(t)

	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateWorkspace)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWorkspaceWebhook_MutateWorkspace_InvalidJSON(t *testing.T) {
	webhook := setupWorkspaceWebhookTest(t)

	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateWorkspace)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
