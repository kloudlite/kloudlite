package repository

import (
	"context"
	"testing"

	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewWorkMachineRepository(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	repo := NewWorkMachineRepository(k8sClient)

	assert.NotNil(t, repo)
}

func TestWorkMachineRepository_GetByOwner(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	tests := []struct {
		name             string
		existingMachines []*machinesv1.WorkMachine
		owner            string
		wantErr          bool
		errContains      string
	}{
		{
			name: "machine found by owner",
			existingMachines: []*machinesv1.WorkMachine{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-machine",
					},
					Spec: machinesv1.WorkMachineSpec{
						OwnedBy:     "test-user",
						MachineType: "standard-4",
					},
				},
			},
			owner:   "test-user",
			wantErr: false,
		},
		{
			name:             "machine not found",
			existingMachines: []*machinesv1.WorkMachine{},
			owner:            "nonexistent-user",
			wantErr:          true,
			errContains:      "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects := make([]runtime.Object, len(tt.existingMachines))
			for i, m := range tt.existingMachines {
				objects[i] = m
			}

			k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
			repo := NewWorkMachineRepository(k8sClient)

			machine, err := repo.GetByOwner(context.Background(), tt.owner)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, machine)
				assert.Equal(t, tt.owner, machine.Spec.OwnedBy)
			}
		})
	}
}

func TestWorkMachineRepository_StartMachine(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:     "test-user",
			MachineType: "standard-4",
			State:       machinesv1.MachineStateStopped,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(machine).Build()
	repo := NewWorkMachineRepository(k8sClient)

	err := repo.StartMachine(context.Background(), "test-machine")
	assert.NoError(t, err)

	// Verify state was updated
	updated, err := repo.Get(context.Background(), "test-machine")
	assert.NoError(t, err)
	assert.Equal(t, machinesv1.MachineStateRunning, updated.Spec.State)
}

func TestWorkMachineRepository_StopMachine(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:     "test-user",
			MachineType: "standard-4",
			State:       machinesv1.MachineStateRunning,
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(machine).Build()
	repo := NewWorkMachineRepository(k8sClient)

	err := repo.StopMachine(context.Background(), "test-machine")
	assert.NoError(t, err)

	// Verify state was updated
	updated, err := repo.Get(context.Background(), "test-machine")
	assert.NoError(t, err)
	assert.Equal(t, machinesv1.MachineStateStopped, updated.Spec.State)
}

func TestWorkMachineRepository_ListByMachineType(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machines := []*machinesv1.WorkMachine{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machine-1",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "user-1",
				MachineType: "standard-4",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machine-2",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "user-2",
				MachineType: "standard-4",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "machine-3",
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "user-3",
				MachineType: "large-8",
			},
		},
	}

	objects := make([]runtime.Object, len(machines))
	for i, m := range machines {
		objects[i] = m
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	repo := NewWorkMachineRepository(k8sClient)

	t.Run("list by specific machine type", func(t *testing.T) {
		list, err := repo.ListByMachineType(context.Background(), "standard-4")
		assert.NoError(t, err)
		assert.Len(t, list.Items, 2)
	})

	t.Run("list by different machine type", func(t *testing.T) {
		list, err := repo.ListByMachineType(context.Background(), "large-8")
		assert.NoError(t, err)
		assert.Len(t, list.Items, 1)
	})

	t.Run("list by non-existent machine type", func(t *testing.T) {
		list, err := repo.ListByMachineType(context.Background(), "nonexistent")
		assert.NoError(t, err)
		assert.Len(t, list.Items, 0)
	})
}

func TestWorkMachineRepository_Create(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	repo := NewWorkMachineRepository(k8sClient)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:     "new-user",
			MachineType: "standard-4",
		},
	}

	err := repo.Create(context.Background(), machine)
	assert.NoError(t, err)

	// Verify creation
	retrieved, err := repo.Get(context.Background(), "new-machine")
	assert.NoError(t, err)
	assert.Equal(t, "new-user", retrieved.Spec.OwnedBy)
}

func TestWorkMachineRepository_Get(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:     "existing-user",
			MachineType: "standard-4",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(machine).Build()
	repo := NewWorkMachineRepository(k8sClient)

	t.Run("get existing machine", func(t *testing.T) {
		retrieved, err := repo.Get(context.Background(), "existing-machine")
		assert.NoError(t, err)
		assert.Equal(t, "existing-user", retrieved.Spec.OwnedBy)
	})

	t.Run("get non-existent machine", func(t *testing.T) {
		_, err := repo.Get(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestWorkMachineRepository_Update(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "update-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:     "original-user",
			MachineType: "standard-4",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(machine).Build()
	repo := NewWorkMachineRepository(k8sClient)

	// Update machine type
	machine.Spec.MachineType = "large-8"
	err := repo.Update(context.Background(), machine)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get(context.Background(), "update-machine")
	assert.NoError(t, err)
	assert.Equal(t, "large-8", retrieved.Spec.MachineType)
}

func TestWorkMachineRepository_Delete(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "delete-machine",
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:     "delete-user",
			MachineType: "standard-4",
		},
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(machine).Build()
	repo := NewWorkMachineRepository(k8sClient)

	err := repo.Delete(context.Background(), "delete-machine")
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.Get(context.Background(), "delete-machine")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestWorkMachineRepository_List(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = machinesv1.AddToScheme(scheme)

	machines := []*machinesv1.WorkMachine{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "machine-1",
				Labels: map[string]string{"team": "engineering"},
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "user-1",
				MachineType: "standard-4",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "machine-2",
				Labels: map[string]string{"team": "sales"},
			},
			Spec: machinesv1.WorkMachineSpec{
				OwnedBy:     "user-2",
				MachineType: "standard-4",
			},
		},
	}

	objects := make([]runtime.Object, len(machines))
	for i, m := range machines {
		objects[i] = m
	}

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	repo := NewWorkMachineRepository(k8sClient)

	t.Run("list all machines", func(t *testing.T) {
		list, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list.Items, 2)
	})

	t.Run("list with label selector", func(t *testing.T) {
		list, err := repo.List(context.Background(), WithLabelSelector("team=engineering"))
		assert.NoError(t, err)
		assert.NotNil(t, list)
	})
}
