package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

func newSession() (*session.Session, error) {
	return session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
		},
	)
}
