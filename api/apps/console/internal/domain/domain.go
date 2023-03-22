package domain

import (
	"encoding/json"
	"fmt"

	t "github.com/kloudlite/operator/agent/types"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type domain struct {
	k8sExtendedClient k8s.ExtendedK8sClient
	k8sYamlClient     *kubectl.YAMLClient

	producer redpanda.Producer

	projectRepo repos.DbRepo[*entities.Project]
	appRepo     repos.DbRepo[*entities.App]
	configRepo  repos.DbRepo[*entities.Config]
	secretRepo  repos.DbRepo[*entities.Secret]
	routerRepo  repos.DbRepo[*entities.Router]
	msvcRepo    repos.DbRepo[*entities.MSvc]
	mresRepo    repos.DbRepo[*entities.MRes]
}

func errAlreadyMarkedForDeletion(label, namespace, name string) error {
	return fmt.Errorf("%s (namespace=%s, name=%s) already marked for deletion", label, namespace, name)
}

func (d *domain) applyK8sResource(ctx ConsoleContext, obj client.Object) error {
	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return err
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.accountName,
		ClusterName: ctx.clusterName,
		Action:      t.ActionApply,
		Object:      m,
	})
	if err != nil {
		return err
	}

	_, err = d.producer.Produce(ctx, ctx.clusterName+"-incoming", obj.GetNamespace(), b)
	return err
}

func (d *domain) deleteK8sResource(ctx ConsoleContext, obj client.Object) error {
	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return err
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.accountName,
		ClusterName: ctx.clusterName,
		Action:      t.ActionDelete,
		Object:      m,
	})
	if err != nil {
		return err
	}
	_, err = d.producer.Produce(ctx, ctx.clusterName+"-incoming", obj.GetNamespace(), b)
	return err
}

var Module = fx.Module("domain",
	fx.Provide(func(
		k8sYamlClient *kubectl.YAMLClient,
		k8sExtendedClient k8s.ExtendedK8sClient,

		producer redpanda.Producer,

		projectRepo repos.DbRepo[*entities.Project],
		appRepo repos.DbRepo[*entities.App],
		configRepo repos.DbRepo[*entities.Config],
		secretRepo repos.DbRepo[*entities.Secret],
		routerRepo repos.DbRepo[*entities.Router],
		msvcRepo repos.DbRepo[*entities.MSvc],
		mresRepo repos.DbRepo[*entities.MRes],
	) Domain {
		return &domain{
			k8sExtendedClient: k8sExtendedClient,
			k8sYamlClient:     k8sYamlClient,

			producer: producer,

			projectRepo: projectRepo,
			appRepo:     appRepo,
			configRepo:  configRepo,
			routerRepo:  routerRepo,
			secretRepo:  secretRepo,
			msvcRepo:    msvcRepo,
			mresRepo:    mresRepo,
		}
	}),
)
