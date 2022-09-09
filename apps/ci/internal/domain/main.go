package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kloudlite.io/pkg/config"

	"github.com/google/go-github/v43/github"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/harbor"
	"kloudlite.io/pkg/logging"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/tekton"
	"kloudlite.io/pkg/types"
	t "kloudlite.io/pkg/types"
)

type HarborHost string

type domainI struct {
	pipelineRepo    repos.DbRepo[*Pipeline]
	authClient      auth.AuthClient
	github          Github
	gitlab          Gitlab
	harborAccRepo   repos.DbRepo[*HarborAccount]
	harborHost      HarborHost
	logger          logging.Logger
	harborCli       *harbor.Client
	gitRepoHookRepo repos.DbRepo[*GitRepositoryHook]

	gitlabWebhookUrl string
	githubWebhookUrl string
}

type Env struct {
	GithubWebhookUrl string `env:"GITHUB_WEBHOOK_URL" required:"true"`
	GitlabWebhookUrl string `env:"GITLAB_WEBHOOK_URL" required:"true"`
}

const (
	GitlabLabel string = "gitlab"
	GithubLabel string = "github"
)

func (d *domainI) HarborImageSearch(ctx context.Context, accountId repos.ID, q string, pagination *t.Pagination) ([]harbor.Repository, error) {
	return d.harborCli.SearchRepositories(
		ctx, accountId, q, harbor.ListOptions{
			Page: func() int64 {
				if pagination == nil {
					return 1
				}
				return int64(pagination.Page)
			}(),
			PageSize: func() int64 {
				if pagination == nil {
					return 20
				}
				return int64(pagination.PerPage)
			}(),
		},
	)
}

func (d *domainI) HarborImageTags(ctx context.Context, imageName string, pagination *t.Pagination) ([]harbor.ImageTag, error) {
	n := strings.SplitN(imageName, "/", 2)
	projectName := n[0]
	resourceName := n[1]

	return d.harborCli.ListTags(
		ctx, projectName, resourceName, harbor.ListTagsOpts{
			ListOptions: harbor.ListOptions{
				Page: func() int64 {
					if pagination == nil {
						return 1
					}
					return int64(pagination.Page)
				}(),
				PageSize: func() int64 {
					if pagination == nil {
						return 10
					}
					return int64(pagination.PerPage)
				}(),
			},
		},
	)
}

