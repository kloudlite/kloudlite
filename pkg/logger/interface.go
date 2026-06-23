package logger

import "go.uber.org/zap"

// Logger is the webhook logger type.
type Logger = *zap.SugaredLogger

func NewZapLogger(logger *zap.Logger) Logger {
	return logger.Sugar()
}
