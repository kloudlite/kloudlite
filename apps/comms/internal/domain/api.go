package domain

import (
	"context"

	"github.com/kloudlite/api/apps/comms/internal/domain/entities"
	"github.com/kloudlite/api/apps/comms/types"
	"github.com/kloudlite/api/pkg/repos"
)

type Domain interface {
	ListNotifications(ctx CommsContext, pagination repos.CursorPagination) (*repos.PaginatedRecord[*types.Notification], error)
	MarkNotificationAsRead(ctx CommsContext, id repos.ID) (*types.Notification, error)

	GetNotificationConfig(ctx CommsContext) (*entities.NotificationConf, error)
	UpdateNotificationConfig(ctx CommsContext, config entities.NotificationConf) (*entities.NotificationConf, error)

	UpdateSubscriptionConfig(ctx CommsContext, id repos.ID, config entities.Subscription) (*entities.Subscription, error)
	GetSubscriptionConfig(ctx CommsContext, id repos.ID) (*entities.Subscription, error)

	Notify(ctx context.Context, notification *types.Notification) error
}

type PublishMsg string

const (
	PublishAdd    PublishMsg = "added"
	PublishDelete PublishMsg = "deleted"
	PublishUpdate PublishMsg = "updated"
)

type ResourceEventPublisher interface {
	// PublishBuildNotification(cluster *types.Notification, msg PublishMsg)
}
