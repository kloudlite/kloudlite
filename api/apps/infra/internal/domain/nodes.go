package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) ListNodes(ctx InfraContext, clusterName string, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Node], error) {
	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.ClusterName: clusterName,
	}

	return d.nodeRepo.FindPaginated(ctx, d.nodeRepo.MergeMatchFilters(filter, matchFilters), pagination)
}

//TODO (nxtcoder17): Deleting node should also be available

func (d *domain) GetNode(ctx InfraContext, clusterName string, nodeName string) (*entities.Node, error) {
	return d.findNode(ctx, clusterName, nodeName)
}

func (d *domain) findNode(ctx InfraContext, clusterName string, nodeName string) (*entities.Node, error) {
	node, err := d.nodeRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: nodeName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if node == nil {
		return nil, errors.Newf("no node with name %q found in cluster %q", nodeName, clusterName)
	}

	return node, nil
}

func (d *domain) OnNodeUpdateMessage(ctx InfraContext, clusterName string, node entities.Node) error {
	if _, err := d.nodeRepo.Upsert(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: node.Name,
	}, &node); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) OnNodeDeleteMessage(ctx InfraContext, clusterName string, node entities.Node) error {
	if err := d.namespaceRepo.DeleteOne(ctx, repos.Filter{
		fields.MetadataName: node.Name,
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
	}); err != nil {
		return errors.NewE(err)
	}
	return nil
}
