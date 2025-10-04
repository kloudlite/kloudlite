package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/environments/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/platform/v1alpha1"
	"github.com/kloudlite/kloudlite/v2/api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	authv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestValidateEnvironment_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

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

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()

	zapLogger, _ := zap.NewDevelopment()
	// Pass nil for clientset in tests - not needed for validation logic
	webhook := NewEnvironmentWebhook(logger.NewZapLogger(zapLogger), k8sClient, nil)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "test-user",
			Activated:       false,
		},
	}

	envBytes, _ := json.Marshal(env)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: envBytes,
			},
			UserInfo: authv1.UserInfo{
				Username: "test-user",
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateEnvironment)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
}

func TestValidateEnvironment_InvalidNamespace_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	activeTrue := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &activeTrue,
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()

	zapLogger, _ := zap.NewDevelopment()
	// Pass nil for clientset in tests - not needed for validation logic
	webhook := NewEnvironmentWebhook(logger.NewZapLogger(zapLogger), k8sClient, nil)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "", // Empty namespace
			CreatedBy:       "test-user",
		},
	}

	envBytes, _ := json.Marshal(env)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: envBytes,
			},
			UserInfo: authv1.UserInfo{
				Username: "test-user",
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateEnvironment)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "namespace name cannot be empty")
}

func TestValidateEnvironment_InvalidNamespace_Reserved(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	activeTrue := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &activeTrue,
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()

	zapLogger, _ := zap.NewDevelopment()
	// Pass nil for clientset in tests - not needed for validation logic
	webhook := NewEnvironmentWebhook(logger.NewZapLogger(zapLogger), k8sClient, nil)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "kube-system", // Reserved namespace
			CreatedBy:       "test-user",
		},
	}

	envBytes, _ := json.Marshal(env)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: envBytes,
			},
			UserInfo: authv1.UserInfo{
				Username: "test-user",
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateEnvironment)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "reserved namespace")
}

func TestValidateEnvironment_InvalidNamespace_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	activeTrue := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &activeTrue,
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()

	zapLogger, _ := zap.NewDevelopment()
	// Pass nil for clientset in tests - not needed for validation logic
	webhook := NewEnvironmentWebhook(logger.NewZapLogger(zapLogger), k8sClient, nil)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "Invalid_Namespace", // Uppercase and underscore
			CreatedBy:       "test-user",
		},
	}

	envBytes, _ := json.Marshal(env)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: envBytes,
			},
			UserInfo: authv1.UserInfo{
				Username: "test-user",
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateEnvironment)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "lower case alphanumeric")
}

func TestValidateEnvironment_NamespaceConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	activeTrue := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &activeTrue,
		},
	}

	existingEnv := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "shared-namespace",
			CreatedBy:       "test-user",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(user, existingEnv).Build()

	zapLogger, _ := zap.NewDevelopment()
	// Pass nil for clientset in tests - not needed for validation logic
	webhook := NewEnvironmentWebhook(logger.NewZapLogger(zapLogger), k8sClient, nil)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "shared-namespace", // Conflict with existing-env
			CreatedBy:       "test-user",
		},
	}

	envBytes, _ := json.Marshal(env)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: envBytes,
			},
			UserInfo: authv1.UserInfo{
				Username: "test-user",
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateEnvironment)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "already used by environment")
}

func TestValidateEnvironment_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	// Pass nil for clientset in tests - not needed for validation logic
	webhook := NewEnvironmentWebhook(logger.NewZapLogger(zapLogger), k8sClient, nil)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "nonexistent-user",
		},
	}

	envBytes, _ := json.Marshal(env)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: envBytes,
			},
			UserInfo: authv1.UserInfo{
				Username: "nonexistent-user",
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateEnvironment)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "does not exist")
}

func TestMutateEnvironment_AddLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	activeTrue := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &activeTrue,
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()

	zapLogger, _ := zap.NewDevelopment()
	// Pass nil for clientset in tests - not needed for validation logic
	webhook := NewEnvironmentWebhook(logger.NewZapLogger(zapLogger), k8sClient, nil)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "test-user",
			Activated:       false,
		},
	}

	envBytes, _ := json.Marshal(env)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: envBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateEnvironment)
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
	assert.NotEmpty(t, patches)

	// Check for environment name label
	foundEnvLabel := false
	foundManagedByLabel := false
	for _, patch := range patches {
		if path, ok := patch["path"].(string); ok {
			if path == "/spec/labels/kloudlite.io~1environment-name" {
				foundEnvLabel = true
				assert.Equal(t, "test-env", patch["value"])
			}
			if path == "/spec/labels/kloudlite.io~1managed-by" {
				foundManagedByLabel = true
				assert.Equal(t, "environment-controller", patch["value"])
			}
		}
	}

	assert.True(t, foundEnvLabel || foundManagedByLabel, "Should have management labels")
}

func TestMutateEnvironment_AddDefaultResourceQuotas(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	activeTrue := true
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-user",
		},
		Spec: platformv1alpha1.UserSpec{
			Email:  "test@example.com",
			Active: &activeTrue,
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(user).Build()

	zapLogger, _ := zap.NewDevelopment()
	// Pass nil for clientset in tests - not needed for validation logic
	webhook := NewEnvironmentWebhook(logger.NewZapLogger(zapLogger), k8sClient, nil)

	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			TargetNamespace: "test-namespace",
			CreatedBy:       "test-user",
			Activated:       true, // Activated without quotas
		},
	}

	envBytes, _ := json.Marshal(env)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Object: runtime.RawExtension{
				Raw: envBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateEnvironment)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)

	var patches []map[string]interface{}
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)

	// Check for resource quotas patch
	foundQuotas := false
	for _, patch := range patches {
		if path, ok := patch["path"].(string); ok {
			if path == "/spec/resourceQuotas" {
				foundQuotas = true
				quotas, ok := patch["value"].(map[string]interface{})
				assert.True(t, ok)
				assert.Contains(t, quotas, "limits.cpu")
				assert.Contains(t, quotas, "limits.memory")
			}
		}
	}

	assert.True(t, foundQuotas, "Should add default resource quotas for activated environment")
}

func TestMutateEnvironment_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = platformv1alpha1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	// Pass nil for clientset in tests - not needed for validation logic
	webhook := NewEnvironmentWebhook(logger.NewZapLogger(zapLogger), k8sClient, nil)

	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateEnvironment)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
