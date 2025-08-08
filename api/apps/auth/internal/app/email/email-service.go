package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"text/template"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/mail"
)

//go:embed templates
var templatesFS embed.FS

type EmailTemplate struct {
	Subject   string
	Html      *template.Template
	PlainText *template.Template
}

type EmailTemplates struct {
	UserVerification *EmailTemplate
	ResetPassword    *EmailTemplate
	Welcome          *EmailTemplate
	AccountInvite    *EmailTemplate
	PlatformInvite   *EmailTemplate
	Alert            *EmailTemplate
	ContactUs        *EmailTemplate
}

type EmailService struct {
	mailer       mail.Mailer
	templates    *EmailTemplates
	supportEmail string
	baseURL      string
}

func NewEmailService(mailer mail.Mailer, supportEmail, baseURL string) (*EmailService, error) {
	templates, err := loadTemplates()
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &EmailService{
		mailer:       mailer,
		templates:    templates,
		supportEmail: supportEmail,
		baseURL:      baseURL,
	}, nil
}

func loadTemplates() (*EmailTemplates, error) {
	userVerification, err := parseTemplate("user-verification", "Verify your email address")
	if err != nil {
		return nil, errors.NewE(err)
	}

	resetPassword, err := parseTemplate("reset-password", "Reset your password")
	if err != nil {
		return nil, errors.NewE(err)
	}

	welcome, err := parseTemplate("welcome", "Welcome to Kloudlite!")
	if err != nil {
		return nil, errors.NewE(err)
	}

	accountInvite, err := parseTemplate("account-invite", "You're Invited to Join Kloudlite")
	if err != nil {
		return nil, errors.NewE(err)
	}

	alert, err := parseTemplate("alert", "Kloudlite Alert")
	if err != nil {
		return nil, errors.NewE(err)
	}

	contactUs, err := parseTemplate("contact-us", "Contact Us Inquiry")
	if err != nil {
		return nil, errors.NewE(err)
	}

	platformInvite, err := parseTemplate("platform-invite", "Invitation to Join Kloudlite Platform Management")
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &EmailTemplates{
		UserVerification: userVerification,
		ResetPassword:    resetPassword,
		Welcome:          welcome,
		AccountInvite:    accountInvite,
		PlatformInvite:   platformInvite,
		Alert:            alert,
		ContactUs:        contactUs,
	}, nil
}

func parseTemplate(name, subject string) (*EmailTemplate, error) {
	txtFile, err := templatesFS.ReadFile(fmt.Sprintf("templates/%s/email.txt", name))
	if err != nil {
		return nil, errors.NewE(err)
	}

	txt, err := template.New("email-text").Parse(string(txtFile))
	if err != nil {
		return nil, errors.NewE(err)
	}

	htmlFile, err := templatesFS.ReadFile(fmt.Sprintf("templates/%s/email.html", name))
	if err != nil {
		return nil, errors.NewE(err)
	}

	html, err := template.New(name).Parse(string(htmlFile))
	if err != nil {
		return nil, errors.NewE(err)
	}

	return &EmailTemplate{
		Subject:   subject,
		Html:      html,
		PlainText: txt,
	}, nil
}

func (s *EmailService) sendEmail(ctx context.Context, toEmail, toName, subject, plainText, htmlContent string) error {
	return s.mailer.SendEmail(ctx, mail.Email{
		FromEmailAddress: s.supportEmail,
		FromName:         "Kloudlite Support",
		Subject:          subject,
		ToEmailAddress:   toEmail,
		ToName:           toName,
		PlainText:        plainText,
		HtmlText:         htmlContent,
	})
}

