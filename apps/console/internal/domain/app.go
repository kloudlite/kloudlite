package domain

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateApp(ctx context.Context, app entities.App) (*entities.App, error) {
	app.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &app.App); err != nil {
		return nil, err
	}

	nApp, err := d.appRepo.Create(ctx, &app)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nApp.App); err != nil {
		return nil, err
	}

	return nApp, nil
}

func (d *domain) DeleteApp(ctx context.Context, namespace string, name string) error {
	app, err := d.findApp(ctx, namespace, name)
	if err != nil {
		return err
	}

	if app.GetDeletionTimestamp() != nil {
		return errAlreadyMarkedForDeletion("app", app.Namespace, app.Name)
	}

	app.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
	if _, err := d.appRepo.UpdateById(ctx, app.Id, app); err != nil {
		return err
	}

	return d.k8sYamlClient.DeleteResource(ctx, &app.App)
}

func (d *domain) UpdateApp(ctx context.Context, app entities.App) (*entities.App, error) {
	app.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &app.App); err != nil {
		return nil, err
	}

	exApp, err := d.findApp(ctx, app.Namespace, app.Name)
	if err != nil {
		return nil, err
	}

	if exApp.GetDeletionTimestamp() != nil {
		return nil, errAlreadyMarkedForDeletion("app", app.Namespace, app.Name)
	}

	status := exApp.Status
	exApp.App = app.App
	exApp.Status = status

	upApp, err := d.appRepo.UpdateById(ctx, exApp.Id, exApp)
	if err != nil {
		return nil, err
	}
	return upApp, nil
}

func (d *domain) GetApps(ctx context.Context, namespace string) ([]*entities.App, error) {
	return d.appRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
}

func (d *domain) GetApp(ctx context.Context, namespace string, name string) (*entities.App, error) {
	return d.findApp(ctx, namespace, name)
}

func (d *domain) findApp(ctx context.Context, namespace string, name string) (*entities.App, error) {
	app, err := d.appRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, fmt.Errorf("no app with name=%s,namespace=%s found", name, namespace)
	}
	return app, nil
}
