package logger

import "go.uber.org/zap"

// Logger is a common logger interface
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}

// ZapLogger wraps zap.Logger to implement the Logger interface
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new ZapLogger from a zap.Logger
func NewZapLogger(logger *zap.Logger) Logger {
	return &ZapLogger{logger: logger}
}

func (z *ZapLogger) Debug(msg string, fields ...interface{}) {
	z.logger.Debug(msg)
}

func (z *ZapLogger) Info(msg string, fields ...interface{}) {
	z.logger.Info(msg)
}

func (z *ZapLogger) Warn(msg string, fields ...interface{}) {
	z.logger.Warn(msg)
}

func (z *ZapLogger) Error(msg string, fields ...interface{}) {
	z.logger.Error(msg)
}

func (z *ZapLogger) Fatal(msg string, fields ...interface{}) {
	z.logger.Fatal(msg)
}