package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) CreateEdge(ctx InfraContext, edge entities.Edge) (*entities.Edge, error) {
	edge.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &edge.Edge); err != nil {
		return nil, err
	}

	edge.IncrementRecordVersion()
	edge.AccountName = ctx.AccountName
	edge.SyncStatus = t.GetSyncStatusForCreation()
	nEdge, err := d.edgeRepo.Create(ctx, &edge)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nEdge.Edge, nEdge.RecordVersion); err != nil {
		return nil, err
	}
	return nEdge, err
}

func (d *domain) ListEdges(ctx InfraContext, clusterName string, providerName *string, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.Edge], error) {
	f := repos.Filter{"spec.clusterName": clusterName}
	if providerName != nil {
		f["spec.providerName"] = providerName
	}
	return d.edgeRepo.FindPaginated(ctx, f, pagination)
}

func (d *domain) GetEdge(ctx InfraContext, clusterName string, name string) (*entities.Edge, error) {
	return d.edgeRepo.FindOne(ctx, repos.Filter{"metadata.name": name, "spec.clusterName": clusterName})
}

func (d *domain) UpdateEdge(ctx InfraContext, edge entities.Edge) (*entities.Edge, error) {
	edge.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &edge.Edge); err != nil {
		return nil, err
	}

	_, err := d.findCluster(ctx, edge.Spec.ClusterName)
	if err != nil {
		return nil, err
	}

	e, err := d.findEdge(ctx, edge.Name)
	if err != nil {
		return nil, err
	}

	e.IncrementRecordVersion()
	e.Spec = edge.Spec
	e.SyncStatus = t.GenSyncStatus(t.SyncActionApply, e.RecordVersion)

	uEdge, err := d.edgeRepo.UpdateById(ctx, e.Id, e)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &uEdge.Edge, uEdge.RecordVersion); err != nil {
		return nil, err
	}
	return uEdge, nil
}

func (d *domain) DeleteEdge(ctx InfraContext, clusterName string, name string) error {
	e, err := d.findEdge(ctx, name)
	if err != nil {
		return err
	}
	e.SyncStatus = t.GetSyncStatusForDeletion(e.Generation)
	return d.deleteK8sResource(ctx, e)
}

func (d *domain) OnDeleteEdgeMessage(ctx InfraContext, edge entities.Edge) error {
	e, err := d.findEdge(ctx, edge.Name)
	if err != nil {
		return err
	}

	return d.edgeRepo.DeleteById(ctx, e.Id)
}

func (d *domain) OnUpdateEdgeMessage(ctx InfraContext, edge entities.Edge) error {
	e, err := d.findEdge(ctx, edge.Name)
	if err != nil {
		return err
	}

	e.Edge = edge.Edge
	e.SyncStatus.LastSyncedAt = time.Now()
	e.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	_, err = d.edgeRepo.UpdateById(ctx, e.Id, e)
	return err
}

func (d *domain) findEdge(ctx InfraContext, edgeName string) (*entities.Edge, error) {
	e, err := d.edgeRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"metadata.name": edgeName,
	})
	if err != nil {
		return nil, err
	}

	if e == nil {
		return nil, fmt.Errorf("edge with name %q not found", edgeName)
	}
	return e, nil
}
