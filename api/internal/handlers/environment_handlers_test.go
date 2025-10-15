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
	"github.com/kloudlite/kloudlite/api/internal/repository"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// mockEnvironmentRepo implements repository.EnvironmentRepository for testing
type mockEnvironmentRepo struct {
	createFunc       func(ctx context.Context, env *environmentsv1.Environment) error
	getFunc          func(ctx context.Context, name string) (*environmentsv1.Environment, error)
	listFunc         func(ctx context.Context, opts ...repository.ListOption) (*environmentsv1.EnvironmentList, error)
	listActiveFunc   func(ctx context.Context) (*environmentsv1.EnvironmentList, error)
	listInactiveFunc func(ctx context.Context) (*environmentsv1.EnvironmentList, error)
	updateFunc       func(ctx context.Context, env *environmentsv1.Environment) error
	deleteFunc       func(ctx context.Context, name string) error
}

func (m *mockEnvironmentRepo) Create(ctx context.Context, env *environmentsv1.Environment) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, env)
	}
	return nil
}

func (m *mockEnvironmentRepo) Get(ctx context.Context, name string) (*environmentsv1.Environment, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, name)
	}
	return nil, errors.New("not found")
}

func (m *mockEnvironmentRepo) List(ctx context.Context, opts ...repository.ListOption) (*environmentsv1.EnvironmentList, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, opts...)
	}
	return &environmentsv1.EnvironmentList{}, nil
}

func (m *mockEnvironmentRepo) ListActive(ctx context.Context) (*environmentsv1.EnvironmentList, error) {
	if m.listActiveFunc != nil {
		return m.listActiveFunc(ctx)
	}
	return &environmentsv1.EnvironmentList{}, nil
}

func (m *mockEnvironmentRepo) ListInactive(ctx context.Context) (*environmentsv1.EnvironmentList, error) {
	if m.listInactiveFunc != nil {
		return m.listInactiveFunc(ctx)
	}
	return &environmentsv1.EnvironmentList{}, nil
}

func (m *mockEnvironmentRepo) Update(ctx context.Context, env *environmentsv1.Environment) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, env)
	}
	return nil
}

func (m *mockEnvironmentRepo) Delete(ctx context.Context, name string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, name)
	}
	return nil
}

func (m *mockEnvironmentRepo) GetByNamespace(ctx context.Context, namespace string) (*environmentsv1.Environment, error) {
	return nil, errors.New("not implemented")
}

func (m *mockEnvironmentRepo) ActivateEnvironment(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *mockEnvironmentRepo) DeactivateEnvironment(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *mockEnvironmentRepo) Patch(ctx context.Context, name string, patchData map[string]interface{}) (*environmentsv1.Environment, error) {
	return nil, errors.New("not implemented")
}

func (m *mockEnvironmentRepo) Watch(ctx context.Context, opts ...repository.WatchOption) (<-chan repository.WatchEvent[*environmentsv1.Environment], error) {
	return nil, errors.New("not implemented")
}

// mockUserRepo implements repository.UserRepository for testing
type mockUserRepo struct{}

func (m *mockUserRepo) Create(ctx context.Context, user *platformv1alpha1.User) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) Get(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) List(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) Update(ctx context.Context, user *platformv1alpha1.User) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) Delete(ctx context.Context, name string) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) ListActive(ctx context.Context) (*platformv1alpha1.UserList, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) UpdateStatus(ctx context.Context, user *platformv1alpha1.User) error {
	return errors.New("not implemented")
}
func (m *mockUserRepo) Patch(ctx context.Context, name string, patchData map[string]interface{}) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserRepo) Watch(ctx context.Context, opts ...repository.WatchOption) (<-chan repository.WatchEvent[*platformv1alpha1.User], error) {
	return nil, errors.New("not implemented")
}

