package entities

import (
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type SecretVariable struct {
	repos.BaseEntity        `json:",inline" graphql:"noinput"`
	AccountName             string            `json:"accountName" graphql:"noinput"`
	Name                    string            `json:"name"`
	StringData              map[string]string `json:"stringData"`
	common.ResourceMetadata `json:",inline"`
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
			{Key: fc.SecretVariableName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
