package repository

import (
	"context"

	workspacesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/workspaces/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkspaceRepository provides operations for Workspace resources (namespace-scoped)
type WorkspaceRepository interface {
	NamespacedRepository[*workspacesv1.Workspace, *workspacesv1.WorkspaceList]

	// Domain-specific methods
	GetByOwner(ctx context.Context, owner string, namespace string) (*workspacesv1.WorkspaceList, error)
	GetByWorkMachine(ctx context.Context, workMachineName string, namespace string) (*workspacesv1.WorkspaceList, error)
	ListActive(ctx context.Context, namespace string) (*workspacesv1.WorkspaceList, error)
	ListSuspended(ctx context.Context, namespace string) (*workspacesv1.WorkspaceList, error)
	ListArchived(ctx context.Context, namespace string) (*workspacesv1.WorkspaceList, error)
	SuspendWorkspace(ctx context.Context, name string, namespace string) error
	ActivateWorkspace(ctx context.Context, name string, namespace string) error
	ArchiveWorkspace(ctx context.Context, name string, namespace string) error
	UpdatePhase(ctx context.Context, name string, namespace string, phase string) error
}

// workspaceRepository implements WorkspaceRepository
type workspaceRepository struct {
	NamespacedRepository[*workspacesv1.Workspace, *workspacesv1.WorkspaceList]
	client client.Client
}

// NewWorkspaceRepository creates a new WorkspaceRepository
func NewWorkspaceRepository(k8sClient client.Client) WorkspaceRepository {
	baseRepo := NewK8sNamespacedRepository(
		k8sClient,
		func() *workspacesv1.Workspace { return &workspacesv1.Workspace{} },
		func() *workspacesv1.WorkspaceList { return &workspacesv1.WorkspaceList{} },
	)

	return &workspaceRepository{
		NamespacedRepository: baseRepo,
		client:               k8sClient,
	}
}

// GetByOwner retrieves all workspaces owned by a specific user
func (r *workspaceRepository) GetByOwner(ctx context.Context, owner string, namespace string) (*workspacesv1.WorkspaceList, error) {
	// Use field selector to find workspaces by owner
	return r.List(ctx, namespace, WithFieldSelector("spec.owner="+owner))
}

// GetByWorkMachine retrieves all workspaces running on a specific WorkMachine
func (r *workspaceRepository) GetByWorkMachine(ctx context.Context, workMachineName string, namespace string) (*workspacesv1.WorkspaceList, error) {
	// List all workspaces and filter by WorkMachine reference
	allWorkspaces, err := r.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := &workspacesv1.WorkspaceList{}
	for _, ws := range allWorkspaces.Items {
		if ws.Spec.WorkMachineRef != nil && ws.Spec.WorkMachineRef.Name == workMachineName {
			result.Items = append(result.Items, ws)
		}
	}

	return result, nil
}

// ListActive retrieves all active workspaces
func (r *workspaceRepository) ListActive(ctx context.Context, namespace string) (*workspacesv1.WorkspaceList, error) {
	// Use field selector to find active workspaces
	return r.List(ctx, namespace, WithFieldSelector("spec.status=active"))
}

// ListSuspended retrieves all suspended workspaces
func (r *workspaceRepository) ListSuspended(ctx context.Context, namespace string) (*workspacesv1.WorkspaceList, error) {
	// Use field selector to find suspended workspaces
	return r.List(ctx, namespace, WithFieldSelector("spec.status=suspended"))
}

// ListArchived retrieves all archived workspaces
func (r *workspaceRepository) ListArchived(ctx context.Context, namespace string) (*workspacesv1.WorkspaceList, error) {
	// Use field selector to find archived workspaces
	return r.List(ctx, namespace, WithFieldSelector("spec.status=archived"))
}

// SuspendWorkspace suspends a workspace by name
func (r *workspaceRepository) SuspendWorkspace(ctx context.Context, name string, namespace string) error {
	// Get the workspace
	workspace, err := r.Get(ctx, namespace, name)
	if err != nil {
		return err
	}

	// Update the status
	workspace.Spec.Status = "suspended"

	// Save the updated workspace
	return r.Update(ctx, workspace)
}

// ActivateWorkspace activates a workspace by name
func (r *workspaceRepository) ActivateWorkspace(ctx context.Context, name string, namespace string) error {
	// Get the workspace
	workspace, err := r.Get(ctx, namespace, name)
	if err != nil {
		return err
	}

	// Update the status
	workspace.Spec.Status = "active"

	// Save the updated workspace
	return r.Update(ctx, workspace)
}

// ArchiveWorkspace archives a workspace by name
func (r *workspaceRepository) ArchiveWorkspace(ctx context.Context, name string, namespace string) error {
	// Get the workspace
	workspace, err := r.Get(ctx, namespace, name)
	if err != nil {
		return err
	}

	// Update the status
	workspace.Spec.Status = "archived"

	// Save the updated workspace
	return r.Update(ctx, workspace)
}

// UpdatePhase updates the phase of a workspace
func (r *workspaceRepository) UpdatePhase(ctx context.Context, name string, namespace string, phase string) error {
	// Get the workspace
	workspace, err := r.Get(ctx, namespace, name)
	if err != nil {
		return err
	}

	// Update the phase in status
	workspace.Status.Phase = phase

	// Use status subresource to update only status
	return r.client.Status().Update(ctx, workspace)
}