package app

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"github.com/kloudlite/api/pkg/mail"
)

type commsSvc struct {
	comms.UnimplementedCommsServer
	mailer mail.Mailer

	supportEmail string

	ev *env.Env

	accountInviteEmail    AccountInviteEmail
	projectInviteEmail    ProjectInviteEmail
	resetPasswordEmail    RestPasswordEmail
	userVerificationEmail UserVerificationEmail
	welcomeEmail          WelcomeEmail
}

func (r *commsSvc) sendSupportEmail(ctx context.Context, subject string, toEmail string, toName string, plainText string, htmlContent string) error {
	err := r.mailer.SendEmail(ctx, mail.Email{
		FromEmailAddress: r.supportEmail,
		FromName:         "Kloudlite Support",
		Subject:          subject,
		ToEmailAddress:   toEmail,
		ToName:           toName,
		PlainText:        plainText,
		HtmlText:         htmlContent,
	})
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (r *commsSvc) SendAccountMemberInviteEmail(ctx context.Context, input *comms.AccountMemberInviteEmailInput) (*comms.Void, error) {
	plainText := new(bytes.Buffer)
	args := map[string]any{
		"Name": func() string {
			if input.Name != "" {
				return input.Name
			}
			return "there"
		}(),
		"InvitedBy":   input.InvitedBy,
		"AccountName": input.AccountName,
		"Link":        fmt.Sprintf("%v?token=%v", r.ev.AccountsWebInviteUrl, input.InvitationToken),
	}

	if err := r.accountInviteEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain-text template")
	}

	html := new(bytes.Buffer)
	if err := r.accountInviteEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.accountInviteEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
		return nil, errors.NewE(err)
	}
	return &comms.Void{}, nil
}

func (r *commsSvc) SendProjectMemberInviteEmail(ctx context.Context, input *comms.ProjectMemberInviteEmailInput) (*comms.Void, error) {
	plainText := new(bytes.Buffer)
	args := map[string]any{
		"Name": func() string {
			if input.Name != "" {
				return input.Name
			}
			return "there"
		}(),
		"InvitedBy":   input.InvitedBy,
		"AccountName": input.ProjectName,
		"Link":        fmt.Sprintf("%v?token=%v", r.ev.ProjectsWebInviteUrl, input.InvitationToken),
	}

	if err := r.projectInviteEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain-text template")
	}

	html := new(bytes.Buffer)
	if err := r.projectInviteEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.projectInviteEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
		return nil, errors.NewE(err)
	}

	return &comms.Void{}, nil
}

func (r *commsSvc) SendPasswordResetEmail(ctx context.Context, input *comms.PasswordResetEmailInput) (*comms.Void, error) {
	plainText := new(bytes.Buffer)
	args := map[string]any{
		"Name": func() string {
			if input.Name != "" {
				return input.Name
			}
			return "there"
		}(),
		"Link": fmt.Sprintf("%v?token=%v", r.ev.ResetPasswordWebUrl, input.ResetToken),
	}

	if err := r.resetPasswordEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := r.resetPasswordEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.resetPasswordEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
		return nil, errors.NewE(err)
	}

	return &comms.Void{}, nil
}

func (r *commsSvc) SendWelcomeEmail(ctx context.Context, input *comms.WelcomeEmailInput) (*comms.Void, error) {
	plainText := new(bytes.Buffer)
	args := map[string]any{
		"Name": func() string {
			if input.Name != "" {
				return input.Name
			}
			return "there"
		}(),
	}

	if err := r.welcomeEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := r.welcomeEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.welcomeEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
		return nil, errors.NewE(err)
	}

	return &comms.Void{}, nil
}

func (r *commsSvc) SendVerificationEmail(ctx context.Context, input *comms.VerificationEmailInput) (*comms.Void, error) {
	plainText := new(bytes.Buffer)
	args := map[string]any{
		"Name": func() string {
			if input.Name != "" {
				return input.Name
			}
			return "there"
		}(),
		"Link": fmt.Sprintf("%v?token=%v", r.ev.VerifyEmailWebUrl, input.VerificationToken),
	}

	if err := r.userVerificationEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := r.userVerificationEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.userVerificationEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
		return nil, errors.NewE(err)
	}
	return &comms.Void{}, nil
}
func newCommsSvc(mailer mail.Mailer, ev *env.Env, ai AccountInviteEmail, pi ProjectInviteEmail, rp RestPasswordEmail, uv UserVerificationEmail, wl WelcomeEmail) comms.CommsServer {
	return &commsSvc{
		mailer:                mailer,
		supportEmail:          ev.SupportEmail,
		ev:                    ev,
		accountInviteEmail:    ai,
		projectInviteEmail:    pi,
		resetPasswordEmail:    rp,
		userVerificationEmail: uv,
		welcomeEmail:          wl,
	}
}
