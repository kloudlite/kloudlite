package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v43/github"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/tekton"
	"kloudlite.io/pkg/types"
	"net/http"
	"net/url"
	"time"
)

type domainI struct {
	pipelineRepo  repos.DbRepo[*Pipeline]
	authClient    auth.AuthClient
	github        Github
	gitlab        Gitlab
	harborAccRepo repos.DbRepo[*HarborAccount]
}

type GitWebhookPayload struct {
	GitProvider string `json:"git_provider,omitempty"`
	RepoUrl     string `json:"repo_url,omitempty"`
	GitRef      string `json:"git_ref,omitempty"`
	CommitHash  string `json:"commit_hash,omitempty"`
}

func (d *domainI) parseGithubHook(eventType string, hookBody []byte) (*GitWebhookPayload, error) {
	hook, err := github.ParseWebHook(eventType, hookBody)
	if err != nil {
		return nil, err
	}
	switch t := hook.(type) {
	case *github.PushEvent:
		{
			payload := GitWebhookPayload{
				GitProvider: common.ProviderGithub,
				RepoUrl:     *t.Repo.HTMLURL,
				GitRef:      *t.Ref,
				CommitHash:  t.GetAfter()[:10],
			}
			return &payload, nil
		}
	}

	if eventType == "ping" {
		var ghPingEvent struct {
			Repo *github.Repository `json:"repository,omitempty"`
		}
		if err := json.Unmarshal(hookBody, &ghPingEvent); err != nil {
			return nil, err
		}
		return &GitWebhookPayload{
			GitProvider: "github",
			RepoUrl:     *ghPingEvent.Repo.HTMLURL,
			GitRef:      "",
			CommitHash:  "",
		}, nil
	}

	return nil, errors.Newf("unknown event type")
}

func (d *domainI) parseGitlabHook(req *tekton.Request) (*GitWebhookPayload, error) {
	eventType := func() gitlab.EventType {
		headers := req.Header["X-Gitlab-Event"]
		if headers != nil {
			return gitlab.EventType(headers[0])
		}
		return ""
	}()
	hook, err := gitlab.ParseWebhook(eventType, []byte(req.Body))
	if err != nil {
		return nil, errors.NewEf(err, "could not parse webhook body for eventType=%s", eventType)
	}
	switch t := hook.(type) {
	case *gitlab.PushEvent:
		{
			payload := &GitWebhookPayload{
				GitProvider: common.ProviderGitlab,
				RepoUrl:     t.Repository.GitHTTPURL,
				GitRef:      t.Ref,
				CommitHash:  t.CheckoutSHA,
			}
			return payload, nil
		}
	}
	return nil, errors.Newf("unknown eventType=", eventType)
}

