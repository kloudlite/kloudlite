package templates

import (
	"github.com/kloudlite/operator/api/v1"
)

type HPASpecParams struct {
	DeploymentName string
	HPA            *v1.AppHPA
}

type AppInterceptPodSpecParams struct {
	DeviceHost      string
	TCPPortMappings map[int32]int32
	UDPPortMappings map[int32]int32
}
