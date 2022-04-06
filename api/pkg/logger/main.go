package logger

import (
	"flag"
	"go.uber.org/zap"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger() Logger {
	isDev := flag.Bool("dev", false, "development mode")
	flag.Parse()

	if isDev != nil && *isDev {
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		return Logger{SugaredLogger: logger.Sugar()}
	}
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return Logger{SugaredLogger: logger.Sugar()}
}
