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

func ParseLogLevelFromString(s string) slog.Level {
	l, err := log.ParseLevel(s)
	if err != nil {
		return slog.LevelInfo
	}
	return slog.Level(l)
}

func ParseLogLevelFromEnv(key string) slog.Level {
	s, ok := os.LookupEnv(key)
	if !ok {
		return slog.LevelInfo
	}
	return ParseLogLevelFromString(s)
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
