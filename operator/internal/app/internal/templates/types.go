package templates

import (
	"github.com/kloudlite/operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type DeploymentParams struct {
	Metadata       metav1.ObjectMeta
	Paused         bool
	Replicas       int
	PodLabels      map[string]string
	PodAnnotations map[string]string
	PodSpec        corev1.PodSpec
}
