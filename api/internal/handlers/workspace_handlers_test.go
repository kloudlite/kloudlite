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
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupWorkspaceHandlerTest() (*WorkspaceHandlers, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = workspacesv1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	wsRepo := repository.NewWorkspaceRepository(k8sClient)
	userRepo := repository.NewUserRepository(k8sClient)
	wmRepo := repository.NewWorkMachineRepository(k8sClient)
	logger, _ := zap.NewDevelopment()

	handlers := NewWorkspaceHandlers(wsRepo, userRepo, wmRepo, k8sClient, logger)
	router := gin.New()

	return handlers, router
}

func TestCreateWorkspace(t *testing.T) {
	t.Run("should create workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test WorkMachine for the user
		workMachine := &machinesv1.WorkMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-wm",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:         "test-user@example.com",
				TargetNamespace: "test-wm-ns",
			},
		}
		_ = handlers.wmRepo.Create(context.Background(), workMachine)

		router.POST("/namespaces/:namespace/workspaces", func(c *gin.Context) {
			c.Set("user_email", "test-user@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.CreateWorkspace)

		reqBody := struct {
			Name string                     `json:"name"`
			Spec workspacesv1.WorkspaceSpec `json:"spec"`
		}{
			Name: "test-workspace",
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Test Workspace",
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("should return 401 when not authenticated", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.POST("/namespaces/:namespace/workspaces", handlers.CreateWorkspace)

		reqBody := struct {
			Name string                     `json:"name"`
			Spec workspacesv1.WorkspaceSpec `json:"spec"`
		}{
			Name: "test-workspace",
			Spec: workspacesv1.WorkspaceSpec{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.POST("/namespaces/:namespace/workspaces", func(c *gin.Context) {
			c.Set("user_email", "test-user@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.CreateWorkspace)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 404 when user has no WorkMachine", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.POST("/namespaces/:namespace/workspaces", func(c *gin.Context) {
			c.Set("user_email", "no-wm-user@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		}, handlers.CreateWorkspace)

		reqBody := struct {
			Name string                     `json:"name"`
			Spec workspacesv1.WorkspaceSpec `json:"spec"`
		}{
			Name: "test-workspace",
			Spec: workspacesv1.WorkspaceSpec{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestGetWorkspace(t *testing.T) {
	t.Run("should get workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspace
		workspace := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-workspace",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Test Workspace",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace)

		router.GET("/namespaces/:namespace/workspaces/:name", handlers.GetWorkspace)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/workspaces/test-workspace", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 for non-existent workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.GET("/namespaces/:namespace/workspaces/:name", handlers.GetWorkspace)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/workspaces/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestListWorkspaces(t *testing.T) {
	t.Run("should list all workspaces", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspaces
		workspace1 := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workspace-1",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Workspace 1",
				Owner:       "user1@example.com",
				Status:      "active",
			},
		}
		workspace2 := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workspace-2",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Workspace 2",
				Owner:       "user2@example.com",
				Status:      "suspended",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace1)
		_ = handlers.wsRepo.Create(context.Background(), workspace2)

		router.GET("/namespaces/:namespace/workspaces", handlers.ListWorkspaces)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/workspaces", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response workspacesv1.WorkspaceList
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response.Items, 2)
	})

	t.Run("should list workspaces by owner", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		workspace1 := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workspace-1",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Workspace 1",
				Owner:       "user1@example.com",
				Status:      "active",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace1)

		router.GET("/namespaces/:namespace/workspaces", handlers.ListWorkspaces)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/workspaces?owner=user1@example.com", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should list workspaces by workMachine", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		workspace2 := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workspace-2",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Workspace 2",
				Owner:       "user2@example.com",
				Status:      "suspended",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace2)

		router.GET("/namespaces/:namespace/workspaces", handlers.ListWorkspaces)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/workspaces?workMachine=wm-1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response workspacesv1.WorkspaceList
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response.Items, 1)
		assert.Equal(t, "workspace-2", response.Items[0].Name)
	})

	t.Run("should list workspaces by status", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		workspace1 := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workspace-1",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Workspace 1",
				Owner:       "user1@example.com",
				Status:      "active",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace1)

		router.GET("/namespaces/:namespace/workspaces", handlers.ListWorkspaces)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/workspaces?status=active", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 400 for invalid status", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.GET("/namespaces/:namespace/workspaces", handlers.ListWorkspaces)

		req := httptest.NewRequest(http.MethodGet, "/namespaces/test-ns/workspaces?status=invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUpdateWorkspace(t *testing.T) {
	t.Run("should update workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspace
		workspace := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "update-workspace",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Original",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace)

		router.PUT("/namespaces/:namespace/workspaces/:name", handlers.UpdateWorkspace)

		reqBody := struct {
			Spec workspacesv1.WorkspaceSpec `json:"spec"`
		}{
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Updated",
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/namespaces/test-ns/workspaces/update-workspace", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 for non-existent workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.PUT("/namespaces/:namespace/workspaces/:name", handlers.UpdateWorkspace)

		reqBody := struct {
			Spec workspacesv1.WorkspaceSpec `json:"spec"`
		}{
			Spec: workspacesv1.WorkspaceSpec{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/namespaces/test-ns/workspaces/nonexistent", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspace first
		workspace := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "update-workspace",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Original",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace)

		router.PUT("/namespaces/:namespace/workspaces/:name", handlers.UpdateWorkspace)

		req := httptest.NewRequest(http.MethodPut, "/namespaces/test-ns/workspaces/update-workspace", bytes.NewBuffer([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDeleteWorkspace(t *testing.T) {
	t.Run("should delete workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspace
		workspace := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "delete-workspace",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Delete Me",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace)

		router.DELETE("/namespaces/:namespace/workspaces/:name", handlers.DeleteWorkspace)

		req := httptest.NewRequest(http.MethodDelete, "/namespaces/test-ns/workspaces/delete-workspace", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("should return 404 for non-existent workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.DELETE("/namespaces/:namespace/workspaces/:name", handlers.DeleteWorkspace)

		req := httptest.NewRequest(http.MethodDelete, "/namespaces/test-ns/workspaces/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestSuspendWorkspace(t *testing.T) {
	t.Run("should suspend workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspace
		workspace := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "suspend-workspace",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Test Workspace",
				Status:      "active",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace)

		router.POST("/namespaces/:namespace/workspaces/:name/suspend", handlers.SuspendWorkspace)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces/suspend-workspace/suspend", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify status changed
		updated, _ := handlers.wsRepo.Get(context.Background(), "test-ns", "suspend-workspace")
		assert.Equal(t, "suspended", updated.Spec.Status)
	})

	t.Run("should return 404 for non-existent workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.POST("/namespaces/:namespace/workspaces/:name/suspend", handlers.SuspendWorkspace)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces/nonexistent/suspend", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestActivateWorkspace(t *testing.T) {
	t.Run("should activate workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspace
		workspace := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "activate-workspace",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Test Workspace",
				Status:      "suspended",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace)

		router.POST("/namespaces/:namespace/workspaces/:name/activate", handlers.ActivateWorkspace)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces/activate-workspace/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify status changed
		updated, _ := handlers.wsRepo.Get(context.Background(), "test-ns", "activate-workspace")
		assert.Equal(t, "active", updated.Spec.Status)
	})

	t.Run("should use default namespace when namespace param is empty", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspace in default namespace
		workspace := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-workspace",
				Namespace: "default",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Test Workspace",
				Status:      "suspended",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace)

		router.POST("/namespaces/:namespace/workspaces/:name/activate", handlers.ActivateWorkspace)

		req := httptest.NewRequest(http.MethodPost, "/namespaces//workspaces/test-workspace/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 when workspace not found", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.POST("/namespaces/:namespace/workspaces/:name/activate", handlers.ActivateWorkspace)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces/nonexistent/activate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestArchiveWorkspace(t *testing.T) {
	t.Run("should archive workspace", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspace
		workspace := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "archive-workspace",
				Namespace: "test-ns",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Test Workspace",
				Status:      "active",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace)

		router.POST("/namespaces/:namespace/workspaces/:name/archive", handlers.ArchiveWorkspace)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces/archive-workspace/archive", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify status changed
		updated, _ := handlers.wsRepo.Get(context.Background(), "test-ns", "archive-workspace")
		assert.Equal(t, "archived", updated.Spec.Status)
	})

	t.Run("should use default namespace when namespace param is empty", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()

		// Create test workspace in default namespace
		workspace := &workspacesv1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-workspace",
				Namespace: "default",
			},
			Spec: workspacesv1.WorkspaceSpec{
				DisplayName: "Test Workspace",
				Status:      "active",
			},
		}
		_ = handlers.wsRepo.Create(context.Background(), workspace)

		router.POST("/namespaces/:namespace/workspaces/:name/archive", handlers.ArchiveWorkspace)

		req := httptest.NewRequest(http.MethodPost, "/namespaces//workspaces/test-workspace/archive", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 404 when workspace not found", func(t *testing.T) {
		handlers, router := setupWorkspaceHandlerTest()
		router.POST("/namespaces/:namespace/workspaces/:name/archive", handlers.ArchiveWorkspace)

		req := httptest.NewRequest(http.MethodPost, "/namespaces/test-ns/workspaces/nonexistent/archive", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
