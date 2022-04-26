package logger

import (
	"flag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger() Logger {
	isDev := flag.Bool("dev", false, "development mode")
	flag.Parse()

	if isDev != nil && *isDev {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, _ := config.Build()
		return Logger{SugaredLogger: logger.Sugar()}
	}
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return Logger{SugaredLogger: logger.Sugar()}
}
