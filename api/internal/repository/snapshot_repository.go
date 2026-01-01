package repository

import (
	"context"

	snapshotv1 "github.com/kloudlite/kloudlite/api/internal/controllers/snapshot/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SnapshotRepository provides operations for Snapshot resources (cluster-scoped)
type SnapshotRepository interface {
	ClusterRepository[*snapshotv1.Snapshot, *snapshotv1.SnapshotList]

	// Domain-specific methods
	ListByEnvironment(ctx context.Context, envName string) (*snapshotv1.SnapshotList, error)
	ListByWorkspace(ctx context.Context, workspaceName string) (*snapshotv1.SnapshotList, error)
	ListByOwner(ctx context.Context, owner string) (*snapshotv1.SnapshotList, error)
	ListReady(ctx context.Context) (*snapshotv1.SnapshotList, error)
}

// snapshotRepository implements SnapshotRepository
type snapshotRepository struct {
	ClusterRepository[*snapshotv1.Snapshot, *snapshotv1.SnapshotList]
	client client.WithWatch
}

// NewSnapshotRepository creates a new SnapshotRepository
func NewSnapshotRepository(k8sClient client.WithWatch) SnapshotRepository {
	baseRepo := NewK8sClusterRepository(
		k8sClient,
		func() *snapshotv1.Snapshot { return &snapshotv1.Snapshot{} },
		func() *snapshotv1.SnapshotList { return &snapshotv1.SnapshotList{} },
	)

	return &snapshotRepository{
		ClusterRepository: baseRepo,
		client:            k8sClient,
	}
}

// ListByEnvironment retrieves all snapshots for an environment
func (r *snapshotRepository) ListByEnvironment(ctx context.Context, envName string) (*snapshotv1.SnapshotList, error) {
	return r.List(ctx, WithLabelSelector("snapshots.kloudlite.io/environment="+envName))
}

// ListByWorkspace retrieves all snapshots for a workspace
func (r *snapshotRepository) ListByWorkspace(ctx context.Context, workspaceName string) (*snapshotv1.SnapshotList, error) {
	return r.List(ctx, WithLabelSelector("snapshots.kloudlite.io/workspace="+workspaceName))
}

// ListByOwner retrieves all snapshots owned by a user
func (r *snapshotRepository) ListByOwner(ctx context.Context, owner string) (*snapshotv1.SnapshotList, error) {
	return r.List(ctx, WithLabelSelector("kloudlite.io/owned-by="+owner))
}

// ListReady retrieves all ready snapshots
func (r *snapshotRepository) ListReady(ctx context.Context) (*snapshotv1.SnapshotList, error) {
	allSnapshots, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	// Filter for ready state (client-side since state is in status, not labels)
	readySnapshots := &snapshotv1.SnapshotList{}
	for _, s := range allSnapshots.Items {
		if s.Status.State == snapshotv1.SnapshotStateReady {
			readySnapshots.Items = append(readySnapshots.Items, s)
		}
	}

	return readySnapshots, nil
}
