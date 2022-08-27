package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"kloudlite.io/pkg/errors"
)

type Logger interface {
	Debugf(msg string, args ...any)
	Infof(msg string, args ...any)
	Errorf(err error, msg string, args ...any)
	Warnf(msg string, args ...any)
	WithName(name string) Logger
	WithKV(key string, value any) Logger
}

type customLogger struct {
	opts      Options
	zapLogger *zap.SugaredLogger
}

func (c customLogger) WithKV(key string, value any) Logger {
	c.zapLogger = c.zapLogger.With(key, value)
	return c
}

func (c customLogger) Debugf(msg string, args ...any) {
	c.zapLogger.Debugf(msg, args...)
}

func (c customLogger) Infof(msg string, args ...any) {
	c.zapLogger.Infof(msg, args...)
}

func (c customLogger) Errorf(err error, msg string, args ...any) {
	c.zapLogger.Errorf(errors.NewEf(err, msg, args...).Error())
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

func New(options *Options) (Logger, error) {
	opts := Options{}
	if options != nil {
		opts = *options
	}

	cfg := func() zap.Config {
		if opts.Dev {
			cfg := zap.NewDevelopmentConfig()
			cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			cfg.EncoderConfig.LineEnding = "\n"
			cfg.EncoderConfig.TimeKey = ""
			cfg.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(fmt.Sprintf("(%s) %s", caller.Function, caller.TrimmedPath()))
			}
			return cfg
		}
		return zap.NewProductionConfig()
	}()
	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	cLogger := &customLogger{zapLogger: logger.Sugar(), opts: opts}
	if opts.Name != "" {
		cLogger.zapLogger = cLogger.zapLogger.Named(opts.Name)
	}
	return cLogger, nil
}
