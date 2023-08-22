package mail

import "context"

type Email struct {
	FromEmailAddress string
	FromName         string
	Subject          string
	ToEmailAddress   string
	ToName           string
	PlainText        string
	HtmlText         string
}

type Mailer interface {
	SendEmail(ctx context.Context, email Email) error
}
