package graph

import (
	"github.com/kloudlite/api/apps/auth/internal/app/graph/model"
	"github.com/kloudlite/api/apps/auth/internal/entities"
)

func mapFromProviderDetail(detail *entities.ProviderDetail) map[string]any {
	if detail == nil {
		return nil
	}
	return map[string]any{
		"token_id": detail.TokenId,
		"avatar":   detail.Avatar,
	}
}

func userModelFromEntity(userEntity *entities.User) *model.User {
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

func inviteCodeModelFromEntity(inviteCodeEntity *entities.InviteCode) *model.InviteCode {
	return &model.InviteCode{
		ID:         inviteCodeEntity.Id,
		Name:       inviteCodeEntity.Name,
		InviteCode: inviteCodeEntity.InviteCode,
	}
}
