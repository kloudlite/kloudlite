package adapters

import (
	"context"
	"time"

	"github.com/kloudlite/operator/pkg/errors"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
)

type accountsSvc struct {
	accountsRPC accounts.AccountsClient
}

// GetAccount implements domain.AccountsSvc.
func (as *accountsSvc) GetAccountRegion(ctx context.Context, userId string, accountName string) (string, error) {
	nctx, cf := context.WithTimeout(ctx, 2*time.Second)
	defer cf()
	out, err := as.accountsRPC.GetAccount(nctx, &accounts.GetAccountIn{
		UserId:      userId,
		AccountName: accountName,
	})
	if err != nil {
		return "", errors.NewE(err)
	}

	return out.GetKloudliteGatewayRegion(), nil
}

func NewAccountsSvc(accountsClient accounts.AccountsClient) domain.AccountsSvc {
	return &accountsSvc{
		accountsRPC: accountsClient,
	}
}

var _ domain.AccountsSvc = (*accountsSvc)(nil)
