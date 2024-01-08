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

func (r *ResourceEventPublisherImpl) PublishProjectManagedServiceEvent(projectManagedService *entities.ProjectManagedService, msg domain.PublishMsg) {
	subject := projectManagedServiceUpdateSubject(projectManagedService)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishAppEvent(app *entities.App, msg domain.PublishMsg) {
	subject := appUpdateSubject(app)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishMresEvent(mres *entities.ManagedResource, msg domain.PublishMsg) {
	subject := mresUpdateSubject(mres)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishProjectEvent(project *entities.Project, msg domain.PublishMsg) {
	subject := projectUpdateSubject(project)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishRouterEvent(router *entities.Router, msg domain.PublishMsg) {
	subject := routerUpdateSubject(router)
	if err := r.cli.Conn.Publish(subject, []byte(msg)); err != nil {
		r.logger.Errorf(err, "failed to publish message to subject %q", subject)
	}
}

func (r *ResourceEventPublisherImpl) PublishWorkspaceEvent(workspace *entities.Environment, msg domain.PublishMsg) {
	subject := environmentUpdateSubject(workspace)
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

func appUpdateSubject(app *entities.App) string {
	return fmt.Sprintf("res-updates.account.%s.project.%s.app.%s", app.AccountName, app.ProjectName, app.Name)
}

func projectManagedServiceUpdateSubject(projectManagedService *entities.ProjectManagedService) string {
	return fmt.Sprintf("res-updates.account.%s.project.%s.app.%s", projectManagedService.AccountName, projectManagedService.ProjectName, projectManagedService.Name)
}

func mresUpdateSubject(mres *entities.ManagedResource) string {
	return fmt.Sprintf("res-updates.account.%s.project.%s.mres.%s", mres.AccountName, mres.ProjectName, mres.Name)
}

func projectUpdateSubject(project *entities.Project) string {
	return fmt.Sprintf("res-updates.account.%s.project.%s.project.%s", project.AccountName, project.Name, project.Name)
}

func routerUpdateSubject(router *entities.Router) string {
	return fmt.Sprintf("res-updates.account.%s.project.%s.router.%s", router.AccountName, router.ProjectName, router.Name)
}

func environmentUpdateSubject(env *entities.Environment) string {
	return fmt.Sprintf("res-updates.account.%s.project.%s.environment.%s", env.AccountName, env.ProjectName, env.Name)
}
