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
	Errorf(msg string, args ...any)
	Warnf(msg string, args ...any)
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

func (c customLogger) Errorf(msg string, args ...any) {
	c.zapLogger.Errorf(msg, args...)
}

func (c customLogger) Warnf(msg string, args ...any) {
	c.zapLogger.Warnf(msg, args...)
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
		// cfg.EncoderConfig.CallerKey = "random"
		cfg.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			// pwd, err := os.Getwd()
			// if err != nil {
			// 	panic(err)
			// }
			// rel, err := filepath.Rel(pwd, caller.FullPath())
			// if err != nil {
			// 	panic(err)
			// }
			// fmt.Println("value: ", filepath.Base(caller.FullPath()), caller.FullPath(), rel)
			// enc.AppendString(filepath.Base(caller.FullPath()))
			// fmt.Println("calleR: ", caller, caller.Defined, caller.TrimmedPath(), caller.Function)
			enc.AppendString(fmt.Sprintf("(%s) %s", caller.Function, caller.TrimmedPath()))
			// enc.AppendString(rel)
		}
		logger, err := cfg.Build(zap.AddCallerSkip(1))
		if err != nil {
			return nil, err
		}
		return &customLogger{zapLogger: logger.Sugar()}, nil
	}
	cfg := zap.NewProductionConfig()
	// cfg.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// 	enc.AppendString(filepath.Base(caller.FullPath()))
	// }
	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &customLogger{zapLogger: logger.Sugar()}, nil
}

func FxProvider() fx.Option {
	return fx.Provide(NewLogger)
}
