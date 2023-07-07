package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/operator/pkg/constants"
	"io"
	"os"
	"strconv"

	t "github.com/kloudlite/operator/agent/types"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/apps/console/internal/env"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	fn "kloudlite.io/pkg/functions"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
	types "kloudlite.io/pkg/types"
)

type domain struct {
	k8sExtendedClient k8s.ExtendedK8sClient
	k8sYamlClient     *kubectl.YAMLClient

	producer redpanda.Producer

	iamClient iam.IAMClient

	projectRepo   repos.DbRepo[*entities.Project]
	workspaceRepo repos.DbRepo[*entities.Workspace]
	appRepo       repos.DbRepo[*entities.App]
	configRepo    repos.DbRepo[*entities.Config]
	secretRepo    repos.DbRepo[*entities.Secret]
	routerRepo    repos.DbRepo[*entities.Router]
	msvcRepo      repos.DbRepo[*entities.ManagedService]
	mresRepo      repos.DbRepo[*entities.ManagedResource]
	ipsRepo       repos.DbRepo[*entities.ImagePullSecret]

	envVars *env.Env

	msvcTemplates    []*entities.MsvcTemplate
	msvcTemplatesMap map[string]map[string]*entities.MsvcTemplateEntry
}

func errAlreadyMarkedForDeletion(label, namespace, name string) error {
	return fmt.Errorf(
		"%s (namespace=%s, name=%s) already marked for deletion",
		label,
		namespace,
		name,
	)
}

func (d *domain) applyK8sResource(ctx ConsoleContext, obj client.Object, recordVersion int) error {
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string, 1)
	}
	ann[constants.RecordVersionKey] = fmt.Sprintf("%d", recordVersion)
	obj.SetAnnotations(ann)

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

	_, err = d.producer.Produce(
		ctx,
		common.GetKafkaTopicName(ctx.AccountName, ctx.ClusterName),
		obj.GetNamespace(),
		b,
	)
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
	_, err = d.producer.Produce(
		ctx,
		common.GetKafkaTopicName(ctx.AccountName, ctx.ClusterName),
		obj.GetNamespace(),
		b,
	)
	return err
}

func (d *domain) resyncK8sResource(ctx ConsoleContext, action types.SyncAction, obj client.Object, rv int) error {
	switch action {
	case types.SyncActionApply:
		{
			return d.applyK8sResource(ctx, obj, rv)
		}
	case types.SyncActionDelete:
		{
			return d.deleteK8sResource(ctx, obj)
		}
	default:
		{
			return fmt.Errorf("unknown sync action %q", action)
		}
	}
}

func (d *domain) parseRecordVersionFromAnnotations(annotations map[string]string) (int, error) {
	annotatedVersion, ok := annotations[constants.RecordVersionKey]
	if !ok {
		return 0, fmt.Errorf("no annotation with record version key (%s), found on the resource", constants.RecordVersionKey)
	}

	annVersion, err := strconv.ParseInt(annotatedVersion, 10, 32)
	if err != nil {
		return 0, err
	}

	return int(annVersion), nil
}

func (d *domain) MatchRecordVersion(annotations map[string]string, rv int) error {
	annVersion, err := d.parseRecordVersionFromAnnotations(annotations)
	if err != nil {
		return err
	}

	if annVersion != rv {
		return fmt.Errorf("record version mismatch, expected %d, got %d", rv, annVersion)
	}

	return nil
}

func (d *domain) canMutateResourcesInProject(ctx ConsoleContext, targetNamespace string) error {
	prj, err := d.findProjectByTargetNs(ctx, targetNamespace)
	if err != nil {
		return err
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, prj.Name),
		},
		Action: string(iamT.MutateResourcesInProject),
	})
	if err != nil {
		return err
	}
	if !co.Status {
		return fmt.Errorf("unauthorized to mutate resources in project %q", prj.Name)
	}
	return nil
}