func (d *domainI) GetAppPipelines(ctx context.Context, appId repos.ID) ([]*Pipeline, error) {
	find, err := d.pipelineRepo.Find(
		ctx, repos.Query{
			Filter: repos.Filter{
				"app_id": appId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return find, nil
}

type GitWebhookPayload struct {
	GitProvider string `json:"git_provider,omitempty"`
	RepoUrl     string `json:"repo_url,omitempty"`
	GitBranch   string `json:"git_branch,omitempty"`
	CommitHash  string `json:"commit_hash,omitempty"`
}

func getBranchFromRef(gitRef string) string {
	// exmaple: gitref => refs/heads/master
	sp := strings.Split(gitRef, "/")
	if len(sp) > 2 {
		return sp[2]
	}
	return ""
}

func (d *domainI) ParseGithubHook(eventType string, hookBody []byte) (*GitWebhookPayload, error) {
	hook, err := github.ParseWebHook(eventType, hookBody)
	if err != nil {
		d.logger.Infof("bad webhook body, dropping message ...")
		return nil, err
	}

	switch h := hook.(type) {
	case *github.PushEvent:
		{
			payload := GitWebhookPayload{
				GitProvider: common.ProviderGithub,
				RepoUrl:     *h.Repo.HTMLURL,
				GitBranch:   getBranchFromRef(h.GetRef()),
				CommitHash:  h.GetAfter()[:int(math.Min(10, float64(len(h.GetAfter()))))],
			}
			return &payload, nil
		}
	default:
		return nil, errors.Newf("event type (%s), currently not supported", eventType)
	}

	// if eventType == "ping" {
	//	var ghPingEvent struct {
	//		Repo *github.Repository `json:"repository,omitempty"`
	//	}
	//	if err := json.Unmarshal(hookBody, &ghPingEvent); err != nil {
	//		return nil, err
	//	}
	//	return &GitWebhookPayload{
	//		GitProvider: "github",
	//		RepoUrl:     *ghPingEvent.Repo.HTMLURL,
	//		GitBranch:      "",
	//		CommitHash:  "",
	//	}, nil
	// }

	// return nil, errors.Newf("unknown event type")
}

func (d *domainI) ParseGitlabHook(eventType string, hookBody []byte) (*GitWebhookPayload, error) {
	hook, err := gitlab.ParseWebhook(gitlab.EventType(eventType), hookBody)
	if err != nil {
		return nil, errors.NewEf(err, "could not parse webhook body for eventType=%s", eventType)
	}
	switch h := hook.(type) {
	case *gitlab.PushEvent:
		{
			payload := &GitWebhookPayload{
				GitProvider: common.ProviderGitlab,
				RepoUrl:     h.Repository.GitHTTPURL,
				GitBranch:   getBranchFromRef(h.Ref),
				CommitHash:  h.CheckoutSHA[:int(math.Min(10, float64(len(h.CheckoutSHA))))],
			}
			return payload, nil
		}
	default:
		return nil, errors.Newf("event type (%s) currently not supported", eventType)
	}
}

func (d *domainI) TektonInterceptorGithub(ctx context.Context, req *tekton.Request) (*TektonVars, *Pipeline, error) {
	reqUrl, err := url.Parse(req.Context.EventURL)
	if err != nil {
		return nil, nil, tekton.NewError(http.StatusBadRequest, err)
	}

	eventType := ""
	headers := req.Header["X-Github-Event"]
	if headers != nil {
		eventType = headers[0]
	}

	if eventType == "" {
		return nil, nil, tekton.NewError(
			http.StatusBadRequest, errors.Newf("could not recognize github event type, aborting ..."),
		)
	}

	hookPayload, err := d.ParseGithubHook(eventType, []byte(req.Body))
	if err != nil {
		return nil, nil, tekton.NewError(
			http.StatusBadRequest,
			errors.NewEf(err, "github (event=%s) is not a push/ping event", eventType),
		)
	}

	pipeline, err := d.pipelineRepo.FindById(ctx, repos.ID(reqUrl.Query().Get("pipelineId")))
	if err != nil {
		return nil, nil, tekton.NewError(http.StatusInternalServerError, err)
	}

	token, err := d.github.GetInstallationToken(ctx, hookPayload.RepoUrl)
	if err != nil {
		return nil, nil, tekton.NewError(http.StatusInternalServerError, err)
	}

	tkVars := TektonVars{
		PipelineId:  pipeline.Id,
		GitRepo:     hookPayload.RepoUrl,
		GitUser:     "x-access-token",
		GitPassword: token,
		// GitBranch:        fmt.Sprintf("refs/heads/%s", pipeline.GitBranch),
		GitBranch:     pipeline.GitBranch,
		GitCommitHash: hookPayload.CommitHash,

		IsDockerBuild: pipeline.DockerBuildInput != nil,
		DockerFile: func() *string {
			if pipeline.DockerBuildInput != nil {
				return &pipeline.DockerBuildInput.DockerFile
			}
			return nil
		}(),
		DockerContextDir: func() *string {
			if pipeline.DockerBuildInput != nil {
				return &pipeline.DockerBuildInput.ContextDir
			}
			return nil
		}(),
		DockerBuildArgs: func() *string {
			if pipeline.DockerBuildInput != nil {
				return &pipeline.DockerBuildInput.BuildArgs
			}
			return nil
		}(),

		BuildBaseImage: func() string {
			if pipeline.Build == nil {
				return ""
			}
			return pipeline.Build.BaseImage
		}(),
		BuildCmd: func() string {
			if pipeline.Build == nil {
				return ""
			}
			return pipeline.Build.Cmd
		}(),
		BuildOutputDir: func() string {
			if pipeline.Build == nil {
				return ""
			}
			return pipeline.Build.OutputDir
		}(),

		RunBaseImage: func() string {
			if pipeline.Run == nil {
				return ""
			}
			return pipeline.Run.Cmd
		}(),
		RunCmd: func() string {
			if pipeline.Run == nil {
				return ""
			}
			return pipeline.Run.Cmd
		}(),
		ArtifactDockerImageName: fmt.Sprintf(
			"%s/%s/%s",
			d.harborHost,
			pipeline.AccountId,
			pipeline.ArtifactRef.DockerImageName,
		),
		ArtifactDockerImageTag: pipeline.ArtifactRef.DockerImageTag,
		TaskNamespace:          pipeline.ProjectName,
	}
	return &tkVars, pipeline, nil
}

func (d *domainI) TektonInterceptorGitlab(ctx context.Context, req *tekton.Request) (*TektonVars, *Pipeline, error) {
	reqUrl, err := url.Parse(req.Context.EventURL)
	if err != nil {
		return nil, nil, tekton.NewError(http.StatusBadRequest, err)
	}

	hookPayload, err := d.ParseGitlabHook(
		func() string {
			if x := req.Header["X-Gitlab-Event"]; len(x) > 0 {
				return x[0]
			}
			return ""
		}(), nil,
	)
	if err != nil {
		return nil, nil, err
		// return tekton.NewResponse(req).errT(err), nil
	}

	pipelineId := repos.ID(reqUrl.Query().Get("pipelineId"))
	pipeline, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil {
		return nil, nil, tekton.NewError(
			http.StatusNotFound, errors.NewEf(err, "could not find pipeline defined by pipelineId=%s", pipelineId),
		)
	}

	if pipeline.AccessTokenId == "" {
		return nil, nil, tekton.NewError(
			http.StatusInternalServerError, errors.NewEf(err, "gitlab tokenId field is null, won't be able to pull repo"),
		)
	}

	token, err := d.gitlabPullToken(ctx, &auth.GetAccessTokenRequest{TokenId: string(pipeline.AccessTokenId)})
	if err != nil {
		return nil, nil, tekton.NewError(
			http.StatusInternalServerError, errors.NewEf(err, "could not retrieve gitlab pull token"),
		)
	}

	tkVars := TektonVars{
		PipelineId:  pipeline.Id,
		GitRepo:     hookPayload.RepoUrl,
		GitUser:     "oauth2",
		GitPassword: token,
		// GitBranch:        fmt.Sprintf("refs/heads/%s", pipeline.GitBranch),
		GitBranch:     pipeline.GitBranch,
		GitCommitHash: hookPayload.CommitHash,

		IsDockerBuild: pipeline.DockerBuildInput != nil,
		DockerFile: func() *string {
			if pipeline.DockerBuildInput != nil {
				return &pipeline.DockerBuildInput.DockerFile
			}
			return nil
		}(),
		DockerContextDir: func() *string {
			if pipeline.DockerBuildInput != nil {
				return &pipeline.DockerBuildInput.ContextDir
			}
			return nil
		}(),
		DockerBuildArgs: func() *string {
			if pipeline.DockerBuildInput != nil {
				return &pipeline.DockerBuildInput.BuildArgs
			}
			return nil
		}(),

		BuildBaseImage: func() string {
			if pipeline.Build == nil {
				return ""
			}
			return pipeline.Build.BaseImage
		}(),
		BuildCmd: func() string {
			if pipeline.Build == nil {
				return ""
			}
			return pipeline.Build.Cmd
		}(),
		BuildOutputDir: func() string {
			if pipeline.Build == nil {
				return ""
			}
			return pipeline.Build.OutputDir
		}(),

		RunBaseImage: func() string {
			if pipeline.Run == nil {
				return ""
			}
			return pipeline.Run.Cmd
		}(),
		RunCmd: func() string {
			if pipeline.Run == nil {
				return ""
			}
			return pipeline.Run.Cmd
		}(),

		ArtifactDockerImageName: fmt.Sprintf(
			"%s/%s/%s",
			d.harborHost,
			pipeline.AccountId,
			pipeline.ArtifactRef.DockerImageName,
		),
		ArtifactDockerImageTag: pipeline.ArtifactRef.DockerImageTag,
		TaskNamespace:          pipeline.ProjectName,
	}

	return &tkVars, pipeline, nil
}

func (d *domainI) gitlabPullToken(ctx context.Context, accTokenReq *auth.GetAccessTokenRequest) (string, error) {
	accessToken, err := d.authClient.GetAccessToken(ctx, accTokenReq)
	if err != nil || accessToken == nil {
		return "", errors.NewEf(err, "could not get gitlab access token")
	}

	accToken := AccessToken{
		UserId:   repos.ID(accessToken.UserId),
		Email:    accessToken.Email,
		Provider: accessToken.Provider,
		Token: &oauth2.Token{
			AccessToken:  accessToken.OauthToken.AccessToken,
			TokenType:    accessToken.OauthToken.TokenType,
			RefreshToken: accessToken.OauthToken.RefreshToken,
			Expiry:       time.UnixMilli(accessToken.OauthToken.Expiry),
		},
		Data: nil,
	}
	repoToken, err := d.gitlab.RepoToken(ctx, &accToken)
	if err != nil || repoToken == nil {
		return "", errors.NewEf(err, "could not get repoToken")
	}
	return repoToken.AccessToken, nil
}

func (d *domainI) GitlabPullToken(ctx context.Context, tokenId repos.ID) (string, error) {
	return d.gitlabPullToken(ctx, &auth.GetAccessTokenRequest{TokenId: string(tokenId)})
}

func (d *domainI) GetPipelines(ctx context.Context, projectId repos.ID) ([]*Pipeline, error) {
	find, err := d.pipelineRepo.Find(
		ctx, repos.Query{
			Filter: repos.Filter{
				"project_id": projectId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return find, nil
}

func (d *domainI) GetTektonRunParams(ctx context.Context, gitProvider string, gitRepoUrl string, gitBranch string) ([]*TektonVars, error) {
	pipelines, err := d.pipelineRepo.Find(
		ctx, repos.Query{
			Filter: repos.Filter{
				"git_provider": gitProvider,
				"git_repo_url": gitRepoUrl,
				"git_branch":   gitBranch,
			},
			Sort: nil,
		},
	)
	if err != nil {
		return nil, err
	}

	tkVars := make([]*TektonVars, 0, len(pipelines))
	for i := range pipelines {
		p := pipelines[i]

		pullToken, err := func() (string, error) {
			if gitProvider == "github" {
				return d.github.GetInstallationToken(ctx, p.GitRepoUrl)
			}
			if gitProvider == "gitlab" {
				return d.gitlabPullToken(ctx, &auth.GetAccessTokenRequest{TokenId: string(p.AccessTokenId)})
			}
			return "", errors.Newf("unknown git provider")
		}()

		if err != nil {
			return nil, err
		}

		tkVars = append(
			tkVars, &TektonVars{
				PipelineId: p.Id,
				GitRepo:    p.GitRepoUrl,
				GitUser: func() string {
					if p.GitProvider == "github" {
						return "x-access-token"
					}
					if p.GitProvider == "gitlab" {
						return "oauth2"
					}
					return ""
				}(),
				GitPassword:   pullToken,
				GitBranch:     p.GitBranch,
				IsDockerBuild: p.DockerBuildInput != nil,
				DockerFile: func() *string {
					if p.DockerBuildInput != nil {
						return &p.DockerBuildInput.DockerFile
					}
					return nil
				}(),
				DockerContextDir: func() *string {
					if p.DockerBuildInput != nil {
						return &p.DockerBuildInput.ContextDir
					}
					return nil
				}(),
				DockerBuildArgs: func() *string {
					if p.DockerBuildInput != nil {
						return &p.DockerBuildInput.BuildArgs
					}
					return nil
				}(),
				BuildBaseImage: func() string {
					if p.Build == nil {
						return ""
					}
					return p.Build.BaseImage
				}(),
				BuildCmd: func() string {
					if p.Build == nil {
						return ""
					}
					return p.Build.Cmd
				}(),
				BuildOutputDir: func() string {
					if p.Build == nil {
						return ""
					}
					return p.Build.OutputDir
				}(),
				RunBaseImage: func() string {
					if p.Run == nil {
						return ""
					}
					return p.Run.Cmd
				}(),
				RunCmd: func() string {
					if p.Run == nil {
						return ""
					}
					return p.Run.Cmd
				}(),
				TaskNamespace:           p.ProjectName,
				ArtifactDockerImageName: fmt.Sprintf("%s/%s/%s", d.harborHost, p.AccountId, p.ArtifactRef.DockerImageName),
				ArtifactDockerImageTag:  p.ArtifactRef.DockerImageTag,
			},
		)
	}

	return tkVars, nil
}

func (d *domainI) TriggerHook(p *Pipeline, latestCommitSHA string) error {
	var req *http.Request
	if p.GitProvider == common.ProviderGithub {
		body := t.M{
			"ref":   fmt.Sprintf("refs/heads/%s", p.GitBranch),
			"after": latestCommitSHA,
			"repository": t.M{
				"html_url": p.GitRepoUrl,
			},
		}

		b, err := json.Marshal(body)
		if err != nil {
			return errors.ErrMarshal(err)
		}
		req, err = http.NewRequest(http.MethodPost, d.githubWebhookUrl, bytes.NewBuffer(b))
		req.Header.Set("X-Github-Event", "push")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Triggered-By", "kloudlite/ci")
		if err != nil {
			return errors.NewEf(err, "could not build http request")
		}
	}

	if p.GitProvider == common.ProviderGitlab {
		body := t.M{
			"ref":          fmt.Sprintf("refs/heads/%s", p.GitBranch),
			"checkout_sha": latestCommitSHA,
			"repository": t.M{
				"git_http_url": p.GitRepoUrl,
			},
		}
		b, err := json.Marshal(body)
		if err != nil {
			return errors.ErrMarshal(err)
		}
		req, err = http.NewRequest(http.MethodPost, d.gitlabWebhookUrl, bytes.NewBuffer(b))
		req.Header.Set("X-Gitlab-Event", "Push Hook")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Triggered-By", "kloudlite/ci")
		if err != nil {
			return errors.NewEf(err, "could not build http request")
		}
	}

	if req != nil {
		r, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.NewEf(err, "while making request")
		}
		if r.StatusCode == http.StatusAccepted {
			return nil
		}
		return errors.Newf("trigger for repo=%s failed as received StatusCode=%s", p.GitRepoUrl, r.StatusCode)
	}
	return errors.Newf("unknown gitProvider=%s, aborting trigger", p.GitProvider)
}

func (d *domainI) GitlabListGroups(ctx context.Context, userId repos.ID, query *string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}
	return d.gitlab.ListGroups(ctx, token, query, pagination)
}

func (d *domainI) GitlabListRepos(ctx context.Context, userId repos.ID, gid string, query *string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}
	return d.gitlab.ListRepos(ctx, token, gid, query, pagination)
}

func (d *domainI) GitlabListBranches(ctx context.Context, userId repos.ID, repoId string, query *string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}
	return d.gitlab.ListBranches(ctx, token, repoId, query, pagination)
}

func (d *domainI) GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string, pipelineId repos.ID) (repos.ID, error) {
	grHook, err := d.gitRepoHookRepo.FindOne(ctx, repos.Filter{"httpUrl": d.gitlabWebhookUrl})
	if err != nil {
		return "", err
	}

	token, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)
	if err != nil {
		return "", err
	}

	if grHook != nil {
		exists, _ := d.gitlab.CheckWebhookExists(ctx, token, repoId, grHook.GitlabWebhookId)
		if exists {
			return grHook.Id, nil
		}
		if err := d.gitRepoHookRepo.DeleteById(ctx, grHook.Id); err != nil {
			return "", err
		}
	}

	webhookId, err := d.gitlab.AddWebhook(ctx, token, repoId, string(pipelineId))
	if err != nil {
		return "", err
	}

	grHook, err = d.gitRepoHookRepo.Create(
		ctx, &GitRepositoryHook{
			HttpUrl:         d.gitlabWebhookUrl,
			GitProvider:     GitlabLabel,
			GitlabWebhookId: webhookId,
		},
	)

	if err != nil {
		return "", err
	}
	return grHook.Id, nil
}

