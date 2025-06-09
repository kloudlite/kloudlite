package grpc

import (
	"fmt"
	"log/slog"
	"time"
)

type logger struct {
	*slog.Logger
	comment string
	t       time.Time
}

func NewRequestLogger(logr *slog.Logger, comment string) *logger {
	logr.Debug("➡️  " + comment)
	return &logger{Logger: logr, comment: comment, t: time.Now()}
}

func (l *logger) End() {
	l.Info("↩️  "+l.comment, "took", fmt.Sprintf("%.2fs", time.Since(l.t).Seconds()))
}
