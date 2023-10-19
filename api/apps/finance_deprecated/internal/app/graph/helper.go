package graph

import (
	"context"
	"fmt"

	"kloudlite.io/apps/finance_deprecated/internal/app/graph/model"
	"kloudlite.io/apps/finance_deprecated/internal/domain"
	fn "kloudlite.io/pkg/functions"
)

func AccountModelFromEntity(account *domain.Account) *model.Account {
	if account == nil {
		return nil
	}
	return &model.Account{
		// ID:   account.Id,
		Name: account.Name,
		Billing: &model.Billing{
			CardholderName: account.Billing.CardholderName,
			Address:        account.Billing.Address,
		},
		IsActive:     fn.DefaultIfNil(account.IsActive, false),
		ContactEmail: account.ContactEmail,
		ReadableID:   account.ReadableId,
		Created:      account.CreatedAt.String(),
	}
}

func toFinanceContext(ctx context.Context) domain.FinanceContext {
	if cc, ok := ctx.Value("kl-finance-ctx").(domain.FinanceContext); ok {
		return cc
	}
	panic(fmt.Errorf("context values '%s' is missing", "kl-finance-ctx"))
}
