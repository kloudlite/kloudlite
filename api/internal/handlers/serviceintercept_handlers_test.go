package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupServiceInterceptHandlerTest() (*ServiceInterceptHandlers, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = interceptsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	logger, _ := zap.NewDevelopment()

	handlers := NewServiceInterceptHandlers(k8sClient, logger)
	router := gin.New()

	return handlers, router
}

func TestCreateServiceIntercept(t *testing.T) {
	t.Run("should create service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.POST("/api/v1/namespaces/:namespace/service-intercepts", handlers.CreateServiceIntercept)

		reqBody := struct {
			Name string                           `json:"name"`
			Spec interceptsv1.ServiceInterceptSpec `json:"spec"`
		}{
			Name: "test-intercept",
			Spec: interceptsv1.ServiceInterceptSpec{
				WorkspaceRef: corev1.ObjectReference{
					Name:      "test-workspace",
					Namespace: "workspace-ns",
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
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/test-ns/service-intercepts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("should return 400 when namespace is missing", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.POST("/api/v1/namespaces/:namespace/service-intercepts", handlers.CreateServiceIntercept)

		reqBody := struct {
			Name string                           `json:"name"`
			Spec interceptsv1.ServiceInterceptSpec `json:"spec"`
		}{
			Name: "test-intercept",
			Spec: interceptsv1.ServiceInterceptSpec{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces//service-intercepts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.POST("/api/v1/namespaces/:namespace/service-intercepts", handlers.CreateServiceIntercept)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/test-ns/service-intercepts", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 409 when service intercept already exists", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		// Create existing service intercept
		existing := &interceptsv1.ServiceIntercept{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existing-intercept",
				Namespace: "test-ns",
			},
			Spec: interceptsv1.ServiceInterceptSpec{
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
		_ = handlers.k8sClient.Create(context.Background(), existing)

		router.POST("/api/v1/namespaces/:namespace/service-intercepts", handlers.CreateServiceIntercept)

		reqBody := struct {
			Name string                           `json:"name"`
			Spec interceptsv1.ServiceInterceptSpec `json:"spec"`
		}{
			Name: "existing-intercept",
			Spec: interceptsv1.ServiceInterceptSpec{
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
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/test-ns/service-intercepts", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestGetServiceIntercept(t *testing.T) {
	t.Run("should get service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		// Create test service intercept
		intercept := &interceptsv1.ServiceIntercept{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-intercept",
				Namespace: "test-ns",
			},
			Spec: interceptsv1.ServiceInterceptSpec{
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
		_ = handlers.k8sClient.Create(context.Background(), intercept)

		router.GET("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.GetServiceIntercept)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/test-ns/service-intercepts/test-intercept", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response interceptsv1.ServiceIntercept
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "test-intercept", response.Name)
	})

	t.Run("should return 404 for non-existent service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.GET("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.GetServiceIntercept)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/test-ns/service-intercepts/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 when namespace is missing", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.GET("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.GetServiceIntercept)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces//service-intercepts/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestListServiceIntercepts(t *testing.T) {
	t.Run("should list all service intercepts", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		// Create test service intercepts
		intercept1 := &interceptsv1.ServiceIntercept{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "intercept-1",
				Namespace: "test-ns",
			},
			Spec: interceptsv1.ServiceInterceptSpec{
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
		intercept2 := &interceptsv1.ServiceIntercept{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "intercept-2",
				Namespace: "test-ns",
			},
			Spec: interceptsv1.ServiceInterceptSpec{
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
		_ = handlers.k8sClient.Create(context.Background(), intercept1)
		_ = handlers.k8sClient.Create(context.Background(), intercept2)

		router.GET("/api/v1/namespaces/:namespace/service-intercepts", handlers.ListServiceIntercepts)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/test-ns/service-intercepts", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(2), response["count"])
	})

	t.Run("should return empty list when no service intercepts exist", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.GET("/api/v1/namespaces/:namespace/service-intercepts", handlers.ListServiceIntercepts)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/empty-ns/service-intercepts", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(0), response["count"])
	})

	t.Run("should return 400 when namespace is missing", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.GET("/api/v1/namespaces/:namespace/service-intercepts", handlers.ListServiceIntercepts)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces//service-intercepts", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateServiceIntercept(t *testing.T) {
	t.Run("should update service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		// Create test service intercept
		intercept := &interceptsv1.ServiceIntercept{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "update-intercept",
				Namespace: "test-ns",
			},
			Spec: interceptsv1.ServiceInterceptSpec{
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
		_ = handlers.k8sClient.Create(context.Background(), intercept)

		router.PUT("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.UpdateServiceIntercept)

		reqBody := struct {
			Spec interceptsv1.ServiceInterceptSpec `json:"spec"`
		}{
			Spec: interceptsv1.ServiceInterceptSpec{
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
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/namespaces/test-ns/service-intercepts/update-intercept", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 for non-existent service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.PUT("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.UpdateServiceIntercept)

		reqBody := struct {
			Spec interceptsv1.ServiceInterceptSpec `json:"spec"`
		}{
			Spec: interceptsv1.ServiceInterceptSpec{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/namespaces/test-ns/service-intercepts/nonexistent", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.PUT("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.UpdateServiceIntercept)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/namespaces/test-ns/service-intercepts/test", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 when namespace is missing", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.PUT("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.UpdateServiceIntercept)

		reqBody := struct {
			Spec interceptsv1.ServiceInterceptSpec `json:"spec"`
		}{
			Spec: interceptsv1.ServiceInterceptSpec{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/namespaces//service-intercepts/test", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDeleteServiceIntercept(t *testing.T) {
	t.Run("should delete service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		// Create test service intercept
		intercept := &interceptsv1.ServiceIntercept{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "delete-intercept",
				Namespace: "test-ns",
			},
			Spec: interceptsv1.ServiceInterceptSpec{
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
		_ = handlers.k8sClient.Create(context.Background(), intercept)

		router.DELETE("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.DeleteServiceIntercept)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/test-ns/service-intercepts/delete-intercept", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 for non-existent service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.DELETE("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.DeleteServiceIntercept)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/test-ns/service-intercepts/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 when namespace is missing", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.DELETE("/api/v1/namespaces/:namespace/service-intercepts/:name", handlers.DeleteServiceIntercept)

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces//service-intercepts/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestActivateServiceIntercept(t *testing.T) {
	t.Run("should activate service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		// Create test service intercept
		intercept := &interceptsv1.ServiceIntercept{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "activate-intercept",
				Namespace: "test-ns",
			},
			Spec: interceptsv1.ServiceInterceptSpec{
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
		_ = handlers.k8sClient.Create(context.Background(), intercept)

		router.POST("/api/v1/namespaces/:namespace/service-intercepts/:name/activate", handlers.ActivateServiceIntercept)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/test-ns/service-intercepts/activate-intercept/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 for non-existent service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.POST("/api/v1/namespaces/:namespace/service-intercepts/:name/activate", handlers.ActivateServiceIntercept)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/test-ns/service-intercepts/nonexistent/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 when namespace is missing", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.POST("/api/v1/namespaces/:namespace/service-intercepts/:name/activate", handlers.ActivateServiceIntercept)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces//service-intercepts/test/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDeactivateServiceIntercept(t *testing.T) {
	t.Run("should deactivate service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		// Create test service intercept
		intercept := &interceptsv1.ServiceIntercept{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deactivate-intercept",
				Namespace: "test-ns",
			},
			Spec: interceptsv1.ServiceInterceptSpec{
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
		_ = handlers.k8sClient.Create(context.Background(), intercept)

		router.POST("/api/v1/namespaces/:namespace/service-intercepts/:name/deactivate", handlers.DeactivateServiceIntercept)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/test-ns/service-intercepts/deactivate-intercept/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 for non-existent service intercept", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.POST("/api/v1/namespaces/:namespace/service-intercepts/:name/deactivate", handlers.DeactivateServiceIntercept)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces/test-ns/service-intercepts/nonexistent/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 when namespace is missing", func(t *testing.T) {
		handlers, router := setupServiceInterceptHandlerTest()

		router.POST("/api/v1/namespaces/:namespace/service-intercepts/:name/deactivate", handlers.DeactivateServiceIntercept)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/namespaces//service-intercepts/test/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
