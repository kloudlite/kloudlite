package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/toolkit/reconciler"
)

type App struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	crdsv1.App `json:",inline"`

	CIBuildId *repos.ID `json:"ciBuildId,omitempty" graphql:"scalar-type=ID"`

	AccountName     string `json:"accountName" graphql:"noinput"`
	EnvironmentName string `json:"environmentName" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (a *App) GetDisplayName() string {
	return a.ResourceMetadata.DisplayName
}

func (a *App) GetGeneration() int64 {
	return a.ObjectMeta.Generation
}

func (a *App) GetStatus() reconciler.Status {
	return a.App.Status
}

func (a *App) GetResourceType() ResourceType {
	return ResourceTypeApp
}

var AppIndexes = []repos.IndexField{
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
	{
		Field: []repos.IndexKey{
			{Key: fc.AppSpecInterceptToDevice, Value: repos.IndexAsc},
		},
	},
}
