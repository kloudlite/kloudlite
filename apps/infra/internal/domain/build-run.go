package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) ListBuildRuns(ctx InfraContext, repoName string, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BuildRun], error) {
	filter := repos.Filter{
		"accountName":             ctx.AccountName,
		"spec.registry.repo.name": repoName,
	}
	return d.buildRunRepo.FindPaginated(ctx, d.nodePoolRepo.MergeMatchFilters(filter, matchFilters), pagination)
}

func (d *domain) GetBuildRun(ctx InfraContext, repoName string, buildRunName string) (*entities.BuildRun, error) {
	brun, err := d.buildRunRepo.FindOne(ctx, repos.Filter{
		"accountName":             ctx.AccountName,
		"metadata.name":           buildRunName,
		"spec.registry.repo.name": repoName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if brun == nil {
		return nil, errors.Newf("build run with name %q not found", buildRunName)
	}
	return brun, nil
}

func (d *domain) OnBuildRunUpdateMessage(ctx InfraContext, clusterName string, buildRun entities.BuildRun) error {
	if _, err := d.buildRunRepo.Upsert(ctx, repos.Filter{
		"metadata.name":      buildRun.Name,
		"metadata.namespace": buildRun.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        clusterName,
	}, &buildRun); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) OnBuildRunDeleteMessage(ctx InfraContext, clusterName string, buildRun entities.BuildRun) error {
	if err := d.buildRunRepo.DeleteOne(ctx, repos.Filter{
		"metadata.name":      buildRun.Name,
		"metadata.namespace": buildRun.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        clusterName,
	}); err != nil {
		return errors.NewE(err)
	}
	return nil
}
