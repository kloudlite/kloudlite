package domain

import (
	"encoding/json"
	"fmt"

	t "github.com/kloudlite/operator/agent/types"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/domain/entities"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
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

	iamClient iam.IAMClient

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
		AccountName: ctx.AccountName,
		ClusterName: ctx.ClusterName,
		Action:      t.ActionApply,
		Object:      m,
	})
	if err != nil {
		return err
	}

	_, err = d.producer.Produce(ctx, ctx.ClusterName+"-incoming", obj.GetNamespace(), b)
	return err
}

func (d *domain) deleteK8sResource(ctx ConsoleContext, obj client.Object) error {
	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return err
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.AccountName,
		ClusterName: ctx.ClusterName,
		Action:      t.ActionDelete,
		Object:      m,
	})
	if err != nil {
		return err
	}
	_, err = d.producer.Produce(ctx, ctx.ClusterName+"-incoming", obj.GetNamespace(), b)
	return err
}

func (d *domain) canMutateResourcesInProject(ctx ConsoleContext, project string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, project),
		},
		Action: string(iamT.MutateResourcesInProject),
	})
	if err != nil {
		return err
	}
	if !co.Status {
		return fmt.Errorf("unauthorized to mutate resources in project %q", project)
	}
	return nil
}

func (d *domain) canReadResourcesInProject(ctx ConsoleContext, project string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, project),
		},
		Action: string(iamT.GetProject),
	})
	if err != nil {
		return err
	}
	if !co.Status {
		return fmt.Errorf("unauthorized to read resources in project %q", project)
	}
	return nil
}

var Module = fx.Module("domain",
	fx.Provide(func(
		k8sYamlClient *kubectl.YAMLClient,
		k8sExtendedClient k8s.ExtendedK8sClient,

		producer redpanda.Producer,

		iamClient iam.IAMClient,

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

			iamClient: iamClient,

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
