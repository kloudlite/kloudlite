package repository

import (
	"context"
	"fmt"

	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineTypeRepository provides access to MachineType resources (cluster-scoped)
type MachineTypeRepository interface {
	ClusterRepository[*machinesv1.MachineType, *machinesv1.MachineTypeList]
	ListActive(ctx context.Context) (*machinesv1.MachineTypeList, error)
	GetByCategory(ctx context.Context, category string) (*machinesv1.MachineTypeList, error)
	GetDefault(ctx context.Context) (*machinesv1.MachineType, error)
}

type machineTypeRepository struct {
	ClusterRepository[*machinesv1.MachineType, *machinesv1.MachineTypeList]
	k8sClient client.Client
}

// NewMachineTypeRepository creates a new MachineType repository
func NewMachineTypeRepository(k8sClient client.Client) MachineTypeRepository {
	baseRepo := NewK8sClusterRepository(
		k8sClient,
		func() *machinesv1.MachineType { return &machinesv1.MachineType{} },
		func() *machinesv1.MachineTypeList { return &machinesv1.MachineTypeList{} },
	)
	return &machineTypeRepository{
		ClusterRepository: baseRepo,
		k8sClient:         k8sClient,
	}
}

// ListActive returns only active machine types
func (r *machineTypeRepository) ListActive(ctx context.Context) (*machinesv1.MachineTypeList, error) {
	list := &machinesv1.MachineTypeList{}
	if err := r.k8sClient.List(ctx, list); err != nil {
		return nil, err
	}

	// Filter for active machine types
	activeList := &machinesv1.MachineTypeList{}
	for _, mt := range list.Items {
		if mt.Spec.Active {
			activeList.Items = append(activeList.Items, mt)
		}
	}

	return activeList, nil
}

// GetByCategory returns machine types of a specific category
func (r *machineTypeRepository) GetByCategory(ctx context.Context, category string) (*machinesv1.MachineTypeList, error) {
	list := &machinesv1.MachineTypeList{}
	if err := r.k8sClient.List(ctx, list); err != nil {
		return nil, err
	}

	// Filter by category
	categoryList := &machinesv1.MachineTypeList{}
	for _, mt := range list.Items {
		if mt.Spec.Category == category {
			categoryList.Items = append(categoryList.Items, mt)
		}
	}

	return categoryList, nil
}

// GetDefault returns the default machine type
func (r *machineTypeRepository) GetDefault(ctx context.Context) (*machinesv1.MachineType, error) {
	list := &machinesv1.MachineTypeList{}
	if err := r.k8sClient.List(ctx, list); err != nil {
		return nil, err
	}

	// Find the default machine type
	for _, mt := range list.Items {
		if mt.Spec.IsDefault && mt.Spec.Active {
			return &mt, nil
		}
	}

	return nil, fmt.Errorf("no default machine type found")
}
