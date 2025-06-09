package domain

import (
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	"github.com/kloudlite/api/apps/iot-console/internal/env"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
)

type domain struct {
	k8sClient k8s.Client
	logger    logging.Logger

	iotProjectRepo         repos.DbRepo[*entities.IOTProject]
	iotDeploymentRepo      repos.DbRepo[*entities.IOTDeployment]
	iotDeviceRepo          repos.DbRepo[*entities.IOTDevice]
	iotDeviceBlueprintRepo repos.DbRepo[*entities.IOTDeviceBlueprint]
	iotAppRepo             repos.DbRepo[*entities.IOTApp]

	envVars *env.Env
}

type IOTConsoleCacheStore kv.BinaryDataRepo

var Module = fx.Module("domain",
	fx.Provide(func(
		k8sClient k8s.Client,
		logger logging.Logger,

		iotProjectRepo repos.DbRepo[*entities.IOTProject],
		iotDeploymentRepo repos.DbRepo[*entities.IOTDeployment],
		iotDeviceRepo repos.DbRepo[*entities.IOTDevice],
		iotDeviceBlueprintRepo repos.DbRepo[*entities.IOTDeviceBlueprint],
		iotAppRepo repos.DbRepo[*entities.IOTApp],

		ev *env.Env,
	) Domain {
		return &domain{
			k8sClient:              k8sClient,
			logger:                 logger,
			iotProjectRepo:         iotProjectRepo,
			iotDeploymentRepo:      iotDeploymentRepo,
			iotDeviceRepo:          iotDeviceRepo,
			iotDeviceBlueprintRepo: iotDeviceBlueprintRepo,
			iotAppRepo:             iotAppRepo,
			envVars:                ev,
		}
	}),
)
