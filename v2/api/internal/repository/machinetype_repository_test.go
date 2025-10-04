package repository

import (
	"context"
	"testing"

	machinesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/machines/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewMachineTypeRepository(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	repo := NewMachineTypeRepository(k8sClient)

	assert.NotNil(t, repo)
}

func TestMachineTypeRepository_ListActive(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machineTypes := []*machinesv1.MachineType{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "active-type-1",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Active Type 1",
				Active:      true,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "active-type-2",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Active Type 2",
				Active:      true,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "inactive-type",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Inactive Type",
				Active:      false,
			},
		},
	}

	objects := make([]runtime.Object, len(machineTypes))
	for i, mt := range machineTypes {
		objects[i] = mt
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	repo := NewMachineTypeRepository(k8sClient)

	list, err := repo.ListActive(context.Background())
	assert.NoError(t, err)
	assert.Len(t, list.Items, 2)
	for _, item := range list.Items {
		assert.True(t, item.Spec.Active)
	}
}

func TestMachineTypeRepository_GetByCategory(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machineTypes := []*machinesv1.MachineType{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "standard-1",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Standard 1",
				Category:    "standard",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "standard-2",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Standard 2",
				Category:    "standard",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gpu-1",
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "GPU 1",
				Category:    "gpu",
			},
		},
	}

	objects := make([]runtime.Object, len(machineTypes))
	for i, mt := range machineTypes {
		objects[i] = mt
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	repo := NewMachineTypeRepository(k8sClient)

	t.Run("get standard category", func(t *testing.T) {
		list, err := repo.GetByCategory(context.Background(), "standard")
		assert.NoError(t, err)
		assert.Len(t, list.Items, 2)
	})

	t.Run("get gpu category", func(t *testing.T) {
		list, err := repo.GetByCategory(context.Background(), "gpu")
		assert.NoError(t, err)
		assert.Len(t, list.Items, 1)
	})

	t.Run("get non-existent category", func(t *testing.T) {
		list, err := repo.GetByCategory(context.Background(), "nonexistent")
		assert.NoError(t, err)
		assert.Len(t, list.Items, 0)
	})
}

func TestMachineTypeRepository_GetDefault(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	tests := []struct {
		name         string
		machineTypes []*machinesv1.MachineType
		wantErr      bool
		wantName     string
	}{
		{
			name: "default machine type exists",
			machineTypes: []*machinesv1.MachineType{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "non-default",
					},
					Spec: machinesv1.MachineTypeSpec{
						DisplayName: "Non Default",
						Active:      true,
						IsDefault:   false,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "default-type",
					},
					Spec: machinesv1.MachineTypeSpec{
						DisplayName: "Default Type",
						Active:      true,
						IsDefault:   true,
					},
				},
			},
			wantErr:  false,
			wantName: "default-type",
		},
		{
			name: "no default machine type",
			machineTypes: []*machinesv1.MachineType{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "type-1",
					},
					Spec: machinesv1.MachineTypeSpec{
						DisplayName: "Type 1",
						Active:      true,
						IsDefault:   false,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "default but inactive",
			machineTypes: []*machinesv1.MachineType{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "inactive-default",
					},
					Spec: machinesv1.MachineTypeSpec{
						DisplayName: "Inactive Default",
						Active:      false,
						IsDefault:   true,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := make([]runtime.Object, len(tt.machineTypes))
			for i, mt := range tt.machineTypes {
				objects[i] = mt
			}

			k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
			repo := NewMachineTypeRepository(k8sClient)

			machineType, err := repo.GetDefault(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantName, machineType.Name)
				assert.True(t, machineType.Spec.IsDefault)
				assert.True(t, machineType.Spec.Active)
			}
		})
	}
}

func TestMachineTypeRepository_Create(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	repo := NewMachineTypeRepository(k8sClient)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "New Type",
			Category:    "general",
			Resources: machinesv1.MachineResources{
				CPU:    "2",
				Memory: "4Gi",
			},
			Active: true,
		},
	}

	err := repo.Create(context.Background(), machineType)
	assert.NoError(t, err)

	// Verify creation
	retrieved, err := repo.Get(context.Background(), "new-type")
	assert.NoError(t, err)
	assert.Equal(t, "New Type", retrieved.Spec.DisplayName)
}

func TestMachineTypeRepository_Get(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Existing Type",
			Category:    "standard",
			Active:      true,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(machineType).Build()
	repo := NewMachineTypeRepository(k8sClient)

	t.Run("get existing machine type", func(t *testing.T) {
		retrieved, err := repo.Get(context.Background(), "existing-type")
		assert.NoError(t, err)
		assert.Equal(t, "Existing Type", retrieved.Spec.DisplayName)
	})

	t.Run("get non-existent machine type", func(t *testing.T) {
		_, err := repo.Get(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestMachineTypeRepository_Update(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "update-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Original Name",
			Category:    "standard",
			Active:      true,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(machineType).Build()
	repo := NewMachineTypeRepository(k8sClient)

	// Update display name
	machineType.Spec.DisplayName = "Updated Name"
	err := repo.Update(context.Background(), machineType)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get(context.Background(), "update-type")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Spec.DisplayName)
}

func TestMachineTypeRepository_Delete(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: "delete-type",
		},
		Spec: machinesv1.MachineTypeSpec{
			DisplayName: "Delete Me",
			Category:    "standard",
			Active:      true,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(machineType).Build()
	repo := NewMachineTypeRepository(k8sClient)

	err := repo.Delete(context.Background(), "delete-type")
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.Get(context.Background(), "delete-type")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMachineTypeRepository_List(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machineTypes := []*machinesv1.MachineType{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "type-1",
				Labels: map[string]string{"provider": "aws"},
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Type 1",
				Active:      true,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "type-2",
				Labels: map[string]string{"provider": "gcp"},
			},
			Spec: machinesv1.MachineTypeSpec{
				DisplayName: "Type 2",
				Active:      true,
			},
		},
	}

	objects := make([]runtime.Object, len(machineTypes))
	for i, mt := range machineTypes {
		objects[i] = mt
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	repo := NewMachineTypeRepository(k8sClient)

	t.Run("list all machine types", func(t *testing.T) {
		list, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list.Items, 2)
	})

	t.Run("list with label selector", func(t *testing.T) {
		list, err := repo.List(context.Background(), WithLabelSelector("provider=aws"))
		assert.NoError(t, err)
		assert.NotNil(t, list)
	})
}
