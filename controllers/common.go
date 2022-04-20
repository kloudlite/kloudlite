package controllers

import (
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
)

func GetLogger(name types.NamespacedName) *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()
	return sugar.With(
		"NAME", name.String(),
	)
}
