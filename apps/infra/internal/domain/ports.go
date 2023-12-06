package domain

import (
	"context"

	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/accounts"
)

type AccountsSvc interface {
	GetAccount(ctx context.Context, userId string, accountName string) (*accounts.GetAccountOut, error)
}
