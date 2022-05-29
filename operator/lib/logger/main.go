package logger

import (
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
)

func New(nn types.NamespacedName) *zap.SugaredLogger {
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
