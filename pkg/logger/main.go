package logger

import "go.uber.org/zap"

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger(isDev bool) Logger {
	logger, err := zap.NewProduction()
	if isDev {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	return Logger{SugaredLogger: logger.Sugar()}
}
