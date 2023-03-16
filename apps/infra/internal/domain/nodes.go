package domain

import (
	"fmt"

	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/constants"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (d *domain) GetNodePools(ctx InfraContext, clusterName string, edgeName string) ([]*entities.NodePool, error) {
	_, err := d.findCluster(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	var nodePools infraV1.NodePoolList
	if err := d.k8sClient.List(ctx, &nodePools, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string{
			//constants.ClusterNameKey: cluster.Name,
			constants.EdgeNameKey: edgeName,
		}),
	}); err != nil {
		return nil, err
	}

	results := make([]*entities.NodePool, len(nodePools.Items))
	for i := range nodePools.Items {
		results[i] = &entities.NodePool{
			NodePool: nodePools.Items[i],
		}
	}
	return results, nil
}

func (d *domain) GetMasterNodes(ctx InfraContext, clusterName string) ([]*entities.MasterNode, error) {
	cluster, err := d.findCluster(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	var mNodesList cmgrV1.MasterNodeList
	if err := d.k8sClient.List(ctx, &mNodesList, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string{constants.ClusterNameKey: cluster.Name}),
	}); err != nil {
		return nil, err
	}

	results := make([]*entities.MasterNode, len(mNodesList.Items))
	for i := range mNodesList.Items {
		results[i] = &entities.MasterNode{
			MasterNode: mNodesList.Items[i],
		}
	}
	return results, nil
}

func (d *domain) GetWorkerNodes(ctx InfraContext, clusterName string, edgeName string) ([]*entities.WorkerNode, error) {
	cluster, err := d.findCluster(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	var wNodes infraV1.WorkerNodeList
	if err := d.k8sClient.List(ctx, &wNodes, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string{
			constants.ClusterNameKey: cluster.Name,
			constants.EdgeNameKey:    edgeName,
		}),
	}); err != nil {
		return nil, err
	}

	results := make([]*entities.WorkerNode, len(wNodes.Items))
	for i := range wNodes.Items {
		results[i] = &entities.WorkerNode{
			WorkerNode: wNodes.Items[i],
		}
	}
	return results, nil
}

func (d *domain) DeleteWorkerNode(ctx InfraContext, clusterName string, edgeName string, name string) (bool, error) {
	cluster, err := d.findCluster(ctx, clusterName)
	if err != nil {
		return false, err
	}

	if err := d.deleteK8sResource(ctx, &cluster.Cluster); err != nil {
		return false, err
	}
	return true, err
}

func (d *domain) OnDeleteWorkerNodeMessage(ctx InfraContext, workerNode entities.WorkerNode) error {
	return d.workerNodeRepo.DeleteOne(ctx, repos.Filter{"metadata.name": workerNode.Name, "spec.edgeName": workerNode.Spec.EdgeName})
}

func (d *domain) OnUpdateWorkerNodeMessage(ctx InfraContext, workerNode entities.WorkerNode) error {
	wn, err := d.workerNodeRepo.FindOne(ctx, repos.Filter{"metadata.name": workerNode.Name})
	if err != nil {
		return err
	}
	if wn == nil {
		return fmt.Errorf("worker node %s not found", workerNode.Name)
	}
	wn.WorkerNode = workerNode.WorkerNode
	_, err = d.workerNodeRepo.UpdateById(ctx, wn.Id, wn)
	return err
}
