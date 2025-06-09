package logging

import (
	"log/slog"

	"github.com/go-logr/logr"
)

type options struct {
	callDepth int
}

func defaultOptions() *options {
	return &options{}
}

type Opt func(opts *options)

func WithCallDepth(depth int) Opt {
	return func(opts *options) {
		opts.callDepth = depth
	}
}

func Slog(logger logr.Logger, opts ...Opt) *slog.Logger {
	options := defaultOptions()
	for i := range opts {
		opts[i](options)
	}

	if options.callDepth > 0 {
		logger = logger.WithCallDepth(options.callDepth)
	}

	return slog.New(logr.ToSlogHandler(logger))
}

func New(logger logr.Logger, opts ...Opt) *slog.Logger {
	return Slog(logger, opts...)
}
