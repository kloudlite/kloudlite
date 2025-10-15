package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockUserServiceForHandlers implements services.UserService for testing
type mockUserServiceForHandlers struct {
	createUserFunc          func(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error)
	getUserFunc             func(ctx context.Context, name string) (*platformv1alpha1.User, error)
	getUserByEmailFunc      func(ctx context.Context, email string) (*platformv1alpha1.User, error)
	listUsersFunc           func(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error)
	updateUserFunc          func(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error)
	deleteUserFunc          func(ctx context.Context, name string) error
	activateUserFunc        func(ctx context.Context, name string) (*platformv1alpha1.User, error)
	deactivateUserFunc      func(ctx context.Context, name string) (*platformv1alpha1.User, error)
	resetUserPasswordFunc   func(ctx context.Context, name, newPassword string) error
	updateUserLastLoginFunc func(ctx context.Context, name string) error
}

func (m *mockUserServiceForHandlers) CreateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user)
	}
	return user, nil
}

func (m *mockUserServiceForHandlers) GetUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, name)
	}
	return nil, errors.New("not found")
}

func (m *mockUserServiceForHandlers) GetUserByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("not found")
}

func (m *mockUserServiceForHandlers) ListUsers(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
	if m.listUsersFunc != nil {
		return m.listUsersFunc(ctx, opts...)
	}
	return &platformv1alpha1.UserList{}, nil
}

func (m *mockUserServiceForHandlers) UpdateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
	if m.updateUserFunc != nil {
		return m.updateUserFunc(ctx, user)
	}
	return user, nil
}

func (m *mockUserServiceForHandlers) DeleteUser(ctx context.Context, name string) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(ctx, name)
	}
	return nil
}

func (m *mockUserServiceForHandlers) GetUserByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockUserServiceForHandlers) ActivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	if m.activateUserFunc != nil {
		return m.activateUserFunc(ctx, name)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserServiceForHandlers) DeactivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	if m.deactivateUserFunc != nil {
		return m.deactivateUserFunc(ctx, name)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserServiceForHandlers) ResetUserPassword(ctx context.Context, name, newPassword string) error {
	if m.resetUserPasswordFunc != nil {
		return m.resetUserPasswordFunc(ctx, name, newPassword)
	}
	return errors.New("not implemented")
}

func (m *mockUserServiceForHandlers) UpdateUserLastLogin(ctx context.Context, name string) error {
	if m.updateUserLastLoginFunc != nil {
		return m.updateUserLastLoginFunc(ctx, name)
	}
	return nil
}

func (m *mockUserServiceForHandlers) ValidatePassword(ctx context.Context, user *platformv1alpha1.User, password string) error {
	return errors.New("not implemented")
}

