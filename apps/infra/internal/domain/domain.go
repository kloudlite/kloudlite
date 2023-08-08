package domain

import (
	"encoding/json"
	"fmt"
	iamT "kloudlite.io/apps/iam/types"
	"strconv"

	"kloudlite.io/apps/infra/internal/entities"
	"kloudlite.io/common"

	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/env"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"context"

	t "github.com/kloudlite/operator/agent/types"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "kloudlite.io/pkg/types"
)

type domain struct {
	env *env.Env

	byocClusterRepo repos.DbRepo[*entities.BYOCCluster]
	clusterRepo     repos.DbRepo[*entities.Cluster]
	nodeRepo        repos.DbRepo[*entities.Node]
	k8sClient       client.Client
	nodePoolRepo    repos.DbRepo[*entities.NodePool]

	secretRepo repos.DbRepo[*entities.CloudProviderSecret]

	producer          redpanda.Producer
	k8sYamlClient     *kubectl.YAMLClient
	k8sExtendedClient k8s.ExtendedK8sClient
	iamClient         iam.IAMClient
}

func (d *domain) applyToTargetCluster(ctx InfraContext, clusterName string, obj client.Object, recordVersion int) error {
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string, 1)
	}
	ann[constants.RecordVersionKey] = fmt.Sprintf("%d", recordVersion)
	obj.SetAnnotations(ann)

	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return err
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.AccountName,
		ClusterName: clusterName,
		Action:      t.ActionApply,
		Object:      m,
	})
	if err != nil {
		return err
	}

	_, err = d.producer.Produce(ctx, common.GetKafkaTopicName(ctx.AccountName, clusterName), obj.GetNamespace(), b)
	return err
}

func (d *domain) deleteFromTargetCluster(ctx InfraContext, clusterName string, obj client.Object) error {
	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return err
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.AccountName,
		ClusterName: clusterName,
		Action:      t.ActionDelete,
		Object:      m,
	})
	if err != nil {
		return err
	}
	_, err = d.producer.Produce(ctx, common.GetKafkaTopicName(ctx.AccountName, clusterName), obj.GetNamespace(), b)
	return err
}

func (d *domain) resyncToTargetCluster(ctx InfraContext, action types.SyncAction, clusterName string, obj client.Object, recordVersion int) error {
	switch action {
	case types.SyncActionApply:
		return d.applyToTargetCluster(ctx, clusterName, obj, recordVersion)
	case types.SyncActionDelete:
		return d.deleteFromTargetCluster(ctx, clusterName, obj)
	}
	return fmt.Errorf("unknonw action: %q", action)
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

func (d *domain) matchRecordVersion(annotations map[string]string, rv int) error {
	annVersion, err := d.parseRecordVersionFromAnnotations(annotations)
	if err != nil {
		return err
	}

	if annVersion != rv {
		return fmt.Errorf("record version mismatch, expected %d, got %d", rv, annVersion)
	}

	return nil
}

func (d *domain) getAccountNamespace(accountName string) string {
	// TODO(nxtcoder17): need to fix this to use accounts-api GRPC, once accounts api is up and running
	return fmt.Sprintf("kl-account-%s", accountName)
}

func (d *domain) ensureNamespaceForAccount(ctx context.Context, accountName string) error {
	namespace := d.getAccountNamespace(accountName)
	var ns corev1.Namespace
	if err := d.k8sClient.Get(ctx, fn.NN("", namespace), &ns); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}

		return d.k8sClient.Create(ctx, &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
				Labels: map[string]string{
					constants.AccountNameKey: accountName,
				},
			},
		})
	}

	return nil
}

func (d *domain) canPerformActionInAccount(ctx InfraContext, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(action),
	})

	if err != nil {
		return err
	}
	if !co.Status {
		return fmt.Errorf("unauthorized to perform action %q in account %q", action, ctx.AccountName)
	}
	return nil
}

var Module = fx.Module("domain",
	fx.Provide(
		func(
			env *env.Env,
			byocClusterRepo repos.DbRepo[*entities.BYOCCluster],
			clusterRepo repos.DbRepo[*entities.Cluster],
			nodeRepo repos.DbRepo[*entities.Node],
			nodePoolRepo repos.DbRepo[*entities.NodePool],
			secretRepo repos.DbRepo[*entities.CloudProviderSecret],

			financeClient finance.FinanceClient,

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
				nodeRepo:        nodeRepo,
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
