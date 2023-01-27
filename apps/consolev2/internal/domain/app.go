package domain

import (
	"context"
	"fmt"
	fn "kloudlite.io/pkg/functions"

	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateApp(ctx context.Context, app entities.App) (*entities.App, error) {
	existingApp, err := d.appRepo.FindOne(ctx, repos.Filter{"metadata.name": app.Name})
	if err != nil {
		return nil, err
	}
	if existingApp != nil {
		return nil, errors.Newf("app %s already exists", app.Name)
	}

	nApp, err := d.appRepo.Create(ctx, &app)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterIdForNamespace(ctx, app.Namespace)
	if err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(nApp.Id), nApp.App); err != nil {
		return nil, err
	}
	return nApp, nil
}

func (d *domain) UpdateApp(ctx context.Context, app entities.App) (*entities.App, error) {
	existingApp, err := d.appRepo.FindOne(ctx, repos.Filter{"metadata.name": app.Name, "metadata.namespace": app.Namespace})
	if err != nil {
		return nil, err
	}
	if existingApp == nil {
		return nil, errors.Newf("app %s does not exist", app.Name)
	}

	clusterId, err := d.getClusterIdForNamespace(ctx, existingApp.Namespace)
	if err != nil {
		return nil, err
	}

	existingApp.App = app.App
	uApp, err := d.appRepo.UpdateById(ctx, existingApp.Id, existingApp)
	if err != nil {
		return nil, err
	}
	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(existingApp.Id), uApp.App); err != nil {
		return nil, err
	}

	return uApp, nil
}

func (d *domain) GetApps(ctx context.Context, namespace string, search *string) ([]*entities.App, error) {
	if search == nil {
		return d.appRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
	}
	return d.appRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace, "metadata.name": fmt.Sprintf("/%s/", *search)}})
}

func (d *domain) GetApp(ctx context.Context, namespace string, name string) (*entities.App, error) {
	return d.appRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
}

func (d *domain) GetInterceptedApps(ctx context.Context, deviceName string) ([]*entities.App, error) {
	return d.appRepo.Find(ctx, repos.Query{Filter: repos.Filter{"spec.interception.forDevice": deviceName}})
}

func (d *domain) FreezeApp(ctx context.Context, appName string) error {
	app, err := d.appRepo.FindOne(ctx, repos.Filter{"metadata.name": appName})
	if err != nil {
		return err
	}
	if app == nil {
		return errors.Newf("no app with name '%s' found", appName)
	}

	clusterId, err := d.getClusterIdForNamespace(ctx, app.Namespace)
	if err != nil {
		return err
	}

	app.Spec.Frozen = true
	return d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(app.Id), app.App)
}

func (d *domain) UnFreezeApp(ctx context.Context, appName string) error {
	app, err := d.appRepo.FindOne(ctx, repos.Filter{"metadata.name": appName})
	if err != nil {
		return err
	}
	if app == nil {
		return errors.Newf("no app with name '%s' found", appName)
	}

	clusterId, err := d.getClusterForProject(ctx, app.Namespace)
	if err != nil {
		return err
	}

	app.Spec.Frozen = false
	return d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(app.Id), app.App)
}

func (d *domain) RestartApp(ctx context.Context, appName string) error {
	app, err := d.appRepo.FindOne(ctx, repos.Filter{"metadata.name": appName})
	if err != nil {
		return err
	}
	if app == nil {
		return errors.Newf("no app with name '%s' found", appName)
	}

	clusterId, err := d.getClusterIdForNamespace(ctx, app.Namespace)
	if err != nil {
		return err
	}

	app.Restart = fn.New(true)
	return d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(app.Id), app.App)
}

func (d *domain) DeleteApp(ctx context.Context, namespace, name string) (bool, error) {
	app, err := d.appRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
	if err != nil {
		return false, err
	}
	if app == nil {
		return true, nil
	}

	clusterId, err := d.getClusterIdForNamespace(ctx, namespace)
	if err != nil {
		return false, err
	}

	if err := d.workloadMessenger.SendAction(ActionDelete, d.getDispatchKafkaTopic(clusterId), string(app.Id), app.App); err != nil {
		return false, err
	}

	if err := d.appRepo.DeleteOne(ctx, repos.Filter{"metadata.name": name, "metadata.namespace": namespace}); err != nil {
		return false, err
	}
	return true, nil
}
