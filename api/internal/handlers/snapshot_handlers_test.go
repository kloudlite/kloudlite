package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"github.com/kloudlite/kloudlite/api/internal/controllers/testutil"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListEnvironmentSnapshots_ReturnsAllSnapshots(t *testing.T) {
	gin.SetMode(gin.TestMode)
	scheme := testutil.NewTestScheme()

	// Create an environment
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env",
		},
		Spec: environmentsv1.EnvironmentSpec{
			Name:            "test-env",
			TargetNamespace: "env-test-env",
			OwnedBy:         "test-user",
		},
	}

	// Create 4 snapshots for the environment - similar to production scenario
	snapshot1 := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env-20260108-081923",
			Labels: map[string]string{
				"snapshots.kloudlite.io/environment": "test-env",
				"kloudlite.io/owned-by":              "test-user",
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			EnvironmentRef: &snapshotv1.EnvironmentReference{
				Name: "test-env",
			},
			OwnedBy: "test-user",
		},
		Status: snapshotv1.SnapshotStatus{
			State: snapshotv1.SnapshotStateReady,
		},
	}

	snapshot2 := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env-20260108-081943",
			Labels: map[string]string{
				"snapshots.kloudlite.io/environment": "test-env",
				"kloudlite.io/owned-by":              "test-user",
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			EnvironmentRef: &snapshotv1.EnvironmentReference{
				Name: "test-env",
			},
			ParentSnapshotRef: &snapshotv1.ParentSnapshotReference{
				Name: "test-env-20260108-081923",
			},
			OwnedBy: "test-user",
		},
		Status: snapshotv1.SnapshotStatus{
			State: snapshotv1.SnapshotStateReady,
		},
	}

	snapshot3 := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env-20260108-082632",
			Labels: map[string]string{
				"snapshots.kloudlite.io/environment": "test-env",
				"kloudlite.io/owned-by":              "test-user",
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			EnvironmentRef: &snapshotv1.EnvironmentReference{
				Name: "test-env",
			},
			ParentSnapshotRef: &snapshotv1.ParentSnapshotReference{
				Name: "test-env-20260108-081923",
			},
			OwnedBy: "test-user",
		},
		Status: snapshotv1.SnapshotStatus{
			State: snapshotv1.SnapshotStateReady,
		},
	}

	snapshot4 := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-env-20260108-082656",
			Labels: map[string]string{
				"snapshots.kloudlite.io/environment": "test-env",
				"kloudlite.io/owned-by":              "test-user",
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			EnvironmentRef: &snapshotv1.EnvironmentReference{
				Name: "test-env",
			},
			ParentSnapshotRef: &snapshotv1.ParentSnapshotReference{
				Name: "test-env-20260108-081923",
			},
			OwnedBy: "test-user",
		},
		Status: snapshotv1.SnapshotStatus{
			State: snapshotv1.SnapshotStateReady,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, snapshot1, snapshot2, snapshot3, snapshot4).Build()

	logger, _ := zap.NewDevelopment()
	snapshotRepo := repository.NewSnapshotRepository(k8sClient)
	envRepo := repository.NewEnvironmentRepository(k8sClient)

	handlers := NewSnapshotHandlers(
		snapshotRepo,
		envRepo,
		nil, // workspaceRepo not needed for this test
		nil, // workmachineRepo not needed for this test
		k8sClient,
		logger,
	)

	// Set up test router
	router := gin.New()
	router.GET("/api/v1/environments/:name/snapshots", func(c *gin.Context) {
		c.Set("user_username", "test-user")
		c.Set("user_email", "test-user@example.com")
		c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
		c.Next()
	}, handlers.ListSnapshots)

	// Make request
	req, _ := http.NewRequest("GET", "/api/v1/environments/test-env/snapshots", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response to verify all 4 snapshots are returned
	body := w.Body.String()
	assert.Contains(t, body, "test-env-20260108-081923")
	assert.Contains(t, body, "test-env-20260108-081943")
	assert.Contains(t, body, "test-env-20260108-082632")
	assert.Contains(t, body, "test-env-20260108-082656")
	assert.Contains(t, body, `"count":4`)
}

func TestListEnvironmentSnapshots_DoubleDashEnvironmentName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	scheme := testutil.NewTestScheme()

	// Create an environment with double-dash in name (like karthik--mine)
	env := &environmentsv1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "karthik--mine",
		},
		Spec: environmentsv1.EnvironmentSpec{
			Name:            "karthik--mine",
			TargetNamespace: "env-karthik--mine",
			OwnedBy:         "test-user",
		},
	}

	// Create snapshots with the same naming pattern
	snapshot1 := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "karthik--mine-20260108-081923",
			Labels: map[string]string{
				"snapshots.kloudlite.io/environment": "karthik--mine",
				"kloudlite.io/owned-by":              "test-user",
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			EnvironmentRef: &snapshotv1.EnvironmentReference{
				Name: "karthik--mine",
			},
			OwnedBy: "test-user",
		},
		Status: snapshotv1.SnapshotStatus{
			State: snapshotv1.SnapshotStateReady,
		},
	}

	snapshot2 := &snapshotv1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: "karthik--mine-20260108-081943",
			Labels: map[string]string{
				"snapshots.kloudlite.io/environment": "karthik--mine",
				"kloudlite.io/owned-by":              "test-user",
			},
		},
		Spec: snapshotv1.SnapshotSpec{
			EnvironmentRef: &snapshotv1.EnvironmentReference{
				Name: "karthik--mine",
			},
			ParentSnapshotRef: &snapshotv1.ParentSnapshotReference{
				Name: "karthik--mine-20260108-081923",
			},
			OwnedBy: "test-user",
		},
		Status: snapshotv1.SnapshotStatus{
			State: snapshotv1.SnapshotStateReady,
		},
	}

	k8sClient := testutil.NewFakeClient(scheme, env, snapshot1, snapshot2).Build()

	logger, _ := zap.NewDevelopment()
	snapshotRepo := repository.NewSnapshotRepository(k8sClient)
	envRepo := repository.NewEnvironmentRepository(k8sClient)

	handlers := NewSnapshotHandlers(
		snapshotRepo,
		envRepo,
		nil,
		nil,
		k8sClient,
		logger,
	)

	router := gin.New()
	router.GET("/api/v1/environments/:name/snapshots", func(c *gin.Context) {
		c.Set("user_username", "test-user")
		c.Set("user_email", "test-user@example.com")
		c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
		c.Next()
	}, handlers.ListSnapshots)

	// Make request with double-dash environment name
	req, _ := http.NewRequest("GET", "/api/v1/environments/karthik--mine/snapshots", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "karthik--mine-20260108-081923")
	assert.Contains(t, body, "karthik--mine-20260108-081943")
	assert.Contains(t, body, `"count":2`)
}
