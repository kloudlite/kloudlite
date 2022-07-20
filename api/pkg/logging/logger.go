package logging

import (
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugf(msg string, args ...any)
	Infof(msg string, args ...any)
	Errorf(err error, msg string, args ...any)
	Warnf(msg string, args ...any)
	WithName(name string) Logger
}

type customLogger struct {
	zapLogger *zap.SugaredLogger
}

func (c customLogger) Debugf(msg string, args ...any) {
	c.zapLogger.Debugf(msg, args...)
}

func (c customLogger) Infof(msg string, args ...any) {
	c.zapLogger.Infof(msg, args...)
}

func (c customLogger) Errorf(err error, msg string, args ...any) {
	c.zapLogger.Errorf("%s AS %+v happened", fmt.Sprintf(msg, args...), err)
}

func (c customLogger) Warnf(msg string, args ...any) {
	c.zapLogger.Warnf(msg, args...)
}

func (c customLogger) WithName(name string) Logger {
	return &customLogger{zapLogger: c.zapLogger.Named(name)}
}

type Options struct {
	Name string
	Dev  bool
}

func NewLogger(options ...Options) (Logger, error) {
	opts := Options{}
	if len(options) > 0 {
		opts = options[0]
	}
	if opts.Dev {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.LineEnding = "\n\n"
		cfg.EncoderConfig.TimeKey = ""
		cfg.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fmt.Sprintf("(%s) %s", caller.Function, caller.TrimmedPath()))
		}
		logger, err := cfg.Build(zap.AddCallerSkip(1))
		if err != nil {
			return nil, err
		}
		return &customLogger{zapLogger: logger.Sugar().Named(opts.Name)}, nil
	}
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("(%s) %s", caller.Function, caller.TrimmedPath()))
	}
	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &customLogger{zapLogger: logger.Sugar().Named(opts.Name)}, nil
}

func FxProvider() fx.Option {
	return fx.Provide(NewLogger)
}
