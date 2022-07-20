package domain

import (
	"context"
	"encoding/json"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/console"
	"kloudlite.io/pkg/repos"
	"strings"
	"time"
)

func (d *domainI) GenerateBillingInvoice(ctx context.Context, accountId repos.ID) (*BillingInvoice, error) {
	currentMonth := time.Now().Month().String()
	currentYear := time.Now().Year()
	one, err := d.invoiceRepo.FindOne(ctx, repos.Filter{
		"account_id": accountId,
		"month":      currentMonth,
		"year":       currentYear,
	})
	if err != nil {
		return nil, err
	}
	if one != nil {
		return one, nil
	}

	accountBillings, err := d.billablesRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"account_id": accountId,
			"month":      nil,
		},
	})

	type ProjStruct struct {
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
	}

	invoice := &BillingInvoice{
		AccountId:  accountId,
		Amount:     0,
		CreatedAt:  time.Time{},
		Month:      currentMonth,
		Year:       currentYear,
		Projects:   map[repos.ID]ProjectInvoice{},
		BaseEntity: repos.BaseEntity{},
	}

	type App struct {
		Name      string   `json:"name" bson:"name"`
		Id        repos.ID `json:"id" bson:"id"`
		ProjectId repos.ID `json:"project_id" bson:"project_id"`
	}

	var billableTotal float64
	for _, ab := range accountBillings {
		if strings.HasPrefix(string(ab.ResourceId), "app-") {
			app := App{}
			appOut, err := d.consoleCli.GetApp(ctx, &console.AppIn{
				AppId: string(ab.Id),
			})
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(appOut.Data.Value, &app)
			if err != nil {
				return nil, err
			}
			projectOut, err := d.consoleCli.GetProjectName(ctx, &console.ProjectIn{
				ProjectId: string(app.ProjectId),
			})
			if err != nil {
				return nil, err
			}
			if _, ok := invoice.Projects[app.ProjectId]; !ok {
				invoice.Projects[app.ProjectId] = ProjectInvoice{
					ProjectName: projectOut.Name,
				}
			}
			if _, ok := invoice.Projects[app.ProjectId].ResourceBillings[ab.ResourceId]; !ok {
				invoice.Projects[app.ProjectId].ResourceBillings[ab.ResourceId] = ResourceInvoice{
					ReadableName: app.Name,
				}
			}
		}

		if strings.HasPrefix(string(ab.ResourceId), "msvc-") {

		}

		if ab.EndTime == nil {
			bill, err := d.calculateBill(ctx, ab.Billables, ab.StartTime, time.Now())
			if err != nil {
				return nil, err
			}
			billableTotal = billableTotal + bill
		} else {
			billableTotal = billableTotal + ab.BillAmount
		}
	}
	return nil, nil

}
