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

func (r *ResourceEventPublisherImpl) PublishMsvcEvent(msvc *entities.ManagedService, msg domain.PublishMsg) {
	subject := msvcUpdateSubject(msvc)
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

func (r *ResourceEventPublisherImpl) PublishWorkspaceEvent(workspace *entities.Workspace, msg domain.PublishMsg) {
	subject := workSpaceUpdateSubject(workspace)
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
	return fmt.Sprintf("res-updates.account.%s.cluster.%s.app.%s", app.AccountName, app.ClusterName, app.Name)
}

func mresUpdateSubject(mres *entities.ManagedResource) string {
	return fmt.Sprintf("res-updates.account.%s.cluster.%s.mres.%s", mres.AccountName, mres.ClusterName, mres.Name)
}

func msvcUpdateSubject(msvc *entities.ManagedService) string {
	return fmt.Sprintf("res-updates.account.%s.cluster.%s.msvc.%s", msvc.AccountName, msvc.ClusterName, msvc.Name)
}

func projectUpdateSubject(project *entities.Project) string {
	return fmt.Sprintf("res-updates.account.%s.cluster.%s.project.%s", project.AccountName, project.ClusterName, project.Name)
}

func routerUpdateSubject(router *entities.Router) string {
	return fmt.Sprintf("res-updates.account.%s.cluster.%s.router.%s", router.AccountName, router.ClusterName, router.Name)
}

func workSpaceUpdateSubject(workspace *entities.Workspace) string {
	return fmt.Sprintf("res-updates.account.%s.cluster.%s.workspace.%s", workspace.AccountName, workspace.ClusterName, workspace.Name)
}
