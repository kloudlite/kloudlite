package domain

import (
	"context"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
)

type AccountsSvc interface {
	GetAccount(ctx context.Context, userId string, accountName string) (*accounts.GetAccountOut, error)
}
