package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

type ExpirationUnit string

const (
	ExpirationUnitHour  ExpirationUnit = "h"
	ExpirationUnitDays  ExpirationUnit = "d"
	ExpirationUnitWeek  ExpirationUnit = "w"
	ExpirationUnitMonth ExpirationUnit = "m"
	ExpirationUnitYear  ExpirationUnit = "y"
)

type RepoAccess string
type Expiration struct {
	Unit  ExpirationUnit `json:"unit"`
	Value int            `json:"value"`
}

const (
	RepoAccessReadOnly  RepoAccess = "read"
	RepoAccessReadWrite RepoAccess = "read_write"
)

type Credential struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	CreatedBy     common.CreatedOrUpdatedBy `json:"createdBy" graphql:"noinput"`
	LastUpdatedBy common.CreatedOrUpdatedBy `json:"lastUpdatedBy" graphql:"noinput"`

	AccountName string     `json:"accountName" graphql:"noinput"`
	Access      RepoAccess `json:"access"`
	Expiration  Expiration `json:"expiration"`
	Name        string     `json:"name"`
	UserName    string     `json:"username"`
	TokenKey    string     `json:"token_key" graphql:"ignore"`
}

var CredentialIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "username", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
