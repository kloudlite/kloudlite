package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupCompositionHandlerTest() (*CompositionHandlers, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	compRepo := repository.NewCompositionRepository(k8sClient)
	logger, _ := zap.NewDevelopment()

	handlers := NewCompositionHandlers(compRepo, k8sClient, logger)
	router := gin.New()

	return handlers, router
}

func TestCreateComposition(t *testing.T) {
	t.Run("should create composition", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()
		router.POST("/namespaces/:namespace/compositions", handlers.CreateComposition)

		reqBody := struct {
			Name string                         `json:"name"`
			Spec environmentsv1.CompositionSpec `json:"spec"`
		}{
			Name: "test-comp",
			Spec: environmentsv1.CompositionSpec{
				DisplayName:    "Test Composition",
				ComposeContent: "version: '3.8'",
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/compositions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("should return 400 when namespace is empty", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()
		router.POST("/namespaces/:namespace/compositions", handlers.CreateComposition)

		req := httptest.NewRequest(http.MethodPost, "/namespaces//compositions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()
		router.POST("/namespaces/:namespace/compositions", handlers.CreateComposition)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/compositions", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetComposition(t *testing.T) {
	t.Run("should get composition", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()

		// Create test composition
		comp := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-comp",
				Namespace: "test-ns",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Test Composition",
			},
		}
		_ = handlers.compRepo.Create(context.Background(), comp)

		router.GET("/namespaces/:namespace/compositions/:name", handlers.GetComposition)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/compositions/test-comp", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 when route not matched", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()
		router.GET("/namespaces/:namespace/compositions/:name", handlers.GetComposition)

		req := httptest.NewRequest(http.MethodGet, "/namespaces//compositions/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestListCompositions(t *testing.T) {
	t.Run("should list all compositions", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()

		// Create test compositions
		comp1 := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "comp-1",
				Namespace: "test-ns",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Composition 1",
			},
		}
		comp2 := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "comp-2",
				Namespace: "test-ns",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Composition 2",
			},
			Status: environmentsv1.CompositionStatus{
				State: environmentsv1.CompositionStateRunning,
			},
		}
		_ = handlers.compRepo.Create(context.Background(), comp1)
		_ = handlers.compRepo.Create(context.Background(), comp2)

		router.GET("/namespaces/:namespace/compositions", handlers.ListCompositions)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/compositions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(2), response["count"])
	})

	t.Run("should list compositions by state", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()

		comp := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "comp-running",
				Namespace: "test-ns",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Running Composition",
			},
			Status: environmentsv1.CompositionStatus{
				State: environmentsv1.CompositionStateRunning,
			},
		}
		_ = handlers.compRepo.Create(context.Background(), comp)

		router.GET("/namespaces/:namespace/compositions", handlers.ListCompositions)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/compositions?state=running", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 400 when namespace is empty", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()
		router.GET("/namespaces/:namespace/compositions", handlers.ListCompositions)

		req := httptest.NewRequest(http.MethodGet, "/namespaces//compositions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateComposition(t *testing.T) {
	t.Run("should update composition", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()

		// Create test composition
		comp := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "update-comp",
				Namespace: "test-ns",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Original",
			},
		}
		_ = handlers.compRepo.Create(context.Background(), comp)

		router.PUT("/namespaces/:namespace/compositions/:name", handlers.UpdateComposition)

		reqBody := struct {
			Spec environmentsv1.CompositionSpec `json:"spec"`
		}{
			Spec: environmentsv1.CompositionSpec{
				DisplayName:    "Updated",
				ComposeContent: "version: '3.8'",
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/namespaces/test-ns/compositions/update-comp", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 when route not matched", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()
		router.PUT("/namespaces/:namespace/compositions/:name", handlers.UpdateComposition)

		req := httptest.NewRequest(http.MethodPut, "/namespaces//compositions/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()
		router.PUT("/namespaces/:namespace/compositions/:name", handlers.UpdateComposition)

		req := httptest.NewRequest(http.MethodPut, "/namespaces/test-ns/compositions/update-comp", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDeleteComposition(t *testing.T) {
	t.Run("should delete composition", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()

		// Create test composition
		comp := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "delete-comp",
				Namespace: "test-ns",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Delete Me",
			},
		}
		_ = handlers.compRepo.Create(context.Background(), comp)

		router.DELETE("/namespaces/:namespace/compositions/:name", handlers.DeleteComposition)

		req := httptest.NewRequest(http.MethodDelete, "/namespaces/test-ns/compositions/delete-comp", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 when route not matched", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()
		router.DELETE("/namespaces/:namespace/compositions/:name", handlers.DeleteComposition)

		req := httptest.NewRequest(http.MethodDelete, "/namespaces//compositions/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetCompositionStatus(t *testing.T) {
	t.Run("should get composition status", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()

		// Create test composition
		comp := &environmentsv1.Composition{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "status-comp",
				Namespace: "test-ns",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Status Composition",
			},
			Status: environmentsv1.CompositionStatus{
				State:   environmentsv1.CompositionStateRunning,
				Message: "Running smoothly",
			},
		}
		_ = handlers.compRepo.Create(context.Background(), comp)

		router.GET("/namespaces/:namespace/compositions/:name/status", handlers.GetCompositionStatus)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/compositions/status-comp/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "status-comp", response["name"])
		assert.Equal(t, "test-ns", response["namespace"])
	})

	t.Run("should return 400 when params are invalid", func(t *testing.T) {
		handlers, router := setupCompositionHandlerTest()
		router.GET("/namespaces/:namespace/compositions/:name/status", handlers.GetCompositionStatus)

		req := httptest.NewRequest(http.MethodGet, "/namespaces//compositions//status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Gin returns 400 for malformed URLs with empty params
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