func (d *domainI) SaveUserAcc(ctx context.Context, acc *HarborAccount) error {
	_, err := d.harborAccRepo.Create(ctx, acc)
	if err != nil {
		return errors.NewEf(err, "[dbRepo] failed to create harbor account")
	}
	return nil
}

func (d *domainI) getAccessTokenByTokenId(ctx context.Context, tokenId repos.ID) (*AccessToken, error) {
	accTokenOut, err := d.authClient.GetAccessToken(
		ctx, &auth.GetAccessTokenRequest{
			TokenId: string(tokenId),
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "finding accessToken")
	}
	return &AccessToken{
		Id:       repos.ID(accTokenOut.Id),
		UserId:   repos.ID(accTokenOut.UserId),
		Email:    accTokenOut.Email,
		Provider: accTokenOut.Provider,
		Token: &oauth2.Token{
			AccessToken:  accTokenOut.OauthToken.AccessToken,
			TokenType:    accTokenOut.OauthToken.TokenType,
			RefreshToken: accTokenOut.OauthToken.RefreshToken,
			Expiry:       time.UnixMilli(accTokenOut.OauthToken.Expiry),
		},
	}, err
}

func (d *domainI) getAccessTokenByUserId(ctx context.Context, provider string, userId repos.ID) (*AccessToken, error) {
	accTokenOut, err := d.authClient.GetAccessToken(
		ctx, &auth.GetAccessTokenRequest{
			UserId:   string(userId),
			Provider: provider,
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "finding accessToken")
	}
	return &AccessToken{
		Id:       repos.ID(accTokenOut.Id),
		UserId:   repos.ID(accTokenOut.UserId),
		Email:    accTokenOut.Email,
		Provider: accTokenOut.Provider,
		Token: &oauth2.Token{
			AccessToken:  accTokenOut.OauthToken.AccessToken,
			TokenType:    accTokenOut.OauthToken.TokenType,
			RefreshToken: accTokenOut.OauthToken.RefreshToken,
			Expiry:       time.UnixMilli(accTokenOut.OauthToken.Expiry),
		},
	}, err
}

