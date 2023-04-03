package domain

import (
	"time"

	"kloudlite.io/pkg/repos"
)

type ProjectInvoice struct {
	ProjectName      string                       `json:"project_name"`
	BillAmount       float64                      `json:"bill_amount" bson:"bill_amount"`
	ResourceBillings map[repos.ID]ResourceInvoice `json:"resource_billings" bson:"resource_billings"`
}

type ResourceInvoice struct {
	ReadableName string                        `json:"readable_name" bson:"readable_name"`
	BillAmount   float64                       `json:"bill_amount" bson:"bill_amount"`
	Consumptions map[string]ConsumptionInvoice `json:"consumptions" bson:"consumptions"`
}

type ConsumptionInvoice struct {
	Plan     float64 `json:"plan" bson:"plan"`
	Size     float64 `json:"size" bson:"size"`
	Quantity float64 `json:"quantity" bson:"quantity"`
	Duration int     `json:"duration" bson:"duration"`
	Amount   float64 `json:"amount" bson:"amount"`
}

type BillingInvoice struct {
	AccountId        repos.ID                    `json:"account_id" bson:"account_id"`
	Amount           float64                     `json:"amount" bson:"amount"`
	CreatedAt        time.Time                   `json:"created_at" bson:"created_at"`
	Month            string                      `json:"month" bson:"month"`
	Year             int                         `json:"year" bson:"year"`
	Projects         map[repos.ID]ProjectInvoice `json:"projects" bson:"projects"`
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
