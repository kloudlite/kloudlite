package domain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kloudlite/operator/agent/types"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/agent"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type domain struct {
	env *env.Env

	clusterRepo     repos.DbRepo[*entities.Cluster]
	byocClusterRepo repos.DbRepo[*entities.BYOCCluster]
	edgeRepo        repos.DbRepo[*entities.Edge]
	providerRepo    repos.DbRepo[*entities.CloudProvider]
	k8sClient       client.Client
	masterNodeRepo  repos.DbRepo[*entities.MasterNode]
	workerNodeRepo  repos.DbRepo[*entities.WorkerNode]
	nodePoolRepo    repos.DbRepo[*entities.NodePool]

	secretRepo repos.DbRepo[*entities.Secret]

	producer          redpanda.Producer
	k8sYamlClient     *kubectl.YAMLClient
	k8sExtendedClient k8s.ExtendedK8sClient
	iamClient         iam.IAMClient
}

func (d *domain) dispatchToTargetAgent(ctx InfraContext, action agent.Action, clusterName string, obj client.Object) error {
	// b, err := fn.K8sObjToYAML(obj)
	// if err != nil {
	// 	return err
	// }
	var m map[string]any
	if err := fn.JsonConversion(obj, &m); err != nil {
		return err
	}

	b, err := json.Marshal(types.AgentMessage{
		AccountName: ctx.AccountName,
		ClusterName: clusterName,
		Action:      types.ActionApply,
		Object:      m,
	})
	if err != nil {
		return err
	}

	_, err = d.producer.Produce(ctx, clusterName+"-incoming", obj.GetNamespace(), b)
	return err
}

func (d *domain) applyK8sResource(ctx InfraContext, obj client.Object, recordVersion int) error {
	if recordVersion > 0 {
		ann := obj.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string, 1)
		}
		ann[constants.RecordVersionKey] = fmt.Sprintf("%d", recordVersion)
		obj.SetAnnotations(ann)
	}

	b, err := fn.K8sObjToYAML(obj)
	if err != nil {
		return err
	}
	if _, err := d.k8sYamlClient.ApplyYAML(ctx, b); err != nil {
		return err
	}

	return nil
}

func (d *domain) deleteK8sResource(ctx InfraContext, obj client.Object) error {
	b, err := fn.K8sObjToYAML(obj)
	if err != nil {
		return err
	}

	if err := d.k8sYamlClient.DeleteYAML(ctx, b); err != nil {
		return err
	}
	return nil
}

func (d *domain) parseRecordVersionFromAnnotations(annotations map[string]string) (int, error) {
	annotatedVersion, ok := annotations[constants.RecordVersionKey]
	if !ok {
		return 0, fmt.Errorf("no annotation with record version key (%s), found on the resource", constants.RecordVersionKey)
	}

	annVersion, err := strconv.ParseInt(annotatedVersion, 10, 32)
	if err != nil {
		return 0, err
	}

	return int(annVersion), nil
}

var Module = fx.Module("domain",
	fx.Provide(
		func(
			env *env.Env,
			clusterRepo repos.DbRepo[*entities.Cluster],
			byocClusterRepo repos.DbRepo[*entities.BYOCCluster],
			providerRepo repos.DbRepo[*entities.CloudProvider],
			edgeRepo repos.DbRepo[*entities.Edge],
			masterNodeRepo repos.DbRepo[*entities.MasterNode],
			workerNodeRepo repos.DbRepo[*entities.WorkerNode],
			nodePoolRepo repos.DbRepo[*entities.NodePool],
			secretRepo repos.DbRepo[*entities.Secret],

			financeClient finance.FinanceClient,

			// agentMessenger agent.Sender,
			producer redpanda.Producer,

			k8sClient client.Client,
			k8sYamlClient *kubectl.YAMLClient,
			k8sExtendedClient k8s.ExtendedK8sClient,
			iamClient iam.IAMClient,
		) Domain {
			return &domain{
				env: env,

				clusterRepo:     clusterRepo,
				byocClusterRepo: byocClusterRepo,
				providerRepo:    providerRepo,
				edgeRepo:        edgeRepo,
				masterNodeRepo:  masterNodeRepo,
				workerNodeRepo:  workerNodeRepo,
				nodePoolRepo:    nodePoolRepo,
				secretRepo:      secretRepo,

				producer: producer,

				k8sClient:         k8sClient,
				k8sYamlClient:     k8sYamlClient,
				k8sExtendedClient: k8sExtendedClient,
				iamClient:         iamClient,
			}
		}),
)
