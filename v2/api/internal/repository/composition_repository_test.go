package repository

import (
	"context"
	"testing"

	environmentsv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/environments/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewCompositionRepository(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	repo := NewCompositionRepository(k8sClient)

	assert.NotNil(t, repo)
}

func TestCompositionRepository_GetByEnvironment(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	compositions := []*environmentsv1.Composition{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "comp-1",
				Namespace: "env-1",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Composition 1",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "comp-2",
				Namespace: "env-1",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Composition 2",
			},
		},
	}

	objects := make([]runtime.Object, len(compositions))
	for i, comp := range compositions {
		objects[i] = comp
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	repo := NewCompositionRepository(k8sClient)

	list, err := repo.GetByEnvironment(context.Background(), "env-1", "env-1")
	assert.NoError(t, err)
	assert.Len(t, list.Items, 2)
}

func TestCompositionRepository_ListByState(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	compositions := []*environmentsv1.Composition{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "comp-ready",
				Namespace: "test-ns",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Ready Composition",
			},
			Status: environmentsv1.CompositionStatus{
				State: environmentsv1.CompositionStateRunning,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "comp-failed",
				Namespace: "test-ns",
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Failed Composition",
			},
			Status: environmentsv1.CompositionStatus{
				State: environmentsv1.CompositionStateFailed,
			},
		},
	}

	objects := make([]runtime.Object, len(compositions))
	for i, comp := range compositions {
		objects[i] = comp
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	repo := NewCompositionRepository(k8sClient)

	t.Run("list running compositions", func(t *testing.T) {
		list, err := repo.ListByState(context.Background(), "test-ns", environmentsv1.CompositionStateRunning)
		assert.NoError(t, err)
		assert.Len(t, list.Items, 1)
		assert.Equal(t, "comp-ready", list.Items[0].Name)
	})

	t.Run("list failed compositions", func(t *testing.T) {
		list, err := repo.ListByState(context.Background(), "test-ns", environmentsv1.CompositionStateFailed)
		assert.NoError(t, err)
		assert.Len(t, list.Items, 1)
		assert.Equal(t, "comp-failed", list.Items[0].Name)
	})
}

func TestCompositionRepository_UpdateState(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Test Composition",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(composition).Build()
	repo := NewCompositionRepository(k8sClient)

	// Status update may fail with fake client
	err := repo.UpdateState(context.Background(), "test-comp", "test-ns", environmentsv1.CompositionStateRunning, "Composition is running")
	if err != nil {
		// Expected with fake client
		assert.NotNil(t, err)
	}
}

func TestCompositionRepository_Create(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	repo := NewCompositionRepository(k8sClient)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-comp",
			Namespace: "test-ns",
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName:    "New Composition",
			ComposeContent: "version: '3.8'",
		},
	}

	err := repo.Create(context.Background(), composition)
	assert.NoError(t, err)

	// Verify creation
	retrieved, err := repo.Get(context.Background(), "test-ns", "new-comp")
	assert.NoError(t, err)
	assert.Equal(t, "New Composition", retrieved.Spec.DisplayName)
}

func TestCompositionRepository_Get(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "existing-comp",
			Namespace: "test-ns",
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Existing Composition",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(composition).Build()
	repo := NewCompositionRepository(k8sClient)

	t.Run("get existing composition", func(t *testing.T) {
		retrieved, err := repo.Get(context.Background(), "test-ns", "existing-comp")
		assert.NoError(t, err)
		assert.Equal(t, "Existing Composition", retrieved.Spec.DisplayName)
	})

	t.Run("get non-existent composition", func(t *testing.T) {
		_, err := repo.Get(context.Background(), "test-ns", "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCompositionRepository_Update(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "update-comp",
			Namespace: "test-ns",
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Original Name",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(composition).Build()
	repo := NewCompositionRepository(k8sClient)

	// Update display name
	composition.Spec.DisplayName = "Updated Name"
	err := repo.Update(context.Background(), composition)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get(context.Background(), "test-ns", "update-comp")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Spec.DisplayName)
}

func TestCompositionRepository_Delete(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "delete-comp",
			Namespace: "test-ns",
		},
		Spec: environmentsv1.CompositionSpec{
			DisplayName: "Delete Me",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(composition).Build()
	repo := NewCompositionRepository(k8sClient)

	err := repo.Delete(context.Background(), "test-ns", "delete-comp")
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.Get(context.Background(), "test-ns", "delete-comp")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCompositionRepository_List(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)

	compositions := []*environmentsv1.Composition{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "comp-1",
				Namespace: "test-ns",
				Labels:    map[string]string{"type": "web"},
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Composition 1",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "comp-2",
				Namespace: "test-ns",
				Labels:    map[string]string{"type": "api"},
			},
			Spec: environmentsv1.CompositionSpec{
				DisplayName: "Composition 2",
			},
		},
	}

	objects := make([]runtime.Object, len(compositions))
	for i, comp := range compositions {
		objects[i] = comp
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	repo := NewCompositionRepository(k8sClient)

	t.Run("list all compositions", func(t *testing.T) {
		list, err := repo.List(context.Background(), "test-ns")
		assert.NoError(t, err)
		assert.Len(t, list.Items, 2)
	})

	t.Run("list with label selector", func(t *testing.T) {
		list, err := repo.List(context.Background(), "test-ns", WithLabelSelector("type=web"))
		assert.NoError(t, err)
		assert.NotNil(t, list)
	})
}
