package logging

import (
	"context"
	"log/slog"
)

type Slogger interface{}

type slogger struct{}

// Enabled implements Slogger.
func (s *slogger) Enabled(context.Context, slog.Level) bool {
	panic("unimplemented")
}

// Handle implements Slogger.
func (s *slogger) Handle(context.Context, slog.Record) error {
	panic("unimplemented")
}

// SkipCaller implements Slogger.
func (s *slogger) SkipCaller() {
	panic("unimplemented")
}

// WithAttrs implements Slogger.
func (s *slogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	panic("unimplemented")
}

// WithGroup implements Slogger.
func (s *slogger) WithGroup(name string) slog.Handler {
	panic("unimplemented")
}

var _ Slogger = (*slogger)(nil)