func (d *domainI) TektonInterceptorGithub(ctx context.Context, req *tekton.Request) *tekton.Response {
	reqUrl, err := url.Parse(req.Context.EventURL)
	if err != nil {
		return tekton.NewResponse(req).Err(err, http.StatusBadRequest)
	}

	if !reqUrl.Query().Has("pipelineId") {
		return tekton.NewResponse(req).Err(
			errors.NewEf(err, "url does not have query params 'pipelineId'"),
			http.StatusBadRequest,
		)
	}

	pipeline, err := d.pipelineRepo.FindById(ctx, repos.ID(reqUrl.Query().Get("pipelineId")))
	if err != nil {
		return tekton.NewResponse(req).Err(err, http.StatusInternalServerError)
	}

	eventType := ""
	headers := req.Header["X-Github-Event"]
	if headers != nil {
		eventType = headers[0]
	}

	if eventType == "" {
		return tekton.NewResponse(req).Err(errors.Newf("could not recognize github event type, aborting ..."))
	}

	hookPayload, err := d.parseGithubHook(eventType, []byte(req.Body))
	if err != nil {
		err := errors.NewEf(err, "github (event=%s) is not a push/ping event", eventType)
		fmt.Printf("ERR occurred: %+v\n", err)
		return tekton.NewResponse(req).Err(err)
	}

	token, err := d.github.GetInstallationToken(ctx, hookPayload.RepoUrl)
	if err != nil {
		return tekton.NewResponse(req).Err(err)
	}

	// if hookPayload.GitRef != pipeline.GitBranch {
	// 	return tekton.NewResponse(req).Err(
	// 		errors.Newf(
	// 			"pipeline is not configured to run on this (ref=%s)", hookPayload.GitRef,
	// 		),
	// 	)
	// }

	tkVars := TektonVars{
		GitRepo:                 hookPayload.RepoUrl,
		GitUser:                 "x-access-token",
		GitPassword:             token,
		GitRef:                  fmt.Sprintf("refs/heads/%s", pipeline.GitBranch),
		GitCommitHash:           hookPayload.CommitHash,
		BuildBaseImage:          pipeline.Build.BaseImage,
		BuildCmd:                pipeline.Build.Cmd,
		RunBaseImage:            pipeline.Run.BaseImage,
		RunCmd:                  pipeline.Run.Cmd,
		ArtifactDockerImageName: pipeline.ArtifactRef.DockerImageName,
		ArtifactDockerImageTag:  pipeline.ArtifactRef.DockerImageTag,
	}

	j, err := tkVars.ToJson()
	if err != nil {
		return tekton.NewResponse(req).Err(err)
	}

	return tekton.NewResponse(req).Extend(j).Ok()
}

func (d *domainI) TektonInterceptorGitlab(ctx context.Context, req *tekton.Request) *tekton.Response {
	reqUrl, err := url.Parse(req.Context.EventURL)
	if err != nil {
		return tekton.NewResponse(req).Err(err, http.StatusBadRequest)
	}

	if !reqUrl.Query().Has("pipelineId") {
		return tekton.NewResponse(req).Err(
			errors.NewEf(err, "url does not have query params 'pipelineId'"),
			http.StatusBadRequest,
		)
	}

	hookPayload, err := d.parseGitlabHook(req)
	if err != nil {
		return tekton.NewResponse(req).Err(err)
	}

	pipeline, err := d.pipelineRepo.FindById(ctx, repos.ID(reqUrl.Query().Get("pipelineId")))
	if err != nil {
		return tekton.NewResponse(req).Err(err, http.StatusInternalServerError)
	}

	token, err := d.gitlabPullToken(ctx, &auth.GetAccessTokenRequest{TokenId: pipeline.GitlabTokenId})
	if err != nil {
		return tekton.NewResponse(req).Err(err)
	}

	tkVars := TektonVars{
		GitRepo:                 hookPayload.RepoUrl,
		GitUser:                 "oauth2",
		GitPassword:             token,
		GitRef:                  fmt.Sprintf("refs/heads/%s", pipeline.GitBranch),
		GitCommitHash:           hookPayload.CommitHash,
		BuildBaseImage:          pipeline.Build.BaseImage,
		BuildCmd:                pipeline.Build.Cmd,
		RunBaseImage:            pipeline.Run.BaseImage,
		RunCmd:                  pipeline.Run.Cmd,
		ArtifactDockerImageName: pipeline.ArtifactRef.DockerImageName,
		ArtifactDockerImageTag:  pipeline.ArtifactRef.DockerImageTag,
	}

	m, err := tkVars.ToJson()
	if err != nil {
		return tekton.NewResponse(req).Err(err)
	}

	return tekton.NewResponse(req).Extend(m).Ok()
}

