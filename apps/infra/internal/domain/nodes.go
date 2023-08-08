package domain

import (
	"fmt"

	"kloudlite.io/apps/infra/internal/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) ListNodes(ctx InfraContext, clusterName string, search *repos.SearchFilter, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.Node], error) {
	filter := repos.Filter{
		"accountName": ctx.AccountName,
		"clusterName": clusterName,
	}

	return d.nodeRepo.FindPaginated(ctx, d.nodeRepo.MergeSearchFilter(filter, search), pagination)
}

//TODO (nxtcoder17): Deleting node should also be available

func (d *domain) GetNode(ctx InfraContext, clusterName string, nodeName string) (*entities.Node, error) {
	return d.findNode(ctx, clusterName, nodeName)
}

func (d *domain) findNode(ctx InfraContext, clusterName string, nodeName string) (*entities.Node, error) {
	node, err := d.nodeRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": nodeName,
	})
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, fmt.Errorf("no node with name %q found in cluster %q", nodeName, clusterName)
	}

	return node, nil
}

func (d *domain) OnNodeUpdateMessage(ctx InfraContext, clusterName string, node entities.Node) error {
	if _, err := d.nodeRepo.Upsert(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": node.Name,
	}, &node); err != nil {
		return err
	}
	return nil
}

func (d *domain) OnNodeDeleteMessage(ctx InfraContext, clusterName string, node entities.Node) error {
	n, err := d.findNode(ctx, clusterName, node.Name)
	if err != nil {
		return err
	}

	return d.nodeRepo.DeleteById(ctx, n.Id)
}
