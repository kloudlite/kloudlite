package app

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"text/template"
)

type email struct {
	Subject   string `json:"subject,omitempty" yaml:"subject,omitempty"`
	PlainText string `json:"plain,omitempty" yaml:"plain,omitempty"`
	HTMLText  string `json:"html,omitempty" yaml:"html,omitempty"`
}

func loadEmailFromYaml(templateName string, params any) (*email, error) {
	file, err := ioutil.ReadFile(fmt.Sprintf("email-templates/%v/email.yaml", templateName))
	htmlFile, err := ioutil.ReadFile(fmt.Sprintf("email-templates/%v/email.html", templateName))
	if err != nil {
		return nil, err
	}
	parse, err := template.New("email-template").Parse(string(file))
	parseHtml, err := template.New("email-html").Parse(string(htmlFile))
	if err != nil {
		return nil, err
	}
	var emailBuffer bytes.Buffer
	var htmlBuffer bytes.Buffer
	err = parse.Execute(&emailBuffer, params)
	err = parseHtml.Execute(&htmlBuffer, params)
	if err != nil {
		return nil, err
	}
	var email email
	parsedEmail, err := ioutil.ReadAll(&emailBuffer)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(parsedEmail, &email)
	email.HTMLText = string(htmlBuffer.Bytes())
	if err != nil {
		return nil, err
	}
	return &email, nil
}

func constructVerificationEmail(name string, token string, emailLinksBaseUrl string) (subject string, plainText string, htmlContent string, err error) {
	email, err := loadEmailFromYaml("user-verification", struct {
		Name string
		Link string
	}{
		Name: name,
		Link: fmt.Sprintf("%v/verify-email/?token=%v", emailLinksBaseUrl, token),
	})
	if err != nil {
		return
	}
	subject = email.Subject
	plainText = email.PlainText
	htmlContent = email.HTMLText
	return
}

func constructResetPasswordEmail(name string, token string, baseUrl string) (subject string, plainText string, htmlContent string, err error) {
	email, err := loadEmailFromYaml("reset-password", struct {
		Name string
		Link string
	}{
		Name: name,
		Link: fmt.Sprintf("%v/reset-password/?token=%v", baseUrl, token),
	})
	fmt.Println(fmt.Sprintf("%v/reset-password/?token=%v", baseUrl, token))
	if err != nil {
		return
	}
	subject = email.Subject
	plainText = email.PlainText
	htmlContent = email.HTMLText
	return
}

func constructAccountInvitationEmail(name string, accountName string, invitationToken string, baseUrl string) (subject string, plainText string, htmlContent string, err error) {
	email, err := loadEmailFromYaml("account-invite", struct {
		Name string
		Link string
	}{
		Name: name,
		Link: fmt.Sprintf("%v/invite-callback/?token=%v", baseUrl, invitationToken),
	})
	fmt.Sprintf("%v/invite-callback/?token=%v", baseUrl, invitationToken)
	if err != nil {
		return
	}
	subject = email.Subject
	plainText = email.PlainText
	htmlContent = email.HTMLText
	return
}
