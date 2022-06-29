package graph

import (
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/apps/finance/internal/domain"
	"kloudlite.io/pkg/repos"
	"time"
)

func AccountModelFromEntity(account *domain.Account) *model.Account {
	return &model.Account{
		ID:   account.Id,
		Name: account.Name,
		Billing: &model.Billing{
			StripeCustomerID: account.Billing.StripeCustomerId,
			CardholderName:   account.Billing.CardholderName,
			Address:          account.Billing.Address,
		},
		IsActive:     account.IsActive,
		ContactEmail: account.ContactEmail,
		ReadableID:   account.ReadableId,
		Created:      account.CreatedAt.String(),
	}
}

func currentMonthBillingModelFromBillables(startDate time.Time, accountId repos.ID, billableEntities []*domain.Billable) *model.CurrentMonthBilling {
	billables := make([]*model.Billable, 0)
	for _, b := range billableEntities {
		startTime := b.StartTime.String()
		var endTimeStr string
		if b.EndTime != nil {
			endTimeStr = b.EndTime.String()
		}
		billables = append(billables, &model.Billable{
			AccountID:    b.AccountId,
			ResourceType: b.ResourceType,
			Quantity:     float64(b.Quantity),
			StartTime:    &startTime,
			EndTime:      &endTimeStr,
		})
	}
	return &model.CurrentMonthBilling{
		AccountID: accountId,
		Billables: billables,
		StartDate: startDate.String(),
	}
}

func ComputeInventoryItemFromEntity(item *domain.InventoryItem) *model.ComputeInventoryItem {
	var modelPricePerHour *model.ItemPrice
	if item.PricePerHour != nil {
		modelPricePerHour = &model.ItemPrice{
			Quantity: item.PricePerHour.Quantity,
			Currency: item.PricePerHour.Currency,
		}
	}

	return &model.ComputeInventoryItem{
		Name:         item.Name,
		Provider:     item.Provider,
		Desc:         item.Desc,
		PricePerHour: modelPricePerHour,
		PricePerMonth: &model.ItemPrice{
			Quantity: item.PricePerMonth.Quantity,
			Currency: item.PricePerMonth.Currency,
		},
	}
}
