package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/entities"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) ListApps(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.App], error) {
	ws, err := d.findWorkspaceByTargetNs(ctx, namespace)
	if err != nil {
		return nil, err
	}

	if err := d.canReadResourcesInWorkspaceOrEnv(ctx, ws.ProjectName, ws); err != nil {
		return nil, err
	}

	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}

	return d.appRepo.FindPaginated(ctx, d.appRepo.MergeMatchFilters(filter, search), pq)
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

func (d *domain) GetApp(ctx ConsoleContext, namespace string, name string) (*entities.App, error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, err
	}

	return d.findApp(ctx, namespace, name)
}

// mutations
func (d *domain) CreateApp(ctx ConsoleContext, app entities.App) (*entities.App, error) {
	ws, err := d.findWorkspaceByTargetNs(ctx, app.Namespace)
	if err != nil {
		return nil, err
	}

	if err := d.canMutateResourcesInWorkspaceOrEnv(ctx, ws.ProjectName, ws); err != nil {
		return nil, err
	}

	app.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &app.App); err != nil {
		return nil, err
	}

	app.IncrementRecordVersion()

	app.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	app.LastUpdatedBy = app.CreatedBy

	app.AccountName = ctx.AccountName
	app.ClusterName = ctx.ClusterName
	app.ProjectName = ws.ProjectName
	app.WorkspaceName = ws.Name
	app.SyncStatus = t.GenSyncStatus(t.SyncActionApply, app.RecordVersion)

	nApp, err := d.appRepo.Create(ctx, &app)
	if err != nil {
		if d.appRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, err
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nApp.App, nApp.RecordVersion); err != nil {
		return nil, err
	}

	return nApp, nil
}

func (d *domain) DeleteApp(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

	app, err := d.findApp(ctx, namespace, name)
	if err != nil {
		return err
	}

	app.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, app.RecordVersion)

	if _, err := d.appRepo.UpdateById(ctx, app.Id, app); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &app.App)
}

func (d *domain) UpdateApp(ctx ConsoleContext, app entities.App) (*entities.App, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, app.Namespace); err != nil {
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

	exApp.IncrementRecordVersion()

	exApp.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	exApp.DisplayName = app.DisplayName

	exApp.Labels = app.Labels
	exApp.Annotations = app.Annotations
	exApp.Spec = app.Spec
	exApp.SyncStatus = t.GenSyncStatus(t.SyncActionApply, exApp.RecordVersion)

	upApp, err := d.appRepo.UpdateById(ctx, exApp.Id, exApp)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upApp.App, upApp.RecordVersion); err != nil {
		return nil, err
	}

	return upApp, nil
}

func (d *domain) OnUpdateAppMessage(ctx ConsoleContext, app entities.App) error {
	exApp, err := d.findApp(ctx, app.Namespace, app.Name)
	if err != nil {
		return err
	}

	annotatedVersion, err := d.parseRecordVersionFromAnnotations(app.Annotations)
	if err != nil {
		return d.resyncK8sResource(ctx, exApp.SyncStatus.Action, &exApp.App, exApp.RecordVersion)
	}

	if annotatedVersion != exApp.RecordVersion {
		return d.resyncK8sResource(ctx, exApp.SyncStatus.Action, &exApp.App, exApp.RecordVersion)
	}

	if err := d.MatchRecordVersion(app.Annotations, exApp.RecordVersion); err != nil {
	}

	exApp.CreationTimestamp = app.CreationTimestamp
	exApp.Labels = app.Labels
	exApp.Annotations = app.Annotations
	exApp.Generation = app.Generation

	exApp.Status = app.Status

	exApp.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	exApp.SyncStatus.RecordVersion = exApp.RecordVersion
	exApp.SyncStatus.Error = nil
	exApp.SyncStatus.LastSyncedAt = time.Now()

	_, err = d.appRepo.UpdateById(ctx, exApp.Id, exApp)
	return err
}

func (d *domain) OnDeleteAppMessage(ctx ConsoleContext, app entities.App) error {
	a, err := d.findApp(ctx, app.Namespace, app.Name)
	if err != nil {
		return err
	}

	if err := d.MatchRecordVersion(app.Annotations, a.RecordVersion); err != nil {
		return err
	}

	return d.appRepo.DeleteById(ctx, a.Id)
}

func (d *domain) OnApplyAppError(ctx ConsoleContext, errMsg string, namespace string, name string) error {
	a, err2 := d.findApp(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	a.SyncStatus.State = t.SyncStateErroredAtAgent
	a.SyncStatus.LastSyncedAt = time.Now()
	a.SyncStatus.Error = &errMsg

	_, err := d.appRepo.UpdateById(ctx, a.Id, a)
	return err
}

func (d *domain) ResyncApp(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return err
	}

	a, err := d.findApp(ctx, namespace, name)
	if err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, a.SyncStatus.Action, &a.App, a.RecordVersion)
}
