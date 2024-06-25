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
  ShowCaller bool

  LogLevel slog.Level
}

func ParseLogLevelFromEnv(key string) slog.Level {
  s, ok := os.LookupEnv(key)
  if !ok {
    return slog.LevelInfo
  }
  l, err := log.ParseLevel(s)
if err != nil {
        return slog.LevelInfo
      }
  return slog.Level(l)
}

func NewSlogLogger(opts SlogOptions) *slog.Logger {
  if opts.Writer == nil {
    opts.Writer = os.Stderr
  }
  log := log.NewWithOptions(opts.Writer, log.Options{ReportCaller: opts.ShowCaller, ReportTimestamp: opts.ShowTimestamp, Prefix: opts.Prefix, Level: log.Level(opts.LogLevel.Level())})
	return slog.New(log)
}
