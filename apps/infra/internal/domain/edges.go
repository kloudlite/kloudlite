package domain

import (
	"fmt"

	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateEdge(ctx InfraContext, edge entities.Edge) (*entities.Edge, error) {
	edge.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &edge.Edge); err != nil {
		return nil, err
	}

	nEdge, err := d.edgeRepo.Create(ctx, &edge)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nEdge.Edge); err != nil {
		return nil, err
	}
	return nEdge, err
}

func (d *domain) ListEdges(ctx InfraContext, clusterName string, providerName *string) ([]*entities.Edge, error) {
	f := repos.Filter{"spec.clusterName": clusterName}
	if providerName != nil {
		f["spec.providerName"] = providerName
	}
	return d.edgeRepo.Find(ctx, repos.Query{Filter: f})
}

func (d *domain) GetEdge(ctx InfraContext, clusterName string, name string) (*entities.Edge, error) {
	return d.edgeRepo.FindOne(ctx, repos.Filter{"metadata.name": name, "spec.clusterName": clusterName})
}

func (d *domain) UpdateEdge(ctx InfraContext, edge entities.Edge) (*entities.Edge, error) {
	edge.EnsureGVK()
	_, err := d.findCluster(ctx, edge.Spec.ClusterName)
	if err != nil {
		return nil, err
	}

	uEdge, err := d.edgeRepo.UpdateOne(ctx, repos.Filter{"id": edge.Id}, &edge)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &uEdge.Edge); err != nil {
		return nil, err
	}
	return uEdge, nil
}

func (d *domain) DeleteEdge(ctx InfraContext, clusterName string, name string) error {
	return d.k8sClient.Delete(ctx, &infraV1.Edge{ObjectMeta: metav1.ObjectMeta{Name: name}})
}

func (d *domain) OnDeleteEdgeMessage(ctx InfraContext, edge entities.Edge) error {
	return d.edgeRepo.DeleteOne(ctx, repos.Filter{"metadata.name": edge.Name})
}

func (d *domain) OnUpdateEdgeMessage(ctx InfraContext, edge entities.Edge) error {
	e, err := d.edgeRepo.FindOne(ctx, repos.Filter{"metadata.name": edge.Name})
	if err != nil {
		return err
	}

	if e == nil {
		return fmt.Errorf("edge %s not found", edge.Name)
	}

	e.Edge = edge.Edge
	_, err = d.edgeRepo.UpdateById(ctx, e.Id, e)
	return err
}
