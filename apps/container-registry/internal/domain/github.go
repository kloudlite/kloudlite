package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/api/pkg/types"
	"golang.org/x/oauth2"
)

type ErrEventNotSupported struct {
	err error
}

func (e *ErrEventNotSupported) Error() string {
	return e.err.Error()
}

type GitWebhookPayload struct {
	GitProvider string `json:"git_provider,omitempty"`
	RepoUrl     string `json:"repo_url,omitempty"`
	GitBranch   string `json:"git_branch,omitempty"`
	CommitHash  string `json:"commit_hash,omitempty"`
}

func getBranchFromRef(gitRef string) string {
	sp := strings.SplitN(gitRef, "/", 3)
	if len(sp) == 3 {
		return sp[2]
	}
	return ""
}

func (d *Impl) ParseGithubHook(eventType string, hookBody []byte) (*GitWebhookPayload, error) {
	hook, err := github.ParseWebHook(eventType, hookBody)
	if err != nil {
		d.logger.Infof("bad webhook body, dropping message ...")
		return nil, err
	}

	switch h := hook.(type) {
	case *github.PushEvent:
		{

			payload := GitWebhookPayload{
				GitProvider: constants.ProviderGithub,
				RepoUrl:     *h.Repo.CloneURL,
				GitBranch:   getBranchFromRef(h.GetRef()),
				CommitHash:  h.GetAfter(),
			}
			return &payload, nil
		}
	default:
		return nil, &ErrEventNotSupported{err: fmt.Errorf("event type (%s), currently not supported", eventType)}
	}

}

func (d *Impl) getAccessTokenByUserId(ctx context.Context, provider string, userId repos.ID) (*entities.AccessToken, error) {
	accTokenOut, err := d.authClient.GetAccessToken(
		ctx, &auth.GetAccessTokenRequest{
			UserId:   string(userId),
			Provider: provider,
		},
	)
	if err != nil {
		return nil, errors.NewEf(err, "finding accessToken")
	}
	return &entities.AccessToken{
		Id:       repos.ID(accTokenOut.Id),
		UserId:   repos.ID(accTokenOut.UserId),
		Email:    accTokenOut.Email,
		Provider: accTokenOut.Provider,
		Token:    &oauth2.Token{AccessToken: accTokenOut.OauthToken.AccessToken, TokenType: accTokenOut.OauthToken.TokenType, RefreshToken: accTokenOut.OauthToken.RefreshToken, Expiry: time.UnixMilli(accTokenOut.OauthToken.Expiry)},
		Data:     map[string]any{},
	}, err
}

func (d *Impl) GithubInstallationToken(ctx context.Context, repoUrl string) (string, error) {
	return d.github.GetInstallationToken(ctx, repoUrl)
}

func (d *Impl) GithubListInstallations(ctx context.Context, userId repos.ID, pagination *types.Pagination) ([]*entities.GithubInstallation, error) {
	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	if err != nil {
		return nil, err
	}
	i, err := d.github.ListInstallations(ctx, token, pagination)
	if err != nil {
		return nil, err
	}

	res := make([]*entities.GithubInstallation, len(i))
	for i2, i3 := range i {
		res[i2] = &entities.GithubInstallation{
			GInst: entities.GInst{
				ID:              i3.ID,
				AppID:           i3.AppID,
				TargetType:      i3.TargetType,
				TargetID:        i3.TargetID,
				NodeID:          i3.NodeID,
				RepositoriesURL: i3.RepositoriesURL,
				Account: entities.GithubUserAccount{
					ID:        i3.Account.ID,
					NodeID:    i3.Account.NodeID,
					Type:      i3.Account.Type,
					Login:     i3.Account.Login,
					AvatarURL: i3.Account.AvatarURL,
				},
			},
		}
	}

	return res, nil
}

func (d *Impl) GithubListRepos(ctx context.Context, userId repos.ID, installationId int64, pagination *types.Pagination) (*entities.GithubListRepository, error) {
	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	if err != nil {
		return nil, err
	}

	i, err := d.github.ListRepos(ctx, token, installationId, pagination)

	repositories := make([]*entities.GithubRepository, len(i.Repositories))

	for i2, r := range i.Repositories {

		repositories[i2] = &entities.GithubRepository{
			ID:                r.ID,
			NodeID:            r.NodeID,
			Name:              r.Name,
			FullName:          r.FullName,
			Description:       r.Description,
			DefaultBranch:     r.DefaultBranch,
			MasterBranch:      r.MasterBranch,
			CreatedAt:         r.CreatedAt.Time,
			PushedAt:          r.PushedAt.Time,
			UpdatedAt:         r.UpdatedAt.Time,
			HTMLURL:           r.HTMLURL,
			CloneURL:          r.CloneURL,
			GitURL:            r.GitURL,
			MirrorURL:         r.MirrorURL,
			Language:          r.Language,
			Size:              r.Size,
			Permissions:       r.Permissions,
			Archived:          r.Archived,
			Disabled:          r.Disabled,
			Private:           r.Private,
			GitignoreTemplate: r.GitignoreTemplate,
			TeamID:            r.TeamID,
			Visibility:        r.Visibility,
			URL:               r.URL,
		}

	}

	return &entities.GithubListRepository{
		TotalCount:   i.TotalCount,
		Repositories: repositories,
	}, err
}

func (d *Impl) GithubSearchRepos(ctx context.Context, userId repos.ID, q, org string, pagination *types.Pagination) (*entities.GithubSearchRepository, error) {

	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	if err != nil {
		return nil, err
	}

	i, err := d.github.SearchRepos(ctx, token, q, org, pagination)

	repositories := make([]*entities.GithubRepository, len(i.Repositories))

	for i2, r := range i.Repositories {

		repositories[i2] = &entities.GithubRepository{
			ID:                r.ID,
			NodeID:            r.NodeID,
			Name:              r.Name,
			FullName:          r.FullName,
			Description:       r.Description,
			DefaultBranch:     r.DefaultBranch,
			MasterBranch:      r.MasterBranch,
			CreatedAt:         r.CreatedAt.Time,
			PushedAt:          r.PushedAt.Time,
			UpdatedAt:         r.UpdatedAt.Time,
			HTMLURL:           r.HTMLURL,
			CloneURL:          r.CloneURL,
			GitURL:            r.GitURL,
			MirrorURL:         r.MirrorURL,
			Language:          r.Language,
			Size:              r.Size,
			Permissions:       r.Permissions,
			Archived:          r.Archived,
			Disabled:          r.Disabled,
			Private:           r.Private,
			GitignoreTemplate: r.GitignoreTemplate,
			TeamID:            r.TeamID,
			Visibility:        r.Visibility,
			URL:               r.URL,
		}
	}

	return &entities.GithubSearchRepository{
		Total:             i.Total,
		Repositories:      repositories,
		IncompleteResults: i.IncompleteResults,
	}, nil
}

func getString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (d *Impl) GithubListBranches(ctx context.Context, userId repos.ID, repoUrl string, pagination *types.Pagination) ([]*entities.GitBranch, error) {
	token, err := d.getAccessTokenByUserId(ctx, "github", userId)
	if err != nil {
		return nil, err
	}

	i, err := d.github.ListBranches(ctx, token, repoUrl, pagination)

	branches := make([]*entities.GitBranch, len(i))

	for i2, b := range i {
		branches[i2] = &entities.GitBranch{
			Name:      *b.Name,
			Protected: *b.Protected,
		}
	}

	return branches, err
}

func (d *Impl) GithubAddWebhook(ctx context.Context, userId repos.ID, repoUrl string) (repos.ID, error) {
	panic("not implemented")
}
