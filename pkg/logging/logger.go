package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// type Logger interface {
// 	Debugf(msg string, args ...any)
// 	Infof(msg string, args ...any)
// 	Errorf(err error, msg string, args ...any)
// 	Warnf(msg string, args ...any)
// 	WithName(name string) Logger
// 	WithKV(keyValuePairs ...any) Logger
// }

type Logger struct {
	zapLogger *zap.SugaredLogger
}

func (c Logger) WithKV(keyValuePairs ...any) Logger {
	return Logger{zapLogger: c.zapLogger.With(keyValuePairs...)}
}

func (c Logger) Debugf(msg string, args ...any) {
	c.zapLogger.Debugf(msg, args...)
}

func (c Logger) Infof(msg string, args ...any) {
	c.zapLogger.Infof(msg, args...)
}

func (c Logger) Errorf(err error, args ...string) {
	if len(args) > 0 {
		c.zapLogger.Error(err, args)
		return
	}
	c.zapLogger.Error(err)
}

func (c Logger) Warnf(msg string, args ...any) {
	c.zapLogger.Warnf(msg, args...)
}

func (c Logger) WithName(name string) Logger {
	return Logger{zapLogger: c.zapLogger.Named(name)}
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

	zapConfig := func() zap.Config {
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
	logger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		return Logger{}, err
	}
	cLogger := Logger{zapLogger: logger.Sugar()}
	if opts.Name != "" {
		cLogger.zapLogger = cLogger.zapLogger.Named(opts.Name)
	}
	return cLogger, nil
}
