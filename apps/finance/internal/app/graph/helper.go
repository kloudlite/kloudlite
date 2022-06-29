package graph

import (
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/apps/finance/internal/domain"
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
