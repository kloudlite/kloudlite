package app

import (
	"context"
	"fmt"
	"kloudlite.io/apps/auth/internal/domain"
)

type messengerI struct {
}

func (m *messengerI) SendVerificationEmail(ctx context.Context, invitationId string, user *domain.User) error {
	fmt.Println("VERIFICATION", invitationId, user)
	return nil
}

func (m *messengerI) SendWelcomeEmail(ctx context.Context, invitationId string, user *domain.User) error {
	fmt.Println("WELCOME", invitationId, user)
	return nil
}

func (m *messengerI) SendResetPasswordEmail(ctx context.Context, invitationId string, user *domain.User) error {
	fmt.Println("RESETPASSWORD", invitationId, user)
	return nil
}

func fxMessenger() domain.Messenger {
	return &messengerI{}
}
