package domain

import (
	"context"
	"encoding/json"
	"fmt"
	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/labels"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type domain struct {
	clusterRepo       repos.DbRepo[*entities.Cluster]
	edgeRepo          repos.DbRepo[*entities.Edge]
	providerRepo      repos.DbRepo[*entities.CloudProvider]
	financeClient     finance.FinanceClient
	k8sClient         client.Client
	masterNodeRepo    repos.DbRepo[*entities.MasterNode]
	workerNodeRepo    repos.DbRepo[*entities.WorkerNode]
	nodePoolRepo      repos.DbRepo[*entities.NodePool]
	agentMessenger    AgentMessenger
	k8sYamlClient     *k8s.YAMLClient
	k8sExtendedClient k8s.ExtendedK8sClient
	secretRepo        repos.DbRepo[*entities.Secret]
}

func (d *domain) CreateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, providerSecret entities.Secret) (*entities.CloudProvider, error) {
	if err := d.k8sExtendedClient.ValidateStruct(ctx, providerSecret.Secret, fmt.Sprintf("%s.%s", fn.RegularPlural(providerSecret.Kind), providerSecret.GroupVersionKind().Group)); err != nil {
		return nil, err
	}
	if err := d.k8sClient.Create(ctx, &providerSecret.Secret); err != nil {
		return nil, err
	}

	cloudProvider.Spec.ProviderSecret.Name = providerSecret.Name
	cloudProvider.Spec.ProviderSecret.Namespace = providerSecret.Namespace

	if err := d.k8sExtendedClient.ValidateStruct(ctx, cloudProvider.CloudProvider, fmt.Sprintf("%s.%s", fn.RegularPlural(cloudProvider.Kind), cloudProvider.GroupVersionKind().Group)); err != nil {
		return nil, err
	}

	cp, err := d.providerRepo.Create(ctx, &cloudProvider)
	if err != nil {
		return nil, err
	}

	if err := d.k8sClient.Create(ctx, &cp.CloudProvider); err != nil {
		return nil, err
	}

	return cp, nil
}

func (d *domain) ListCloudProviders(ctx context.Context, accountName string) ([]*entities.CloudProvider, error) {
	return d.providerRepo.Find(ctx, repos.Query{Filter: repos.Filter{"spec.accountName": accountName}})
}

func (d *domain) GetCloudProvider(ctx context.Context, accountName string, name string) (*entities.CloudProvider, error) {
	return d.providerRepo.FindOne(ctx, repos.Filter{"metadata.name": name, "spec.accountName": accountName})
}

func (d *domain) UpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, providerSecret *entities.Secret) (*entities.CloudProvider, error) {
	cp, err := d.providerRepo.FindOne(ctx, repos.Filter{"metadata.name": cloudProvider.Name})
	if err != nil {
		return nil, err
	}

	if cp == nil {
		return nil, fmt.Errorf("cloud provider %s not found", cloudProvider.Name)
	}

	if providerSecret != nil {
		ps, err := d.secretRepo.FindOne(ctx, repos.Filter{"metadata.name": providerSecret.Name, "metadata.namespace": providerSecret.Namespace})
		if err != nil {
			return nil, err
		}
		if ps == nil {
			return nil, fmt.Errorf("provider %s does not exist", providerSecret.Name)
		}

		uSecret, err := d.secretRepo.UpdateById(ctx, ps.Id, providerSecret)
		if err != nil {
			return nil, err
		}

		b, err := json.Marshal(uSecret.Secret)
		if err != nil {
			return nil, err
		}

		if err := d.k8sYamlClient.ApplyYAML(ctx, b); err != nil {
			return nil, err
		}

		cloudProvider.Spec.ProviderSecret.Name = providerSecret.Name
		cloudProvider.Spec.ProviderSecret.Namespace = providerSecret.Namespace
	}

	uProvider, err := d.providerRepo.UpdateOne(ctx, repos.Filter{"id": cp.Id}, &cloudProvider)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(uProvider)
	if err != nil {
		return nil, err
	}

	if err := d.k8sYamlClient.ApplyYAML(ctx, b); err != nil {
		return nil, err
	}

	return nil, err
}

func (d *domain) DeleteCloudProvider(ctx context.Context, accountName string, name string) error {
	return d.providerRepo.DeleteOne(ctx, repos.Filter{"metadata.name": name, "spec.accountName": accountName})
}

func (d *domain) findCluster(ctx context.Context, clusterName string) (*entities.Cluster, error) {
	cluster, err := d.clusterRepo.FindOne(ctx, repos.Filter{"metadata.name": clusterName})
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster %q not found", clusterName)
	}
	return cluster, nil
}

