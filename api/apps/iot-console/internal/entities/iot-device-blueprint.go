package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type BluePrintType string

const (
	SingletonBlueprint BluePrintType = "singleton-blueprint"
	GroupBlueprint     BluePrintType = "group-blueprint"
)

type IOTDeviceBlueprint struct {
	repos.BaseEntity        `json:",inline" graphql:"noinput"`
	common.ResourceMetadata `json:",inline"`

	Name            string         `json:"name"`
	AccountName     string         `json:"accountName" graphql:"noinput"`
	ProjectName     string         `json:"projectName" graphql:"noinput"`
	EnvironmentName string         `json:"environmentName" graphql:"noinput"`
	BluePrintType   *BluePrintType `json:"bluePrintType"`
	Version         string         `json:"version"`
}

var IOTDeviceBlueprintIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{
				Key:   "name",
				Value: repos.IndexAsc,
			},
		},
		Unique: true,
	},
}
