package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/v2/api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestValidateConfigMap_Success tests validating a config envvar with no duplicates
func TestValidateConfigMap_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvConfigMapName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string]string{
			"KEY1": "value1",
		},
	}

	configMapBytes, _ := json.Marshal(configMap)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: configMapBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateConfigMap)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
}

// TestValidateConfigMap_DuplicateKey tests rejection when key exists in secret
func TestValidateConfigMap_DuplicateKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	// Create existing secret with the same key
	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvSecretName,
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"KEY1": []byte("secret-value"),
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingSecret).Build()
	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvConfigMapName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string]string{
			"KEY1": "value1", // Duplicate key
		},
	}

	configMapBytes, _ := json.Marshal(configMap)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: configMapBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateConfigMap)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "already exists as a secret")
}

// TestValidateConfigMap_WrongName tests rejection when wrong ConfigMap uses envvars label
func TestValidateConfigMap_WrongName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wrong-name",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string]string{
			"KEY1": "value1",
		},
	}

	configMapBytes, _ := json.Marshal(configMap)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: configMapBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateConfigMap)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Only ConfigMap named")
}

// TestValidateConfigMap_OtherConfigMapAllowed tests allowing other ConfigMaps without envvars label
func TestValidateConfigMap_OtherConfigMapAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-configmap",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"KEY1": "value1",
		},
	}

	configMapBytes, _ := json.Marshal(configMap)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: configMapBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateConfigMap)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
}

// TestValidateSecret_Success tests validating a secret envvar with no duplicates
func TestValidateSecret_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvSecretName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string][]byte{
			"PASSWORD": []byte("secret"),
		},
	}

	secretBytes, _ := json.Marshal(secret)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: secretBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
}

// TestValidateSecret_DuplicateKey tests rejection when key exists in configmap
func TestValidateSecret_DuplicateKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	// Create existing configmap with the same key
	existingConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvConfigMapName,
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"PASSWORD": "config-value",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingConfigMap).Build()
	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvSecretName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string][]byte{
			"PASSWORD": []byte("secret"), // Duplicate key
		},
	}

	secretBytes, _ := json.Marshal(secret)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: secretBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "already exists as a config")
}

// TestValidateSecret_WrongName tests rejection when wrong Secret uses envvars label
func TestValidateSecret_WrongName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wrong-name",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string][]byte{
			"PASSWORD": []byte("secret"),
		},
	}

	secretBytes, _ := json.Marshal(secret)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: secretBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Only Secret named")
}

// TestValidateSecret_OtherSecretAllowed tests allowing other Secrets without envvars label
func TestValidateSecret_OtherSecretAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-secret",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"PASSWORD": []byte("secret"),
		},
	}

	secretBytes, _ := json.Marshal(secret)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: secretBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
}

// TestValidateConfigMap_InvalidJSON tests handling invalid JSON
func TestValidateConfigMap_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateConfigMap)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to unmarshal")
}

// TestValidateSecret_InvalidJSON tests handling invalid JSON
func TestValidateSecret_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to unmarshal")
}

// TestValidateConfigMap_UpdateOperation tests UPDATE operation
func TestValidateConfigMap_UpdateOperation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvConfigMapName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string]string{
			"KEY1": "updated-value",
		},
	}

	configMapBytes, _ := json.Marshal(configMap)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Update,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: configMapBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateConfigMap)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
}

// TestValidateConfigMap_EmptyKey tests rejection of empty key
func TestValidateConfigMap_EmptyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvConfigMapName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string]string{
			"": "value1", // Empty key
		},
	}

	configMapBytes, _ := json.Marshal(configMap)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: configMapBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateConfigMap)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "cannot be empty")
}

// TestValidateConfigMap_InvalidKeyFormat tests rejection of invalid key format
func TestValidateConfigMap_InvalidKeyFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	testCases := []struct {
		name        string
		key         string
		description string
	}{
		{"starts with number", "123KEY", "key starting with number"},
		{"contains dash", "KEY-NAME", "key with dash"},
		{"contains dot", "KEY.NAME", "key with dot"},
		{"contains space", "KEY NAME", "key with space"},
		{"starts with special char", "@KEY", "key starting with special char"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      EnvConfigMapName,
					Namespace: "test-namespace",
					Labels: map[string]string{
						"kloudlite.io/config-type": "envvars",
					},
				},
				Data: map[string]string{
					tc.key: "value1",
				},
			}

			configMapBytes, _ := json.Marshal(configMap)
			admissionReview := admissionv1.AdmissionReview{
				Request: &admissionv1.AdmissionRequest{
					UID:       "test-uid",
					Operation: admissionv1.Create,
					Namespace: "test-namespace",
					Object: runtime.RawExtension{
						Raw: configMapBytes,
					},
				},
			}

			body, _ := json.Marshal(admissionReview)
			req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			router := gin.New()
			router.POST("/validate", webhook.ValidateConfigMap)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response admissionv1.AdmissionReview
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response.Response.Allowed, "Should reject %s", tc.description)
			assert.Contains(t, response.Response.Result.Message, "invalid environment variable key")
		})
	}
}

// TestValidateConfigMap_EmptyValue tests rejection of empty value
func TestValidateConfigMap_EmptyValue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvConfigMapName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string]string{
			"KEY1": "", // Empty value
		},
	}

	configMapBytes, _ := json.Marshal(configMap)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: configMapBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateConfigMap)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "cannot be empty")
}

// TestValidateConfigMap_EmptyData tests rejection of empty ConfigMap data
func TestValidateConfigMap_EmptyData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvConfigMapName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string]string{}, // Empty data
	}

	configMapBytes, _ := json.Marshal(configMap)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: configMapBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateConfigMap)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "ConfigMap data cannot be empty")
}

// TestValidateSecret_EmptyKey tests rejection of empty key in Secret
func TestValidateSecret_EmptyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvSecretName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string][]byte{
			"": []byte("value1"), // Empty key
		},
	}

	secretBytes, _ := json.Marshal(secret)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: secretBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "cannot be empty")
}

// TestValidateSecret_InvalidKeyFormat tests rejection of invalid key format in Secret
func TestValidateSecret_InvalidKeyFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvSecretName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string][]byte{
			"123-INVALID": []byte("secret"), // Invalid key
		},
	}

	secretBytes, _ := json.Marshal(secret)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: secretBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "invalid environment variable key")
}

// TestValidateSecret_EmptyData tests rejection of empty Secret data
func TestValidateSecret_EmptyData(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewEnvVarWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      EnvSecretName,
			Namespace: "test-namespace",
			Labels: map[string]string{
				"kloudlite.io/config-type": "envvars",
			},
		},
		Data: map[string][]byte{}, // Empty data
	}

	secretBytes, _ := json.Marshal(secret)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: secretBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateSecret)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Secret data cannot be empty")
}
