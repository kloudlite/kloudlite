package domain

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"golang.org/x/oauth2"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	"time"
)

type Domain interface {
	GetPipeline(ctx context.Context, pipelineId repos.ID) (*Pipeline, error)
	CretePipeline(ctx context.Context, pipeline Pipeline) (*Pipeline, error)

	GithubInstallationToken(ctx context.Context, repoUrl string) (string, error)
	GithubListInstallations(ctx context.Context) (any, error)
	GithubListRepos(ctx context.Context, installationId int64, page, size int) (any, error)
	GithubSearchRepos(ctx context.Context, q string, org string, page, size int) (any, error)
	GithubListBranches(ctx context.Context, repoUrl string, page, size int) (any, error)
	GithubAddWebhook(ctx context.Context, repoUrl string) error
}

type domainI struct {
	pipelineRepo repos.DbRepo[*Pipeline]
	authClient   auth.AuthClient
	github       Github
	gitlab       Gitlab
}

func (d *domainI) getAccessToken(ctx context.Context, provider string) (*AccessToken, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("no session")
	}
	accTokenOut, err := d.authClient.GetAccessToken(ctx, &auth.GetAccessTokenRequest{
		UserId:   string(session.UserId),
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

func (d *domainI) GithubInstallationToken(ctx context.Context, repoUrl string) (string, error) {
	token, err := d.getAccessToken(ctx, "github")
	if err != nil {
		return "", err
	}
	return d.github.GetInstallationToken(ctx, token, repoUrl)
}

func (d *domainI) GithubListBranches(ctx context.Context, repoUrl string, page int, size int) (any, error) {
	token, err := d.getAccessToken(ctx, "github")
	if err != nil {
		return "", err
	}
	return d.github.ListBranches(ctx, token, repoUrl, page, size)
}

func (d *domainI) GithubAddWebhook(ctx context.Context, repoUrl string) error {
	token, err := d.getAccessToken(ctx, "github")
	if err != nil {
		return err
	}
	return d.github.AddWebhook(ctx, token, repoUrl)
}

func (d *domainI) GithubSearchRepos(ctx context.Context, q string, org string, page int, size int) (any, error) {
	token, err := d.getAccessToken(ctx, "github")
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, errors.NewEf(err, "while finding accessToken")
	}
	return d.github.SearchRepos(ctx, token, q, org, page, size)
}

func (d *domainI) GithubListRepos(ctx context.Context, instId int64, page int, size int) (any, error) {
	token, err := d.getAccessToken(ctx, "github")
	if err != nil {
		return nil, err
	}
	return d.github.ListRepos(ctx, token, instId, page, size)
}

func (d *domainI) GithubListInstallations(ctx context.Context) (any, error) {
	token, err := d.getAccessToken(ctx, "github")
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

func (d *domainI) CretePipeline(ctx context.Context, pipeline Pipeline) (*Pipeline, error) {
	return d.pipelineRepo.Create(ctx, &pipeline)
}

func (d *domainI) GetPipeline(ctx context.Context, pipelineId repos.ID) (*Pipeline, error) {
	id, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func fxDomain(pipelineRepo repos.DbRepo[*Pipeline], authClient auth.AuthClient) Domain {
	return &domainI{
		authClient:   authClient,
		pipelineRepo: pipelineRepo,
	}
}

var Module = fx.Module("domain",
	fx.Provide(fxDomain),
)
