package domain

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"

	fc "github.com/kloudlite/api/apps/container-registry/internal/domain/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"

	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/container-registry-authorizer/admin"
	common_types "github.com/kloudlite/operator/apis/common-types"
	distributionv1 "github.com/kloudlite/operator/apis/distribution/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	t2 "github.com/kloudlite/operator/operators/resource-watcher/types"
	fn "github.com/kloudlite/operator/pkg/functions"
)

func (d *Impl) ListBuildRuns(ctx RegistryContext, buildId repos.ID, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BuildRun], error) {
	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
		fc.BuildRunBuildId: buildId,
	}
	paginated, err := d.buildRunRepo.FindPaginated(ctx, d.buildRunRepo.MergeMatchFilters(filter, search), pagination)
	return paginated, err
}

func (d *Impl) GetBuildRun(ctx RegistryContext, buildId repos.ID, runName string) (*entities.BuildRun, error) {
	brun, err := d.buildRunRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: runName,
		fc.BuildRunBuildId:  buildId,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if brun == nil {
		return nil, errors.Newf("build run with id=%q not found", buildId)
	}
	return brun, nil
}

func (d *Impl) GetLatestBuildRun(ctx RegistryContext, buildId repos.ID) (*entities.BuildRun, error) {
	bruns, err := d.buildRunRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			fields.AccountName: ctx.AccountName,
			fc.BuildRunBuildId: buildId,
		},
		Sort: map[string]any{
			"creationTime": -1,
		},
		Limit: fn.New(int64(1)),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if len(bruns) == 0 {
		return nil, nil
	}

	if len(bruns) != 1 {
		return nil, errors.Newf("failed to find latest build run")
	}

	return bruns[0], nil
}

func (d *Impl) parseRecordVersionFromAnnotations(annotations map[string]string) (int, error) {
	annotatedVersion, ok := annotations[constants.RecordVersionKey]
	if !ok {
		return 0, errors.Newf("no annotation with record version key (%s), found on the resource", constants.RecordVersionKey)
	}

	annVersion, err := strconv.ParseInt(annotatedVersion, 10, 32)
	if err != nil {
		return 0, errors.NewE(err)
	}

	return int(annVersion), nil
}

func (d *Impl) MatchRecordVersion(annotations map[string]string, rv int) (int, error) {
	annVersion, err := d.parseRecordVersionFromAnnotations(annotations)
	if err != nil {
		return -1, errors.NewE(err)
	}

	if annVersion != rv {
		return -1, errors.Newf("record version mismatch, expected %d, got %d", rv, annVersion)
	}

	return annVersion, nil
}

func (d *Impl) OnBuildRunUpdateMessage(ctx RegistryContext, buildRun entities.BuildRun, status t2.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xBr, err := d.buildRunRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:       ctx.AccountName,
		fields.MetadataName:      buildRun.Name,
		fields.MetadataNamespace: buildRun.Namespace,
		fields.ClusterName:       buildRun.ClusterName,
	})
	if err != nil {
		return errors.NewE(err)
	}
	if xBr == nil {
		return errors.Newf("build run with name %q not found", buildRun.Name)
	}

	if _, err := d.MatchRecordVersion(buildRun.Annotations, xBr.RecordVersion); err != nil {
		d.logger.Warnf("error while matching record version: %s, ignoring silently", err.Error())
		return nil
	}

	if _, err = d.buildRunRepo.PatchById(
		ctx,
		xBr.Id,
		common.PatchForSyncFromAgent(&buildRun, 0, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		})); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishBuildRunEvent(&buildRun, PublishAdd)

	return nil
}

func (d *Impl) OnBuildRunDeleteMessage(ctx RegistryContext, buildRun entities.BuildRun) error {
	if err := d.buildRunRepo.DeleteOne(ctx, repos.Filter{
		fields.MetadataName:      buildRun.Name,
		fields.MetadataNamespace: buildRun.Namespace,
		fields.AccountName:       ctx.AccountName,
		fields.ClusterName:       buildRun.ClusterName,
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishBuildRunEvent(&buildRun, PublishDelete)
	return nil
}

func (d *Impl) OnBuildRunApplyErrorMessage(ctx RegistryContext, clusterName string, name string, errorMsg string) error {
	upBuildRun, err := d.buildRunRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: name,
			fields.ClusterName:  clusterName,
		},
		common.PatchForErrorFromAgent(
			errorMsg,
			common.PatchOpts{
				MessageTimestamp: time.Time{},
			},
		),
	)

	d.resourceEventPublisher.PublishBuildRunEvent(upBuildRun, PublishUpdate)
	return errors.NewE(err)
}

func getUniqueKey(build *entities.Build, hook *GitWebhookPayload, seed string) string {
	uid := fmt.Sprint(build.Id, hook.CommitHash, seed)
	return fmt.Sprintf("%x", md5.Sum([]byte(uid)))
}

func (d *Impl) CreateBuildRun(ctx RegistryContext, build *entities.Build, hook *GitWebhookPayload, pullToken string, seed string) error {
	uniqueKey := getUniqueKey(build, hook, seed)
	i, err := admin.GetExpirationTime(fmt.Sprintf("%d%s", 1, "d"))
	if err != nil {
		return errors.NewE(err)
	}
	token, err := admin.GenerateToken(KL_ADMIN, build.Spec.AccountName, string("read_write"), i, d.envs.RegistrySecretKey+build.Spec.AccountName)
	if err != nil {
		return errors.NewE(err)
	}

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
		Registry: distributionv1.Registry{
			Repo: distributionv1.Repo{
				Name: build.Spec.Registry.Repo.Name,
				Tags: build.Spec.Registry.Repo.Tags,
			},
		},
		CacheKeyName: build.Spec.CacheKeyName,
		GitRepo: distributionv1.GitRepo{
			Url:    hook.RepoUrl,
			Branch: hook.CommitHash,
		},
		Resource: build.Spec.Resource,
		CredentialsRef: common_types.SecretRef{
			Name:      fmt.Sprint("build-run-", uniqueKey),
			Namespace: d.envs.JobBuildNamespace,
		},
	})
	if err != nil {
		return err
	}

	brRaw := distributionv1.BuildRun{}
	err = yaml.Unmarshal(b, &brRaw)
	if err != nil {
		d.logger.Errorf(err, "could not unmarshal build run")
		return errors.NewE(err)
	}
	br := entities.BuildRun{
		BuildRun:    brRaw,
		BuildId:     build.Id,
		SyncStatus:  t.GenSyncStatus(t.SyncActionApply, build.RecordVersion),
		AccountName: build.Spec.AccountName,
		ClusterName: build.BuildClusterName,
	}
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

	if cbr.Annotations == nil {
		cbr.Annotations = make(map[string]string)
	}

	cbr.Annotations[constants.ObservabilityTrackingKey] = string(cbr.Id)

	if err = d.dispatcher.ApplyToTargetCluster(ctx, build.BuildClusterName, &cbr.BuildRun, 0); err != nil {
		d.logger.Errorf(err, "could not apply build run")
		return errors.NewE(err)
	}
	return nil
}
