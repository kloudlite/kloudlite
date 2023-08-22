package mail

import (
	"context"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendgridMailer struct {
	client *sendgrid.Client
}

func (s SendgridMailer) SendEmail(ctx context.Context, email Email) error {
	if _, err := s.client.SendWithContext(
		ctx,
		mail.NewSingleEmail(
			mail.NewEmail(email.FromName, email.FromEmailAddress),
			email.Subject,
			mail.NewEmail(email.ToName, email.ToEmailAddress),
			email.PlainText, email.HtmlText,
		),
	); err != nil {
		return err
	}
	return nil
}

func NewSendgridMailer(apiKey string) Mailer {
	return SendgridMailer{client: sendgrid.NewSendClient(apiKey)}
}
