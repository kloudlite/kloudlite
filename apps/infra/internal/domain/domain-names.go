package domain

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/plugin/modelgen/out"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/apps/infra/internal/entities"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

func (d *domain) ListDomainEntry(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.DomainEntry], error) {
	filters := map[string]any{
		"accountName": ctx.AccountName,
	}
	return d.domainEntryRepo.FindPaginated(ctx, d.domainEntryRepo.MergeMatchFilters(filters, search), pagination)
}

func (d *domain) GetDomainEntry(ctx InfraContext, domainName string) (*entities.DomainEntry, error) {
	filters := repos.Filter{
		"accountName": ctx.AccountName,
		"domain":      domainName,
	}
	return d.domainEntryRepo.FindOne(ctx, filters)
}

func (d *domain) CreateDomainEntry(ctx InfraContext, de entities.DomainEntry) (*entities.DomainEntry, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateDomainEntry); err != nil {
		return nil, err
	}
	de.AccountName = ctx.AccountName
	de.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	de.LastUpdatedBy = de.CreatedBy

	nde, err := d.domainEntryRepo.Create(ctx, &de)
	if err != nil {
		return nil, err
	}

	return nde, nil
}

func (d *domain) UpdateDomainEntry(ctx InfraContext, de entities.DomainEntry) (*entities.DomainEntry, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateDomainEntry); err != nil {
		return nil, err
	}

	existing, err := d.findDomainEntry(ctx, ctx.AccountName, de.ClusterName)
	if err != nil {
		return nil, err
	}

	existing.DisplayName = de.DisplayName
}

func (d *domain) DeleteDomainEntry(ctx InfraContext, name string) error {
}

func (d *domain) findDomainEntry(ctx context.Context, accountName string, clusterName string) (*entities.DomainEntry, error) {
	filters := repos.Filter{
		"accountName": accountName,
		"clusterName": clusterName,
	}
	one, err := d.domainEntryRepo.FindOne(ctx, filters)
	if err != nil {
		return nil, err
	}

	if one == nil {
		return nil, fmt.Errorf("domain entry not found")
	}

	return one, nil
}
