package repository

import (
	"context"
	"fmt"
	"log"

	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkMachineRepository provides access to WorkMachine resources (cluster-scoped)
type WorkMachineRepository interface {
	ClusterRepository[*machinesv1.WorkMachine, *machinesv1.WorkMachineList]
	GetByOwner(ctx context.Context, owner string) (*machinesv1.WorkMachine, error)
	StartMachine(ctx context.Context, name string) error
	StopMachine(ctx context.Context, name string) error
	ListByMachineType(ctx context.Context, machineType string) (*machinesv1.WorkMachineList, error)
}

type workMachineRepository struct {
	ClusterRepository[*machinesv1.WorkMachine, *machinesv1.WorkMachineList]
	k8sClient client.Client
}

// NewWorkMachineRepository creates a new WorkMachine repository
func NewWorkMachineRepository(k8sClient client.Client) WorkMachineRepository {
	baseRepo := NewK8sClusterRepository(
		k8sClient,
		func() *machinesv1.WorkMachine { return &machinesv1.WorkMachine{} },
		func() *machinesv1.WorkMachineList { return &machinesv1.WorkMachineList{} },
	)
	return &workMachineRepository{
		ClusterRepository: baseRepo,
		k8sClient:         k8sClient,
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

	// Return a proper Kubernetes NotFound error
	return nil, apierrors.NewNotFound(
		schema.GroupResource{Group: "machines.kloudlite.io", Resource: "workmachines"},
		fmt.Sprintf("owner:%s", owner),
	)
}

// StartMachine starts a WorkMachine
func (r *workMachineRepository) StartMachine(ctx context.Context, name string) error {
	log.Printf("[DEBUG] StartMachine called for machine: %s", name)

	machine, err := r.Get(ctx, name)
	if err != nil {
		log.Printf("[ERROR] Failed to get machine %s: %v", name, err)
		return err
	}

	log.Printf("[DEBUG] Current machine state - Name: %s, Current: %s, Desired: %s",
		machine.Name, machine.Status.State, machine.Spec.DesiredState)

	// Update desired state to running
	oldDesiredState := machine.Spec.DesiredState
	machine.Spec.DesiredState = machinesv1.MachineStateRunning
	log.Printf("[DEBUG] Updating desired state from %s to %s", oldDesiredState, machine.Spec.DesiredState)

	err = r.Update(ctx, machine)
	if err != nil {
		log.Printf("[ERROR] Failed to update machine %s: %v", name, err)
		return err
	}

	log.Printf("[DEBUG] Successfully updated machine %s", name)
	return nil
}

// StopMachine stops a WorkMachine
func (r *workMachineRepository) StopMachine(ctx context.Context, name string) error {
	log.Printf("[DEBUG] StopMachine called for machine: %s", name)

	machine, err := r.Get(ctx, name)
	if err != nil {
		log.Printf("[ERROR] Failed to get machine %s: %v", name, err)
		return err
	}

	log.Printf("[DEBUG] Current machine state - Name: %s, Current: %s, Desired: %s",
		machine.Name, machine.Status.State, machine.Spec.DesiredState)

	// Update desired state to stopped
	oldDesiredState := machine.Spec.DesiredState
	machine.Spec.DesiredState = machinesv1.MachineStateStopped
	log.Printf("[DEBUG] Updating desired state from %s to %s", oldDesiredState, machine.Spec.DesiredState)

	err = r.Update(ctx, machine)
	if err != nil {
		log.Printf("[ERROR] Failed to update machine %s: %v", name, err)
		return err
	}

	log.Printf("[DEBUG] Successfully updated machine %s", name)
	return nil
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
