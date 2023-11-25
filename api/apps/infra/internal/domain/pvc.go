package domain

import (
	"fmt"
	"kloudlite.io/apps/infra/internal/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) ListPVCs(ctx InfraContext, clusterName string, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.PersistentVolumeClaim], error) {
	filter := repos.Filter{
		"accountName": ctx.AccountName,
		"clusterName": clusterName,
	}
	return d.pvcRepo.FindPaginated(ctx, d.nodePoolRepo.MergeMatchFilters(filter, matchFilters), pagination)
}

func (d *domain) GetPVC(ctx InfraContext, clusterName string, buildRunName string) (*entities.PersistentVolumeClaim, error) {
	pvc, err := d.pvcRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": buildRunName,
	})
	if err != nil {
		return nil, err
	}

	if pvc == nil {
		return nil, fmt.Errorf("persistent volume claim with name %q not found", buildRunName)
	}
	return pvc, nil
}

func (d *domain) OnPVCUpdateMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim) error {
	if _, err := d.pvcRepo.Upsert(ctx, repos.Filter{
		"metadata.name":      pvc.Name,
		"metadata.namespace": pvc.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        clusterName,
	}, &pvc); err != nil {
		return err
	}
	return nil
}

func (d *domain) OnPVCDeleteMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim) error {
	if err := d.pvcRepo.DeleteOne(ctx, repos.Filter{
		"metadata.name":      pvc.Name,
		"metadata.namespace": pvc.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        clusterName,
	}); err != nil {
		return err
	}
	return nil
}
