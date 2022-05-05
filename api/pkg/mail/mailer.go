package mail

type Mailer interface {
	SendEmail(
		from string,
		fromName string,
		subject string,
		to string,
		toName string,
		plaintextContent string,
		htmlContent string,
	) error
}
