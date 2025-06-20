package domain

import (
	"embed"
	"github.com/kloudlite/api/apps/comms/internal/domain/entities"
	"github.com/kloudlite/api/apps/comms/internal/env"
	"github.com/kloudlite/api/apps/comms/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"log/slog"

	"github.com/kloudlite/api/pkg/mail"
	"github.com/kloudlite/api/pkg/repos"
	"go.uber.org/fx"
)

//go:embed email-templates
var TemplatesDir embed.FS

type Impl struct {
	notificationRepo       repos.DbRepo[*types.Notification]
	subscriptionRepo       repos.DbRepo[*entities.Subscription]
	notificationConfigRepo repos.DbRepo[*entities.NotificationConf]

	iamClient iam.IAMClient
	envs      *env.CommsEnv
	logger    *slog.Logger

	eTemplates *EmailTemplates

	resourceEventPublisher ResourceEventPublisher

	mailer mail.Mailer

	// CommsServer CommsServer
}

var Module = fx.Module("domain", fx.Provide(func(e *env.CommsEnv,
	notificationRepo repos.DbRepo[*types.Notification],
	subscriptionRepo repos.DbRepo[*entities.Subscription],
	notificationConfigRepo repos.DbRepo[*entities.NotificationConf],

	logger *slog.Logger,

	resourceEventPublisher ResourceEventPublisher,

	mailer mail.Mailer,
) (Domain, error) {
	eTemplates, err := GetEmailTemplates(EmailTemplatesDir{
		FS: TemplatesDir,
	})
	if err != nil {
		return nil, err
	}
	return &Impl{
		envs:                   e,
		logger:                 logger,
		resourceEventPublisher: resourceEventPublisher,

		notificationRepo:       notificationRepo,
		subscriptionRepo:       subscriptionRepo,
		notificationConfigRepo: notificationConfigRepo,
		mailer:                 mailer,
		eTemplates:             eTemplates,
	}, nil
}))
