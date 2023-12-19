package domain

import (
	"context"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccountsSvc interface {
	GetAccount(ctx context.Context, userId string, accountName string) (*accounts.GetAccountOut, error)
}




type ResourceDispatcher interface {
	ApplyToTargetCluster(ctx InfraContext, clusterName string, obj client.Object, recordVersion int) error
	DeleteFromTargetCluster(ctx InfraContext, clusterName string, obj client.Object) error
}