package domain

import (
	"context"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/accounts"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccountsSvc interface {
	GetAccount(ctx context.Context, userId string, accountName string) (*accounts.GetAccountOut, error)
}

type ResourceDispatcher interface {
	ApplyToTargetCluster(ctx InfraContext, dispatchAddr *entities.DispatchAddr, obj client.Object, recordVersion int) error
	DeleteFromTargetCluster(ctx InfraContext, dispatchAddr *entities.DispatchAddr, obj client.Object) error
}

type PublishMsg string

const (
	PublishAdd    PublishMsg = "added"
	PublishDelete PublishMsg = "deleted"
	PublishUpdate PublishMsg = "updated"
)

type ResourceEventPublisher interface {
	PublishInfraEvent(ctx InfraContext, resourceType ResourceType, resName string, update PublishMsg)
	PublishResourceEvent(ctx InfraContext, clusterName string, resourceType ResourceType, resName string, update PublishMsg)
}
