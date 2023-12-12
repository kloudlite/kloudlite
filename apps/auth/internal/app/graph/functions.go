package graph

import (
	"github.com/kloudlite/api/apps/auth/internal/app/graph/model"
	"github.com/kloudlite/api/apps/auth/internal/domain"
	"github.com/kloudlite/api/common"
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

func sessionModelFromAuthSession(session *common.AuthSession) *model.Session {
	return &model.Session{
		ID:           session.Id,
		UserID:       session.UserId,
		UserEmail:    session.UserEmail,
		LoginMethod:  session.LoginMethod,
		UserVerified: session.UserVerified,
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
