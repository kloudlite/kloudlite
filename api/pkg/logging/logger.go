package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugf(msg string, args ...any)
	Infof(msg string, args ...any)
	Errorf(err error, args ...any)
	Warnf(msg string, args ...any)
	WithName(name string) Logger
	WithKV(keyValuePairs ...any) Logger
}

type logger struct {
	zapLogger *zap.SugaredLogger
}

func (c *logger) WithKV(keyValuePairs ...any) Logger {
	return &logger{zapLogger: c.zapLogger.With(keyValuePairs...)}
}

func (c *logger) Debugf(msg string, args ...any) {
	c.zapLogger.Debugf(msg, args...)
}

func (c *logger) Infof(msg string, args ...any) {
	c.zapLogger.Infof(msg, args...)
}

func (c *logger) Errorf(err error, args ...any) {
	if len(args) > 0 {
		c.zapLogger.Error(err, args)
		return
	}
	c.zapLogger.Error(err)
}

func (c *logger) Warnf(msg string, args ...any) {
	c.zapLogger.Warnf(msg, args...)
}

func (c *logger) WithName(name string) Logger {
	return &logger{zapLogger: c.zapLogger.Named(name)}
}

type Options struct {
	Name        string
	Dev         bool
	CallerTrace bool
}

var EmptyLogger *logger

func New(options *Options) (Logger, error) {
	opts := Options{}
	if options != nil {
		opts = *options
	}

	cfg := func() zapcore.EncoderConfig {
		if opts.Dev {
			cfg := zap.NewDevelopmentEncoderConfig()
			cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
			cfg.LineEnding = "\n"
			cfg.TimeKey = ""

			return cfg
		}
		pcfg := zap.NewProductionEncoderConfig()
		pcfg.TimeKey = ""
		return pcfg
	}()

	// if !opts.Dev {
	// 	cfg.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
	// 		encoder.AppendString(ts.UTC().Format(time.RFC3339))
	// 	}
	// }

	loglevel := zapcore.InfoLevel
	if opts.Dev {
		loglevel = zapcore.DebugLevel
	}

	zapOpts := make([]zap.Option, 0, 3)
	zapOpts = append(zapOpts, zap.AddStacktrace(zap.DPanicLevel))

	if !opts.Dev {
		opts.CallerTrace = true
	}

	if opts.CallerTrace {
		zapOpts = append(zapOpts, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	lgr := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), os.Stdout, loglevel), zapOpts...)

	cLogger := &logger{
		zapLogger: lgr.Sugar(),
	}
	return cLogger, nil
}
