package entities

import "github.com/kloudlite/api/pkg/repos"

type KLoudliteEdgeCluster struct {
	repos.BaseEntity `json:",inline"`
	Region           string `json:"region"`
	Name             string `json:"name"`

	NumAccounts    int `json:"num_accounts"`
	MaxNumAccounts int `json:"max_num_accounts"`

	Comments string `json:"comments"`
}
