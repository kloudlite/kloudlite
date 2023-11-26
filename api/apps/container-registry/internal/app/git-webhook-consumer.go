package app

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kloudlite/container-registry-authorizer/admin"
	t "github.com/kloudlite/operator/agent/types"
	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"

	common_types "github.com/kloudlite/operator/apis/common-types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/container-registry/internal/domain"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/common"
	"kloudlite.io/constants"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/kafka"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/types"
	"sigs.k8s.io/yaml"
)

const (
	GithubEventHeader string = "X-Github-Event"
	GitlabEventHeader string = "X-Gitlab-Event"
)

func getUniqueKey(build *entities.Build, hook *domain.GitWebhookPayload) string {
	uid := fmt.Sprint(build.Id, hook.CommitHash)

	return fmt.Sprintf("%x", md5.Sum([]byte(uid)))
}

func invokeProcessGitWebhooks(d domain.Domain, consumer kafka.Consumer, producer kafka.Producer, logr logging.Logger, envs *env.Env) {
	consumer.StartConsuming(func(ctx kafka.ConsumerContext, topic string, msg []byte, metadata kafka.RecordMetadata) error {
		logger := ctx.Logger.WithName("ci-webhook")
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
				fmt.Println(gitHook.GitProvider)
				logger.Infof(err.Error())
				return nil
			}
			logger.Errorf(err, "could not extract gitHook")
			return err
		}

		logger = logger.WithKV("repo", hook.RepoUrl, "provider", hook.GitProvider, "branch", hook.GitBranch)

		logger.Infof("repo: %s, branch: %s, gitprovider: %s", hook.RepoUrl, hook.GitBranch, hook.GitProvider)

		builds, err := d.ListBuildsByGit(ctx, hook.RepoUrl, hook.GitBranch, hook.GitProvider)
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

		for _, build := range builds {
			if hook.GitProvider == constants.ProviderGitlab {
				pullToken, err = d.GitlabPullToken(ctx, build.CredUser.UserId)
				if err != nil {
					errorMessage := fmt.Sprintf("could not get pull token for build, Error: %s", err.Error())
					if build.ErrorMessages == nil {
						build.ErrorMessages = make(map[string]string)
					}
					if build.ErrorMessages["access-error"] != errorMessage {
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

			// pullUrl, err := domain.BuildUrl(hook.RepoUrl, pullToken)
			// if err != nil {
			// 	logger.Errorf(err, "could not build pull url")
			// 	continue
			// }

			if pullToken == "" {
				logger.Warnf("pull token is empty")
				continue
			}

			// fmt.Println("pullUrl", len(builds), pullUrl)

			i, err := admin.GetExpirationTime(fmt.Sprintf("%d%s", 1, "d"))
			if err != nil {
				return err
			}

			token, err := admin.GenerateToken(domain.KL_ADMIN, build.Spec.AccountName, string("read_write"), i, envs.RegistrySecretKey+build.Spec.AccountName)
			if err != nil {
				logger.Errorf(err, "could not generate pull-token")
				continue
			}

			uniqueKey := getUniqueKey(build, hook)

			b, err := d.GetBuildTemplate(domain.BuildJobTemplateData{
				AccountName: build.Spec.AccountName,
				Name:        uniqueKey,
				Namespace:   envs.JobBuildNamespace,
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
					// Password: token,
					// Username: domain.KL_ADMIN,
					// Host:     envs.RegistryHost,
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
					Name:      uniqueKey,
					Namespace: envs.JobBuildNamespace,
				},
			})
			if err != nil {
				logger.Errorf(err, "could not get build template")
				return err
			}

			sec := corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      uniqueKey,
					Namespace: envs.JobBuildNamespace,
				},
				StringData: map[string]string{
					"registry-admin": domain.KL_ADMIN,
					"registry-host":  envs.RegistryHost,
					"registry-token": token,
					"github-token":   pullToken,
				},
			}

			var m map[string]any
			if err := yaml.Unmarshal(b, &m); err != nil {
				return err
			}

			b1, err := json.Marshal(t.AgentMessage{
				AccountName: envs.BuildClusterAccountName,
				ClusterName: envs.BuildClusterName,
				Action:      t.ActionApply,
				Object:      m,
			})
			if err != nil {
				return err
			}

			b, err = yaml.Marshal(sec)
			if err != nil {
				return err
			}

			var m2 map[string]any
			if err := yaml.Unmarshal(b, &m2); err != nil {
				return err
			}

			b2, err := json.Marshal(t.AgentMessage{
				AccountName: envs.BuildClusterAccountName,
				ClusterName: envs.BuildClusterName,
				Action:      t.ActionApply,
				Object:      m2,
			})
			if err != nil {
				return err
			}
			po1, err := producer.Produce(ctx, constants.MSGTO_TargetWaitQueueTopicName, b1, kafka.MessageArgs{
				Key: []byte(build.Spec.AccountName),
				Headers: map[string][]byte{
					"topic": []byte(common.GetKafkaTopicName(envs.BuildClusterAccountName, envs.BuildClusterName)),
				},
			})
			if err != nil {
				return err
			}
			logger.Infof("produced message to topic=%s, offset=%d", po1.Topic, po1.Offset)

			po2, err := producer.Produce(ctx, constants.MSGTO_TargetWaitQueueTopicName, b2, kafka.MessageArgs{
				Key: []byte(build.Spec.AccountName),
				Headers: map[string][]byte{
					"topic": []byte(common.GetKafkaTopicName(envs.BuildClusterAccountName, envs.BuildClusterName)),
				},
			})
			if err != nil {
				return err
			}

			build.Status = entities.BuildStatusQueued
			d.UpdateBuildInternal(ctx, build)

			logger.Infof("produced message to topic=%s, offset=%d", po2.Topic, po2.Offset)
		}
		return nil
	})
}
