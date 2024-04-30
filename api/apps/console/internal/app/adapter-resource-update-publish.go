package app

import (
	"fmt"

	"github.com/kloudlite/api/apps/console/internal/domain"
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
)

type ResourceEventPublisherImpl struct {
	cli    *nats.Client
	logger logging.Logger
}

func (r *ResourceEventPublisherImpl) PublishEnvironmentResourceEvent(ctx domain.ConsoleContext, envName string, resourceType entities.ResourceType, name string, update domain.PublishMsg) {
	subject := fmt.Sprintf("res-updates.account.%s.environment.%s.%s.%s", ctx.AccountName, envName, resourceType, name)
	r.publish(subject, update)
}

func (r *ResourceEventPublisherImpl) PublishConsoleEvent(ctx domain.ConsoleContext, resourceType entities.ResourceType, name string, update domain.PublishMsg) {
	subject := fmt.Sprintf("res-updates.account.%s.%s.%s", ctx.AccountName, resourceType, name)
	r.publish(subject, update)
}

func (r *ResourceEventPublisherImpl) PublishResourceEvent(ctx domain.ResourceContext, resourceType entities.ResourceType, name string, update domain.PublishMsg) {
	subject := fmt.Sprintf("res-updates.account.%s.environment.%s.%s.%s", ctx.AccountName, ctx.EnvironmentName, resourceType, name)
	r.publish(subject, update)
}

func (r *ResourceEventPublisherImpl) publish(subject string, msg domain.PublishMsg) {
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func NewResourceEventPublisher(cli *nats.Client, logger logging.Logger) domain.ResourceEventPublisher {
	return &ResourceEventPublisherImpl{cli, logger}
}
