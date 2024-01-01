package types

import (
	"strings"

	"github.com/kloudlite/operator/pkg/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceUpdate struct {
	AuthToken   string         `json:"authToken"`
	AccountName string         `json:"accountName"`
	ClusterName string         `json:"clusterName"`
	Object      map[string]any `json:"object"`
}

type ResourceStatus string

func (rs ResourceStatus) String() string {
	return string(rs)
}

const (
	ResourceStatusUpdated  ResourceStatus = "updated"
	ResourceStatusDeleting ResourceStatus = "deleting"
	ResourceStatusDeleted  ResourceStatus = "deleted"
)

const (
	ResourceStatusKey string = "resource-watcher-resource-status"
)

func HasOtherKloudliteFinalizers(obj client.Object) bool {
	hasOtherKloudliteFinalizers := false

	for _, f := range obj.GetFinalizers() {
		if strings.Contains(f, ".kloudlite.io") {
			if f == constants.StatusWatcherFinalizer {
				continue
			}
			hasOtherKloudliteFinalizers = true
		}
	}

	return hasOtherKloudliteFinalizers
}