func (m *mockUserServiceForHandlers) HashPassword(password string) (string, error) {
	return "", errors.New("not implemented")
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("super admin should create user with any roles", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			createUserFunc: func(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
				user.Name = "new-admin"
				return user, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		// Set super admin context
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "superadmin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleSuperAdmin})
			c.Next()
		})

		router.POST("/users", handlers.CreateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "newadmin@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("admin should create regular user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			createUserFunc: func(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
				user.Name = "new-user"
				return user, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users", handlers.CreateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "newuser@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("admin should not create admin user", func(t *testing.T) {
		handlers := NewUserHandlers(&mockUserServiceForHandlers{}, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users", handlers.CreateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "newadmin@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "Insufficient permissions")
	})

	t.Run("regular user should not create any user", func(t *testing.T) {
		handlers := NewUserHandlers(&mockUserServiceForHandlers{}, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "user@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		})

		router.POST("/users", handlers.CreateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "newuser@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should reject request without auth", func(t *testing.T) {
		handlers := NewUserHandlers(&mockUserServiceForHandlers{}, logger)
		router := gin.New()
		router.POST("/users", handlers.CreateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "test@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject invalid request body", func(t *testing.T) {
		handlers := NewUserHandlers(&mockUserServiceForHandlers{}, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users", handlers.CreateUser)

		body := []byte(`{invalid json}`)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 when webhook validation fails", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			createUserFunc: func(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
				return nil, errors.New("admission webhook denied the request: email already exists")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users", handlers.CreateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "duplicate@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 500 when creation fails", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			createUserFunc: func(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
				return nil, errors.New("database connection error")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users", handlers.CreateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "newuser@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	activeTrue := true

	t.Run("should get user by name", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.GET("/users/:name", handlers.GetUser)

		req, _ := http.NewRequest("GET", "/users/test-user", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var user platformv1alpha1.User
		err := json.Unmarshal(w.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Equal(t, "test-user", user.Name)
	})

	t.Run("should return 404 for non-existent user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, errors.New("user not found")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.GET("/users/:name", handlers.GetUser)

		req, _ := http.NewRequest("GET", "/users/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetUserByEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	activeTrue := true

	t.Run("should get user by email", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserByEmailFunc: func(ctx context.Context, email string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:  email,
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.GET("/users/by-email", handlers.GetUserByEmail)

		req, _ := http.NewRequest("GET", "/users/by-email?email=test@example.com", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 for non-existent email", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserByEmailFunc: func(ctx context.Context, email string) (*platformv1alpha1.User, error) {
				return nil, errors.New("user not found")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.GET("/users/by-email", handlers.GetUserByEmail)

		req, _ := http.NewRequest("GET", "/users/by-email?email=nonexistent@example.com", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should list all users", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			listUsersFunc: func(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
				return &platformv1alpha1.UserList{
					Items: []platformv1alpha1.User{
						{ObjectMeta: metav1.ObjectMeta{Name: "user1"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "user2"}},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.GET("/users", handlers.ListUsers)

		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var userList platformv1alpha1.UserList
		err := json.Unmarshal(w.Body.Bytes(), &userList)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(userList.Items))
	})

	t.Run("should handle service error", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			listUsersFunc: func(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
				return nil, errors.New("service error")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.GET("/users", handlers.ListUsers)

		req, _ := http.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUpdateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	activeTrue := true

	t.Run("super admin should update any user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "admin@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin},
					},
				}, nil
			},
			updateUserFunc: func(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
				return user, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "superadmin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleSuperAdmin})
			c.Next()
		})

		router.PUT("/users/:name", handlers.UpdateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "updated@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("PUT", "/users/test-admin", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("admin should not update admin user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "otheradmin@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.PUT("/users/:name", handlers.UpdateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "updated@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("PUT", "/users/other-admin", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should return 404 for non-existent user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, errors.New("user not found")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.PUT("/users/:name", handlers.UpdateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "updated@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("PUT", "/users/nonexistent", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 401 when no user roles", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{}
		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.PUT("/users/:name", handlers.UpdateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "updated@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("PUT", "/users/test-user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.PUT("/users/:name", handlers.UpdateUser)

		req, _ := http.NewRequest("PUT", "/users/test-user", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 403 when trying to assign higher roles", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.PUT("/users/:name", handlers.UpdateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "user@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleSuperAdmin},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("PUT", "/users/test-user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should return 400 when webhook validation fails", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
			updateUserFunc: func(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
				return nil, errors.New("admission webhook denied the request: email already exists")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.PUT("/users/:name", handlers.UpdateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "duplicate@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("PUT", "/users/test-user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 500 when update fails", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
			updateUserFunc: func(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
				return nil, errors.New("database error")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.PUT("/users/:name", handlers.UpdateUser)

		userSpec := platformv1alpha1.UserSpec{
			Email: "updated@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
		}
		body, _ := json.Marshal(userSpec)
		req, _ := http.NewRequest("PUT", "/users/test-user", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	activeTrue := true

	t.Run("super admin should delete any user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
			deleteUserFunc: func(ctx context.Context, name string) error {
				return nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "superadmin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleSuperAdmin})
			c.Next()
		})

		router.DELETE("/users/:name", handlers.DeleteUser)

		req, _ := http.NewRequest("DELETE", "/users/test-user", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("admin should not delete admin user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "admin@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "otheradmin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.DELETE("/users/:name", handlers.DeleteUser)

		req, _ := http.NewRequest("DELETE", "/users/admin-user", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("regular user should not delete any user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "otheruser@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "user@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		})

		router.DELETE("/users/:name", handlers.DeleteUser)

		req, _ := http.NewRequest("DELETE", "/users/other-user", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should return 401 when no user roles", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{}
		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.DELETE("/users/:name", handlers.DeleteUser)

		req, _ := http.NewRequest("DELETE", "/users/test-user", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 when user not found", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, errors.New("user not found")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.DELETE("/users/:name", handlers.DeleteUser)

		req, _ := http.NewRequest("DELETE", "/users/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 500 when delete fails", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
			deleteUserFunc: func(ctx context.Context, name string) error {
				return errors.New("database error")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.DELETE("/users/:name", handlers.DeleteUser)

		req, _ := http.NewRequest("DELETE", "/users/test-user", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestActivateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	activeFalse := false

	t.Run("should activate inactive user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeFalse,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
			activateUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				activeTrue := true
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Active: &activeTrue,
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users/:name/activate", handlers.ActivateUser)

		req, _ := http.NewRequest("POST", "/users/test-user/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 401 when no user roles", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{}
		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.POST("/users/:name/activate", handlers.ActivateUser)

		req, _ := http.NewRequest("POST", "/users/test-user/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 when user not found", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, errors.New("user not found")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users/:name/activate", handlers.ActivateUser)

		req, _ := http.NewRequest("POST", "/users/nonexistent/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 403 when insufficient permissions", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "admin@example.com",
						Active: &activeFalse,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "user@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		})

		router.POST("/users/:name/activate", handlers.ActivateUser)

		req, _ := http.NewRequest("POST", "/users/admin-user/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should return 500 when activation fails", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeFalse,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
			activateUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, errors.New("activation failed")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users/:name/activate", handlers.ActivateUser)

		req, _ := http.NewRequest("POST", "/users/test-user/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestDeactivateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	activeTrue := true

	t.Run("should deactivate active user", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
			deactivateUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				activeFalse := false
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Active: &activeFalse,
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users/:name/deactivate", handlers.DeactivateUser)

		req, _ := http.NewRequest("POST", "/users/test-user/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 401 when no user roles", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{}
		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.POST("/users/:name/deactivate", handlers.DeactivateUser)

		req, _ := http.NewRequest("POST", "/users/test-user/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 404 when user not found", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, errors.New("user not found")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users/:name/deactivate", handlers.DeactivateUser)

		req, _ := http.NewRequest("POST", "/users/nonexistent/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 403 when insufficient permissions", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "admin@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin},
					},
				}, nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "user@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		})

		router.POST("/users/:name/deactivate", handlers.DeactivateUser)

		req, _ := http.NewRequest("POST", "/users/admin-user/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should return 500 when deactivation fails", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			getUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: name},
					Spec: platformv1alpha1.UserSpec{
						Email:  "user@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
			deactivateUserFunc: func(ctx context.Context, name string) (*platformv1alpha1.User, error) {
				return nil, errors.New("deactivation failed")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()

		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})

		router.POST("/users/:name/deactivate", handlers.DeactivateUser)

		req, _ := http.NewRequest("POST", "/users/test-user/deactivate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestResetUserPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should reset user password successfully", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			resetUserPasswordFunc: func(ctx context.Context, name, newPassword string) error {
				assert.Equal(t, "test-user", name)
				assert.Equal(t, "newpassword123", newPassword)
				return nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.POST("/users/:name/reset-password", handlers.ResetUserPassword)

		reqBody := map[string]string{
			"newPassword": "newpassword123",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/users/test-user/reset-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Password reset successfully")
	})

	t.Run("should return 400 with invalid request body", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{}
		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.POST("/users/:name/reset-password", handlers.ResetUserPassword)

		req, _ := http.NewRequest("POST", "/users/test-user/reset-password", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 when password is too short", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{}
		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.POST("/users/:name/reset-password", handlers.ResetUserPassword)

		reqBody := map[string]string{
			"newPassword": "short",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/users/test-user/reset-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 500 when service fails", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			resetUserPasswordFunc: func(ctx context.Context, name, newPassword string) error {
				return errors.New("service error")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.POST("/users/:name/reset-password", handlers.ResetUserPassword)

		reqBody := map[string]string{
			"newPassword": "newpassword123",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/users/test-user/reset-password", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUpdateUserLastLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should update last login successfully", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			updateUserLastLoginFunc: func(ctx context.Context, name string) error {
				assert.Equal(t, "test-user", name)
				return nil
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.POST("/users/:name/update-last-login", handlers.UpdateUserLastLogin)

		req, _ := http.NewRequest("POST", "/users/test-user/update-last-login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Last login updated successfully")
	})

	t.Run("should return 500 when service fails", func(t *testing.T) {
		userService := &mockUserServiceForHandlers{
			updateUserLastLoginFunc: func(ctx context.Context, name string) error {
				return errors.New("service error")
			},
		}

		handlers := NewUserHandlers(userService, logger)
		router := gin.New()
		router.POST("/users/:name/update-last-login", handlers.UpdateUserLastLogin)

		req, _ := http.NewRequest("POST", "/users/test-user/update-last-login", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGetNamespace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should get namespace from query parameter", func(t *testing.T) {
		router := gin.New()
		var capturedNamespace string
		router.GET("/test", func(c *gin.Context) {
			capturedNamespace = getNamespace(c)
			c.JSON(http.StatusOK, gin.H{"namespace": capturedNamespace})
		})

		req, _ := http.NewRequest("GET", "/test?namespace=test-ns", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "test-ns", capturedNamespace)
	})

	t.Run("should get namespace from header when query is not set", func(t *testing.T) {
		router := gin.New()
		var capturedNamespace string
		router.GET("/test", func(c *gin.Context) {
			capturedNamespace = getNamespace(c)
			c.JSON(http.StatusOK, gin.H{"namespace": capturedNamespace})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Namespace", "header-ns")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "header-ns", capturedNamespace)
	})

	t.Run("should return default namespace when neither query nor header is set", func(t *testing.T) {
		router := gin.New()
		var capturedNamespace string
		router.GET("/test", func(c *gin.Context) {
			capturedNamespace = getNamespace(c)
			c.JSON(http.StatusOK, gin.H{"namespace": capturedNamespace})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "default", capturedNamespace)
	})

	t.Run("should prefer query parameter over header", func(t *testing.T) {
		router := gin.New()
		var capturedNamespace string
		router.GET("/test", func(c *gin.Context) {
			capturedNamespace = getNamespace(c)
			c.JSON(http.StatusOK, gin.H{"namespace": capturedNamespace})
		})

		req, _ := http.NewRequest("GET", "/test?namespace=query-ns", nil)
		req.Header.Set("X-Namespace", "header-ns")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "query-ns", capturedNamespace)
	})
}
