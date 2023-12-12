package app

import (
	"embed"
	"fmt"
	"text/template"

	"github.com/kloudlite/api/pkg/grpc"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/comms"
	"go.uber.org/fx"
)

type CommsGrpcServer grpc.Server

type EmailTemplatesDir struct {
	embed.FS
}

type EmailTemplate struct {
	Subject   string
	Html      *template.Template
	PlainText *template.Template
}

type AccountInviteEmail *EmailTemplate
type ProjectInviteEmail *EmailTemplate
type RestPasswordEmail *EmailTemplate
type UserVerificationEmail *EmailTemplate
type WelcomeEmail *EmailTemplate

func parseTemplate(et EmailTemplatesDir, templateName string, subject string) (*EmailTemplate, error) {
	txtFile, err := et.ReadFile(fmt.Sprintf("email-templates/%v/email.txt", templateName))
	if err != nil {
		return nil, err
	}
	txt, err := template.New("email-text").Parse(string(txtFile))
	if err != nil {
		return nil, err
	}

	htmlFile, err := et.ReadFile(fmt.Sprintf("email-templates/%v/email.html", templateName))
	if err != nil {
		return nil, err
	}
	html, err := template.New(templateName).Parse(string(htmlFile))
	if err != nil {
		return nil, err
	}

	return &EmailTemplate{
		Subject:   subject,
		Html:      html,
		PlainText: txt,
	}, nil
}

var Module = fx.Module("app",
	fx.Provide(func(et EmailTemplatesDir) (AccountInviteEmail, error) {
		return parseTemplate(et, "account-invite", "[Kloudlite] Account Invite")
	}),
	fx.Provide(func(et EmailTemplatesDir) (ProjectInviteEmail, error) {
		return parseTemplate(et, "project-invite", "[Kloudlite] Project Invite")
	}),
	fx.Provide(func(et EmailTemplatesDir) (RestPasswordEmail, error) {
		return parseTemplate(et, "reset-password", "[Kloudlite] Reset Password")
	}),
	fx.Provide(func(et EmailTemplatesDir) (UserVerificationEmail, error) {
		return parseTemplate(et, "user-verification", "[Kloudlite] Verify Email")
	}),
	fx.Provide(func(et EmailTemplatesDir) (WelcomeEmail, error) {
		return parseTemplate(et, "welcome", "[Kloudlite] Welcome to Kloudlite")
	}),

	fx.Provide(newCommsSvc),

	fx.Invoke(func(server CommsGrpcServer, commsServer comms.CommsServer) {
		comms.RegisterCommsServer(server, commsServer)
	}),
)
