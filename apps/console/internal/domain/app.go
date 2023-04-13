package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) CreateApp(ctx ConsoleContext, app entities.App) (*entities.App, error) {
	if err := d.canMutateResourcesInProject(ctx, app.Namespace); err != nil {
		return nil, err
	}

	app.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &app.App); err != nil {
		return nil, err
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, ""),
		},
		Action: string(iamT.MutateResourcesInProject),
	})
	if err != nil {
		return nil, err
	}

	if co.Status {
		return nil, fmt.Errorf("Unauthorized")
	}

	app.AccountName = ctx.AccountName
	app.ClusterName = ctx.ClusterName
	app.SyncStatus = t.GetSyncStatusForCreation()

	nApp, err := d.appRepo.Create(ctx, &app)
	if err != nil {
		if d.appRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("app with name %q already exists", app.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nApp.App); err != nil {
		return nil, err
	}

	return nApp, nil
}

func (d *domain) DeleteApp(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return err
	}

	app, err := d.findApp(ctx, namespace, name)
	if err != nil {
		return err
	}

	app.SyncStatus = t.GetSyncStatusForDeletion(app.Generation)

	if _, err := d.appRepo.UpdateById(ctx, app.Id, app); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &app.App)
}

func (d *domain) UpdateApp(ctx ConsoleContext, app entities.App) (*entities.App, error) {
	if err := d.canMutateResourcesInProject(ctx, app.Namespace); err != nil {
		return nil, err
	}

	app.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &app.App); err != nil {
		return nil, err
	}

	exApp, err := d.findApp(ctx, app.Namespace, app.Name)
	if err != nil {
		return nil, err
	}

	exApp.Spec = app.Spec
	exApp.SyncStatus = t.GetSyncStatusForUpdation(app.Generation + 1)

	upApp, err := d.appRepo.UpdateById(ctx, exApp.Id, exApp)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upApp.App); err != nil {
		return nil, err
	}

	return upApp, nil
}

func (d *domain) ListApps(ctx ConsoleContext, namespace string) ([]*entities.App, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}

	return d.appRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}})
}

func (d *domain) GetApp(ctx ConsoleContext, namespace string, name string) (*entities.App, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}

	return d.findApp(ctx, namespace, name)
}

func (d *domain) findApp(ctx ConsoleContext, namespace string, name string) (*entities.App, error) {
	app, err := d.appRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, fmt.Errorf("no app with name=%q,namespace=%q found", name, namespace)
	}
	return app, nil
}

func (d *domain) OnDeleteAppMessage(ctx ConsoleContext, app entities.App) error {
	a, err := d.findApp(ctx, app.Namespace, app.Name)
	if err != nil {
		return err
	}

	return d.appRepo.DeleteById(ctx, a.Id)
}

func (d *domain) OnUpdateAppMessage(ctx ConsoleContext, app entities.App) error {
	a, err := d.findApp(ctx, app.Namespace, app.Name)
	if err != nil {
		return err
	}

	a.Status = app.Status
	a.SyncStatus.LastSyncedAt = time.Now()
	a.SyncStatus.State = t.ParseSyncState(app.Status.IsReady)

	_, err = d.appRepo.UpdateById(ctx, a.Id, a)
	return err
}

func (d *domain) OnApplyAppError(ctx ConsoleContext, err error, namespace, name string) error {
	a, err2 := d.findApp(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	a.SyncStatus.Error = err.Error()
	_, err2 = d.appRepo.UpdateById(ctx, a.Id, a)
	return err2
}
