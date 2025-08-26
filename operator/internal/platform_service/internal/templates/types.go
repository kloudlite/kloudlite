package templates

import (
	"github.com/kloudlite/operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PluginResourceTemplateArgs struct {
	Metadata metav1.ObjectMeta
	Plugin   *v1.PluginTemplate
}
