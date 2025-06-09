package domain

import (
	"github.com/kloudlite/api/apps/comms/internal/domain/entities"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/apps/comms/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	// "github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/mail"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
)

type Impl struct {
	notificationRepo       repos.DbRepo[*types.Notification]
	subscriptionRepo       repos.DbRepo[*entities.Subscription]
	notificationConfigRepo repos.DbRepo[*entities.NotificationConf]

	iamClient iam.IAMClient
	envs      *env.Env
	logger    logging.Logger
	// cacheClient kv.BinaryDataRepo
	// authClient  auth.AuthClient

	eTemplates *EmailTemplates

	resourceEventPublisher ResourceEventPublisher

	mailer mail.Mailer

	// CommsServer CommsServer
}

var Module = fx.Module("domain", fx.Provide(func(e *env.Env,
	notificationRepo repos.DbRepo[*types.Notification],
	subscriptionRepo repos.DbRepo[*entities.Subscription],
	notificationConfigRepo repos.DbRepo[*entities.NotificationConf],

	logger logging.Logger,
	iamClient iam.IAMClient,
	// cacheClient kv.BinaryDataRepo,
	// authClient auth.AuthClient,
	resourceEventPublisher ResourceEventPublisher,

	eTemplates *EmailTemplates,
	mailer mail.Mailer,
) (Domain, error) {
	return &Impl{
		iamClient: iamClient,
		envs:      e,
		logger:    logger,
		// cacheClient: cacheClient,
		// authClient:             authClient,
		resourceEventPublisher: resourceEventPublisher,

		notificationRepo:       notificationRepo,
		subscriptionRepo:       subscriptionRepo,
		notificationConfigRepo: notificationConfigRepo,
		mailer:                 mailer,
		eTemplates:             eTemplates,
	}, nil
}))
