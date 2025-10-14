package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestValidateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	activeTrue := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &activeTrue,
			Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	userBytes, _ := json.Marshal(user)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
}

func TestValidateUser_MissingEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "", // Missing email
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	userBytes, _ := json.Marshal(user)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Email is required")
}

func TestValidateUser_InvalidEmailFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "invalid-email", // Invalid format
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	userBytes, _ := json.Marshal(user)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Invalid email format")
}

func TestValidateUser_EmailAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)

	activeTrue := true
	existingUser := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-user",
			Labels: map[string]string{
				"kloudlite.io/user-email-hash": hashEmail("test@example.com"),
			},
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &activeTrue,
			Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingUser).Build()
	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	newUser := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com", // Duplicate email
			Active: &activeTrue,
			Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	userBytes, _ := json.Marshal(newUser)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "already exists")
}

func TestValidateUser_EmailChangeNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	activeTrue := true
	oldUser := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "old@example.com",
			Active: &activeTrue,
			Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	newUser := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "new@example.com", // Changed email
			Active: &activeTrue,
			Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	oldUserBytes, _ := json.Marshal(oldUser)
	newUserBytes, _ := json.Marshal(newUser)

	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Update,
			Object: runtime.RawExtension{
				Raw: newUserBytes,
			},
			OldObject: runtime.RawExtension{
				Raw: oldUserBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Email cannot be changed")
}

func TestValidateUser_InvalidKubernetesName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "InvalidName", // Uppercase not allowed
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	userBytes, _ := json.Marshal(user)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Invalid name format")
}

func TestValidateUser_NoRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
			Roles: []platformv1alpha1.RoleType{}, // No roles
		},
	}

	userBytes, _ := json.Marshal(user)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "At least one role is required")
}

func TestValidateUser_InvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
			Roles: []platformv1alpha1.RoleType{"invalid-role"},
		},
	}

	userBytes, _ := json.Marshal(user)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Invalid role")
}

func TestValidateUser_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	// Invalid JSON body
	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to unmarshal")
}

// MUTATION TESTS

func TestMutateUser_AddGenerateName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			// No name or generateName
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	userBytes, _ := json.Marshal(user)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.NotNil(t, response.Response.Patch)

	// Verify patch contains generateName
	var patches []patchOperation
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)

	foundGenerateName := false
	for _, patch := range patches {
		if patch.Path == "/metadata/generateName" {
			foundGenerateName = true
			assert.Equal(t, "add", patch.Op)
			assert.Equal(t, "user-", patch.Value)
		}
	}
	assert.True(t, foundGenerateName)
}

func TestMutateUser_AddDefaultActive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
			// Active is nil
		},
	}

	userBytes, _ := json.Marshal(user)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)

	var patches []patchOperation
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)

	foundActive := false
	for _, patch := range patches {
		if patch.Path == "/spec/active" {
			foundActive = true
			assert.Equal(t, "add", patch.Op)
			assert.Equal(t, true, patch.Value)
		}
	}
	assert.True(t, foundActive)
}

func TestMutateUser_AddEmailLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewUserWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email: "test@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		},
	}

	userBytes, _ := json.Marshal(user)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: userBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateUser)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)

	var patches []patchOperation
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)

	// Verify patches exist (webhook always adds labels and other fields)
	assert.NotEmpty(t, patches, "Should have mutation patches")

	// Check if email labels were added
	for _, patch := range patches {
		if patch.Path == "/metadata/labels/kloudlite.io~1user-email" {
			assert.Equal(t, "test-at-example-dot-com", patch.Value)
		}
		if patch.Path == "/metadata/labels/kloudlite.io~1user-email-hash" {
			assert.NotEmpty(t, patch.Value)
		}
	}
}

// Helper function tests

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name@domain.co.uk", true},
		{"invalid", false},
		{"@example.com", true}, // Simple check allows this
		{"test@", false},       // No dot after @
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsValidKubernetesName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"valid-name", true},
		{"valid123", true},
		{"123valid", true},
		{"Invalid-Name", false}, // Uppercase not allowed
		{"invalid_name", false}, // Underscore not allowed
		{"invalid.name", false}, // Dot not allowed
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidKubernetesName(tt.name)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestSanitizeEmailForLabel(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"test@example.com", "test-at-example-dot-com"},
		{"user.name@domain.com", "user-dot-name-at-domain-dot-com"},
		{"Test@Example.COM", "test-at-example-dot-com"}, // Should lowercase
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := sanitizeEmailForLabel(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashEmail(t *testing.T) {
	email1 := "test@example.com"
	email2 := "test@example.com"
	email3 := "different@example.com"

	hash1 := hashEmail(email1)
	hash2 := hashEmail(email2)
	hash3 := hashEmail(email3)

	// Same email should produce same hash
	assert.Equal(t, hash1, hash2)

	// Different emails should produce different hashes
	assert.NotEqual(t, hash1, hash3)

	// Hash should be 16 characters
	assert.Equal(t, 16, len(hash1))
}
