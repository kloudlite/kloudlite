package logging

import (
	"os"

	"github.com/rs/zerolog"
)

type HttpLogger interface {
}

type httpLogger struct {
	zerolog.Logger
}

func NewHttpLogger() HttpLogger {
	z := zerolog.New(os.Stdout).With().Caller().Timestamp().Logger()
	return &httpLogger{
		Logger: z,
	}
}