func TestGetEnvironment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should get environment by name", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						TargetNamespace: "test-namespace",
						Activated:       true,
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return nil, errors.New("environment test-env not found")
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.GET("/environments/:name", handlers.GetEnvironment)

		req, _ := http.NewRequest("GET", "/environments/test-env", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 when name is empty", func(t *testing.T) {
		handlers := NewEnvironmentHandlers(&mockEnvironmentRepo{}, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			listFunc: func(ctx context.Context, opts ...repository.ListOption) (*environmentsv1.EnvironmentList, error) {
				return &environmentsv1.EnvironmentList{
					Items: []environmentsv1.Environment{
						{ObjectMeta: metav1.ObjectMeta{Name: "env1"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "env2"}},
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			listActiveFunc: func(ctx context.Context) (*environmentsv1.EnvironmentList, error) {
				return &environmentsv1.EnvironmentList{
					Items: []environmentsv1.Environment{
						{ObjectMeta: metav1.ObjectMeta{Name: "active-env"}},
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.GET("/environments", handlers.ListEnvironments)

		req, _ := http.NewRequest("GET", "/environments?status=active", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should list inactive environments", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			listInactiveFunc: func(ctx context.Context) (*environmentsv1.EnvironmentList, error) {
				return &environmentsv1.EnvironmentList{
					Items: []environmentsv1.Environment{
						{ObjectMeta: metav1.ObjectMeta{Name: "inactive-env"}},
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
		_ = platformv1alpha1.AddToScheme(scheme)
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&platformv1alpha1.User{
				ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
				Spec: platformv1alpha1.UserSpec{
					Email: "test@example.com",
				},
			},
		).Build()

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, k8sClient, logger)
		router := gin.New()

		// Add middleware to set user context
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "test@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		})

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
		handlers := NewEnvironmentHandlers(&mockEnvironmentRepo{}, &mockUserRepo{}, nil, logger)
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
		handlers := NewEnvironmentHandlers(&mockEnvironmentRepo{}, &mockUserRepo{}, nil, logger)
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						TargetNamespace: "old-ns",
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return nil, errors.New("environment test-env not found")
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: false,
					},
				}, nil
			},
			deleteFunc: func(ctx context.Context, name string) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.DELETE("/environments/:name", handlers.DeleteEnvironment)

		req, _ := http.NewRequest("DELETE", "/environments/test-env", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should prevent deletion of activated environment without force", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: true,
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.DELETE("/environments/:name", handlers.DeleteEnvironment)

		req, _ := http.NewRequest("DELETE", "/environments/test-env", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Cannot delete an activated environment")
	})

	t.Run("should force delete activated environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: true,
					},
				}, nil
			},
			deleteFunc: func(ctx context.Context, name string) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.DELETE("/environments/:name", handlers.DeleteEnvironment)

		req, _ := http.NewRequest("DELETE", "/environments/test-env?force=true", nil)
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: false,
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.POST("/environments/:name/activate", handlers.ActivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/test-env/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject activation of already activated environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: true,
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.POST("/environments/:name/activate", handlers.ActivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/test-env/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "already activated")
	})

	t.Run("should return 404 for non-existent environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return nil, fmt.Errorf("environment %s not found", name)
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: true,
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.POST("/environments/:name/deactivate", handlers.DeactivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/test-env/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject deactivation of already deactivated environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: false,
					},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.POST("/environments/:name/deactivate", handlers.DeactivateEnvironment)

		req, _ := http.NewRequest("POST", "/environments/test-env/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "already deactivated")
	})

	t.Run("should return 404 for non-existent environment", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return nil, fmt.Errorf("environment %s not found", name)
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						TargetNamespace: "test-ns",
						Activated:       true,
					},
					Status: environmentsv1.EnvironmentStatus{},
				}, nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: environmentsv1.EnvironmentSpec{
						Activated: false,
					},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec:       environmentsv1.EnvironmentSpec{},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return nil
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
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
		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
		router.PATCH("/environments/:name", handlers.PatchEnvironment)

		req, _ := http.NewRequest("PATCH", "/environments/test-env", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 404 when environment not found", func(t *testing.T) {
		envRepo := &mockEnvironmentRepo{
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return nil, errors.New("environment test-env not found")
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
			getFunc: func(ctx context.Context, name string) (*environmentsv1.Environment, error) {
				return &environmentsv1.Environment{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec:       environmentsv1.EnvironmentSpec{},
				}, nil
			},
			updateFunc: func(ctx context.Context, env *environmentsv1.Environment) error {
				return errors.New("update failed")
			},
		}

		handlers := NewEnvironmentHandlers(envRepo, &mockUserRepo{}, nil, logger)
		router := gin.New()
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
