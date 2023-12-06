package app

import (
	"context"
	"errors"
	"time"

	"kloudlite.io/apps/infra/internal/domain"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/accounts"
)

type accountsSvc struct {
	accountsClient accounts.AccountsClient
}

// GetAccount implements domain.AccountsSvc.
func (as *accountsSvc) GetAccount(ctx context.Context, userId string, accountName string) (*accounts.GetAccountOut, error) {
	ctx2, cf := context.WithTimeout(ctx, 2*time.Second)
	defer cf()
	out, err := as.accountsClient.GetAccount(ctx2, &accounts.GetAccountIn{
		UserId:      userId,
		AccountName: accountName,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, domain.ErrGRPCCall{Err: err}
		}
		return nil, err
	}

	return out, nil
}

func NewAccountsSvc(accountsClient accounts.AccountsClient) domain.AccountsSvc {
	return &accountsSvc{
		accountsClient: accountsClient,
	}
}
