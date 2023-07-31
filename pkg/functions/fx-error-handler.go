package functions

import (
	"go.uber.org/fx"
	"kloudlite.io/pkg/logging"
)

type ErrH struct {
	Logger logging.Logger
}

func (e *ErrH) HandleError(err error) {
	e.Logger.Errorf(err)
}

func (e *ErrH) String() string {
	return "err-handler"
}

func FxErrorHandler() fx.Option {
	return fx.Provide(
		func(logger logging.Logger) fx.Option {
			return fx.ErrorHook(&ErrH{Logger: logger.WithKV("component", "fx-error-handler")})
		},
	)
}
