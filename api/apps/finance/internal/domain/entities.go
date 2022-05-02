package domain

import (
	"time"

	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type Billing struct {
	StripeCustomerId    string
	StripeSetupIntentId string
	CardholderName      string
	Address             map[string]any
}

type Account struct {
	repos.BaseEntity `bson:",inline" json:"repos_._base_entity"`
	Name             string    `json:"name,omitempty" bson:"name,omitempty"`
	ContactEmail     string    `bson:"contact_email" json:"contact_email,omitempty"`
	Billing          Billing   `json:"billing" bson:"billing"`
	IsActive         bool      `json:"is_active,omitempty" bson:"is_active"`
	IsDeleted        bool      `json:"is_deleted,omitempty" bson:"is_deleted"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
	ReadableId       repos.ID  `json:"readable_id" bson:"readable_id"`
}

type Membership struct {
	AccountId repos.ID
	UserId    repos.ID
	Role      common.Role
	Accepted  bool
}

var AccountIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
