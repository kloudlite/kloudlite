package domain

import (
	"kloudlite.io/constants"
	"time"

	"kloudlite.io/pkg/repos"
)

type AccountInviteToken struct {
	Token       string   `json:"token"`
	UserId      repos.ID `json:"user_id"`
	Role        string   `json:"role"`
	AccountName repos.ID `json:"account_name"`
}

type Billing struct {
	StripeCustomerId string         `json:"stripe_customer_id" bson:"stripe_customer_id"`
	PaymentMethodId  string         `json:"payment_method_id" bson:"payment_method_id"`
	CardholderName   string         `json:"cardholder_name" bson:"cardholder_name"`
	Address          map[string]any `json:"address" bson:"address"`
}

type Account struct {
	repos.BaseEntity `bson:",inline"`
	Name             string    `json:"name,omitempty" bson:"name,omitempty"`
	ContactEmail     string    `bson:"contact_email" json:"contact_email,omitempty"`
	Billing          Billing   `json:"billing" bson:"billing"`
	IsActive         bool      `json:"is_active,omitempty" bson:"is_active"`
	IsDeleted        bool      `json:"is_deleted,omitempty" bson:"is_deleted"`
	CreatedAt        time.Time `json:"created_at" bson:"created_at"`
	ReadableId       repos.ID  `json:"readable_id" bson:"readable_id"`
	ClusterID        repos.ID  `json:"cluster_id" bson:"cluster_id"`
}

var AccountIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

type Membership struct {
	AccountName string
	UserId      repos.ID
	Role        constants.Role
	Accepted    bool
}

type Billable struct {
	ResourceType string  `json:"resource_type" bson:"resource_type"`
	Plan         string  `json:"plan" bson:"plan"`
	IsShared     bool    `json:"is_shared" bson:"is_shared"`
	Quantity     float64 `json:"quantity" bson:"quantity"`
	Count        int     `json:"count" bson:"count"`
}

type AccountBilling struct {
	repos.BaseEntity `bson:",inline"`
	AccountName      string     `json:"account_id" bson:"account_id"`
	ProjectId        repos.ID   `json:"project_id" bson:"project_id"`
	ResourceId       repos.ID   `json:"resource_id" bson:"resource_id"`
	Billables        []Billable `json:"billables" bson:"billables"`
	StartTime        time.Time  `json:"start_time" bson:"start_time"`
	EndTime          *time.Time `json:"end_time" bson:"end_time"`
	BillAmount       float64    `json:"bill_amount" bson:"bill_amount"`
	Month            *string    `json:"month" bson:"month"`
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
			{Key: "account_name", Value: repos.IndexAsc},
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

type StoragePlan struct {
	Name       string  `yaml:"name"`
	PricePerGB float64 `yaml:"pricePerGB"`
}

type BillingEvent struct {
	Key     string `json:"key"`
	Stage   string `json:"stage"`
	Billing struct {
		Name  string `json:"name"`
		Items []struct {
			Type     string  `json:"type"`
			Count    int     `json:"count"`
			Plan     string  `json:"plan"`
			PlanQ    float64 `json:"planQuantity"`
			IsShared string  `json:"isShared"`
		} `json:"items"`
	} `json:"billing"`
	Metadata struct {
		ClusterId        string `json:"clusterId"`
		AccountName      string `json:"accountName"`
		ProjectId        string `json:"projectId"`
		ResourceId       string `json:"resourceId"`
		GroupVersionKind struct {
			Group   string `json:"Group"`
			Version string `json:"Version"`
			Kind    string `json:"Kind"`
		} `json:"groupVersionKind"`
	} `json:"metadata"`
}
