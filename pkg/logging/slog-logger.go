package logging

import (
	"io"
	"log/slog"
	"os"

	"github.com/charmbracelet/log"
)

type SlogOptions struct {
	Writer io.Writer
	Prefix string

	ShowTimestamp bool
	ShowCaller    bool
	ShowDebugLogs bool
}

func NewSlogLogger(opts SlogOptions) *slog.Logger {
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}

	level := log.InfoLevel
	if opts.ShowDebugLogs {
		level = log.DebugLevel
	}

	logger := log.NewWithOptions(opts.Writer, log.Options{ReportCaller: opts.ShowCaller, ReportTimestamp: opts.ShowTimestamp, Prefix: opts.Prefix, Level: level})

	return slog.New(logger)
}
