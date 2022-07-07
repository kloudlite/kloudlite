package domain

import "kloudlite.io/pkg/repos"

type AccountDNS struct {
	repos.BaseEntity `bson:",inline"`
	AccountId        repos.ID `bson:"accountId",json:"accountId"`
	Hosts            []string `bson:"hosts",json:"hosts"`
}

type SiteClaim struct {
	repos.BaseEntity `bson:",inline"`
	AccountId        repos.ID `bson:"accountId,omitempty" json:"accountId,omitempty"`
	SiteId           repos.ID `bson:"siteId,omitempty" json:"siteId,omitempty"`
	Attached         bool     `bson:"attached,omitempty" json:"attached,omitempty"`
}

type Site struct {
	repos.BaseEntity `bson:",inline"`
	AccountId        repos.ID `bson:"accountId,omitempty" json:"accountId,omitempty"`
	Domain           string   `bson:"host,omitempty" json:"host,omitempty"`
}

type Record struct {
	repos.BaseEntity `bson:",inline"`
	SiteId           repos.ID `bson:"siteId,omitempty" json:"siteId,omitempty"`
	Type             string   `bson:"type,omitempty" json:"type,omitempty"`
	Host             string   `bson:"host,omitempty" json:"host,omitempty"`
	Answer           string   `bson:"answer,omitempty" json:"answer,omitempty"`
	TTL              uint32   `bson:"ttl,omitempty" json:"ttl,omitempty"`
	Priority         int64    `bson:"priority,omitempty" json:"priority,omitempty"`
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
			{Key: "host", Value: repos.IndexAsc},
			{Key: "type", Value: repos.IndexAsc},
			{Key: "answer", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

var SiteIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "domain", Value: repos.IndexAsc},
		},
	},
}

var AccountDNSIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountId", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

var SiteClaimIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountId", Value: repos.IndexAsc},
			{Key: "siteId", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
