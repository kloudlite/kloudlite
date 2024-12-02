package logging

import (
	"io"
	"os"
	"time"

	"github.com/kloudlite/operator/toolkit/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type slogLike interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
	Fatal(msg string, kv ...any)
}

type Logger interface {
	slogLike

	// Skip adds
	Skip() Logger

	Debugf(msg string, args ...any)
	Infof(msg string, args ...any)
	Errorf(err error, msg string, args ...any)
	// Error(err error)
	Warnf(msg string, args ...any)
	WithKV(key string, value any) Logger
	// WithName(name string) Logger
	WithOptions(options ...zap.Option) Logger

	// Slog() *slog.Logger
}

type logger struct {
	opts Options
	sl   *zap.SugaredLogger
}

// Skip implements Logger.
func (c *logger) Skip() Logger {
	lg := c.sl.Desugar().WithOptions(zap.AddCallerSkip(1))
	return &logger{sl: lg.Sugar(), opts: c.opts}
}

// Debug implements Logger.
func (c *logger) Debug(msg string, kv ...any) {
	c.sl.Debugw(msg, kv...)
}

// Error implements Logger.
func (c *logger) Error(msg string, kv ...any) {
	c.sl.Errorw(msg, kv...)
}

// Info implements Logger.
func (c *logger) Info(msg string, kv ...any) {
	c.sl.Infow(msg, kv...)
}

// Fatal implements Logger.
func (c *logger) Fatal(msg string, kv ...any) {
	c.sl.Fatalw(msg, kv...)
}

// Warn implements Logger.
func (c *logger) Warn(msg string, kv ...any) {
	c.sl.Warnw(msg, kv...)
}

func (c logger) WithOptions(options ...zap.Option) Logger {
	lg := c.sl.Desugar().WithOptions(options...)
	return &logger{sl: lg.Sugar(), opts: c.opts}
}

func (c logger) Debugf(msg string, args ...any) {
	c.sl.Debugf(msg, args...)
}

func (c logger) Infof(msg string, args ...any) {
	c.sl.Infof(msg, args...)
}

func (c logger) Errorf(err error, msg string, args ...any) {
	c.sl.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Errorf(errors.NewEf(err, msg, args...).Error())
}

// func (c logger) Error(err error) {
// 	c.sl.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Errorf(err.Error())
// }

func (c logger) Warnf(msg string, args ...any) {
	c.sl.Warnf(msg, args...)
}

func (c logger) WithKV(key string, value any) Logger {
	c.sl = c.sl.With(key, value)
	return &c
}

// func (c logger) WithName(name string) Logger {
// 	if c.opts.DevelopmentMode {
// 		return &logger{sl: c.sl.Named(decorateName(name)), opts: c.opts}
// 	}
// 	return &logger{sl: c.sl.Named(decorateName(name)), opts: c.opts}
// }

type Options struct {
	Writer io.Writer
	Prefix string

	ShowTimestamp bool
	ShowCaller    bool
	ShowDebugLogs bool

	DevelopmentMode bool
}

func New(options *Options) (Logger, error) {
	opts := Options{}
	if options != nil {
		opts = *options
	}

	cfg := func() zapcore.EncoderConfig {
		if opts.DevelopmentMode {
			cfg := zap.NewDevelopmentEncoderConfig()
			cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
			cfg.LineEnding = "\n"
			cfg.TimeKey = ""

			return cfg
		}
		ec := zap.NewProductionEncoderConfig()
		ec.TimeKey = "" // because k8s logs will always have timestamps
		return ec
	}()

	if !opts.DevelopmentMode {
		cfg.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(ts.UTC().Format(time.RFC3339))
		}
	}

	loglevel := zapcore.InfoLevel
	if opts.DevelopmentMode {
		loglevel = zapcore.DebugLevel
	}

	zapOpts := make([]zap.Option, 0, 3)
	zapOpts = append(zapOpts, zap.AddStacktrace(zap.DPanicLevel))
	if opts.ShowCaller {
		zapOpts = append(zapOpts, zap.AddCaller(), zap.AddCallerSkip(1))
	}

	logr := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), os.Stdout, loglevel), zapOpts...)

	cLogger := &logger{sl: logr.Sugar(), opts: opts}
	return cLogger, nil
}

func NewOrDie(options *Options) Logger {
	logger, err := New(options)
	if err != nil {
		panic(err)
	}
	return logger
}
