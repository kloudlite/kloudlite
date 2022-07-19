package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
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

	//invoice := &BillingInvoice{
	//	AccountId:  accountId,
	//	Amount:     0,
	//	CreatedAt:  time.Time{},
	//	Month:      currentMonth,
	//	Year:       currentYear,
	//	Projects:   nil,
	//	BaseEntity: repos.BaseEntity{},
	//}

	var billableTotal float64
	for _, ab := range accountBillings {
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
