package functions

import (
	"go.uber.org/fx"
	"kloudlite.io/pkg/logging"
)

type errH struct {
	logger logging.Logger
}

func (e errH) HandleError(err error) {
	e.logger.Errorf(err)
}

func FxErrorHandler() fx.Option {
	return fx.Provide(
		func(logger logging.Logger) fx.Option {
			return fx.ErrorHook(errH{logger: logger})
		},
	)
}
