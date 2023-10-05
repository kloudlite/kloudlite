package domain

import (
	"context"

	"kloudlite.io/apps/container-registry/internal/domain/entities"
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

func (d *Impl) GitlabAddWebhook(ctx context.Context, userId repos.ID, repoId string, pipelineId repos.ID) (repos.ID, error) {
	panic("not implemented")
}
func (d *Impl) GitlabPullToken(ctx context.Context, tokenId repos.ID) (string, error) {
	panic("not implemented")
}

func (d *Impl) ParseGitlabHook(eventType string, hookBody []byte) (*GitWebhookPayload, error) {
	panic("not implemented")
}
