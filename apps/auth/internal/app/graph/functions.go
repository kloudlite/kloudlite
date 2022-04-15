package graph

import (
	"kloudlite.io/apps/auth/internal/app/graph/model"
	"kloudlite.io/apps/auth/internal/domain"
)

func mapFromProviderDetail(detail *domain.ProviderDetail) map[string]any {
	if detail == nil {
		return nil
	}
	return map[string]any{
		"token_id": detail.TokenId,
		"avatar":   detail.Avatar,
	}
}

func userModelFromEntity(userEntity *domain.User) *model.User {

	return &model.User{
		ID:             userEntity.Id,
		Name:           userEntity.Name,
		Email:          userEntity.Email,
		Avatar:         userEntity.Avatar,
		Invite:         string(userEntity.InvitationStatus),
		Verified:       userEntity.Verified,
		Metadata:       userEntity.Metadata,
		Joined:         userEntity.Joined.String(),
		ProviderGitlab: mapFromProviderDetail(userEntity.ProviderGitlab),
		ProviderGithub: mapFromProviderDetail(userEntity.ProviderGithub),
		ProviderGoogle: mapFromProviderDetail(userEntity.ProviderGoogle),
	}
}
