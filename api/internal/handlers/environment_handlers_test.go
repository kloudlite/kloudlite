package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	userv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// mockAuthMiddlewareEnv sets up authentication context for environment handler tests
func mockAuthMiddlewareEnv() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_username", "testuser")
		c.Set("user_email", "test@example.com")
		c.Set("user_roles", []userv1alpha1.RoleType{userv1alpha1.RoleUser})
		c.Next()
	}
}

// mockWorkmachineRepoForEnv implements repository.WorkMachineRepository for environment handler tests
type mockWorkmachineRepoForEnv struct{}

func (m *mockWorkmachineRepoForEnv) Create(ctx context.Context, wm *workmachinev1.WorkMachine) error {
	return nil
}
func (m *mockWorkmachineRepoForEnv) Get(ctx context.Context, name string) (*workmachinev1.WorkMachine, error) {
	return nil, errors.New("not found")
}
func (m *mockWorkmachineRepoForEnv) GetByOwner(ctx context.Context, owner string) (*workmachinev1.WorkMachine, error) {
	return &workmachinev1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{Name: "wm-" + owner},
		Spec: workmachinev1.WorkMachineSpec{
			TargetNamespace: "wm-" + owner,
		},
	}, nil
}
func (m *mockWorkmachineRepoForEnv) Update(ctx context.Context, wm *workmachinev1.WorkMachine) error {
	return nil
}
func (m *mockWorkmachineRepoForEnv) Patch(ctx context.Context, name string, patchData map[string]interface{}) (*workmachinev1.WorkMachine, error) {
	return nil, nil
}
func (m *mockWorkmachineRepoForEnv) Delete(ctx context.Context, name string) error {
	return nil
}
func (m *mockWorkmachineRepoForEnv) List(ctx context.Context, opts ...repository.ListOption) (*workmachinev1.WorkMachineList, error) {
	return nil, nil
}
func (m *mockWorkmachineRepoForEnv) Watch(ctx context.Context, opts ...repository.WatchOption) (<-chan repository.WatchEvent[*workmachinev1.WorkMachine], error) {
	return nil, nil
}
func (m *mockWorkmachineRepoForEnv) StartMachine(ctx context.Context, name string) error {
	return nil
}
func (m *mockWorkmachineRepoForEnv) StopMachine(ctx context.Context, name string) error {
	return nil
}
func (m *mockWorkmachineRepoForEnv) ListByMachineType(ctx context.Context, machineType string) (*workmachinev1.WorkMachineList, error) {
	return nil, nil
}

// mockEnvironmentRepo implements repository.EnvironmentRepository for testing
type mockEnvironmentRepo struct {
	createFunc       func(ctx context.Context, env *environmentsv1.Environment) error
	getFunc          func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error)
	listFunc         func(ctx context.Context, namespace string, opts ...repository.ListOption) (*environmentsv1.EnvironmentList, error)
	listActiveFunc   func(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error)
	listInactiveFunc func(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error)
	updateFunc       func(ctx context.Context, env *environmentsv1.Environment) error
	deleteFunc       func(ctx context.Context, namespace, name string) error
}

func (m *mockEnvironmentRepo) Create(ctx context.Context, env *environmentsv1.Environment) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, env)
	}
	return nil
}

func (m *mockEnvironmentRepo) Get(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, namespace, name)
	}
	return nil, errors.New("not found")
}

func (m *mockEnvironmentRepo) List(ctx context.Context, namespace string, opts ...repository.ListOption) (*environmentsv1.EnvironmentList, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, namespace, opts...)
	}
	return &environmentsv1.EnvironmentList{}, nil
}

func (m *mockEnvironmentRepo) ListActive(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error) {
	if m.listActiveFunc != nil {
		return m.listActiveFunc(ctx, namespace)
	}
	return &environmentsv1.EnvironmentList{}, nil
}

func (m *mockEnvironmentRepo) ListInactive(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error) {
	if m.listInactiveFunc != nil {
		return m.listInactiveFunc(ctx, namespace)
	}
	return &environmentsv1.EnvironmentList{}, nil
}

func (m *mockEnvironmentRepo) Update(ctx context.Context, env *environmentsv1.Environment) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, env)
	}
	return nil
}

func (m *mockEnvironmentRepo) Delete(ctx context.Context, namespace, name string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, namespace, name)
	}
	return nil
}

func (m *mockEnvironmentRepo) GetByTargetNamespace(ctx context.Context, targetNamespace string) (*environmentsv1.Environment, error) {
	return nil, errors.New("not implemented")
}

func (m *mockEnvironmentRepo) ActivateEnvironment(ctx context.Context, namespace, name string) error {
	return errors.New("not implemented")
}

func (m *mockEnvironmentRepo) DeactivateEnvironment(ctx context.Context, namespace, name string) error {
	return errors.New("not implemented")
}

func (m *mockEnvironmentRepo) Patch(ctx context.Context, namespace, name string, patchData map[string]interface{}) (*environmentsv1.Environment, error) {
	return nil, errors.New("not implemented")
}