func (d *domainI) gitlabPullToken(ctx context.Context, accTokenReq *auth.GetAccessTokenRequest) (string, error) {
	accessToken, err := d.authClient.GetAccessToken(ctx, accTokenReq)
	fmt.Println("accessToken: ", accessToken, "err:", err)

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

func (d *domainI) GitlabPullToken(ctx context.Context, pipelineId repos.ID) (string, error) {
	pipeline, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil {
		return "", err
	}
	return d.gitlabPullToken(ctx, &auth.GetAccessTokenRequest{TokenId: pipeline.GitlabTokenId})
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

func (d *domainI) GitlabListGroups(ctx context.Context, userId repos.ID, query *string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessToken(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}
	return d.gitlab.ListGroups(ctx, token, query, pagination)
}

func (d *domainI) GitlabListRepos(ctx context.Context, userId repos.ID, gid string, query *string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessToken(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}
	return d.gitlab.ListRepos(ctx, token, gid, query, pagination)
}

func (d *domainI) GitlabListBranches(ctx context.Context, userId repos.ID, repoId string, query *string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessToken(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}
	return d.gitlab.ListBranches(ctx, token, repoId, query, pagination)
}

func (d *domainI) GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string, pipelineId repos.ID) error {
	token, err := d.getAccessToken(ctx, "gitlab", userId)
	if err != nil {
		return err
	}
	webhook, err := d.gitlab.AddWebhook(ctx, token, repoId, string(pipelineId))
	_, err = d.pipelineRepo.UpdateById(
		ctx, pipelineId, &Pipeline{
			GitlabWebhookId: &webhook.ID,
		},
	)
	return err
}

func (d *domainI) SaveUserAcc(ctx context.Context, acc *HarborAccount) error {
	_, err := d.harborAccRepo.Create(ctx, acc)
	if err != nil {
		return errors.NewEf(err, "[dbRepo] failed to create harbor account")
	}
	return nil
}

func (d *domainI) getAccessToken(ctx context.Context, provider string, userId repos.ID) (*AccessToken, error) {
	accTokenOut, err := d.authClient.GetAccessToken(
		ctx, &auth.GetAccessTokenRequest{
			UserId:   string(userId),
			Provider: provider,
		},
	)
	if err != nil {
		return nil, err
	}
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

func (d *domainI) GithubInstallationToken(ctx context.Context, pipelineId repos.ID) (string, error) {
	pipeline, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil || pipeline == nil {
		return "", err
	}
	return d.github.GetInstallationToken(ctx, "")
}

func (d *domainI) GithubListBranches(ctx context.Context, userId repos.ID, repoUrl string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return "", err
	}
	return d.github.ListBranches(ctx, token, repoUrl, pagination)
}

func (d *domainI) GithubAddWebhook(ctx context.Context, userId repos.ID, pipelineId repos.ID, repoUrl string) error {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return err
	}
	hookId, err := d.github.AddWebhook(ctx, token, string(pipelineId), repoUrl)
	if err != nil {
		return err
	}

	_, err = d.pipelineRepo.UpdateById(
		ctx, pipelineId, &Pipeline{
			GithubWebhookId: hookId,
		},
	)
	return err
}

func (d *domainI) GithubSearchRepos(ctx context.Context, userId repos.ID, q, org string, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, errors.NewEf(err, "while finding accessToken")
	}
	return d.github.SearchRepos(ctx, token, q, org, pagination)
}

func (d *domainI) GithubListRepos(ctx context.Context, userId repos.ID, installationId int64, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	return d.github.ListRepos(ctx, token, installationId, pagination)
}

func (d *domainI) GithubListInstallations(ctx context.Context, userId repos.ID, pagination *types.Pagination) (any, error) {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	i, err := d.github.ListInstallations(ctx, token, pagination)
	if err != nil {
		return nil, err
	}
	fmt.Printf("item: %+v\n", i[0])
	return i, nil
}

