package app

import (
	"fmt"
	"github.com/kloudlite/api/apps/container-registry/internal/domain"
	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/nats"
)

type ResourceEventPublisherImpl struct {
	cli    *nats.Client
	logger logging.Logger
}

func (r *ResourceEventPublisherImpl) PublishBuildRunEvent(buildrun *entities.BuildRun, msg domain.PublishMsg) {
	subject := clusterBuildRunUpdateSubject(buildrun)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishBuildCacheEvent(buildCache *entities.BuildCacheKey, msg domain.PublishMsg) {
	subject := clusterBuildCacheUpdateSubject(buildCache)
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

func clusterBuildRunUpdateSubject(buildRun *entities.BuildRun) string {
	return fmt.Sprintf("res-updates.account.%s.repo.%s.build-run.%s",
		buildRun.AccountName,
		buildRun.Spec.Registry.Repo.Name,
		buildRun.Id,
	)
}

func clusterBuildCacheUpdateSubject(buildCache *entities.BuildCacheKey) string {
	return fmt.Sprintf("res-updates.account.%s.id.%s",
		buildCache.AccountName,
		buildCache.Id,
	)
}
