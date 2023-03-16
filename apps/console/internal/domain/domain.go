package domain

import (
	"fmt"

	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/agent"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type domain struct {
	k8sExtendedClient k8s.ExtendedK8sClient
	k8sYamlClient     *kubectl.YAMLClient

	agentSender agent.Sender

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
	b, err := fn.K8sObjToYAML(obj)
	if err != nil {
		return err
	}

	_, err = d.agentSender.Dispatch(ctx, ctx.clusterName+"-incoming", obj.GetNamespace(), agent.Message{
		Action: agent.Apply,
		Yamls:  b,
	})
	return err
}

func (d *domain) deleteK8sResource(ctx ConsoleContext, obj client.Object) error {
	b, err := fn.K8sObjToYAML(obj)
	if err != nil {
		return err
	}

	_, err = d.agentSender.Dispatch(ctx, ctx.clusterName+"-incoming", obj.GetNamespace(), agent.Message{
		Action: agent.Delete,
		Yamls:  b,
	})
	return err
}

var Module = fx.Module("domain",
	fx.Provide(func(
		k8sYamlClient *kubectl.YAMLClient,
		k8sExtendedClient k8s.ExtendedK8sClient,

		agentSender agent.Sender,

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

			agentSender: agentSender,

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
