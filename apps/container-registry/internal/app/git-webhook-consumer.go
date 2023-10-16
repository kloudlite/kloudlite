package app

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kloudlite/container-registry-authorizer/admin"
	"go.uber.org/fx"
	"kloudlite.io/apps/container-registry/internal/domain"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/constants"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/types"
)

const (
	GithubEventHeader string = "X-Github-Event"
	GitlabEventHeader string = "X-Gitlab-Event"
)

func getUniqueKey(build *entities.Build, hook *domain.GitWebhookPayload) string {
	uid := fmt.Sprint(build.Id, build.UpdateTime.Format(time.RFC3339), hook.CommitHash)

	return fmt.Sprintf("%x", md5.Sum([]byte(uid)))
}

func fxInvokeProcessGitWebhooks() fx.Option {
	return fx.Options(

		fx.Invoke(
			func(d domain.Domain, consumer redpanda.Consumer, producer redpanda.Producer, logr logging.Logger, envs *env.Env) {
				consumer.StartConsuming(
					func(msg []byte, _ time.Time, offset int64) error {
						logger := logr.WithName("ci-webhook").WithKV("offset", offset)
						logger.Infof("started processing")
						defer func() {
							logger.Infof("finished processing")
						}()

						var gitHook types.GitHttpHook
						if err := json.Unmarshal(msg, &gitHook); err != nil {
							logger.Errorf(err, "could not unmarshal into *GitWebhookPayload")
							return err
						}

						hook, err := func() (*domain.GitWebhookPayload, error) {
							if gitHook.GitProvider == constants.ProviderGithub {
								return d.ParseGithubHook(gitHook.Headers[GithubEventHeader], gitHook.Body)
							}
							if gitHook.GitProvider == constants.ProviderGitlab {
								return d.ParseGitlabHook(gitHook.Headers[GitlabEventHeader], gitHook.Body)
							}
							return nil, errors.New("unknown git provider")
						}()

						if err != nil {
							if _, ok := err.(*domain.ErrEventNotSupported); ok {
								logger.Infof(err.Error())
								return nil
							}
							logger.Errorf(err, "could not extract gitHook")
							return err
						}

						logger = logger.WithKV("repo", hook.RepoUrl, "provider", hook.GitProvider, "branch", hook.GitBranch)

						ctx := context.TODO()

						builds, err := d.ListBuildsByGit(ctx, hook.RepoUrl, hook.GitBranch, hook.GitProvider)
						if err != nil {
							return err
						}

						var pullToken string

						switch hook.GitProvider {

						case constants.ProviderGithub:
							pullToken, err = d.GithubInstallationToken(ctx, hook.RepoUrl)
							if err != nil {
								return err
							}

						default:
							return fmt.Errorf("provider %s not supported", hook.GitProvider)
						}

						pullUrl, err := domain.BuildUrl(hook.RepoUrl, hook.CommitHash, pullToken)
						if err != nil {
							return err
						}

						for _, build := range builds {

							i, err := admin.GetExpirationTime(fmt.Sprintf("%d%s", 1, "d"))
							if err != nil {
								return err
							}

							token, err := admin.GenerateToken(domain.KL_ADMIN, build.AccountName, string("read_write"), i, envs.RegistrySecretKey+build.AccountName)

							uniqueKey := getUniqueKey(build, hook)

							b, err := d.GetBuildTemplate(domain.BuildJobTemplateObject{
								KlAdmin:          domain.KL_ADMIN,
								AccountName:      build.AccountName,
								Registry:         envs.RegistryHost,
								Name:             uniqueKey,
								Tag:              build.Tag,
								RegistryRepoName: fmt.Sprintf("%s/%s", build.AccountName, build.Repository),
								DockerPassword:   token,
								Namespace:        "kl-core",
								PullUrl:          pullUrl,
								Labels: map[string]string{
									"kloudlite.io/build-id": string(build.Id),
									"kloudlite.io/account":  build.AccountName,
									"github.com/commit":     hook.CommitHash,
								},
								Annotations: map[string]string{
									"kloudlite.io/build-id": string(build.Id),
									"kloudlite.io/account":  build.AccountName,
									"github.com/commit":     hook.CommitHash,
									"github.com/repository": hook.RepoUrl,
									"github.com/branch":     hook.GitBranch,
									"kloudlite.io/repo":     build.Repository,
									"kloudlite.io/tag":      build.Tag,
								},
							})

							if err != nil {
								logger.Errorf(err, "could not get build template")
								return err
							}

							po, err := producer.Produce(ctx, envs.RegistryTopic, build.AccountName, b)
							if err != nil {
								return err
							}

							logger.Infof("produced message to topic=%s, offset=%d", po.Topic, po.Offset)
						}

						return nil
					},
				)
			},
		),
	)
}