func (d *domain) GetMasterNodes(ctx context.Context, clusterName string) ([]*entities.MasterNode, error) {
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

func (d *domain) GetWorkerNodes(ctx context.Context, clusterName string, edgeName string) ([]*entities.WorkerNode, error) {
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

func (d *domain) DeleteWorkerNode(ctx context.Context, clusterName string, edgeName string, name string) (bool, error) {
	cluster, err := d.findCluster(ctx, clusterName)
	if err != nil {
		return false, err
	}

	wNode, err := d.workerNodeRepo.FindOne(ctx, repos.Filter{"metadata.name": name, "spec.edgeName": edgeName})
	if err != nil {
		return false, err
	}

	if wNode == nil {
		return false, fmt.Errorf("worker node %s not found", name)
	}

	if err := d.agentMessenger.SendAction(ctx, ActionDelete, clusterName, clusterName, cluster.Cluster); err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) GetNodePools(ctx context.Context, clusterName string) ([]*entities.NodePool, error) {
	cluster, err := d.findCluster(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	var nodePools infraV1.NodePoolList
	if err := d.k8sClient.List(ctx, &nodePools, &client.ListOptions{
		LabelSelector: labels.SelectorFromValidatedSet(map[string]string{
			constants.ClusterNameKey: cluster.Name,
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

func (d *domain) CreateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error) {
	if err := d.k8sExtendedClient.ValidateStruct(ctx, cluster.Cluster, fmt.Sprintf("%s.%s", fn.RegularPlural(cluster.Kind), cluster.GroupVersionKind().Group)); err != nil {
		return nil, err
	}

	nCluster, err := d.clusterRepo.Create(ctx, &cluster)
	if err != nil {
		return nil, err
	}

	if err := d.k8sClient.Create(ctx, &nCluster.Cluster); err != nil {
		return nil, err
	}

	return nCluster, nil
}

func (d *domain) ListClusters(ctx context.Context, accountName string) ([]*entities.Cluster, error) {
	return d.clusterRepo.Find(ctx, repos.Query{Filter: repos.Filter{"spec.accountName": accountName}})
}

func (d *domain) GetCluster(ctx context.Context, name string) (*entities.Cluster, error) {
	return d.clusterRepo.FindOne(ctx, repos.Filter{"metadata.name": name})
}

func (d *domain) UpdateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error) {
	clus, err := d.findCluster(ctx, cluster.Name)
	if err != nil {
		return nil, err
	}
	uCluster, err := d.clusterRepo.UpdateById(ctx, clus.Id, &cluster)

	b, err := json.Marshal(uCluster.Cluster)
	if err != nil {
		return nil, err
	}

	if err := d.k8sYamlClient.ApplyYAML(ctx, b); err != nil {
		return nil, err
	}

	return uCluster, nil
}

func (d *domain) DeleteCluster(ctx context.Context, name string) error {
	cluster, err := d.findCluster(ctx, name)
	if err != nil {
		return err
	}

	if err := d.k8sClient.Delete(ctx, &cluster.Cluster); err != nil {
		return err
	}
	return nil
}

func (d *domain) CreateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error) {
	if err := d.k8sExtendedClient.ValidateStruct(ctx, edge.Edge, fmt.Sprintf("%s.%s", fn.RegularPlural(edge.Kind), edge.GroupVersionKind().Group)); err != nil {
		return nil, err
	}

	nEdge, err := d.edgeRepo.Create(ctx, &edge)
	if err != nil {
		return nil, err
	}

	if err := d.k8sClient.Create(ctx, &nEdge.Edge); err != nil {
		return nil, err
	}
	return nEdge, err
}

func (d *domain) ListEdges(ctx context.Context, clusterName string, providerName string) ([]*entities.Edge, error) {
	return d.edgeRepo.Find(ctx, repos.Query{Filter: repos.Filter{"spec.clusterName": clusterName, "spec.providerName": providerName}})
}

func (d *domain) GetEdge(ctx context.Context, clusterName string, name string) (*entities.Edge, error) {
	return d.edgeRepo.FindOne(ctx, repos.Filter{"metadata.name": name, "spec.clusterName": clusterName})
}

func (d *domain) UpdateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error) {
	_, err := d.findCluster(ctx, edge.Spec.ClusterName)
	if err != nil {
		return nil, err
	}

	uEdge, err := d.edgeRepo.UpdateOne(ctx, repos.Filter{"id": edge.Id}, &edge)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(edge.Edge)
	if err != nil {
		return nil, err
	}

	if err := d.k8sYamlClient.ApplyYAML(ctx, b); err != nil {
		return nil, err
	}
	return uEdge, nil
}

func (d *domain) DeleteEdge(ctx context.Context, clusterName string, name string) error {
	return d.edgeRepo.DeleteOne(ctx, repos.Filter{"metadata.name": name, "spec.clusterName": clusterName})
}

var Module = fx.Module("domain",
	fx.Provide(
		func(
			clusterRepo repos.DbRepo[*entities.Cluster],
			providerRepo repos.DbRepo[*entities.CloudProvider],
			edgeRepo repos.DbRepo[*entities.Edge],
			masterNodeRepo repos.DbRepo[*entities.MasterNode],
			workerNodeRepo repos.DbRepo[*entities.WorkerNode],
			nodePoolRepo repos.DbRepo[*entities.NodePool],
			secretRepo repos.DbRepo[*entities.Secret],

			financeClient finance.FinanceClient,
			agentMessenger AgentMessenger,

			k8sClient client.Client,
			k8sYamlClient *k8s.YAMLClient,
			k8sExtendedClient k8s.ExtendedK8sClient,
		) Domain {
			return &domain{
				clusterRepo:    clusterRepo,
				providerRepo:   providerRepo,
				edgeRepo:       edgeRepo,
				masterNodeRepo: masterNodeRepo,
				workerNodeRepo: workerNodeRepo,
				nodePoolRepo:   nodePoolRepo,
				secretRepo:     secretRepo,

				financeClient:  financeClient,
				agentMessenger: agentMessenger,

				k8sClient:         k8sClient,
				k8sYamlClient:     k8sYamlClient,
				k8sExtendedClient: k8sExtendedClient,
			}
		}),
)
