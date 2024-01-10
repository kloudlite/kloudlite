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

func (r *ResourceEventPublisherImpl) publish(subject string, msg domain.PublishMsg) {
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishProjectManagedServiceEvent(projectManagedService *entities.ProjectManagedService, msg domain.PublishMsg) {
	subject := fmt.Sprintf(
		"res-updates.account.%s.project.%s.app.%s",
		projectManagedService.AccountName,
		projectManagedService.ProjectName,
		projectManagedService.Name,
	)

	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishAppEvent(app *entities.App, msg domain.PublishMsg) {
	subject := fmt.Sprintf("res-updates.account.%s.project.%s.app.%s", app.AccountName, app.ProjectName, app.Name)
	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishMresEvent(mres *entities.ManagedResource, msg domain.PublishMsg) {
	subject := fmt.Sprintf("res-updates.account.%s.project.%s.mres.%s", mres.AccountName, mres.ProjectName, mres.Name)
	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishProjectEvent(project *entities.Project, msg domain.PublishMsg) {
	subject := fmt.Sprintf("res-updates.account.%s.project.%s.project.%s", project.AccountName, project.Name, project.Name)
	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishRouterEvent(router *entities.Router, msg domain.PublishMsg) {
	subject := fmt.Sprintf("res-updates.account.%s.project.%s.router.%s", router.AccountName, router.ProjectName, router.Name)

	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishWorkspaceEvent(env *entities.Environment, msg domain.PublishMsg) {
	subject := fmt.Sprintf("res-updates.account.%s.project.%s.environment.%s", env.AccountName, env.ProjectName, env.Name)
	r.publish(subject, msg)
}

func (r *ResourceEventPublisherImpl) PublishVpnDeviceEvent(device *entities.ConsoleVPNDevice, msg domain.PublishMsg) {
	subject := fmt.Sprintf("res-updates.account.%s.vpn-device.%s", device.AccountName, device.Name)

	r.publish(subject, msg)
}

func NewResourceEventPublisher(cli *nats.Client, logger logging.Logger) domain.ResourceEventPublisher {
	return &ResourceEventPublisherImpl{
		cli,
		logger,
	}
}