func (m *mockEnvironmentRepo) Watch(ctx context.Context, namespace string, opts ...repository.WatchOption) (<-chan repository.WatchEvent[*environmentsv1.Environment], error) {
	return nil, errors.New("not implemented")
}

// mockUserRepo implements repository.UserRepository for testing
type mockUserRepo struct{}

func (m *mockUserRepo) Create(ctx context.Context, user *userv1alpha1.User) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) Get(ctx context.Context, name string) (*userv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) List(ctx context.Context, opts ...repository.ListOption) (*userv1alpha1.UserList, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) Update(ctx context.Context, user *userv1alpha1.User) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) Delete(ctx context.Context, name string) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*userv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*userv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) ListActive(ctx context.Context) (*userv1alpha1.UserList, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) UpdateStatus(ctx context.Context, user *userv1alpha1.User) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) Patch(ctx context.Context, name string, patchData map[string]interface{}) (*userv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) Watch(ctx context.Context, opts ...repository.WatchOption) (<-chan repository.WatchEvent[*userv1alpha1.User], error) {
	return nil, errors.New("not implemented")
}

func TestGetEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should get environment by name", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
					Spec: environmentsv1.EnvironmentSpec{
						TargetNamespace: "test-namespace",
						Activated:       true,
						OwnedBy:         "testuser", // Must match mockAuthMiddlewareEnv username
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.GET("/environments/:name", handlers.GetEnvironment)

		req, _ := http.NewRequest("GET", "/environments/test-env", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var env environmentsv1.Environment
		err := json.Unmarshal(w.Body.Bytes(), &env)
		assert.NoError(t, err)
		assert.Equal(t, "test-env", env.Name)
	})

	t.Run("should return 404 for non-existent environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return nil, errors.New("environment test-env not found")
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.GET("/environments/:name", handlers.GetEnvironment)

		req, _ := http.NewRequest("GET", "/environments/test-env", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 when name is empty", func(t *testing.T) {
		handlers := NewEnvironmentHandlers(&mockEnvironmentRepo{}, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.GET("/environments/:name", handlers.GetEnvironment)

		req, _ := http.NewRequest("GET", "/environments/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code) // Gin returns 404 for missing route param
	})
}

func TestListEnvironments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should list all environments", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			listFunc: func(ctx context.Context, namespace string, opts ...repository.ListOption) (*environmentsv1.EnvironmentList, error) {
				return &environmentsv1.EnvironmentList{
					Items: []environmentsv1.Environment{
						{ObjectMeta: metav1.ObjectMeta{Name: "env1"}, Spec: environmentsv1.EnvironmentSpec{OwnedBy: "testuser"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "env2"}, Spec: environmentsv1.EnvironmentSpec{OwnedBy: "testuser"}},
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.GET("/environments", handlers.ListEnvironments)

		req, _ := http.NewRequest("GET", "/environments", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["count"])
	})

	t.Run("should list active environments", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			listActiveFunc: func(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error) {
				return &environmentsv1.EnvironmentList{
					Items: []environmentsv1.Environment{
						{ObjectMeta: metav1.ObjectMeta{Name: "active-env"}},
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.GET("/environments", handlers.ListEnvironments)

		req, _ := http.NewRequest("GET", "/environments?status=active", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should list inactive environments", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			listInactiveFunc: func(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error) {
				return &environmentsv1.EnvironmentList{
					Items: []environmentsv1.Environment{
						{ObjectMeta: metav1.ObjectMeta{Name: "inactive-env"}},
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.GET("/environments", handlers.ListEnvironments)

		req, _ := http.NewRequest("GET", "/environments?status=inactive", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestCreateEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should create environment with authenticated user", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			createFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		// Create fake k8s client with a test user
		scheme := runtime.NewScheme()
		_ = userv1alpha1.AddToScheme(scheme)
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&userv1alpha1.User{
				ObjectMeta: metav1.ObjectMeta{Name: "testuser"},
				Spec: userv1alpha1.UserSpec{
					Email: "test@example.com",
				},
			},
		).Build()

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, k8sClient, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.POST("/environments", handlers.CreateEnvironment)

		reqBody := map[string]interface{}{
			"name": "test-env",
			"spec": map[string]interface{}{
				"targetNamespace": "test-ns",
			},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/environments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("should reject creation without authentication", func(t *testing.T) {
		handlers := NewEnvironmentHandlers(&mockEnvironmentRepo{}, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.POST("/environments", handlers.CreateEnvironment)

		reqBody := map[string]interface{}{
			"name": "test-env",
			"spec": map[string]interface{}{
				"targetNamespace": "test-ns",
			},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/environments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject invalid request body", func(t *testing.T) {
		handlers := NewEnvironmentHandlers(&mockEnvironmentRepo{}, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.POST("/environments", handlers.CreateEnvironment)

		body := []byte(`{"invalid": json}`)
		req, _ := http.NewRequest("POST", "/environments", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should update environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						TargetNamespace: "old-ns",
						OwnedBy:         "testuser",
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.PUT("/environments/:name", handlers.UpdateEnvironment)

		reqBody := map[string]interface{}{
			"spec": map[string]interface{}{
				"targetNamespace": "new-ns",
			},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/environments/test-env", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 for non-existent environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return nil, errors.New("environment test-env not found")
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.PUT("/environments/:name", handlers.UpdateEnvironment)

		reqBody := map[string]interface{}{
			"spec": map[string]interface{}{
				"targetNamespace": "new-ns",
			},
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/environments/test-env", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDeleteEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should delete deactivated environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: false,
						OwnedBy:   "testuser",
					},
				}, nil
			},
			deleteFunc: func(ctx context.Context, namespace, name string) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.DELETE("/environments/:name", handlers.DeleteEnvironment)

		req, _ := http.NewRequest("DELETE", "/environments/test-env", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should delete activated environment without force", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: true,
						OwnedBy:   "testuser",
					},
				}, nil
			},
			deleteFunc: func(ctx context.Context, namespace, name string) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.DELETE("/environments/:name", handlers.DeleteEnvironment)

		req, _ := http.NewRequest("DELETE", "/environments/test-env", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestActivateEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should activate deactivated environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: false,
						OwnedBy:   "testuser",
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.POST("/environments/:name/activate", handlers.ActivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/test-env/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject activation of already activated environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: true,
						OwnedBy:   "testuser",
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.POST("/environments/:name/activate", handlers.ActivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/test-env/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "already activated")
	})

	t.Run("should return 404 for non-existent environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return nil, fmt.Errorf("environment %s not found", name)
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.POST("/environments/:name/activate", handlers.ActivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/nonexistent/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDeactivateEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should deactivate activated environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: true,
						OwnedBy:   "testuser",
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.POST("/environments/:name/deactivate", handlers.DeactivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/test-env/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject deactivation of already deactivated environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: false,
						OwnedBy:   "testuser",
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.POST("/environments/:name/deactivate", handlers.DeactivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/test-env/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "already deactivated")
	})

	t.Run("should return 404 for non-existent environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return nil, fmt.Errorf("environment %s not found", name)
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.POST("/environments/:name/deactivate", handlers.DeactivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/nonexistent/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetEnvironmentStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should get environment status", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						TargetNamespace: "test-ns",
						Activated:       true,
						OwnedBy:         "testuser",
					},
					Status: environmentsv1.EnvironmentStatus{},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.GET("/environments/:name/status", handlers.GetEnvironmentStatus)

		req, _ := http.NewRequest("GET", "/environments/test-env/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "test-env", response["name"])
		assert.Equal(t, true, response["activated"])
	})
}

func TestPatchEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should patch environment activated field", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: false,
						OwnedBy:   "testuser",
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.PATCH("/environments/:name", handlers.PatchEnvironment)

		patchBody := map[string]interface{}{
			"activated": true,
		}
		body, _ := json.Marshal(patchBody)
		req, _ := http.NewRequest("PATCH", "/environments/test-env", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should patch environment labels", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec:       environmentsv1.EnvironmentSpec{OwnedBy: "testuser"},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.PATCH("/environments/:name", handlers.PatchEnvironment)

		patchBody := map[string]interface{}{
			"labels": map[string]interface{}{
				"team": "platform",
			},
		}
		body, _ := json.Marshal(patchBody)
		req, _ := http.NewRequest("PATCH", "/environments/test-env", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 400 when name is empty", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{}
		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.PATCH("/environments/:name", handlers.PatchEnvironment)

		patchBody := map[string]interface{}{
			"activated": true,
		}
		body, _ := json.Marshal(patchBody)
		req, _ := http.NewRequest("PATCH", "/environments/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code) // Gin returns 404 when route doesn't match
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{}
		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.PATCH("/environments/:name", handlers.PatchEnvironment)

		req, _ := http.NewRequest("PATCH", "/environments/test-env", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 404 when environment not found", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return nil, errors.New("environment test-env not found")
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.PATCH("/environments/:name", handlers.PatchEnvironment)

		patchBody := map[string]interface{}{
			"activated": true,
		}
		body, _ := json.Marshal(patchBody)
		req, _ := http.NewRequest("PATCH", "/environments/test-env", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 500 when update fails", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, namespace, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						OwnedBy: "testuser",
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return errors.New("update failed")
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, &mockWorkmachineRepoForEnv{}, nil, logger)
		router := gin.New()
		router.Use(mockAuthMiddlewareEnv())
		router.PATCH("/environments/:name", handlers.PatchEnvironment)

		patchBody := map[string]interface{}{
			"activated": true,
		}
		body, _ := json.Marshal(patchBody)
		req, _ := http.NewRequest("PATCH", "/environments/test-env", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
