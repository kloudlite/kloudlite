package main

import (
	"os"

	"github.com/kloudlite/operator/toolkit/logging"
)

func main() {
	logger := logging.NewOrDie(&logging.Options{
		Writer:          os.Stderr,
		Prefix:          "",
		ShowTimestamp:   false,
		ShowCaller:      true,
		ShowDebugLogs:   false,
		DevelopmentMode: true,
	})

	logger.Debugf("hello world")
	logger.Infof("hello world")
	logger.Warnf("hello world")
}
