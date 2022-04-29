package domain

import (
	"context"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/types"
	"time"
)

type domainI struct {
	pipelineRepo  repos.DbRepo[*Pipeline]
	authClient    auth.AuthClient
	github        Github
	gitlab        Gitlab
	harborAccRepo repos.DbRepo[*HarborAccount]
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

func (d *domainI) GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string, pipelineId string) (*gitlab.ProjectHook, error) {
	token, err := d.getAccessToken(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}
	return d.gitlab.AddWebhook(ctx, token, repoId, pipelineId)
}

func (d *domainI) SaveUserAcc(ctx context.Context, acc *HarborAccount) error {
	acc, err := d.harborAccRepo.Create(ctx, acc)
	if err != nil {
		return errors.NewEf(err, "[dbRepo] failed to create harbor account")
	}
	return nil
}

func (d *domainI) getAccessToken(ctx context.Context, provider string, userId repos.ID) (*AccessToken, error) {
	accTokenOut, err := d.authClient.GetAccessToken(ctx, &auth.GetAccessTokenRequest{
		UserId:   string(userId),
		Provider: provider,
	})
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, errors.NewEf(err, "finding accessToken")
	}
	return &AccessToken{
		UserId:   repos.ID(accTokenOut.UserId),
		Email:    accTokenOut.Email,
		Provider: accTokenOut.Provider,
		Token: &oauth2.Token{
			AccessToken:  accTokenOut.OauthToken.AccessToken,
			TokenType:    accTokenOut.OauthToken.TokenType,
			RefreshToken: accTokenOut.OauthToken.RefreshToken,
			Expiry:       time.UnixMilli(accTokenOut.OauthToken.Expiry),
		},
		Data: nil,
	}, err
}

func (d *domainI) GithubInstallationToken(ctx context.Context, repoUrl string, instId int64) (string, error) {
	return d.github.GetInstallationToken(ctx, repoUrl, instId)
}

func (d *domainI) GithubListBranches(ctx context.Context, userId repos.ID, repoUrl string, page int, size int) (any, error) {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return "", err
	}
	return d.github.ListBranches(ctx, token, repoUrl, page, size)
}

func (d *domainI) GithubAddWebhook(ctx context.Context, userId repos.ID, refId string, repoUrl string) error {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return err
	}
	return d.github.AddWebhook(ctx, token, refId, repoUrl)
}

func (d *domainI) GithubSearchRepos(ctx context.Context, userId repos.ID, q string, org string, page int, size int) (any, error) {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, errors.NewEf(err, "while finding accessToken")
	}
	return d.github.SearchRepos(ctx, token, q, org, page, size)
}

func (d *domainI) GithubListRepos(ctx context.Context, userId repos.ID, instId int64, page int, size int) (any, error) {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	return d.github.ListRepos(ctx, token, instId, page, size)
}

func (d *domainI) GithubListInstallations(ctx context.Context, userId repos.ID) (any, error) {
	token, err := d.getAccessToken(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	i, err := d.github.ListInstallations(ctx, token)
	if err != nil {
		return nil, err
	}
	fmt.Printf("item: %+v\n", i[0])
	return i, nil
}

func (d *domainI) CretePipeline(ctx context.Context, userId repos.ID, pipeline Pipeline) (*Pipeline, error) {
	pipeline.Id = d.pipelineRepo.NewId()
	err := d.GithubAddWebhook(ctx, userId, string(pipeline.Id), pipeline.GitRepoUrl)
	if err != nil {
		return nil, err
	}
	return d.pipelineRepo.Create(ctx, &pipeline)
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

var Module = fx.Module("domain",
	fx.Provide(fxDomain),
)
