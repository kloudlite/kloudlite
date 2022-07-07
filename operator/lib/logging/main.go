package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/types"
	"operators.kloudlite.io/lib/errors"
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
	if opts.Dev {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.LineEnding = "\n\n"
		cfg.EncoderConfig.TimeKey = ""
		logger, err := cfg.Build(zap.AddCallerSkip(1))
		if err != nil {
			return nil, err
		}
		if opts.Name != "" {
			return &customLogger{zapLogger: logger.Sugar().Named(opts.Name)}, nil
		}
		return &customLogger{zapLogger: logger.Sugar()}, nil
	}
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	if opts.Name != "" {
		return &customLogger{zapLogger: logger.Sugar().Named(opts.Name)}, nil
	}
	return &customLogger{zapLogger: logger.Sugar()}, nil
}

func Must(l Logger, err error) Logger {
	if err != nil {
		panic(err)
	}
	return l
}

func NewZapLogger(nn types.NamespacedName) *zap.SugaredLogger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.LineEnding = "\n\n"
	cfg.EncoderConfig.TimeKey = ""
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()
	return sugar.With("REF", nn.String())
}
