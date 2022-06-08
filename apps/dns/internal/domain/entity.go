package domain

import "kloudlite.io/pkg/repos"

type Verification struct {
	repos.BaseEntity `bson:",inline"`
	AccountId        repos.ID `json:"accountId,omitempty"`
	SiteId           repos.ID `json:"siteId,omitempty"`
	VerifyText       string   `json:"verifyText,omitempty"`
}

type Site struct {
	repos.BaseEntity `bson:",inline"`
	AccountId        repos.ID `json:"accountId,omitempty"`
	Domain           string   `json:"host,omitempty"`
	Verified         bool     `json:"verified,omitempty"`
}

type Record struct {
	repos.BaseEntity `bson:",inline"`
	SiteId           repos.ID `json:"siteId,omitempty"`
	Type             string   `json:"type,omitempty"`
	Host             string   `json:"host,omitempty"`
	Answer           string   `json:"answer,omitempty"`
	TTL              uint32   `json:"ttl,omitempty"`
	Priority         int64    `json:"priority,omitempty"`
}

var RecordIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "Host", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
