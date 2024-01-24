package app

import (
	"fmt"
	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
)

type ResourceEventPublisherImpl struct {
	cli    *nats.Client
	logger logging.Logger
}

func (r *ResourceEventPublisherImpl) PublishInfraEvent(ctx domain.InfraContext, resourceType domain.ResourceType, resName string, update domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.resourceType.%s.%s",
		ctx.AccountName, resourceType, resName,
	)

	r.publish(subject, update)
}

func (r *ResourceEventPublisherImpl) PublishResourceEvent(ctx domain.InfraContext, clusterName string, resourceType domain.ResourceType, resName string, update domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.cluster.%s.%s.%s",
		ctx.AccountName, clusterName, resourceType, resName,
	)

	r.publish(subject, update)
}

func (r *ResourceEventPublisherImpl) publish(subject string, msg domain.PublishMsg) {
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func NewResourceEventPublisher(cli *nats.Client, logger logging.Logger) domain.ResourceEventPublisher {
	return &ResourceEventPublisherImpl{
		cli,
		logger,
	}
}
