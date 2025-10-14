package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/managers"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupWorkMachineHandlerTest() (*WorkMachineHandlers, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	wmRepo := repository.NewWorkMachineRepository(k8sClient)
	mtRepo := repository.NewMachineTypeRepository(k8sClient)

	mgr := &managers.Manager{
		WorkMachineRepository: wmRepo,
		MachineTypeRepository: mtRepo,
	}

	handlers := NewWorkMachineHandlers(mgr)
	router := gin.New()

	return handlers, router
}

func TestGetMyWorkMachine(t *testing.T) {
	t.Run("should get user's work machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create test machine
		machine := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-machine",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "test-user",
				MachineType: "standard-4",
			},
		}
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine)

		router.GET("/my-machine", func(c *gin.Context) {
			c.Set("user_email", "test-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.GetMyWorkMachine)

		req := httptest.NewRequest(http.MethodGet, "/my-machine", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response machinesv1.WorkMachine
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test-user", response.Spec.OwnedBy)
	})

	t.Run("should return 401 when user not authenticated", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()
		router.GET("/my-machine", handlers.GetMyWorkMachine)

		req := httptest.NewRequest(http.MethodGet, "/my-machine", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 when user has no machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()
		router.GET("/my-machine", func(c *gin.Context) {
			c.Set("user_email", "no-machine-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.GetMyWorkMachine)

		req := httptest.NewRequest(http.MethodGet, "/my-machine", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestStartMyWorkMachine(t *testing.T) {
	t.Run("should start user's work machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create test machine
		machine := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-machine-start",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:      "test-user",
				MachineType:  "standard-4",
				DesiredState: machinesv1.MachineStateStopped,
			},
		}
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine)

		router.POST("/start", func(c *gin.Context) {
			c.Set("user_email", "test-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.StartMyWorkMachine)

		req := httptest.NewRequest(http.MethodPost, "/start", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify machine state changed
		updated, _ := handlers.manager.WorkMachineRepository.Get(context.Background(), "test-machine-start")
		assert.Equal(t, machinesv1.MachineStateRunning, updated.Spec.DesiredState)
	})

	t.Run("should return 401 when user not authenticated", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()
		router.POST("/start", handlers.StartMyWorkMachine)

		req := httptest.NewRequest(http.MethodPost, "/start", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 when user has no machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		router.POST("/start", func(c *gin.Context) {
			c.Set("user_email", "no-machine-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.StartMyWorkMachine)

		req := httptest.NewRequest(http.MethodPost, "/start", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestStopMyWorkMachine(t *testing.T) {
	t.Run("should stop user's work machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create test machine
		machine := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-machine-stop",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:      "test-user",
				MachineType:  "standard-4",
				DesiredState: machinesv1.MachineStateRunning,
			},
		}
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine)

		router.POST("/stop", func(c *gin.Context) {
			c.Set("user_email", "test-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.StopMyWorkMachine)

		req := httptest.NewRequest(http.MethodPost, "/stop", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify machine state changed
		updated, _ := handlers.manager.WorkMachineRepository.Get(context.Background(), "test-machine-stop")
		assert.Equal(t, machinesv1.MachineStateStopped, updated.Spec.DesiredState)
	})

	t.Run("should return 401 when user not authenticated", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()
		router.POST("/stop", handlers.StopMyWorkMachine)

		req := httptest.NewRequest(http.MethodPost, "/stop", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 when user has no machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		router.POST("/stop", func(c *gin.Context) {
			c.Set("user_email", "no-machine-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.StopMyWorkMachine)

		req := httptest.NewRequest(http.MethodPost, "/stop", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestCreateMyWorkMachine(t *testing.T) {
	t.Run("should create work machine with default type", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create default machine type
		defaultMT := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Default Type",
				IsDefault:   true,
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), defaultMT)

		router.POST("/create", func(c *gin.Context) {
			c.Set("user_email", "new-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.CreateMyWorkMachine)

		reqBody := WorkMachineCreateRequest{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("should return 401 when user not authenticated", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()
		router.POST("/create", handlers.CreateMyWorkMachine)

		reqBody := WorkMachineCreateRequest{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create default machine type needed by CreateMyWorkMachine
		defaultMT := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Default Type",
				IsDefault:   true,
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), defaultMT)

		router.POST("/create", func(c *gin.Context) {
			c.Set("user_email", "test-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.CreateMyWorkMachine)

		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 409 when user already has a machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create existing machine for user
		existingMachine := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "existing-machine",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "existing-user",
				MachineType: "standard-4",
			},
		}
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), existingMachine)

		router.POST("/create", func(c *gin.Context) {
			c.Set("user_email", "existing-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.CreateMyWorkMachine)

		reqBody := WorkMachineCreateRequest{
			MachineType: "standard-4",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("should return 400 when no machine type and no default", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Don't create any default machine type

		router.POST("/create", func(c *gin.Context) {
			c.Set("user_email", "new-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.CreateMyWorkMachine)

		reqBody := WorkMachineCreateRequest{
			// No MachineType specified
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateMyWorkMachine(t *testing.T) {
	t.Run("should update work machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create test machine
		machine := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-machine-update",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "test-user",
				MachineType: "standard-4",
			},
		}
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine)

		// Create new machine type
		newMT := &machinesv1.MachineType{
			ObjectMeta: metav1.ObjectMeta{
				Name: "large-8",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Large 8",
				Active:      true,
			},
		}
		_ = handlers.manager.MachineTypeRepository.Create(context.Background(), newMT)

		router.PUT("/update", func(c *gin.Context) {
			c.Set("user_email", "test-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.UpdateMyWorkMachine)

		reqBody := WorkMachineUpdateRequest{
			MachineType: "large-8",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/update", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify update
		updated, _ := handlers.manager.WorkMachineRepository.Get(context.Background(), "test-machine-update")
		assert.Equal(t, "large-8", updated.Spec.MachineType)
	})

	t.Run("should return 401 when user not authenticated", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()
		router.PUT("/update", handlers.UpdateMyWorkMachine)

		reqBody := WorkMachineUpdateRequest{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/update", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		router.PUT("/update", func(c *gin.Context) {
			c.Set("user_email", "test-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.UpdateMyWorkMachine)

		req := httptest.NewRequest(http.MethodPut, "/update", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 404 when user has no machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		router.PUT("/update", func(c *gin.Context) {
			c.Set("user_email", "no-machine-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.UpdateMyWorkMachine)

		reqBody := WorkMachineUpdateRequest{
			MachineType: "large-8",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/update", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDeleteMyWorkMachine(t *testing.T) {
	t.Run("should delete work machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create test machine
		machine := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-machine-delete",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:      "test-user",
				MachineType:  "standard-4",
				DesiredState: machinesv1.MachineStateStopped,
			},
		}
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine)

		router.DELETE("/delete", func(c *gin.Context) {
			c.Set("user_email", "test-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.DeleteMyWorkMachine)

		req := httptest.NewRequest(http.MethodDelete, "/delete", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 401 when user not authenticated", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()
		router.DELETE("/delete", handlers.DeleteMyWorkMachine)

		req := httptest.NewRequest(http.MethodDelete, "/delete", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 when user has no machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		router.DELETE("/delete", func(c *gin.Context) {
			c.Set("user_email", "no-machine-user")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.DeleteMyWorkMachine)

		req := httptest.NewRequest(http.MethodDelete, "/delete", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestListAllWorkMachines(t *testing.T) {
	t.Run("should list all work machines", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create test machines
		machine1 := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machine-1",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "user1",
				MachineType: "standard-4",
			},
		}
		machine2 := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machine-2",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "user2",
				MachineType: "large-8",
			},
		}
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine1)
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine2)

		router.GET("/machines", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.ListAllWorkMachines)

		req := httptest.NewRequest(http.MethodGet, "/machines", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(2), response["count"])
	})

	t.Run("should list work machines filtered by machine type", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create test machines
		machine1 := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machine-standard",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "user1",
				MachineType: "standard-4",
			},
		}
		machine2 := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machine-large",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "user2",
				MachineType: "large-8",
			},
		}
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine1)
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine2)

		router.GET("/machines", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.ListAllWorkMachines)

		req := httptest.NewRequest(http.MethodGet, "/machines?machineType=standard-4", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, float64(1), response["count"])
	})

	t.Run("should return 401 when user not authenticated", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()
		router.GET("/machines", handlers.ListAllWorkMachines)

		req := httptest.NewRequest(http.MethodGet, "/machines", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetWorkMachine(t *testing.T) {
	t.Run("should get work machine by name", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		// Create test machine
		machine := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-machine",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "test-user",
				MachineType: "standard-4",
			},
		}
		_ = handlers.manager.WorkMachineRepository.Create(context.Background(), machine)

		router.GET("/machines/:name", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.GetWorkMachine)

		req := httptest.NewRequest(http.MethodGet, "/machines/test-machine", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response machinesv1.WorkMachine
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "test-machine", response.Name)
	})

	t.Run("should return 404 for non-existent machine", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()

		router.GET("/machines/:name", func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		}, handlers.GetWorkMachine)

		req := httptest.NewRequest(http.MethodGet, "/machines/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 401 when user not authenticated", func(t *testing.T) {
		handlers, router := setupWorkMachineHandlerTest()
		router.GET("/machines/:name", handlers.GetWorkMachine)

		req := httptest.NewRequest(http.MethodGet, "/machines/test-machine", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
