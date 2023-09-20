package entities

import (
	"kloudlite.io/pkg/repos"
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
	AccountName      string     `json:"accountName" graphql:"noinput"`
	Token            string     `json:"token" graphql:"noinput"`
	Access           RepoAccess `json:"access"`
	Expiration       Expiration `json:"expiration"`
	Name             string     `json:"name"`
	UserName         string     `json:"username"`
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
