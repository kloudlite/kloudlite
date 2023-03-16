package domain

import (
	"fmt"

	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error) {
	cluster.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &cluster.Cluster); err != nil {
		return nil, err
	}

	cluster.AccountName = ctx.AccountName
	nCluster, err := d.clusterRepo.Create(ctx, &cluster)
	if err != nil {
		if d.clusterRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("cluster with name '%s' already exists", cluster.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nCluster.Cluster); err != nil {
		return nil, err
	}

	return nCluster, nil
}

func (d *domain) ListClusters(ctx InfraContext) ([]*entities.Cluster, error) {
	return d.clusterRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"accountName": ctx.AccountName,
		},
	})
}

func (d *domain) GetCluster(ctx InfraContext, name string) (*entities.Cluster, error) {
	return d.clusterRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"metadata.name": name,
	})
}

func (d *domain) UpdateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error) {
	cluster.EnsureGVK()
	clus, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return nil, err
	}

	clus.Cluster = cluster.Cluster
	uCluster, err := d.clusterRepo.UpdateById(ctx, clus.Id, clus)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &uCluster.Cluster); err != nil {
		return nil, err
	}

	return uCluster, nil
}

func (d *domain) DeleteCluster(ctx InfraContext, name string) error {
	return d.k8sClient.Delete(ctx, &cmgrV1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: name}})
}

func (d *domain) OnDeleteClusterMessage(ctx InfraContext, cluster entities.Cluster) error {
	return d.clusterRepo.DeleteOne(ctx, repos.Filter{"metadata.name": cluster.Name})
}

func (d *domain) OnUpdateClusterMessage(ctx InfraContext, cluster entities.Cluster) error {
	c, err := d.clusterRepo.FindOne(ctx, repos.Filter{"metadata.name": cluster.Name})
	if err != nil {
		return err
	}

	if c == nil {
		return fmt.Errorf("cluster %s not found", cluster.Name)
	}

	c.Cluster = cluster.Cluster
	_, err = d.clusterRepo.UpdateById(ctx, c.Id, c)
	return err
}

func (d *domain) findCluster(ctx InfraContext, clusterName string) (*entities.Cluster, error) {
	cluster, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"metadata.name": clusterName,
	})
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster with name %q not found", clusterName)
	}
	return cluster, nil
}
