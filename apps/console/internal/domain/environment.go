package domain

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/constants"
	"kloudlite.io/pkg/repos"
	"time"

	t "kloudlite.io/pkg/types"
)

func (d *domain) findEnvironment(ctx ConsoleContext, namespace, name string) (*entities.Environment, error) {
	ws, err := d.environmentRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})

	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, fmt.Errorf("no environment with name=%q, namespace=%q found", name, namespace)
	}
	return ws, nil
}

func (d *domain) ListEnvironments(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Environment], error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}

	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}

	return d.environmentRepo.FindPaginated(ctx, d.environmentRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) GetEnvironment(ctx ConsoleContext, namespace, name string) (*entities.Environment, error) {
	return d.findEnvironment(ctx, namespace, name)
}

func (d *domain) CreateEnvironment(ctx ConsoleContext, env entities.Environment) (*entities.Environment, error) {
	env.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &env.Environment); err != nil {
		return nil, err
	}

	if err := d.canMutateResourcesInProject(ctx, env.Namespace); err != nil {
		return nil, err
	}

	env.IncrementRecordVersion()
	env.AccountName = ctx.AccountName
	env.ClusterName = ctx.ClusterName
	env.SyncStatus = t.GenSyncStatus(t.SyncActionApply, env.RecordVersion)

	nEv, err := d.environmentRepo.Create(ctx, &env)
	if err != nil {
		if d.environmentRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, err
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &nEv.Environment, nEv.RecordVersion); err != nil {
		return nil, err
	}

	return nEv, nil
}

func (d *domain) UpdateEnvironment(ctx ConsoleContext, env entities.Environment) (*entities.Environment, error) {
	env.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &env.Environment); err != nil {
		return nil, err
	}

	if err := d.canMutateResourcesInProject(ctx, env.Namespace); err != nil {
		return nil, err
	}

	exEnv, err := d.findEnvironment(ctx, env.Namespace, env.Name)
	if err != nil {
		return nil, err
	}

	if exEnv.GetDeletionTimestamp() != nil {
		return nil, errAlreadyMarkedForDeletion("environment", "", env.Name)
	}

	exEnv.Labels = env.Labels
	exEnv.Annotations = env.Annotations
	exEnv.Spec = env.Spec
	exEnv.SyncStatus = t.GenSyncStatus(t.SyncActionApply, exEnv.RecordVersion)

	upWs, err := d.environmentRepo.UpdateById(ctx, exEnv.Id, exEnv)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upWs.Environment, upWs.RecordVersion); err != nil {
		return nil, err
	}

	return upWs, nil
}

func (d *domain) DeleteEnvironment(ctx ConsoleContext, namespace, name string) error {
	ev, err := d.findEnvironment(ctx, namespace, name)
	if err != nil {
		return err
	}

	if err := d.canMutateResourcesInProject(ctx, ev.Namespace); err != nil {
		return err
	}

	ev.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, ev.RecordVersion)
	if _, err := d.environmentRepo.UpdateById(ctx, ev.Id, ev); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &ev.Environment)
}

func (d *domain) ResyncEnvironment(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInProject(ctx, namespace); err != nil {
		return err
	}

	envt, err := d.findEnvironment(ctx, namespace, name)
	if err != nil {
		return err
	}

	if err := d.resyncK8sResource(ctx, t.SyncActionApply, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: envt.Spec.TargetNamespace,
			Labels: map[string]string{
				constants.EnvNameKey: envt.Name,
			},
		},
	}, 0); err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, envt.SyncStatus.Action, &envt.Environment, envt.RecordVersion)
}

func (d *domain) OnApplyEnvironmentError(ctx ConsoleContext, errMsg, namespace, name string) error {
	envt, err2 := d.findEnvironment(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	envt.SyncStatus.State = t.SyncStateErroredAtAgent
	envt.SyncStatus.LastSyncedAt = time.Now()
	envt.SyncStatus.Error = &errMsg
	_, err := d.environmentRepo.UpdateById(ctx, envt.Id, envt)
	return err
}

func (d *domain) OnDeleteEnvironmentMessage(ctx ConsoleContext, envt entities.Environment) error {
	exEnvt, err := d.findEnvironment(ctx, envt.Namespace, envt.Name)
	if err != nil {
		return err
	}

	if err := d.MatchRecordVersion(envt.Annotations, exEnvt.RecordVersion); err != nil {
		return err
	}

	return d.environmentRepo.DeleteById(ctx, exEnvt.Id)
}

func (d *domain) OnUpdateEnvironmentMessage(ctx ConsoleContext, envt entities.Environment) error {
	exEnvt, err := d.findEnvironment(ctx, envt.Namespace, envt.Name)
	if err != nil {
		return err
	}

	annotatedVersion, err := d.parseRecordVersionFromAnnotations(exEnvt.Annotations)
	if err != nil {
		return d.resyncK8sResource(ctx, exEnvt.SyncStatus.Action, &exEnvt.Environment, exEnvt.RecordVersion)
	}

	if annotatedVersion != exEnvt.RecordVersion {
		if err := d.resyncK8sResource(ctx, exEnvt.SyncStatus.Action, &exEnvt.Environment, exEnvt.RecordVersion); err != nil {
			return err
		}
		return nil
	}

	exEnvt.CreationTimestamp = envt.CreationTimestamp
	exEnvt.Labels = envt.Labels
	exEnvt.Annotations = envt.Annotations
	exEnvt.Generation = envt.Generation

	exEnvt.Status = envt.Status

	exEnvt.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	exEnvt.SyncStatus.RecordVersion = annotatedVersion
	exEnvt.SyncStatus.Error = nil
	exEnvt.SyncStatus.LastSyncedAt = time.Now()

	_, err = d.environmentRepo.UpdateById(ctx, exEnvt.Id, exEnvt)
	return err
}