func (d *domain) canMutateResourcesInWorkspace(ctx ConsoleContext, targetNamespace string) error {
	ws, err := d.findWorkspaceByTargetNs(ctx, targetNamespace)
	if err != nil {
		return err
	}

	wsp, err := d.findWorkspace(ctx, ws.Namespace, ws.Name)
	if err != nil {
		return err
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, wsp.Spec.ProjectName),
		},
		Action: string(iamT.MutateResourcesInProject),
	})
	if err != nil {
		return err
	}
	if !co.Status {
		return fmt.Errorf("unauthorized to mutate resources in workspace %q", wsp.Name)
	}
	return nil
}

func (d *domain) canReadResourcesInWorkspace(ctx ConsoleContext, targetNamespace string) error {
	ws, err := d.findWorkspaceByTargetNs(ctx, targetNamespace)
	if err != nil {
		return err
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, ws.Spec.ProjectName),
		},
		Action: string(iamT.GetProject),
	})
	if err != nil {
		return err
	}
	if !co.Status {
		return fmt.Errorf("unauthorized to read resources in project %q", ws.Spec.ProjectName)
	}
	return nil
}

func (d *domain) canReadResourcesInProject(ctx ConsoleContext, targetNamespace string) error {
	prj, err := d.findProjectByTargetNs(ctx, targetNamespace)
	if err != nil {
		return err
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, prj.Name),
		},
		Action: string(iamT.GetProject),
	})
	if err != nil {
		return err
	}
	if !co.Status {
		return fmt.Errorf("unauthorized to read resources in project %q", prj.Name)
	}
	return nil
}

func (d *domain) canMutateSecretsInAccount(ctx context.Context, userId string, accountName string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: userId,
		ResourceRefs: []string{
			iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		},
		Action: string(iamT.CreateSecretsInAccount),
	})
	if err != nil {
		return err
	}
	if !co.Status {
		return fmt.Errorf("unauthorized to mutate secrets in account %q", accountName)
	}
	return nil
}

func (d *domain) canReadSecretsFromAccount(ctx context.Context, userId string, accountName string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: userId,
		ResourceRefs: []string{
			iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		},
		Action: string(iamT.ReadSecretsFromAccount),
	})
	if err != nil {
		return err
	}
	if !co.Status {
		return fmt.Errorf("unauthorized to read secrets from account  %q", accountName)
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
		environmentRepo repos.DbRepo[*entities.Workspace],
		appRepo repos.DbRepo[*entities.App],
		configRepo repos.DbRepo[*entities.Config],
		secretRepo repos.DbRepo[*entities.Secret],
		routerRepo repos.DbRepo[*entities.Router],
		msvcRepo repos.DbRepo[*entities.ManagedService],
		mresRepo repos.DbRepo[*entities.ManagedResource],
		ipsRepo repos.DbRepo[*entities.ImagePullSecret],

		ev *env.Env,
	) (Domain, error) {
		open, err := os.Open(ev.MsvcTemplateFilePath)
		if err != nil {
			return nil, err
		}

		b, err := io.ReadAll(open)
		if err != nil {
			return nil, err
		}

		var templates []*entities.MsvcTemplate

		if err := yaml.Unmarshal(b, &templates); err != nil {
			return nil, err
		}

		msvcTemplatesMap := map[string]map[string]*entities.MsvcTemplateEntry{}

		for _, t := range templates {
			if _, ok := msvcTemplatesMap[t.Category]; !ok {
				msvcTemplatesMap[t.Category] = make(map[string]*entities.MsvcTemplateEntry, len(t.Items))
			}
			for i := range t.Items {
				msvcTemplatesMap[t.Category][t.Items[i].Name] = &t.Items[i]
			}
		}

		return &domain{
			k8sExtendedClient: k8sExtendedClient,
			k8sYamlClient:     k8sYamlClient,

			producer: producer,

			iamClient: iamClient,

			projectRepo:   projectRepo,
			workspaceRepo: environmentRepo,
			appRepo:       appRepo,
			configRepo:    configRepo,
			routerRepo:    routerRepo,
			secretRepo:    secretRepo,
			msvcRepo:      msvcRepo,
			mresRepo:      mresRepo,
			ipsRepo:       ipsRepo,

			envVars: ev,

			msvcTemplates:    templates,
			msvcTemplatesMap: msvcTemplatesMap,
		}, nil
	}),
)
