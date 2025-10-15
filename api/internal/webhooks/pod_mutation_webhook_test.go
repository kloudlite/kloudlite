package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPodMutationWebhook_MutatePod_NoIntercept(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewPodMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: "nginx",
				},
			},
		},
	}

	podBytes, _ := json.Marshal(pod)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: podBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-pod", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-pod", webhook.MutatePod)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.Nil(t, response.Response.Patch, "Pod should not be patched when no intercept exists")
}

func TestPodMutationWebhook_MutatePod_SkipWorkspacePods(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an active intercept
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase: "Active",
			OriginalServiceSelector: map[string]string{
				"app": "test-app",
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewPodMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	// Workspace pod with workspace label
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workspace-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"workspaces.kloudlite.io/workspace-name": "my-workspace",
				"app": "test-app",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "workspace",
					Image: "workspace:latest",
				},
			},
		},
	}

	podBytes, _ := json.Marshal(pod)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: podBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-pod", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-pod", webhook.MutatePod)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.Nil(t, response.Response.Patch, "Workspace pods should not be patched")
}

func TestPodMutationWebhook_MutatePod_HoldMatchingPod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an active intercept
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			
			ServiceRef: corev1.ObjectReference{
				Name:      "test-service",
				Namespace: "test-namespace",
			},
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase: "Active",
			OriginalServiceSelector: map[string]string{
				"app": "test-app",
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewPodMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	// Pod matching the intercept's selector
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: "nginx",
				},
			},
		},
	}

	podBytes, _ := json.Marshal(pod)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: podBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-pod", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-pod", webhook.MutatePod)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.NotNil(t, response.Response.Patch, "Matching pod should be patched")

	// Verify patches
	var patches []map[string]interface{}
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)
	assert.NotEmpty(t, patches)

	// Check for node selector patch
	foundNodeSelector := false
	foundAnnotation := false
	for _, patch := range patches {
		if path, ok := patch["path"].(string); ok {
			if path == "/spec/nodeSelector" || path == "/spec/nodeSelector/kloudlite.io~1intercept-hold" {
				foundNodeSelector = true
				if value, ok := patch["value"].(string); ok {
					assert.Equal(t, "non-existing", value)
				}
			}
			if path == "/metadata/annotations/intercepts.kloudlite.io~1held-by" {
				foundAnnotation = true
				assert.Equal(t, "test-intercept", patch["value"])
			}
		}
	}

	assert.True(t, foundNodeSelector, "Should add node selector to hold pod")
	assert.True(t, foundAnnotation, "Should add annotation to track intercept")
}

func TestPodMutationWebhook_MutatePod_SkipInactiveIntercept(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an inactive intercept
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase: "Inactive",
			OriginalServiceSelector: map[string]string{
				"app": "test-app",
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewPodMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: "nginx",
				},
			},
		},
	}

	podBytes, _ := json.Marshal(pod)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: podBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-pod", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-pod", webhook.MutatePod)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.Nil(t, response.Response.Patch, "Inactive intercepts should not hold pods")
}

func TestPodMutationWebhook_MutatePod_SkipDeletedIntercept(t *testing.T) {
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
			Finalizers:        []string{"intercepts.kloudlite.io/finalizer"}, // Required for deletion
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase: "Active",
			OriginalServiceSelector: map[string]string{
				"app": "test-app",
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewPodMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: "nginx",
				},
			},
		},
	}

	podBytes, _ := json.Marshal(pod)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: podBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-pod", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-pod", webhook.MutatePod)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.Nil(t, response.Response.Patch, "Deleted intercepts should not hold pods")
}

func TestPodMutationWebhook_MutatePod_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewPodMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	body := []byte(`{invalid json}`)
	req, _ := http.NewRequest("POST", "/mutate-pod", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-pod", webhook.MutatePod)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPodMutationWebhook_MutatePod_WithExistingNodeSelector(t *testing.T) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	// Create an active intercept
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-intercept",
			Namespace: "test-namespace",
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			
		},
		Status: interceptsv1.ServiceInterceptStatus{
			Phase: "Active",
			OriginalServiceSelector: map[string]string{
				"app": "test-app",
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(intercept).Build()

	zapLogger, _ := zap.NewDevelopment()
	webhook := NewPodMutationWebhook(logger.NewZapLogger(zapLogger), k8sClient)

	// Pod with existing node selector
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"app": "test-app",
			},
		},
		Spec: corev1.PodSpec{
			NodeSelector: map[string]string{
				"existing": "selector",
			},
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: "nginx",
				},
			},
		},
	}

	podBytes, _ := json.Marshal(pod)
	admissionReview := admissionv1.AdmissionReview{
		Request: &admissionv1.AdmissionRequest{
			UID:       "test-uid",
			Operation: admissionv1.Create,
			Namespace: "test-namespace",
			Object: runtime.RawExtension{
				Raw: podBytes,
			},
		},
	}

	body, _ := json.Marshal(admissionReview)
	req, _ := http.NewRequest("POST", "/mutate-pod", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/mutate-pod", webhook.MutatePod)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response admissionv1.AdmissionReview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Response.Allowed)
	assert.NotNil(t, response.Response.Patch)

	// Verify the patch adds to existing node selector
	var patches []map[string]interface{}
	err = json.Unmarshal(response.Response.Patch, &patches)
	assert.NoError(t, err)

	foundHoldSelector := false
	for _, patch := range patches {
		if path, ok := patch["path"].(string); ok {
			if path == "/spec/nodeSelector/kloudlite.io~1intercept-hold" {
				foundHoldSelector = true
				assert.Equal(t, "non-existing", patch["value"])
			}
		}
	}

	assert.True(t, foundHoldSelector, "Should add hold selector to existing node selector")
}
