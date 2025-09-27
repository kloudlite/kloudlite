package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/kloudlite/api/pkg/errors"
)

type MailtrapMailer struct {
	apiToken string
	apiURL   string
	from     string
}

type mailtrapRecipient struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type mailtrapRequest struct {
	To      []mailtrapRecipient `json:"to"`
	From    mailtrapRecipient   `json:"from"`
	Subject string              `json:"subject"`
	Text    string              `json:"text"`
	HTML    string              `json:"html"`
}

func (m MailtrapMailer) SendEmail(ctx context.Context, email Email) error {
	// Prepare the request body
	reqBody := mailtrapRequest{
		To: []mailtrapRecipient{
			{
				Email: email.ToEmailAddress,
				Name:  email.ToName,
			},
		},
		From: mailtrapRecipient{
			Email: email.FromEmailAddress,
			Name:  email.FromName,
		},
		Subject: email.Subject,
		Text:    email.PlainText,
		HTML:    email.HtmlText,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return errors.NewEf(err, "failed to marshal email request")
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", m.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.NewEf(err, "failed to create request")
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Api-Token", m.apiToken)
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.NewEf(err, "failed to send email via Mailtrap")
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return errors.Newf("mailtrap error: status=%d, response=%v", resp.StatusCode, errorResp)
		}
		return errors.Newf("mailtrap error: status=%d", resp.StatusCode)
	}

	return nil
}

func NewMailtrapMailer(apiToken string, fromEmail string) Mailer {
	return MailtrapMailer{
		apiToken: apiToken,
		apiURL:   "https://send.api.mailtrap.io/api/send",
		from:     fromEmail,
	}
}