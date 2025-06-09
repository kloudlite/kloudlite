package templates

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/toolkit/templates"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CommonManagedServiceParams struct {
	Metadata       metav1.ObjectMeta
	PluginTemplate crdsv1.PluginTemplate
}

var ParseBytes = templates.ParseBytes
