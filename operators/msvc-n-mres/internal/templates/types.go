package templates

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/pkg/plugin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CommonManagedServiceParams struct {
	APIVersion string
	Kind       string

	Metadata       metav1.ObjectMeta
	PluginTemplate crdsv1.PluginTemplate
	Export         plugin.Export
}

type CommonManagedResourceParams struct {
	APIVersion string
	Kind       string

	Metadata       metav1.ObjectMeta
	PluginTemplate crdsv1.PluginTemplate
	Export         plugin.Export
}
