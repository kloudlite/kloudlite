package entities

import (
	"kloudlite.io/pkg/repos"
)

type Tag struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	Tags        []string        `json:"tags" graphql:"noinput"`
	AccountName string          `json:"accountName" graphql:"noinput"`
	Repository  string          `json:"repository" graphql:"noinput"`
	Actor       string          `json:"actor" graphql:"noinput"`
	Digest      string          `json:"digest" graphql:"noinput"`
	Size        int             `json:"size" graphql:"noinput"`
	Length      int             `json:"length" graphql:"noinput"`
	MediaType   string          `json:"mediaType" graphql:"noinput"`
	URL         string          `json:"url" graphql:"noinput"`
	References  []RepoReference `json:"references" graphql:"noinput"`
	Deleting    bool            `json:"deleting" graphql:"noinput"`
}

var TagIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "digest", Value: repos.IndexAsc},
			{Key: "repository", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
