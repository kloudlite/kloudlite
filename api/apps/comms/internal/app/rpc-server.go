package app

import (
	"context"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/comms"
	"kloudlite.io/pkg/mail"
)

type rpcImpl struct {
	comms.UnimplementedCommsServer
	mailer            mail.Mailer
	supportEmail      string
	emailLinksBaseUrl string
}

func (r *rpcImpl) sendSupportEmail(
	subject string,
	email string,
	name string,
	plainText string,
	htmlContent string,
) error {
	err := r.mailer.SendEmail(
		r.supportEmail,
		"Support",
		subject,
		email,
		name,
		plainText,
		htmlContent,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *rpcImpl) SendAccountMemberInviteEmail(ctx context.Context, input *comms.AccountMemberInviteEmailInput) (*comms.Void, error) {
	subject, plainText, htmlContent, err := constructAccountInvitationEmail(
		input.Name, input.AccountName, input.InvitationToken, r.emailLinksBaseUrl,
	)
	if err != nil {
		return nil, err
	}
	err = r.sendSupportEmail(subject, input.Email, input.Name, plainText, htmlContent)
	if err != nil {
		return nil, err
	}
	return &comms.Void{}, nil
}

func (r *rpcImpl) SendProjectMemberInviteEmail(ctx context.Context, input *comms.ProjectMemberInviteEmailInput) (*comms.Void, error) {
	subject, plainText, htmlContent, err := constructProjectInvitationEmail(
		input.Name, input.ProjectName, input.InvitationToken, r.emailLinksBaseUrl,
	)
	if err != nil {
		return nil, err
	}
	err = r.sendSupportEmail(subject, input.Email, input.Name, plainText, htmlContent)
	if err != nil {
		return nil, err
	}
	return &comms.Void{}, nil
}

func (r *rpcImpl) SendPasswordResetEmail(_ context.Context, input *comms.PasswordResetEmailInput) (*comms.Void, error) {
	subject, plainText, htmlContent, err := constructResetPasswordEmail(input.Name, input.ResetToken, r.emailLinksBaseUrl)
	if err != nil {
		return nil, err
	}
	err = r.sendSupportEmail(subject, input.Email, input.Name, plainText, htmlContent)
	if err != nil {
		return nil, err
	}
	return &comms.Void{}, nil
}

func (r *rpcImpl) SendWelcomeEmail(ctx context.Context, input *comms.WelcomeEmailInput) (*comms.Void, error) {
	subject, plainText, htmlContent, err := constructWelcomeEmail(input.Name)
	if err != nil {
		return nil, err
	}
	err = r.sendSupportEmail(subject, input.Email, input.Name, plainText, htmlContent)
	if err != nil {
		return nil, err
	}
	return &comms.Void{}, nil
}

func (r *rpcImpl) SendVerificationEmail(_ context.Context, input *comms.VerificationEmailInput) (*comms.Void, error) {
	subject, plainText, htmlContent, err := constructVerificationEmail(
		input.Name,
		input.VerificationToken,
		r.emailLinksBaseUrl,
	)
	if err != nil {
		return nil, err
	}
	err = r.sendSupportEmail(subject, input.Email, input.Name, plainText, htmlContent)
	if err != nil {
		return nil, err
	}
	return &comms.Void{}, nil
}

func fxRPCServer(mailer mail.Mailer, env *Env) comms.CommsServer {
	return &rpcImpl{
		mailer:            mailer,
		supportEmail:      env.SupportEmail,
		emailLinksBaseUrl: env.EmailsBaseUrl,
	}
}
