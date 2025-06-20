package grpc

import (
	"embed"
	"fmt"
	"text/template"

	"github.com/kloudlite/api/pkg/errors"
)

type EmailTemplatesDir struct {
	embed.FS
}

type EmailTemplate struct {
	Subject   string
	Html      *template.Template
	PlainText *template.Template
}

type EmailTemplates struct {
	AccountInviteEmail    *EmailTemplate
	ProjectInviteEmail    *EmailTemplate
	ResetPasswordEmail    *EmailTemplate
	UserVerificationEmail *EmailTemplate
	WelcomeEmail          *EmailTemplate
	WaitingEmail          *EmailTemplate
	AlertEmail            *EmailTemplate
	ContactUsEmail        *EmailTemplate
}

func parseMailTemplate(et EmailTemplatesDir, templateName string, subject string) (*EmailTemplate, error) {
	txtFile, err := et.ReadFile(fmt.Sprintf("email-templates/%v/email.txt", templateName))
	if err != nil {
		return nil, errors.NewE(err)
	}
	txt, err := template.New("email-text").Parse(string(txtFile))
	if err != nil {
		return nil, errors.NewE(err)
	}

	htmlFile, err := et.ReadFile(fmt.Sprintf("email-templates/%v/email.html", templateName))
	if err != nil {
		return nil, errors.NewE(err)
	}
	html, err := template.New(templateName).Parse(string(htmlFile))
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &EmailTemplate{
		Subject:   subject,
		Html:      html,
		PlainText: txt,
	}, nil
}

func GetEmailTemplates(et EmailTemplatesDir) (*EmailTemplates, error) {
	accountInvite, err := parseMailTemplate(et, "account-invite", "[Kloudlite] Account Invite")
	if err != nil {
		return nil, err
	}

	restPassword, err := parseMailTemplate(et, "reset-password", "[Kloudlite] Reset Password")
	if err != nil {
		return nil, err
	}

	userVerification, err := parseMailTemplate(et, "user-verification", "[Kloudlite] Verify Email")
	if err != nil {
		return nil, err
	}

	welcome, err := parseMailTemplate(et, "welcome", "[Kloudlite] Welcome to Kloudlite")
	if err != nil {
		return nil, err
	}

	contactUs, err := parseMailTemplate(et, "contact-us", "[Kloudlite] Contact Us")
	if err != nil {
		return nil, err
	}

	alert, err := parseMailTemplate(et, "alert", "[Kloudlite] Console Notification")
	if err != nil {
		return nil, err
	}

	return &EmailTemplates{
		AccountInviteEmail:    accountInvite,
		ResetPasswordEmail:    restPassword,
		UserVerificationEmail: userVerification,
		WelcomeEmail:          welcome,
		ContactUsEmail:        contactUs,
		AlertEmail:            alert,
	}, nil
}
