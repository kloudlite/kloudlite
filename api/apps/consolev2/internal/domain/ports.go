package domain

import (
	"context"

	fWebsocket "github.com/gofiber/websocket/v2"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	SetupAccount(ctx context.Context, accountId repos.ID) (bool, error)

	CreateCloudProvider(ctx context.Context, cloudProvider *entities.CloudProvider, creds entities.SecretData) (*entities.CloudProvider, error)
	UpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, creds entities.SecretData) (*entities.CloudProvider, error)
	DeleteCloudProvider(ctx context.Context, name string) error
	ListCloudProviders(ctx context.Context, accountId repos.ID) ([]*entities.CloudProvider, error)
	GetCloudProvider(ctx context.Context, name string) (*entities.CloudProvider, error)

	GetSocketCtx(
		conn *fWebsocket.Conn,
		cacheClient AuthCacheClient,
		cookieName,
		cookieDomain string,
		sessionKeyPrefix string,
	) context.Context
}

type AuthCacheClient cache.Client
type WorkloadMessenger interface {
	SendAction(action string, kafkaTopic string, resId string, res any) error
}
