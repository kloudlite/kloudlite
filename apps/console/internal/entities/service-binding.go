package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/repos"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
)

type InterceptStatus struct {
	Intercepted  *bool                             `json:"intercepted"`
	ToAddr       string                            `json:"toAddr"`
	PortMappings []crdsv1.SvcInterceptPortMappings `json:"portMappings,omitempty"`
}

type ServiceBinding struct {
	repos.BaseEntity            `json:",inline"`
	networkingv1.ServiceBinding `json:",inline"`

	AccountName string `json:"accountName"`
	ClusterName string `json:"clusterName"`

	EnvironmentName      string `json:"environmentName" graphql:"noinput"`
	EnvironmentNamespace string `json:"environmentNamespace" graphql:"noinput"`

	InterceptStatus *InterceptStatus `json:"interceptStatus"`
}

var ServiceBindingIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.AccountName, Value: repos.IndexAsc},
			// {Key: fc.ClusterName, Value: repos.IndexAsc},
			// {Key: fc.MetadataName, Value: repos.IndexAsc},
			{Key: fc.ServiceBindingSpecHostname, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ServiceBindingSpecServiceRefName, Value: repos.IndexAsc},
			{Key: fc.EnvironmentNamespace, Value: repos.IndexAsc},
		},
	},
}
