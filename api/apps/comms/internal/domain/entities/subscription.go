package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

//	type SubscriptionConf struct {
//		AlertsEnabled        bool `json:"alertsEnabled"`
//		NotificationsEnabled bool `json:"notificationsEnabled"`
//	}

type Subscription struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	CreatedBy     common.CreatedOrUpdatedBy `json:"createdBy" graphql:"noinput"`
	LastUpdatedBy common.CreatedOrUpdatedBy `json:"lastUpdatedBy" graphql:"noinput"`

	AccountName string `json:"accountName" graphql:"noinput"`
	MailAddress string `json:"mailAddress"`

	// Configurations *SubscriptionConf `json:"configurations"`
	Enabled bool `json:"enabled"`
}

var SubscriptionIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
