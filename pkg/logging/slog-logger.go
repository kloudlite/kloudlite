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

func NewSlogLogger(opts SlogOptions) *slog.Logger {
	if opts.Writer == nil {
		opts.Writer = os.Stderr
	}

	level := log.InfoLevel
	if opts.ShowDebugLogs {
		level = log.DebugLevel
	}

	logger := log.NewWithOptions(opts.Writer, log.Options{ReportCaller: opts.ShowCaller, ReportTimestamp: opts.ShowTimestamp, Prefix: opts.Prefix, Level: level})

	styles := log.DefaultStyles()
	styles.Levels[log.DebugLevel] = styles.Levels[log.DebugLevel].Foreground(lipgloss.Color("#5b717f"))

	// styles.Key = lipgloss.NewStyle().Background(lipgloss.Color("#083e54")).Foreground(lipgloss.Color("#9dbdc9")).Bold(true)
	styles.Key = lipgloss.NewStyle().Foreground(lipgloss.Color("#36cbfa")).Bold(true)

	logger.SetStyles(styles)

	return slog.New(logger)
}
