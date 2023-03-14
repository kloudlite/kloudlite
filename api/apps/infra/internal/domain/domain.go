package domain

import (
	"context"
	"fmt"

	cmgrV1 "github.com/kloudlite/cluster-operator/apis/cmgr/v1"
	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type domain struct {
	clusterRepo       repos.DbRepo[*entities.Cluster]
	edgeRepo          repos.DbRepo[*entities.Edge]
	providerRepo      repos.DbRepo[*entities.CloudProvider]
	k8sClient         client.Client
	masterNodeRepo    repos.DbRepo[*entities.MasterNode]
	workerNodeRepo    repos.DbRepo[*entities.WorkerNode]
	nodePoolRepo      repos.DbRepo[*entities.NodePool]
	agentMessenger    AgentMessenger
	k8sYamlClient     *kubectl.YAMLClient
	k8sExtendedClient k8s.ExtendedK8sClient
	secretRepo        repos.DbRepo[*entities.Secret]
}

func (d *domain) applyK8sResource(ctx context.Context, obj client.Object) error {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	if _, err := d.k8sYamlClient.ApplyYAML(ctx, b); err != nil {
		return err
	}
	return nil
}

func (d *domain) CreateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, providerSecret entities.Secret) (*entities.CloudProvider, error) {
	cloudProvider.EnsureGVK()
	providerSecret.EnsureGVK()

	if err := d.k8sExtendedClient.ValidateStruct(ctx, &providerSecret.Secret); err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &providerSecret.Secret); err != nil {
		return nil, err
	}

	cloudProvider.Spec.ProviderSecret.Name = providerSecret.Name
	cloudProvider.Spec.ProviderSecret.Namespace = providerSecret.Namespace

	if err := d.k8sExtendedClient.ValidateStruct(ctx, &cloudProvider.CloudProvider); err != nil {
		return nil, err
	}

	cp, err := d.providerRepo.Create(ctx, &cloudProvider)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &cp.CloudProvider); err != nil {
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
	cloudProvider.EnsureGVK()
	providerSecret.EnsureGVK()

	if err := d.k8sExtendedClient.ValidateStruct(ctx, &cloudProvider.CloudProvider); err != nil {
		return nil, err
	}

	cp, err := d.providerRepo.FindOne(ctx, repos.Filter{"metadata.name": cloudProvider.Name})
	if err != nil {
		return nil, err
	}

	if cp == nil {
		return nil, fmt.Errorf("cloud provider %s not found", cloudProvider.Name)
	}

	if providerSecret != nil {
		if err := d.k8sExtendedClient.ValidateStruct(ctx, &providerSecret.Secret); err != nil {
			return nil, err
		}

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

		if err := d.applyK8sResource(ctx, &uSecret.Secret); err != nil {
			return nil, err
		}

		cloudProvider.Spec.ProviderSecret.Name = providerSecret.Name
		cloudProvider.Spec.ProviderSecret.Namespace = providerSecret.Namespace
	}

	uProvider, err := d.providerRepo.UpdateById(ctx, cp.Id, &cloudProvider)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &uProvider.CloudProvider); err != nil {
		return nil, err
	}

	return nil, err
}

func (d *domain) DeleteCloudProvider(ctx context.Context, accountName string, name string) error {
	return d.k8sClient.Delete(ctx, &infraV1.CloudProvider{ObjectMeta: metav1.ObjectMeta{Name: name}})
}

func (d *domain) OnDeleteCloudProviderMessage(ctx context.Context, cloudProvider entities.CloudProvider) error {
	return d.providerRepo.DeleteOne(ctx, repos.Filter{"metadata.name": cloudProvider.Name})
}

func (d *domain) OnUpdateCloudProviderMessage(ctx context.Context, cloudProvider entities.CloudProvider) error {
	cp, err := d.providerRepo.FindOne(ctx, repos.Filter{"metadata.name": cloudProvider.Name})
	if err != nil {
		return err
	}

	if cp == nil {
		return fmt.Errorf("no cloud provider named %s found", cloudProvider.Name)
	}

	cp.CloudProvider = cloudProvider.CloudProvider
	_, err = d.providerRepo.UpdateById(ctx, cp.Id, cp)
	return err
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

	if err := d.agentMessenger.SendAction(ctx, ActionDelete, clusterName, clusterName, cluster.Cluster); err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) OnDeleteWorkerNodeMessage(ctx context.Context, workerNode entities.WorkerNode) error {
	return d.workerNodeRepo.DeleteOne(ctx, repos.Filter{"metadata.name": workerNode.Name, "spec.edgeName": workerNode.Spec.EdgeName})
}

func (d *domain) OnUpdateWorkerNodeMessage(ctx context.Context, workerNode entities.WorkerNode) error {
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

func (d *domain) GetNodePools(ctx context.Context, clusterName string, edgeName string) ([]*entities.NodePool, error) {
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

func (d *domain) CreateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error) {
	cluster.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &cluster.Cluster); err != nil {
		return nil, err
	}

	nCluster, err := d.clusterRepo.Create(ctx, &cluster)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nCluster.Cluster); err != nil {
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

func (d *domain) DeleteCluster(ctx context.Context, name string) error {
	return d.k8sClient.Delete(ctx, &cmgrV1.Cluster{ObjectMeta: metav1.ObjectMeta{Name: name}})
}

func (d *domain) OnDeleteClusterMessage(ctx context.Context, cluster entities.Cluster) error {
	return d.clusterRepo.DeleteOne(ctx, repos.Filter{"metadata.name": cluster.Name})
}

func (d *domain) OnUpdateClusterMessage(ctx context.Context, cluster entities.Cluster) error {
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

func (d *domain) CreateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error) {
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

func (d *domain) ListEdges(ctx context.Context, clusterName string, providerName *string) ([]*entities.Edge, error) {
	f := repos.Filter{"spec.clusterName": clusterName}
	if providerName != nil {
		f["spec.providerName"] = providerName
	}
	return d.edgeRepo.Find(ctx, repos.Query{Filter: f})
}

func (d *domain) GetEdge(ctx context.Context, clusterName string, name string) (*entities.Edge, error) {
	return d.edgeRepo.FindOne(ctx, repos.Filter{"metadata.name": name, "spec.clusterName": clusterName})
}

func (d *domain) UpdateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error) {
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

func (d *domain) DeleteEdge(ctx context.Context, clusterName string, name string) error {
	return d.k8sClient.Delete(ctx, &infraV1.Edge{ObjectMeta: metav1.ObjectMeta{Name: name}})
}

func (d *domain) OnDeleteEdgeMessage(ctx context.Context, edge entities.Edge) error {
	return d.edgeRepo.DeleteOne(ctx, repos.Filter{"metadata.name": edge.Name})
}

func (d *domain) OnUpdateEdgeMessage(ctx context.Context, edge entities.Edge) error {
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
			k8sYamlClient *kubectl.YAMLClient,
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

				agentMessenger: agentMessenger,

				k8sClient:         k8sClient,
				k8sYamlClient:     k8sYamlClient,
				k8sExtendedClient: k8sExtendedClient,
			}
		}),
)