func (s *EmailService) SendVerificationEmail(ctx context.Context, email, name, verificationToken string) error {
	args := map[string]any{
		"Name": getNameOrDefault(name),
		"Link": fmt.Sprintf("%s/auth/verify-email?token=%s", s.baseURL, verificationToken),
	}

	plainText := new(bytes.Buffer)
	if err := s.templates.UserVerification.PlainText.Execute(plainText, args); err != nil {
		return errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := s.templates.UserVerification.Html.Execute(html, args); err != nil {
		return errors.NewEf(err, "failed to execute html template")
	}

	return s.sendEmail(ctx, email, name, s.templates.UserVerification.Subject, plainText.String(), html.String())
}

func (s *EmailService) SendPasswordResetEmail(ctx context.Context, email, name, resetToken string) error {
	args := map[string]any{
		"Name": getNameOrDefault(name),
		"Link": fmt.Sprintf("%s/auth/reset-password?token=%s", s.baseURL, resetToken),
	}

	plainText := new(bytes.Buffer)
	if err := s.templates.ResetPassword.PlainText.Execute(plainText, args); err != nil {
		return errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := s.templates.ResetPassword.Html.Execute(html, args); err != nil {
		return errors.NewEf(err, "failed to execute html template")
	}

	return s.sendEmail(ctx, email, name, s.templates.ResetPassword.Subject, plainText.String(), html.String())
}

func (s *EmailService) SendWelcomeEmail(ctx context.Context, email, name string) error {
	args := map[string]any{
		"Name": getNameOrDefault(name),
		"Link": s.baseURL,
	}

	plainText := new(bytes.Buffer)
	if err := s.templates.Welcome.PlainText.Execute(plainText, args); err != nil {
		return errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := s.templates.Welcome.Html.Execute(html, args); err != nil {
		return errors.NewEf(err, "failed to execute html template")
	}

	return s.sendEmail(ctx, email, name, s.templates.Welcome.Subject, plainText.String(), html.String())
}

func (s *EmailService) SendAccountInviteEmail(ctx context.Context, email, name, invitedBy, accountName, inviteLink string) error {
	args := map[string]any{
		"Name":        getNameOrDefault(name),
		"InvitedBy":   invitedBy,
		"AccountName": accountName,
		"Link":        inviteLink,
	}

	plainText := new(bytes.Buffer)
	if err := s.templates.AccountInvite.PlainText.Execute(plainText, args); err != nil {
		return errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := s.templates.AccountInvite.Html.Execute(html, args); err != nil {
		return errors.NewEf(err, "failed to execute html template")
	}

	return s.sendEmail(ctx, email, name, s.templates.AccountInvite.Subject, plainText.String(), html.String())
}

func (s *EmailService) SendAlertEmail(ctx context.Context, email, alertTitle, alertMessage string, alertData map[string]any) error {
	args := map[string]any{
		"Title":   alertTitle,
		"Message": alertMessage,
		"Data":    alertData,
	}

	plainText := new(bytes.Buffer)
	if err := s.templates.Alert.PlainText.Execute(plainText, args); err != nil {
		return errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := s.templates.Alert.Html.Execute(html, args); err != nil {
		return errors.NewEf(err, "failed to execute html template")
	}

	return s.sendEmail(ctx, email, "", s.templates.Alert.Subject, plainText.String(), html.String())
}

func (s *EmailService) SendContactUsEmail(ctx context.Context, customerEmail, customerName, subject, message string) error {
	args := map[string]any{
		"CustomerEmail": customerEmail,
		"CustomerName":  customerName,
		"Subject":       subject,
		"Message":       message,
	}

	plainText := new(bytes.Buffer)
	if err := s.templates.ContactUs.PlainText.Execute(plainText, args); err != nil {
		return errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := s.templates.ContactUs.Html.Execute(html, args); err != nil {
		return errors.NewEf(err, "failed to execute html template")
	}

	// Send to support email
	return s.sendEmail(ctx, s.supportEmail, "Support Team", fmt.Sprintf("Contact Us: %s", subject), plainText.String(), html.String())
}

func (s *EmailService) SendPlatformInviteEmail(ctx context.Context, email, name, invitedBy, role, inviteLink string) error {
	args := map[string]any{
		"Name":       getNameOrDefault(name),
		"InvitedBy":  invitedBy,
		"Role":       role,
		"Link":       inviteLink,
	}

	plainText := new(bytes.Buffer)
	if err := s.templates.PlatformInvite.PlainText.Execute(plainText, args); err != nil {
		return errors.NewEf(err, "failed to execute plain text template")
	}

	html := new(bytes.Buffer)
	if err := s.templates.PlatformInvite.Html.Execute(html, args); err != nil {
		return errors.NewEf(err, "failed to execute html template")
	}

	return s.sendEmail(ctx, email, name, s.templates.PlatformInvite.Subject, plainText.String(), html.String())
}

func getNameOrDefault(name string) string {
	if name != "" {
		return name
	}
	return "there"
}