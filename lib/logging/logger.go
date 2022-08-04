package logging

import (
	"fmt"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"operators.kloudlite.io/lib/errors"
)

type Logger interface {
	Debugf(msg string, args ...any)
	Infof(msg string, args ...any)
	Errorf(err error, msg string, args ...any)
	Error(err error)
	Warnf(msg string, args ...any)
	WithKV(key string, value any) Logger
	WithName(name string) Logger
}

type customLogger struct {
	opts   Options
	logger *zap.SugaredLogger
}

func (c customLogger) Debugf(msg string, args ...any) {
	c.logger.Debugf(msg, args...)
}

func (c customLogger) Infof(msg string, args ...any) {
	c.logger.Infof(msg, args...)
}

func (c customLogger) Errorf(err error, msg string, args ...any) {
	c.logger.Errorf(errors.NewEf(err, msg, args...).Error())
}

func (c customLogger) Error(err error) {
	c.logger.Errorf(err.Error())
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
	Name string
	Dev  bool
}

var magenta = color.New(color.FgCyan).SprintFunc()
func decorateName(name string) string {
	return fmt.Sprintf("(%s)", magenta(name))
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

			if len(opts.Name) > 0 {
				opts.Name = decorateName(opts.Name)
			}
			return cfg
		}
		return zap.NewProductionConfig()
	}()

	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
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

