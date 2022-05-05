package mail

import (
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.uber.org/fx"
)

type SendgridMailer struct {
	client *sendgrid.Client
}

func (s SendgridMailer) SendEmail(
	from string,
	fromName string,
	subject string,
	to string,
	toName string,
	plaintextContent string,
	htmlContent string,
) error {
	_, err := s.client.Send(mail.NewSingleEmail(mail.NewEmail(fromName, from), subject, mail.NewEmail(toName, to), plaintextContent, htmlContent))
	fmt.Println(
		from,
		fromName,
		subject,
		to,
		toName,
		plaintextContent,
		htmlContent, err)
	if err != nil {
		return err
	}
	return nil
}

func NewSendgridMailer(apiKey string) Mailer {
	return SendgridMailer{client: sendgrid.NewSendClient(apiKey)}
}

type SendgridMailerOptions interface {
	GetSendGridApiKey() string
}

func NewSendGridMailerFx[T SendgridMailerOptions]() fx.Option {
	return fx.Module("mailer", fx.Provide(func(o T) Mailer {
		fmt.Println("NewSendGridMailerFx", o.GetSendGridApiKey())
		return NewSendgridMailer(o.GetSendGridApiKey())
	}))
}
