package logging

import (
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/kloudlite/operator/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugf(msg string, args ...any)
	Infof(msg string, args ...any)
	Errorf(err error, msg string, args ...any)
	Error(err error)
	Warnf(msg string, args ...any)
	WithKV(key string, value any) Logger
	WithName(name string) Logger
	WithOptions(options ...zap.Option) Logger
}

type customLogger struct {
	opts   Options
	logger *zap.SugaredLogger
}

func (c customLogger) WithOptions(options ...zap.Option) Logger {
	lg := c.logger.Desugar().WithOptions(options...)
	return &customLogger{logger: lg.Sugar(), opts: c.opts}
}

func (c customLogger) Debugf(msg string, args ...any) {
	c.logger.Debugf(msg, args...)
}

func (c customLogger) Infof(msg string, args ...any) {
	c.logger.Infof(msg, args...)
}

func (c customLogger) Errorf(err error, msg string, args ...any) {
	c.logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Errorf(errors.NewEf(err, msg, args...).Error())
}

func (c customLogger) Error(err error) {
	c.logger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Errorf(err.Error())
}

func (c customLogger) Warnf(msg string, args ...any) {
	c.logger.Warnf(msg, args...)
}

func (c customLogger) WithKV(key string, value any) Logger {
	c.logger = c.logger.With(key, value)
	return c
}

func (c customLogger) WithName(name string) Logger {
	if c.opts.Dev {
		return &customLogger{logger: c.logger.Named(decorateName(name)), opts: c.opts}
	}
	return &customLogger{logger: c.logger.Named(decorateName(name)), opts: c.opts}
}

type Options struct {
	Name        string
	Dev         bool
	CallerTrace bool
}

func decorateName(name string) string {
	return color.New(color.FgHiCyan, color.Bold).SprintFunc()(name)
}

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

			if len(opts.Name) > 0 {
				opts.Name = decorateName(opts.Name)
			}
			return cfg
		}
		return zap.NewProductionEncoderConfig()
	}()

	if !opts.Dev {
		cfg.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(ts.UTC().Format(time.RFC3339))
		}
	}

	loglevel := zapcore.InfoLevel
	if opts.Dev {
		loglevel = zapcore.DebugLevel
	}

	zapOpts := make([]zap.Option, 0, 2)
	if opts.CallerTrace {
		zapOpts = append(zapOpts, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), os.Stdout, loglevel), zapOpts...)

	cLogger := &customLogger{logger: logger.Sugar(), opts: opts}
	if opts.Name != "" {
		cLogger.logger = cLogger.logger.Named(opts.Name)
	}
	return cLogger, nil
}

func NewOrDie(options *Options) Logger {
	logger, err := New(options)
	if err != nil {
		panic(err)
	}
	return logger
}
