package domain

import "kloudlite.io/pkg/repos"

type MessageOfficeToken struct {
	repos.BaseEntity `json:",inline"`
	AccountName      string `json:"accountName"`
	ClusterName      string `json:"clusterName"`
	Token            string `json:"token"`
}

var MOTokenIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "clusterName", Value: repos.IndexAsc},
			{Key: "token", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

type AccessToken struct {
	repos.BaseEntity `json:",inline"`
	AccountName      string `json:"accountName"`
	ClusterName      string `json:"clusterName"`
	AccessToken      string `json:"accessToken"`
}


var AccessTokenIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "clusterName", Value: repos.IndexAsc},
			{Key: "accessToken", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
