package domain

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/constants"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
	"kloudlite.io/pkg/types"
)

func (d *Impl) GitlabListGroups(ctx context.Context, userId repos.ID, query *string, pagination *types.Pagination) ([]*entities.GitlabGroup, error) {
	token, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}

	return d.gitlab.ListGroups(ctx, token, query, pagination)
}

func (d *Impl) GitlabListRepos(ctx context.Context, userId repos.ID, gid string, query *string, pagination *types.Pagination) ([]*entities.GitlabProject, error) {
	token, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}
	repos, err := d.gitlab.ListRepos(ctx, token, gid, query, pagination)
	if err != nil {
		return nil, err
	}

	res := make([]*entities.GitlabProject, len(repos))

	for i, r := range repos {
		res[i] = &entities.GitlabProject{
			ID:                r.ID,
			Description:       r.Description,
			DefaultBranch:     r.DefaultBranch,
			Public:            r.Public,
			SSHURLToRepo:      r.SSHURLToRepo,
			HTTPURLToRepo:     r.HTTPURLToRepo,
			WebURL:            r.WebURL,
			TagList:           r.TagList,
			Topics:            r.Topics,
			Name:              r.Name,
			NameWithNamespace: r.NameWithNamespace,
			Path:              r.Path,
			PathWithNamespace: r.PathWithNamespace,
			CreatedAt:         r.CreatedAt,
			LastActivityAt:    r.LastActivityAt,
			CreatorID:         r.CreatorID,
			EmptyRepo:         r.EmptyRepo,
			Archived:          r.Archived,
			AvatarURL:         r.AvatarURL,
		}
	}

	return res, nil
}

func (d *Impl) GitlabListBranches(ctx context.Context, userId repos.ID, repoId string, query *string, pagination *types.Pagination) ([]*entities.GitBranch, error) {

	token, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}

	branches, err := d.gitlab.ListBranches(ctx, token, repoId, query, pagination)

	if err != nil {
		return nil, err
	}

	res := make([]*entities.GitBranch, len(branches))

	for i, b := range branches {
		res[i] = &entities.GitBranch{
			Name:      b.Name,
			Protected: b.Protected,
		}
	}

	return res, nil
}

func (d *Impl) GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string) (*int, error) {

	// grHook, err := d.gitRepoHookRepo.FindOne(ctx, repos.Filter{"httpUrl": d.envs.GitlabWebhookUrl})
	// if err != nil {
	// 	return "", err
	// }

	token, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)
	if err != nil {
		return nil, err
	}

	// if grHook != nil {
	// 	exists, _ := d.gitlab.CheckWebhookExists(ctx, token, repoId, grHook.GitlabWebhookId)
	// 	if exists {
	// 		return grHook.Id, nil
	// 	}
	// 	if err := d.gitRepoHookRepo.DeleteById(ctx, grHook.Id); err != nil {
	// 		return "", err
	// 	}
	// }

	webhookId, err := d.gitlab.AddWebhook(ctx, token, repoId)
	if err != nil {
		return nil, err
	}

	// grHook, err = d.gitRepoHookRepo.Create(
	// 	ctx, &GitRepositoryHook{
	// 		HttpUrl:         d.gitlabWebhookUrl,
	// 		GitProvider:     GitlabProvider,
	// 		GitlabWebhookId: webhookId,
	// 	},
	// )

	return webhookId, nil
}

func (d *Impl) gitlabPullToken(ctx context.Context, accTokenReq *auth.GetAccessTokenRequest) (string, error) {
	accessToken, err := d.authClient.GetAccessToken(ctx, accTokenReq)
	if err != nil || accessToken == nil {
		return "", errors.NewEf(err, "could not get gitlab access token")
	}

	accToken := entities.AccessToken{
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

func (d *Impl) GitlabPullToken(ctx context.Context, userId repos.ID) (string, error) {

	at, err := d.getAccessTokenByUserId(ctx, "gitlab", userId)

	if err != nil {
		return "", err
	}

	return d.gitlabPullToken(ctx, &auth.GetAccessTokenRequest{TokenId: string(at.Token.AccessToken)})
}

func (d *Impl) ParseGitlabHook(eventType string, hookBody []byte) (*GitWebhookPayload, error) {
	hook, err := gitlab.ParseWebhook(gitlab.EventType(eventType), hookBody)
	if err != nil {
		return nil, errors.NewEf(err, "could not parse webhook body for eventType=%s", eventType)
	}
	switch h := hook.(type) {
	case *gitlab.PushEvent:
		{
			payload := &GitWebhookPayload{
				GitProvider: constants.ProviderGitlab,
				RepoUrl:     h.Repository.GitHTTPURL,
				GitBranch:   getBranchFromRef(h.Ref),
				CommitHash:  h.CheckoutSHA[:int(math.Min(10, float64(len(h.CheckoutSHA))))],
			}
			return payload, nil
		}
	default:
		return nil, &ErrEventNotSupported{err: fmt.Errorf("event type (%s) currently not supported", eventType)}
	}
}
