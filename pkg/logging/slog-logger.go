package logging

import (
	"io"
	"log/slog"
	"os"

	"github.com/charmbracelet/lipgloss"
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
	// INFO: force colored output, otherwise honor the env-var `CLICOLOR_FORCE`
	if _, ok := os.LookupEnv("CLICOLOR_FORCE"); !ok {
		os.Setenv("CLICOLOR_FORCE", "1")
	}

	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}

	level := log.InfoLevel
	if opts.ShowDebugLogs {
		level = log.DebugLevel
	}

	logger := log.NewWithOptions(opts.Writer, log.Options{
		ReportCaller:    opts.ShowCaller,
		ReportTimestamp: opts.ShowTimestamp,
		Prefix:          opts.Prefix,
		Level:           level,
	})

	styles := log.DefaultStyles()
	styles.Levels[log.DebugLevel] = styles.Levels[log.DebugLevel].Foreground(lipgloss.Color("#5b717f"))
	styles.Levels[log.InfoLevel] = styles.Levels[log.InfoLevel].Foreground(lipgloss.Color("#36cbfa"))

	styles.Key = lipgloss.NewStyle().Foreground(lipgloss.Color("#36cbfa")).Bold(true)

	logger.SetStyles(styles)

	l := slog.New(logger)

	return l
}
