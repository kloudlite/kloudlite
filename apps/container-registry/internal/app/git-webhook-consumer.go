package app

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
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
	uid := fmt.Sprint(build.Id, hook.CommitHash)

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
						ctx := context.TODO()
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
								fmt.Println(gitHook.GitProvider)
								logger.Infof(err.Error())
								return nil
							}
							logger.Errorf(err, "could not extract gitHook")
							return err
						}

						log.Println("here...................1", hook.RepoUrl, hook.GitBranch, hook.GitProvider)

						logger = logger.WithKV("repo", hook.RepoUrl, "provider", hook.GitProvider, "branch", hook.GitBranch)

						log.Println(hook.RepoUrl, hook.GitBranch, hook.GitProvider)

						builds, err := d.ListBuildsByGit(ctx, hook.RepoUrl, hook.GitBranch, hook.GitProvider)

						log.Println(builds, err, "bbbbbbbbbbb")

						if err != nil {
							return err
						}

						var pullToken string

						switch hook.GitProvider {

						case constants.ProviderGithub:
							pullToken, err = d.GithubInstallationToken(ctx, hook.RepoUrl)
							if err != nil {
								fmt.Println(err)
								return err
							}

						case constants.ProviderGitlab:
							pullToken = ""

						default:
							fmt.Println("provider not supported", hook.GitProvider)
							return fmt.Errorf("provider %s not supported", hook.GitProvider)
						}

						log.Println("here...................2", "builds", len(builds), hook.RepoUrl, hook.GitBranch, hook.GitProvider)

						for _, build := range builds {

							if hook.GitProvider == constants.ProviderGitlab {
								pullToken, err = d.GitlabPullToken(ctx, build.CredUser.UserId)
								if err != nil {
									errorMessage := fmt.Sprintf("could not get pull token for build, Error: %s", err.Error())
									if build.ErrorMessages["access-error"] != errorMessage {
										if build.ErrorMessages == nil {
											build.ErrorMessages = make(map[string]string)
										}
										build.ErrorMessages["access-error"] = errorMessage
										_, err := d.UpdateBuildInternal(ctx, build)
										if err != nil {
											return err
										}
									}

									continue
								} else {
									if build.ErrorMessages["access-error"] != "" {
										delete(build.ErrorMessages, "access-error")
										_, err := d.UpdateBuildInternal(ctx, build)
										if err != nil {
											return err
										}
									}
								}
							}

							pullUrl, err := domain.BuildUrl(hook.RepoUrl, hook.CommitHash, pullToken)
							if err != nil {
								logger.Errorf(err, "could not build pull url")
								continue
							}

							if pullToken == "" {
								logger.Warnf("pull token is empty")
								continue
							}

							fmt.Println("pullUrl", len(builds), pullUrl)

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

							build.Status = entities.BuildStatusQueued
							d.UpdateBuildInternal(ctx, build)

							logger.Infof("produced message to topic=%s, offset=%d", po.Topic, po.Offset)
						}

						return nil
					},
				)
			},
		),
	)
}
