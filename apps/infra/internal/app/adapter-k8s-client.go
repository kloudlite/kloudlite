package app

import (
	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

func NewK8sClient(config *rest.Config, scheme *runtime.Scheme) (domain.K8sClient, error) {
	return k8s.NewClient(config, scheme)
}