package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	interceptsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/intercepts/v1"
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

func TestServiceMutationWebhook_MutateService_NoIntercept(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test-app",
			},
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

	serviceBytes, _ := json.Marshal(service)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: serviceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-service", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-service", webhook.MutateService)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.Nil(t, response.Response.Patch, "Service should not be patched when no intercept exists")
}

func TestServiceMutationWebhook_MutateService_ActiveIntercept(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an active intercept with workspace pod name
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"intercepts.kloudlite.io/service-name": "test-service",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-namespace",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:            "Active",
			WorkspacePodName: "workspace-pod-123",
			OriginalServiceSelector: map[string]string{
				"app": "test-app",
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test-app",
			},
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

	serviceBytes, _ := json.Marshal(service)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: serviceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-service", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-service", webhook.MutateService)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.NotNil(t, response.Response.Patch, "Service should be patched when active intercept exists")

	// Verify patches
	var patches []patchOperation
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)
	assert.NotEmpty(t, patches)

	// Verify selector patch
	foundSelectorPatch := false
	foundAnnotationPatch := false
	foundPortsPatch := false

	for _, patch := range patches {
		switch patch.Path {
		case "/spec/selector":
			foundSelectorPatch = true
			selector := patch.Value.(map[string]interface{})
			assert.Equal(t, "test-workspace", selector["workspaces.kloudlite.io/workspace-name"])
		case "/metadata/annotations/intercepts.kloudlite.io~1intercepted-by":
			foundAnnotationPatch = true
			assert.Equal(t, "test-intercept", patch.Value)
		case "/spec/ports":
			foundPortsPatch = true
			// Verify ports are updated with workspace port
		}
	}

	assert.True(t, foundSelectorPatch, "Should patch service selector")
	assert.True(t, foundAnnotationPatch, "Should add intercept annotation")
	assert.True(t, foundPortsPatch, "Should patch service ports")
}

func TestServiceMutationWebhook_MutateService_NoWorkspacePodName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an active intercept without workspace pod name
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"intercepts.kloudlite.io/service-name": "test-service",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:            "Active",
			WorkspacePodName: "", // Empty workspace pod name
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test-app",
			},
		},
	}

	serviceBytes, _ := json.Marshal(service)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: serviceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-service", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-service", webhook.MutateService)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.Nil(t, response.Response.Patch, "Service should not be patched when workspace pod name is empty")
}

func TestServiceMutationWebhook_MutateService_InactiveIntercept(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an inactive intercept
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"intercepts.kloudlite.io/service-name": "test-service",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "inactive", // Inactive
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:            "Inactive",
			WorkspacePodName: "workspace-pod-123",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test-app",
			},
		},
	}

	serviceBytes, _ := json.Marshal(service)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: serviceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-service", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-service", webhook.MutateService)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.Nil(t, response.Response.Patch, "Inactive intercept should not patch service")
}

func TestServiceMutationWebhook_MutateService_DeletedIntercept(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an intercept being deleted
	now := metav1.Now()
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-intercept",
			Namespace:         "test-namespace",
			DeletionTimestamp: &now, // Being deleted
			Finalizers:        []string{"intercepts.kloudlite.io/finalizer"},
			Labels: map[string]string{
				"intercepts.kloudlite.io/service-name": "test-service",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:            "Active",
			WorkspacePodName: "workspace-pod-123",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test-app",
			},
		},
	}

	serviceBytes, _ := json.Marshal(service)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: serviceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-service", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-service", webhook.MutateService)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.Nil(t, response.Response.Patch, "Deleted intercept should not patch service")
}

func TestServiceMutationWebhook_MutateService_MultiplePortMappings(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an active intercept with multiple port mappings
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"intercepts.kloudlite.io/service-name": "test-service",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					ServicePort:   443,
					WorkspacePort: 3443,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:            "Active",
			WorkspacePodName: "workspace-pod-123",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test-app",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "https",
					Port:       443,
					TargetPort: intstr.FromInt(8443),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	serviceBytes, _ := json.Marshal(service)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: serviceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-service", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-service", webhook.MutateService)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.NotNil(t, response.Response.Patch)

	// Verify patches include updated ports
	var patches []patchOperation
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)

	foundPortsPatch := false
	for _, patch := range patches {
		if patch.Path == "/spec/ports" {
			foundPortsPatch = true
			ports := patch.Value.([]interface{})
			assert.Equal(t, 2, len(ports), "Should have both ports")
		}
	}

	assert.True(t, foundPortsPatch, "Should patch service ports with multiple mappings")
}

func TestServiceMutationWebhook_MutateService_WithExistingAnnotations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an active intercept
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"intercepts.kloudlite.io/service-name": "test-service",
			},
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			Status: "active",
			WorkspaceRef: corev1.ObjectReference{
				Name:      "test-workspace",
				Namespace: "workspace-namespace",
			},
			PortMappings: []interceptsv1.PortMapping{
				{
					ServicePort:   80,
					WorkspacePort: 3000,
					Protocol:      corev1.ProtocolTCP,
				},
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase:            "Active",
			WorkspacePodName: "workspace-pod-123",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	// Service with existing annotations
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "test-namespace",
			Annotations: map[string]string{
				"existing": "annotation",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "test-app",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
		},
	}

	serviceBytes, _ := json.Marshal(service)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: serviceBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-service", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-service", webhook.MutateService)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.NotNil(t, response.Response.Patch)

	// Verify annotation patch
	var patches []patchOperation
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)

	foundAnnotationPatch := false
	for _, patch := range patches {
		if patch.Path == "/metadata/annotations/intercepts.kloudlite.io~1intercepted-by" {
			foundAnnotationPatch = true
			assert.Equal(t, "test-intercept", patch.Value)
		}
	}

	assert.True(t, foundAnnotationPatch, "Should add intercept annotation to existing annotations")
}

func TestServiceMutationWebhook_MutateService_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewServiceMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/mutate-service", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-service", webhook.MutateService)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