func (d *domainI) GithubInstallationToken(ctx context.Context, repoUrl string) (string, error) {
	return d.github.GetInstallationToken(ctx, repoUrl)
}

func (d *domainI) GithubListBranches(ctx context.Context, userId repos.ID, repoUrl string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	if err != nil {
		return "", err
	}
	return d.github.ListBranches(ctx, token, repoUrl, pagination)
}

func (d *domainI) GithubAddWebhook(ctx context.Context, userId repos.ID, repoUrl string) (repos.ID, error) {
	grHook, err := d.gitRepoHookRepo.FindOne(ctx, repos.Filter{"httpUrl": d.githubWebhookUrl})
	if err != nil {
		return "", err
	}

	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	if err != nil {
		return "", err
	}

	if grHook != nil {
		exists, _ := d.github.CheckWebhookExists(ctx, token, repoUrl, grHook.GithubWebhookId)
		if exists {
			return grHook.Id, nil
		}
		if err := d.gitRepoHookRepo.DeleteById(ctx, grHook.Id); err != nil {
			return "", err
		}
	}

	webhookId, err := d.github.AddWebhook(ctx, token, repoUrl, d.githubWebhookUrl)
	if err != nil {
		return "", err
	}

	grHook, err = d.gitRepoHookRepo.Create(
		ctx, &GitRepositoryHook{
			HttpUrl:         d.githubWebhookUrl,
			GitProvider:     GithubLabel,
			GithubWebhookId: webhookId,
		},
	)

	if err != nil {
		return "", err
	}
	return grHook.Id, nil
}

