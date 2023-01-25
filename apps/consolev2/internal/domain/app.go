package domain

import (
	"context"
	fn "kloudlite.io/pkg/functions"

	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (d *domain) InstallApp(ctx context.Context, app entities.App) (*entities.App, error) {
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

	clusterId, err := d.getClusterForProject(ctx, nApp.Spec.ProjectName)
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

	clusterId, err := d.getClusterForProject(ctx, existingApp.Spec.ProjectName)
	if err != nil {
		return nil, err
	}

	existingApp.App = app.App
	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(existingApp.Id), existingApp.App); err != nil {
		return nil, err
	}

	return existingApp, nil
}

func (d *domain) GetApps(ctx context.Context, projectName string) ([]*entities.App, error) {
	return d.appRepo.Find(ctx, repos.Query{Filter: repos.Filter{"spec.projectName": projectName}})
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

	clusterId, err := d.getClusterForProject(ctx, app.Spec.ProjectName)
	if err != nil {
		return err
	}

	app.Spec.Frozen = true

	return d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(app.Id), app)
}

func (d *domain) UnFreezeApp(ctx context.Context, appName string) error {
	app, err := d.appRepo.FindOne(ctx, repos.Filter{"metadata.name": appName})
	if err != nil {
		return err
	}
	if app == nil {
		return errors.Newf("no app with name '%s' found", appName)
	}

	clusterId, err := d.getClusterForProject(ctx, app.Spec.ProjectName)
	if err != nil {
		return err
	}

	app.Spec.Frozen = false
	return d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(app.Id), app)
}

func (d *domain) RestartApp(ctx context.Context, appName string) error {
	app, err := d.appRepo.FindOne(ctx, repos.Filter{"metadata.name": appName})
	if err != nil {
		return err
	}
	if app == nil {
		return errors.Newf("no app with name '%s' found", appName)
	}

	clusterId, err := d.getClusterForProject(ctx, app.Spec.ProjectName)
	if err != nil {
		return err
	}

	app.Restart = fn.New(true)

	return d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(app.Id), app)
}

func (d *domain) GetApp(ctx context.Context, appName string) (*entities.App, error) {
	return d.appRepo.FindOne(ctx, repos.Filter{"metadata.name": appName})
}

func (d *domain) DeleteApp(ctx context.Context, appName string) (bool, error) {
	if err := d.appRepo.DeleteOne(ctx, repos.Filter{"metadata.name": appName}); err != nil {
		return false, err
	}
	return true, nil
}
