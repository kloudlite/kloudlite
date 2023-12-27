package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/api/pkg/messaging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"

	"github.com/kloudlite/api/apps/container-registry/internal/domain"
	"github.com/kloudlite/api/apps/container-registry/internal/env"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/types"
)

const (
	GithubEventHeader string = "X-Github-Event"
	GitlabEventHeader string = "X-Gitlab-Event"
)

type (
	GitWebhookConsumer messaging.Consumer
	BuildRunProducer   messaging.Producer
)


func processGitWebhooks(ctx context.Context, d domain.Domain, consumer GitWebhookConsumer, producer BuildRunProducer, logr logging.Logger, envs *env.Env) error {
	err := consumer.Consume(func(msg *msgTypes.ConsumeMsg) error {
		logger := logr.WithName("ci-webhook")
		logger.Infof("started processing")
		defer func() {
			logger.Infof("finished processing")
		}()
		var gitHook types.GitHttpHook
		if err := json.Unmarshal(msg.Payload, &gitHook); err != nil {
			logger.Errorf(err, "could not unmarshal into *GitWebhookPayload")
			return errors.NewE(err)
		}
		hook, err := func() (*domain.GitWebhookPayload, error) {
			if gitHook.GitProvider == constants.ProviderGithub {
				return d.ParseGithubHook(gitHook.Headers[GithubEventHeader][0], gitHook.Body)
			}
			if gitHook.GitProvider == constants.ProviderGitlab {
				return d.ParseGitlabHook(gitHook.Headers[GitlabEventHeader][0], gitHook.Body)
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
			return errors.NewE(err)
		}
		logger = logger.WithKV("repo", hook.RepoUrl, "provider", hook.GitProvider, "branch", hook.GitBranch)
		logger.Infof("repo: %s, branch: %s, gitprovider: %s", hook.RepoUrl, hook.GitBranch, hook.GitProvider)
		builds, err := d.ListBuildsByGit(ctx, hook.RepoUrl, hook.GitBranch, hook.GitProvider)
		if err != nil {
			return errors.NewE(err)
		}

		var pullToken string

		switch hook.GitProvider {

		case constants.ProviderGithub:
			pullToken, err = d.GithubInstallationToken(ctx, hook.RepoUrl)
			if err != nil {
				fmt.Println(err)
				return errors.NewE(err)
			}

		case constants.ProviderGitlab:
			pullToken = ""

		default:
			fmt.Println("provider not supported", hook.GitProvider)
			return errors.Newf("provider %s not supported", hook.GitProvider)
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
							return errors.NewE(err)
						}
					}

					continue
				} else {
					if build.ErrorMessages["access-error"] != "" {
						delete(build.ErrorMessages, "access-error")
						_, err := d.UpdateBuildInternal(ctx, build)
						if err != nil {
							return errors.NewE(err)
						}
					}
				}
			}

			if pullToken == "" {
				logger.Warnf("pull token is empty")
				continue
			}


			if err != nil {
				logger.Errorf(err, "could not generate pull-token")
				continue
			}

			dctx := domain.RegistryContext{
				Context:     context.TODO(),
				UserId:      "sys-user:error-on-apply-worker",
				UserEmail:   "",
				UserName:    "",
				AccountName: build.AccountName,
			}

			err := d.CreateBuildRun(dctx, build, hook, pullToken)
			if err != nil {
				logger.Errorf(err, "could not create build run")
			}
		}
		return nil
	}, msgTypes.ConsumeOpts{})
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}
