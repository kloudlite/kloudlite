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
	StripeCustomerId string         `json:"stripe_customer_id" bson:"stripe_customer_id"`
	PaymentMethodId  string         `json:"payment_method_id" bson:"payment_method_id"`
	CardholderName   string         `json:"cardholder_name" bson:"cardholder_name"`
	Address          map[string]any `json:"address" bson:"address"`
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
	ResourceType string  `json:"resource_type" bson:"resource_type"`
	Plan         string  `json:"plan" bson:"plan"`
	IsShared     bool    `json:"is_shared" bson:"is_shared"`
	Quantity     float64 `json:"quantity" bson:"quantity"`
	Count        int     `json:"count" bson:"count"`
}

type AccountBilling struct {
	repos.BaseEntity `bson:",inline"`
	AccountId        repos.ID   `json:"account_id" bson:"account_id"`
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
			Type     string `json:"type"`
			Count    int    `json:"count"`
			Plan     string `json:"plan"`
			PlanQ    string `json:"planQ"`
			IsShared string `json:"isShared"`
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

type BillingInvoice struct {
	AccountId repos.ID  `json:"account_id" bson:"account_id"`
	Amount    float64   `json:"amount" bson:"amount"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	Month     string    `json:"month" bson:"month"`
	Year      int       `json:"year" bson:"year"`
	Projects  map[repos.ID]struct {
		BillAmount       float64 `json:"bill_amount" bson:"bill_amount"`
		ResourceBillings map[repos.ID]struct {
			ReadableName string  `json:"readable_name" bson:"readable_name"`
			BillAmount   float64 `json:"bill_amount" bson:"bill_amount"`
			Consumptions map[string]struct {
				Plan     float64 `json:"plan" bson:"plan"`
				Size     float64 `json:"size" bson:"size"`
				Quantity float64 `json:"quantity" bson:"quantity"`
				Duration int     `json:"duration" bson:"duration"`
				Amount   float64 `json:"amount" bson:"amount"`
			} `json:"consumptions" bson:"consumptions"`
		} `json:"resource_billings" bson:"resource_billings"`
	} `json:"projects" bson:"projects"`
	repos.BaseEntity `bson:",inline"`
}

var BillingInvoiceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "account_id", Value: repos.IndexAsc},
			{Key: "month", Value: repos.IndexAsc},
			{Key: "year", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "account_id", Value: repos.IndexAsc},
		},
	},
}
