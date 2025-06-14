package app

import (
	"bytes"
	"context"
	"fmt"

	"github.com/kloudlite/api/apps/comms/internal/domain"
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

	eTemplattes *domain.EmailTemplates
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
		"Link":        fmt.Sprintf("%v?token=%v", r.ev.BaseUrl, input.InvitationToken),
	}

	if err := r.eTemplattes.AccountInviteEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain-text template")
	}

	html := new(bytes.Buffer)
	if err := r.eTemplattes.AccountInviteEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.eTemplattes.AccountInviteEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
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
		"Link": fmt.Sprintf("%v?token=%v", r.ev.BaseUrl, input.ResetToken),
	}

	if err := r.eTemplattes.ResetPasswordEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := r.eTemplattes.ResetPasswordEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.eTemplattes.ResetPasswordEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
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

	if err := r.eTemplattes.WelcomeEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := r.eTemplattes.WelcomeEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.eTemplattes.WelcomeEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
		return nil, errors.NewE(err)
	}

	return &comms.Void{}, nil
}

func (r *commsSvc) SendWaitingEmail(ctx context.Context, input *comms.WelcomeEmailInput) (*comms.Void, error) {
	plainText := new(bytes.Buffer)
	args := map[string]any{
		"Name": func() string {
			if input.Name != "" {
				return input.Name
			}
			return "there"
		}(),
	}

	if err := r.eTemplattes.WaitingEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := r.eTemplattes.WaitingEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.eTemplattes.WaitingEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
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
		"Link": fmt.Sprintf("%v?token=%v", r.ev.BaseUrl, input.VerificationToken),
	}

	if err := r.eTemplattes.UserVerificationEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := r.eTemplattes.UserVerificationEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.eTemplattes.UserVerificationEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
		return nil, errors.NewE(err)
	}
	return &comms.Void{}, nil
}

func (r *commsSvc) SendContactUsEmail(ctx context.Context, input *comms.SendContactUsEmailInput) (*comms.Void, error) {
	plainText := new(bytes.Buffer)
	args := map[string]any{
		"Name": func() string {
			if input.Name != "" {
				return input.Name
			}
			return "there"
		}(),
		"CompanyName": input.CompanyName,
		"Country":     input.Country,
		"Message":     input.Message,
	}

	if err := r.eTemplattes.ContactUsEmail.PlainText.Execute(plainText, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := r.eTemplattes.ContactUsEmail.Html.Execute(html, args); err != nil {
		return nil, errors.NewEf(err, "failed to execute html template")
	}

	if err := r.sendSupportEmail(ctx, r.eTemplattes.ContactUsEmail.Subject, input.Email, input.Name, plainText.String(), html.String()); err != nil {
		return nil, errors.NewE(err)
	}
	return &comms.Void{}, nil
}

func newCommsSvc(mailer mail.Mailer, ev *env.Env, et *domain.EmailTemplates) comms.CommsServer {
	return &commsSvc{
		mailer:       mailer,
		supportEmail: ev.SupportEmail,
		ev:           ev,
		eTemplattes:  et,
	}
}
