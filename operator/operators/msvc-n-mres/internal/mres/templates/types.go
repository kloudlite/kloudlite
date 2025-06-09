package templates

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/toolkit/templates"
	"github.com/kloudlite/operator/toolkit/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CommonManagedResourceParams struct {
	Metadata          metav1.ObjectMeta
	PluginTemplate    crdsv1.PluginTemplate
	ManagedServiceRef types.ObjectReference
}

var ParseBytes = templates.ParseBytes
