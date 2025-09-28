package repository

import (
	"context"
	"fmt"

	machinesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/machines/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkMachineRepository provides access to WorkMachine resources
type WorkMachineRepository interface {
	Repository[*machinesv1.WorkMachine, *machinesv1.WorkMachineList]
	GetByOwner(ctx context.Context, owner string) (*machinesv1.WorkMachine, error)
	StartMachine(ctx context.Context, name string) error
	StopMachine(ctx context.Context, name string) error
	ListByMachineType(ctx context.Context, machineType string) (*machinesv1.WorkMachineList, error)
}

type workMachineRepository struct {
	Repository[*machinesv1.WorkMachine, *machinesv1.WorkMachineList]
	k8sClient client.Client
}

// NewWorkMachineRepository creates a new WorkMachine repository
func NewWorkMachineRepository(k8sClient client.Client) WorkMachineRepository {
	baseRepo := NewK8sRepository(
		k8sClient,
		func() *machinesv1.WorkMachine { return &machinesv1.WorkMachine{} },
		func() *machinesv1.WorkMachineList { return &machinesv1.WorkMachineList{} },
	)
	return &workMachineRepository{
		Repository: baseRepo,
		k8sClient:  k8sClient,
	}
}

// GetByOwner returns the WorkMachine owned by a specific user
func (r *workMachineRepository) GetByOwner(ctx context.Context, owner string) (*machinesv1.WorkMachine, error) {
	list := &machinesv1.WorkMachineList{}
	if err := r.k8sClient.List(ctx, list); err != nil {
		return nil, err
	}

	// Find machine owned by the user (each user should have only one)
	for _, machine := range list.Items {
		if machine.Spec.OwnedBy == owner {
			return &machine, nil
		}
	}

	return nil, fmt.Errorf("no machine found for owner %s", owner)
}

// StartMachine starts a WorkMachine
func (r *workMachineRepository) StartMachine(ctx context.Context, name string) error {
	machine, err := r.Get(ctx, "", name)
	if err != nil {
		return err
	}

	// Update desired state to running
	machine.Spec.DesiredState = machinesv1.MachineStateRunning
	return r.Update(ctx, machine)
}

// StopMachine stops a WorkMachine
func (r *workMachineRepository) StopMachine(ctx context.Context, name string) error {
	machine, err := r.Get(ctx, "", name)
	if err != nil {
		return err
	}

	// Update desired state to stopped
	machine.Spec.DesiredState = machinesv1.MachineStateStopped
	return r.Update(ctx, machine)
}

// ListByMachineType returns WorkMachines using a specific machine type
func (r *workMachineRepository) ListByMachineType(ctx context.Context, machineType string) (*machinesv1.WorkMachineList, error) {
	list := &machinesv1.WorkMachineList{}
	if err := r.k8sClient.List(ctx, list); err != nil {
		return nil, err
	}

	// Filter by machine type
	filteredList := &machinesv1.WorkMachineList{}
	for _, machine := range list.Items {
		if machine.Spec.MachineType == machineType {
			filteredList.Items = append(filteredList.Items, machine)
		}
	}

	return filteredList, nil
}