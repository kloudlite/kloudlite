package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type SecretVariable struct {
	repos.BaseEntity        `json:",inline" graphql:"noinput"`
	AccountName             string            `json:"accountName" graphql:"noinput"`
	Metadata                Metadata          `json:"metadata"`
	StringData              map[string]string `json:"stringData"`
	common.ResourceMetadata `json:",inline"`
}

type Metadata struct {
	Name string `json:"name"`
}

var SecretVariableIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fields.MetadataName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
