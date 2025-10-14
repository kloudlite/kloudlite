package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestServiceInterceptWebhook_ValidateServiceIntercept_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = environmentsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create running workspace
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-ns",
		},
		Spec: workspacesv1.WorkspaceSpec{
			DisplayName: "Test Workspace",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
		},
	}

	// Create service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-ns",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workspace, service).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "test-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-ns",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_MissingWorkspaceName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status:       "active",
			WorkspaceRef: corev1.ObjectReference{}, // Empty workspace ref
			ServiceRef: corev1.ObjectReference{
				Name: "test-service",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "WorkspaceRef.Name is required")
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_WorkspaceNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "nonexistent-workspace",
				Namespace: "test-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name: "test-service",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "not found")
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_WorkspaceNotRunning(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create workspace that is not running
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stopped-workspace",
			Namespace: "test-ns",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Stopped",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workspace).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "stopped-workspace",
				Namespace: "test-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name: "test-service",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "is not running")
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_ServiceNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create running workspace
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-ns",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workspace).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "test-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "nonexistent-service",
				Namespace: "test-ns",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "not found")
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_InvalidServicePort(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create running workspace
	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-ns",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
		},
	}

	// Create service with port 80
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-ns",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workspace, service).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "test-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-ns",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   443, // Port not in service
					WorkspacePort: 3000,
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Service port 443 not found")
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_NoPortMappings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-ns",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-ns",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 80},
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workspace, service).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "test-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-ns",
			},
			PortMappings: []interceptsv1.PortMapping{}, // Empty
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "At least one port mapping is required")
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_InvalidPortNumbers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-ns",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-ns",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 80},
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workspace, service).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "test-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-ns",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   0, // Invalid
					WorkspacePort: 70000, // Invalid (> 65535)
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "Invalid")
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_DuplicateActiveIntercept(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-ns",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-ns",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 80},
			},
		},
	}

	// Existing active intercept
	existingIntercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-intercept",
			Namespace: "test-ns",
			Labels: map[string]string{
				"intercepts.kloudlite.io/service-name": "test-service",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name: "other-workspace",
			},
			ServiceRef: corev1.ObjectReference{
				Name: "test-service",
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase: "Active",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workspace, service, existingIntercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "test-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-ns",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "already being intercepted")
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	workspace := &workspacesv1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-ns",
		},
		Status: workspacesv1.WorkspaceStatus{
			Phase: "Running",
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-ns",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 80},
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workspace, service).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "invalid-status", // Invalid
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "test-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-ns",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Response.Allowed)
	assert.Contains(t, response.Response.Result.Message, "must be either 'active' or 'inactive'")
}

func TestServiceInterceptWebhook_MutateServiceIntercept_AddsLabelsAndFinalizer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-ns",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-ns",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "service-ns",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
				},
			},
		},
	}

	interceptBytes, _ := json.Marshal(intercept)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-ns",
			Object: runtime.RawExtension{
				Raw: interceptBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.NotNil(t, response.Response.Patch)

	// Verify patches
	var patches []patchOperation
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)
	assert.NotEmpty(t, patches)

	// Check for expected labels and finalizer
	foundServiceLabel := false
	foundWorkspaceLabel := false
	foundFinalizer := false

	for _, patch := range patches {
		path := patch.Path
		if path == "/metadata/labels" {
			labels := patch.Value.(map[string]interface{})
			if labels["intercepts.kloudlite.io/service-name"] == "test-service" {
				foundServiceLabel = true
			}
			if labels["intercepts.kloudlite.io/workspace-name"] == "test-workspace" {
				foundWorkspaceLabel = true
			}
		}
		if path == "/metadata/finalizers" {
			foundFinalizer = true
		}
	}

	assert.True(t, foundServiceLabel || foundWorkspaceLabel, "Should add service/workspace labels")
	assert.True(t, foundFinalizer, "Should add finalizer")
}

func TestServiceInterceptWebhook_MutateServiceIntercept_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/mutate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate", webhook.MutateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestServiceInterceptWebhook_ValidateServiceIntercept_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceInterceptWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/validate", webhook.ValidateServiceIntercept)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
