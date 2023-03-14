package domain

import (
	"context"
	"fmt"

	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/repos"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type domain struct {
	projectRepo       repos.DbRepo[*entities.Project]
	appRepo           repos.DbRepo[*entities.App]
	configRepo        repos.DbRepo[*entities.Config]
	secretRepo        repos.DbRepo[*entities.Secret]
	routerRepo        repos.DbRepo[*entities.Router]
	msvcRepo          repos.DbRepo[*entities.MSvc]
	mresRepo          repos.DbRepo[*entities.MRes]
	k8sExtendedClient k8s.ExtendedK8sClient
	k8sYamlClient     *kubectl.YAMLClient
}

func errAlreadyMarkedForDeletion(label, namespace, name string) error {
	return fmt.Errorf("app (namespace=%s, name=%s) already marked for deletion", namespace, name)
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

var Module = fx.Module("domain",
	fx.Provide(func(
		k8sYamlClient *kubectl.YAMLClient,
		k8sExtendedClient k8s.ExtendedK8sClient,

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
			projectRepo:       projectRepo,
			appRepo:           appRepo,
			configRepo:        configRepo,
			routerRepo:        routerRepo,
			secretRepo:        secretRepo,
			msvcRepo:          msvcRepo,
			mresRepo:          mresRepo,
		}
	}),
)
