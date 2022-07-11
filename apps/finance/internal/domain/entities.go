package domain

import (
	"time"

	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type AccountInviteToken struct {
	Token     string   `json:"token"`
	UserId    repos.ID `json:"user_id"`
	Role      string   `json:"role"`
	AccountId repos.ID `json:"account_id"`
}

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

type Billable struct {
	ResourceType string `json:"resource_type" bson:"resource_type"`
	Plan         string `json:"plan" bson:"plan"`
	Quantity     string `json:"quantity" bson:"quantity"`
}

type AccountBilling struct {
	repos.BaseEntity `bson:",inline"`
	AccountId        repos.ID   `json:"account_id" bson:"account_id"`
	ProjectId        repos.ID   `json:"project_id" bson:"project_id"`
	ResourceId       repos.ID   `json:"resource_id" bson:"resource_id"`
	Billables        []Billable `json:"billables" bson:"billables"`
	StartTime        time.Time  `json:"start_time" bson:"start_time"`
	EndTime          *time.Time `json:"end_time" bson:"end_time"`
}

var BillableIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "account_id", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "resource_id", Value: repos.IndexAsc},
		},
	},
}

type ComputePlan struct {
	Name           string  `yaml:"name"`
	SharedPrice    float64 `yaml:"sharedPrice"`
	DedicatedPrice float64 `yaml:"dedicatedPrice"`
}

type LamdaPlan struct {
	Name         string  `yaml:"name"`
	PricePerGBHr float64 `yaml:"pricePerGBHr"`
	FreeTire     int     `yaml:"freeTire"`
}

type BillingEvent struct {
	Key         string `json:"key"`
	Termination bool   `json:"to_be_deleted"`
	Billing     struct {
		Name  string `json:"name"`
		Items []struct {
			Type  string `json:"type"`
			Count int    `json:"count"`
			Plan  string `json:"plan"`
			PlanQ string `json:"planQ"`
		} `json:"items"`
	} `json:"billing"`
	Metadata struct {
		ClusterId        string `json:"clusterId"`
		AccountId        string `json:"accountId"`
		ProjectId        string `json:"projectId"`
		ResourceId       string `json:"resourceId"`
		GroupVersionKind struct {
			Group   string `json:"Group"`
			Version string `json:"Version"`
			Kind    string `json:"Kind"`
		} `json:"groupVersionKind"`
	} `json:"metadata"`
}
