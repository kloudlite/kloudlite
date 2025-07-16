package app

import (
	"github.com/kloudlite/api/apps/comms/internal/domain"
	"github.com/kloudlite/api/pkg/nats"
	"log/slog"
)

type ResourceEventPublisherImpl struct {
	cli    *nats.Client
	logger *slog.Logger
}

func (r *ResourceEventPublisherImpl) publish(subject string, msg domain.PublishMsg) {
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Error(err.Error(), "failed to publish message to subject %q", subject)
	}
}

func NewResourceEventPublisher(cli *nats.Client, logger *slog.Logger) domain.ResourceEventPublisher {
	return &ResourceEventPublisherImpl{cli, logger}
}