func (d *domainI) CreatePipeline(ctx context.Context, userId repos.ID, pipeline Pipeline) (*Pipeline, error) {
	exP, err := d.pipelineRepo.FindOne(
		ctx, repos.Filter{
			"git_repo_url": pipeline.GitRepoUrl,
			"git_branch":   pipeline.GitBranch,
			"git_provider": pipeline.GitProvider,
			"project_id":   pipeline.ProjectId,
			"app_id":       pipeline.AppId,
		},
	)
	if err != nil {
		return nil, err
	}

	if exP == nil {
		pipeline.Id = exP.Id
	} else {
		pipeline.Id = d.pipelineRepo.NewId()
	}

	latestCommit := ""
	if pipeline.GitProvider == common.ProviderGithub {
		err := d.GithubAddWebhook(ctx, userId, pipeline.Id, pipeline.GitRepoUrl)
		if err != nil {
			return nil, err
		}
		commit, err := d.github.GetLatestCommit(ctx, pipeline.GitRepoUrl, pipeline.GitBranch)
		if err != nil {
			return nil, err
		}
		latestCommit = commit
	}

	if pipeline.GitProvider == common.ProviderGitlab {
		token, err := d.getAccessToken(ctx, pipeline.GitProvider, userId)
		if err != nil {
			return nil, err
		}
		pipeline.GitlabTokenId = string(token.Id)
		// TODO check webhook id
		err = d.GitlabAddWebhook(ctx, userId, d.gitlab.GetRepoId(pipeline.GitRepoUrl), pipeline.Id)
		if err != nil {
			return nil, err
		}

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
	if err = p.TriggerHook(latestCommit); err != nil {
		return nil, errors.NewEf(err, "failed to trigger webhook")
	}
	return p, nil
}

func (d *domainI) DeletePipeline(ctx context.Context, userId, pipelineId repos.ID) error {
	pipeline, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil {
		return err
	}

	if pipeline.GitProvider == common.ProviderGithub {
		token, err := d.getAccessToken(ctx, "github", userId)
		if err != nil {
			return err
		}
		if err := d.github.DeleteWebhook(ctx, token, pipeline.GitRepoUrl, *pipeline.GithubWebhookId); err != nil {
			return err
		}
	}

	if pipeline.GitProvider == common.ProviderGitlab {
		token, err := d.getAccessToken(ctx, "gitlab", userId)
		if err != nil {
			return err
		}
		if err := d.gitlab.DeleteWebhook(ctx, token, pipeline.GitRepoUrl, *pipeline.GitlabWebhookId); err != nil {
			return err
		}
	}

	return d.pipelineRepo.DeleteById(ctx, pipelineId)
}

func (d *domainI) TriggerPipeline(ctx context.Context, userId repos.ID, pipelineId repos.ID) error {
	pipeline, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil {
		return err
	}

	var latestCommit string
	if pipeline.GitProvider == common.ProviderGithub {
		latestCommit, err = d.github.GetLatestCommit(ctx, pipeline.GitRepoUrl, pipeline.GitBranch)
		if err != nil {
			return errors.NewEf(err, "getting latest commit")
		}
	}

	if pipeline.GitProvider == common.ProviderGitlab {
		token, err := d.getAccessToken(ctx, pipeline.GitProvider, userId)
		if err != nil {
			return err
		}
		latestCommit, err = d.gitlab.GetLatestCommit(ctx, token, pipeline.GitRepoUrl, pipeline.GitBranch)
		if err != nil {
			return errors.NewEf(err, "getting latest commit")
		}
	}

	return pipeline.TriggerHook(latestCommit)
}

func (d *domainI) GetPipeline(ctx context.Context, pipelineId repos.ID) (*Pipeline, error) {
	id, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func fxDomain(pipelineRepo repos.DbRepo[*Pipeline], harborAccRepo repos.DbRepo[*HarborAccount], authClient auth.AuthClient, gitlab Gitlab, github Github) (Domain, Harbor) {
	d := domainI{
		authClient:    authClient,
		pipelineRepo:  pipelineRepo,
		gitlab:        gitlab,
		github:        github,
		harborAccRepo: harborAccRepo,
	}
	return &d, &d
}

var Module = fx.Module(
	"domain",
	fx.Provide(fxDomain),
)
