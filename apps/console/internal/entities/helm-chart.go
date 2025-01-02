package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/toolkit/reconciler"
)

type HelmChart struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	crdsv1.HelmChart `json:",inline"`

	AccountName     string `json:"accountName" graphql:"noinput"`
	EnvironmentName string `json:"environmentName" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (a *HelmChart) GetDisplayName() string {
	return a.ResourceMetadata.DisplayName
}

func (a *HelmChart) GetGeneration() int64 {
	return a.ObjectMeta.Generation
}

func (a *HelmChart) GetStatus() reconciler.Status {
	return a.HelmChart.Status.Status
}

func (a *HelmChart) GetResourceType() ResourceType {
	return ResourceTypeApp
}

var HelmChartIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fields.MetadataName, Value: repos.IndexAsc},
			{Key: fields.MetadataNamespace, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fields.EnvironmentName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
