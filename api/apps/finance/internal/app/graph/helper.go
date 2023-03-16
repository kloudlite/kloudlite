package graph

import (
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/apps/finance/internal/domain"
)

func AccountModelFromEntity(account *domain.Account) *model.Account {
	if account == nil {
		return nil
	}
	return &model.Account{
		ID:   account.Id,
		Name: account.Name,
		Billing: &model.Billing{
			CardholderName: account.Billing.CardholderName,
			Address:        account.Billing.Address,
		},
		IsActive:     account.IsActive,
		ContactEmail: account.ContactEmail,
		ReadableID:   account.ReadableId,
		Created:      account.CreatedAt.String(),
	}
}
