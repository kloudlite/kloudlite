package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/v2/api/internal/managers"
	"github.com/kloudlite/kloudlite/v2/api/internal/repository"
	machinesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/machines/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/platform/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupMachineTypeHandlerTest() (*MachineTypeHandlers, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	mtRepo := repository.NewMachineTypeRepository(k8sClient)

	mgr := &managers.Manager{
		MachineTypeRepository: mtRepo,
	}

	handlers := NewMachineTypeHandlers(mgr)
	router := gin.New()

	return handlers, router
}

func TestListMachineTypes(t *testing.T) {
	t.Run("should list all machine types", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create test machine types
		mt1 := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "type-1",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Type 1",
				Category:    "general",
				Active:      true,
			},
		}
		mt2 := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "type-2",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Type 2",
				Category:    "compute-optimized",
				Active:      false,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt1)
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt2)

		router.GET("/machine-types", handlers.ListMachineTypes)

		req := httptest.NewRequest(http.MethodGet, "/machine-types", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(2), response["count"])
	})

	t.Run("should list active machine types only", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create test machine types
		mt1 := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "type-active",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Active Type",
				Category:    "general",
				Active:      true,
			},
		}
		mt2 := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "type-inactive",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Inactive Type",
				Category:    "general",
				Active:      false,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt1)
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt2)

		router.GET("/machine-types", handlers.ListMachineTypes)

		req := httptest.NewRequest(http.MethodGet, "/machine-types?active=true", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["count"])
	})

	t.Run("should list machine types by category", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create test machine types
		mt1 := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "type-general",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "General Type",
				Category:    "general",
				Active:      true,
			},
		}
		mt2 := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "type-compute",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Compute Type",
				Category:    "compute-optimized",
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt1)
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt2)

		router.GET("/machine-types", handlers.ListMachineTypes)

		req := httptest.NewRequest(http.MethodGet, "/machine-types?category=general", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["count"])
	})
}

func TestGetMachineType(t *testing.T) {
	t.Run("should get machine type", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create test machine type
		mt := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Test Type",
				Category:    "general",
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt)

		router.GET("/machine-types/:name", handlers.GetMachineType)

		req := httptest.NewRequest(http.MethodGet, "/machine-types/test-type", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response machinesv1.MachineType
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "test-type", response.Name)
	})

	t.Run("should return 404 for non-existent machine type", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.GET("/machine-types/:name", handlers.GetMachineType)

		req := httptest.NewRequest(http.MethodGet, "/machine-types/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCreateMachineType(t *testing.T) {
	t.Run("should create machine type with auth", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.POST("/machine-types", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.CreateMachineType)

		reqBody := MachineTypeCreateRequest{
			Name: "new-type",
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "New Type",
				Category:    "general",
				Resources: machinesv1.MachineResources{
					CPU:    "4",
					Memory: "8Gi",
				},
				Active: true,
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/machine-types", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("should return 401 when not authenticated", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.POST("/machine-types", handlers.CreateMachineType)

		reqBody := MachineTypeCreateRequest{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/machine-types", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.POST("/machine-types", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.CreateMachineType)

		req := httptest.NewRequest(http.MethodPost, "/machine-types", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateMachineType(t *testing.T) {
	t.Run("should update machine type", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create test machine type
		mt := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "update-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Original",
				Category:    "general",
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt)

		router.PUT("/machine-types/:name", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.UpdateMachineType)

		reqBody := MachineTypeUpdateRequest{
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Updated",
				Category:    "general",
				Active:      false,
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/machine-types/update-type", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 401 when not authenticated", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create test machine type
		mt := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "update-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Original",
				Category:    "general",
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt)

		router.PUT("/machine-types/:name", handlers.UpdateMachineType)

		reqBody := MachineTypeUpdateRequest{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/machine-types/update-type", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestDeleteMachineType(t *testing.T) {
	t.Run("should delete machine type", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create test machine type
		mt := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "delete-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Delete Me",
				Category:    "general",
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt)

		router.DELETE("/machine-types/:name", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.DeleteMachineType)

		req := httptest.NewRequest(http.MethodDelete, "/machine-types/delete-type", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 401 when not authenticated", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create test machine type
		mt := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "delete-type-unauth",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Delete Me",
				Category:    "general",
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt)

		router.DELETE("/machine-types/:name", handlers.DeleteMachineType)

		req := httptest.NewRequest(http.MethodDelete, "/machine-types/delete-type", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestActivateMachineType(t *testing.T) {
	t.Run("should activate inactive machine type", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create inactive machine type
		mt := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "inactive-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Inactive Type",
				Active:      false,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt)

		router.POST("/machine-types/:name/activate", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.ActivateMachineType)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/inactive-type/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify machine type was activated
		updated, _ := handlers.manager.MachineTypeRepository.Get(context.Background(), "inactive-type")
		assert.True(t, updated.Spec.Active)
	})

	t.Run("should return 401 when not authenticated", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.POST("/machine-types/:name/activate", handlers.ActivateMachineType)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/test-type/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 for non-existent machine type", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.POST("/machine-types/:name/activate", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.ActivateMachineType)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/nonexistent/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDeactivateMachineType(t *testing.T) {
	t.Run("should deactivate active machine type", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create active machine type
		mt := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "active-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Active Type",
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt)

		router.POST("/machine-types/:name/deactivate", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.DeactivateMachineType)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/active-type/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify machine type was deactivated
		updated, _ := handlers.manager.MachineTypeRepository.Get(context.Background(), "active-type")
		assert.False(t, updated.Spec.Active)
	})

	t.Run("should return 401 when not authenticated", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.POST("/machine-types/:name/deactivate", handlers.DeactivateMachineType)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/test-type/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 for non-existent machine type", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.POST("/machine-types/:name/deactivate", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.DeactivateMachineType)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/nonexistent/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestToggleMachineTypeActive(t *testing.T) {
	t.Run("should toggle active machine type to inactive", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create active machine type
		mt := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "toggle-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Toggle Type",
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt)

		router.POST("/machine-types/:name/toggle", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.ToggleMachineTypeActive)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/toggle-type/toggle", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify machine type was toggled to inactive
		updated, _ := handlers.manager.MachineTypeRepository.Get(context.Background(), "toggle-type")
		assert.False(t, updated.Spec.Active)
	})

	t.Run("should toggle inactive machine type to active", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()

		// Create inactive machine type
		mt := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "toggle-type-2",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Toggle Type 2",
				Active:      false,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), mt)

		router.POST("/machine-types/:name/toggle", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.ToggleMachineTypeActive)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/toggle-type-2/toggle", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify machine type was toggled to active
		updated, _ := handlers.manager.MachineTypeRepository.Get(context.Background(), "toggle-type-2")
		assert.True(t, updated.Spec.Active)
	})

	t.Run("should return 401 when not authenticated", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.POST("/machine-types/:name/toggle", handlers.ToggleMachineTypeActive)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/test-type/toggle", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 for non-existent machine type", func(t *testing.T) {
		handlers, router := setupMachineTypeHandlerTest()
		router.POST("/machine-types/:name/toggle", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.ToggleMachineTypeActive)

		req := httptest.NewRequest(http.MethodPost, "/machine-types/nonexistent/toggle", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
