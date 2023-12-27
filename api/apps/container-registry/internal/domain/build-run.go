package domain

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/container-registry-authorizer/admin"
	common_types "github.com/kloudlite/operator/apis/common-types"
	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	distributionv1 "github.com/kloudlite/operator/apis/distribution/v1"
	"strings"
	"time"
)

func (d *Impl) ListBuildRuns(ctx RegistryContext, repoName string, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BuildRun], error) {
	filter := repos.Filter{
		"accountName":             ctx.AccountName,
		"spec.registry.repo.name": repoName,
	}
	return d.buildRunRepo.FindPaginated(ctx, d.buildRunRepo.MergeMatchFilters(filter, matchFilters), pagination)
}

func (d *Impl) GetBuildRun(ctx RegistryContext, repoName string, buildRunName string) (*entities.BuildRun, error) {
	brun, err := d.buildRunRepo.FindOne(ctx, repos.Filter{
		"accountName":             ctx.AccountName,
		"metadata.name":           buildRunName,
		"spec.registry.repo.name": repoName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if brun == nil {
		return nil, errors.Newf("build run with name %q not found", buildRunName)
	}
	return brun, nil
}

func (d *Impl) OnBuildRunUpdateMessage(ctx RegistryContext, buildRun entities.BuildRun) error {
	if _, err := d.buildRunRepo.Upsert(ctx, repos.Filter{
		"metadata.name":      buildRun.Name,
		"metadata.namespace": buildRun.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        buildRun.ClusterName,
	}, &buildRun); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *Impl) OnBuildRunDeleteMessage(ctx RegistryContext, buildRun entities.BuildRun) error {
	if err := d.buildRunRepo.DeleteOne(ctx, repos.Filter{
		"metadata.name":      buildRun.Name,
		"metadata.namespace": buildRun.Namespace,
		"accountName":        ctx.AccountName,
		"clusterName":        buildRun.ClusterName,
	}); err != nil {
		return errors.NewE(err)
	}
	//d.natCli.Conn.Publish(fmt.Scan(buildRun.BuildRun, ))
	return nil
}

func (d *Impl) OnBuildRunApplyErrorMessage(ctx RegistryContext,clusterName string, name string, errorMsg string) error{
	buildRun, err := d.buildRunRepo.FindOne(ctx, repos.Filter{
		"accountName": ctx.AccountName,
		"metadata.name": name,
		"clusterName":  clusterName,
	})
	if err != nil {
		return errors.NewE(err)
	}

	buildRun.SyncStatus.State = t.SyncStateErroredAtAgent
	buildRun.SyncStatus.LastSyncedAt = time.Now()
	buildRun.SyncStatus.Error = &errorMsg

	_, err = d.buildRunRepo.UpdateById(ctx, buildRun.Id, buildRun)
	d.resourceEventPublisher.PublishBuildRunEvent(buildRun, PublishUpdate)
	return errors.NewE(err)
}

func getUniqueKey(build *entities.Build, hook *GitWebhookPayload) string {
	uid := fmt.Sprint(build.Id, hook.CommitHash)
	return fmt.Sprintf("%x", md5.Sum([]byte(uid)))
}

func (d *Impl) CreateBuildRun(ctx RegistryContext, build *entities.Build, hook *GitWebhookPayload, pullToken string) error{
	uniqueKey := getUniqueKey(build, hook)
	i, err := admin.GetExpirationTime(fmt.Sprintf("%d%s", 1, "d"))
	if err != nil {
		return errors.NewE(err)
	}
	token, err := admin.GenerateToken(KL_ADMIN, build.Spec.AccountName, string("read_write"), i, d.envs.RegistrySecretKey+build.Spec.AccountName)

	sec := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprint("build-run-", uniqueKey),
			Namespace: d.envs.JobBuildNamespace,
			Annotations: map[string]string{
				"kloudlite.io/build-run.name": uniqueKey,
			},
		},
		StringData: map[string]string{
			"registry-admin": KL_ADMIN,
			"registry-host":  d.envs.RegistryHost,
			"registry-token": token,
			"github-token":   pullToken,
		},
	}
	var secretCreationError error
	if err = d.dispatcher.ApplyToTargetCluster(ctx, build.BuildClusterName, &sec, 0); err != nil {
		d.logger.Errorf(err, "could not apply secret")
		secretCreationError = err
	}

	b, err := d.GetBuildTemplate(BuildJobTemplateData{
		AccountName: build.Spec.AccountName,
		Name:        uniqueKey,
		Namespace:   d.envs.JobBuildNamespace,
		Labels: map[string]string{
			"kloudlite.io/build-id": string(build.Id),
			"kloudlite.io/account":  build.Spec.AccountName,
			"github.com/commit":     hook.CommitHash,
		},
		Annotations: map[string]string{
			"kloudlite.io/build-id": string(build.Id),
			"kloudlite.io/account":  build.Spec.AccountName,
			"github.com/commit":     hook.CommitHash,
			"github.com/repository": hook.RepoUrl,
			"github.com/branch":     hook.GitBranch,
			"kloudlite.io/repo":     build.Spec.Registry.Repo.Name,
			"kloudlite.io/tag":      strings.Join(build.Spec.Registry.Repo.Tags, ","),
		},
		BuildOptions: build.Spec.BuildOptions,
		Registry: dbv1.Registry{
			Repo: dbv1.Repo{
				Name: build.Spec.Registry.Repo.Name,
				Tags: build.Spec.Registry.Repo.Tags,
			},
		},
		CacheKeyName: build.Spec.CacheKeyName,
		GitRepo: dbv1.GitRepo{
			Url:    hook.RepoUrl,
			Branch: hook.CommitHash,
		},
		Resource: build.Spec.Resource,
		CredentialsRef: common_types.SecretRef{
			Name:      fmt.Sprint("build-run-", uniqueKey),
			Namespace: d.envs.JobBuildNamespace,
		},
	})
	brRaw := distributionv1.BuildRun{}
	err = json.Unmarshal(b, &brRaw)
	if err != nil {
		d.logger.Errorf(err, "could not unmarshal build run")
		return errors.NewE(err)
	}
	br := entities.BuildRun{
		BuildRun: brRaw,
		BuildName: build.Name,
		SyncStatus: t.SyncStatus{},
	}
	br.AccountName = build.Spec.AccountName
	br.ClusterName = build.BuildClusterName
	if secretCreationError != nil {
		msg := secretCreationError.Error()
		br.SyncStatus.Error = &msg
	}
	cbr, err := d.buildRunRepo.Create(ctx, &br)
	if err != nil {
		d.logger.Errorf(err, "could not create build run")
		return errors.NewE(err)
	}
	if secretCreationError != nil {
		return errors.NewE(secretCreationError)
	}

	if err != nil {
		d.logger.Errorf(err, "could not get build template")
		return errors.NewE(err)
	}

	if err = d.dispatcher.ApplyToTargetCluster(ctx, build.BuildClusterName, cbr, 0); err != nil {
		d.logger.Errorf(err, "could not apply build run")
		return errors.NewE(err)
	}
	return nil
}