func (d *domainI) GithubSearchRepos(ctx context.Context, userId repos.ID, q, org string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, errors.NewEf(err, "while finding accessToken")
	}
	return d.github.SearchRepos(ctx, token, q, org, pagination)
}

func (d *domainI) GithubListRepos(ctx context.Context, userId repos.ID, installationId int64, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	return d.github.ListRepos(ctx, token, installationId, pagination)
}

func (d *domainI) GithubListInstallations(ctx context.Context, userId repos.ID, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	i, err := d.github.ListInstallations(ctx, token, pagination)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (d *domainI) CreatePipeline(ctx context.Context, userId repos.ID, pipeline Pipeline) (*Pipeline, error) {
	exP, err := d.pipelineRepo.FindOne(
		ctx, repos.Filter{
			"project_id":     pipeline.ProjectId,
			"app_id":         pipeline.AppId,
			"container_name": pipeline.ContainerName,
		},
	)
	if err != nil {
		return nil, err
	}

	if exP != nil {
		pipeline.Id = exP.Id
	} else {
		pipeline.Id = d.pipelineRepo.NewId()
	}

	latestCommit := ""
	if pipeline.GitProvider == common.ProviderGithub {
		token, err := d.getAccessTokenByUserId(ctx, "github", userId)
		if err != nil {
			return nil, err
		}

		pipeline.AccessTokenId = token.Id
		hookId, err := d.GithubAddWebhook(ctx, userId, pipeline.GitRepoUrl)
		if err != nil {
			return nil, err
		}
		pipeline.WebhookId = hookId

		// pipeline.WebhookId = hookId
		commit, err := d.github.GetLatestCommit(ctx, token, pipeline.GitRepoUrl, pipeline.GitBranch)
		if err != nil {
			return nil, err
		}
		latestCommit = commit
	}

	if pipeline.GitProvider == common.ProviderGitlab {
		token, err := d.getAccessTokenByUserId(ctx, pipeline.GitProvider, userId)
		if err != nil {
			return nil, err
		}
		pipeline.AccessTokenId = token.Id
		hookId, err := d.GitlabAddWebhook(ctx, userId, d.gitlab.GetRepoId(pipeline.GitRepoUrl), pipeline.Id)
		if err != nil {
			return nil, err
		}
		pipeline.WebhookId = hookId

		commit, err := d.gitlab.GetLatestCommit(ctx, token, pipeline.GitRepoUrl, pipeline.GitBranch)
		if err != nil {
			return nil, err
		}
		latestCommit = commit
	}

	p, err := d.pipelineRepo.Upsert(ctx, repos.Filter{"id": pipeline.Id}, &pipeline)
	if err != nil {
		return nil, err
	}
	if err = d.TriggerHook(p, latestCommit); err != nil {
		return nil, errors.NewEf(err, "failed to trigger webhook")
	}
	return p, nil
}

func (d *domainI) DeletePipeline(ctx context.Context, pipelineId repos.ID) error {
	// TODO: now not deleting github/gitlab webhook on our pipeline delete
	return d.pipelineRepo.DeleteById(ctx, pipelineId)

	// if pipeline.GitProvider == common.ProviderGithub {
	//	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	//	if err != nil {
	//		return err
	//	}
	//	//if err := d.github.DeleteWebhook(ctx, token, pipeline.GitRepoUrl, *pipeline.WebhookId); err != nil {
	//	//	return err
	//	//}
	// }
	//
	// if pipeline.GitProvider == common.ProviderGitlab {
	//	token, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)
	//	if err != nil {
	//		return err
	//	}
	//	//if err := d.gitlab.DeleteWebhook(ctx, token, pipeline.GitRepoUrl, *pipeline.WebhookId); err != nil {
	//	//	return err
	//	//}
	// }

}

func (d *domainI) TriggerPipeline(ctx context.Context, userId repos.ID, pipelineId repos.ID) error {
	pipeline, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil {
		return err
	}

	token, err := d.getAccessTokenByUserId(ctx, pipeline.GitProvider, userId)
	if err != nil {
		return err
	}

	var latestCommit string
	if pipeline.GitProvider == common.ProviderGithub {
		latestCommit, err = d.github.GetLatestCommit(ctx, token, pipeline.GitRepoUrl, pipeline.GitBranch)
		if err != nil {
			return errors.NewEf(err, "getting latest commit")
		}
	}
	if pipeline.GitProvider == common.ProviderGitlab {
		latestCommit, err = d.gitlab.GetLatestCommit(ctx, token, pipeline.GitRepoUrl, pipeline.GitBranch)
		if err != nil {
			return errors.NewEf(err, "getting latest commit")
		}
	}
	return d.TriggerHook(pipeline, latestCommit)
}

func (d *domainI) GetPipeline(ctx context.Context, pipelineId repos.ID) (*Pipeline, error) {
	id, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil {
		return nil, err
	}
	return id, nil
}

var Module = fx.Module(
	"domain",

	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(
		func(
			pipelineRepo repos.DbRepo[*Pipeline],
			authClient auth.AuthClient,
			gitlab Gitlab,
			github Github,
			harborHost HarborHost,
			logger logging.Logger,
			harborCli *harbor.Client,
			gitRepoHookRepo repos.DbRepo[*GitRepositoryHook],
			env *Env,
		) Domain {
			return &domainI{
				authClient:       authClient,
				pipelineRepo:     pipelineRepo,
				gitlab:           gitlab,
				github:           github,
				harborHost:       harborHost,
				logger:           logger.WithName("[ci]:domain/main.go"),
				harborCli:        harborCli,
				gitRepoHookRepo:  gitRepoHookRepo,
				gitlabWebhookUrl: env.GitlabWebhookUrl,
				githubWebhookUrl: env.GithubWebhookUrl,
			}
		},
	),
)
