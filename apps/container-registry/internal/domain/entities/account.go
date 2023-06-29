package entities

import "kloudlite.io/pkg/repos"

type HarborProject struct {
	repos.BaseEntity  `json:",inline" graphql:"noinput"`
	HarborProjectName string            `json:"harborProjectName"`
	AccountName       string            `json:"accountName"`
	Credentials       HarborCredentials `json:"credentials"`
}

type HarborCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var HarborProjectIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "project_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "account_name", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
