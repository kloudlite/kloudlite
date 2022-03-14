package app

import (
	"fmt"
	"net/http"

	"kloudlite.io/apps/message-consumer/internal/domain"
	"kloudlite.io/pkg/errors"
)

const (
	ResourceProject    = "project"
	ResourceAccount    = "account"
	ResourceApp        = "app"
	ResourceConfig     = "config"
	ResourceSecret     = "secret"
	ResourceRouter     = "secret"
	ResourceManagedSvc = "managed-svc"
	ResourceManagedRes = "managed-res"
)

const (
	ActionCreate = "create"
	ActionDelete = "delete"
	ActionUpdate = "update"
)

type appI struct {
	svc     domain.DomainSvc
	httpCli *http.Client
	gql     *GqlClient
}

func (app *appI) Handle(msg *Message) (e error) {
	defer errors.HandleErr(&e)

	switch msg.ResourceType {

	case ResourceAccount:
		fmt.Println("Currently, does not suppport account resource")
		return

	case ResourceProject:
		switch msg.Action {
		case ActionCreate, ActionUpdate:
			return app.svc.ApplyProject(msg.ResourceId)
		case ActionDelete:
			return app.svc.DeleteProject(msg.ResourceId)
		default:
			return fmt.Errorf("Unknown action (%s)", msg.Action)
		}

	case ResourceApp:
		switch msg.Action {
		case ActionCreate, ActionUpdate:
			return app.svc.ApplyApp(msg.ResourceId)
		case ActionDelete:
			return app.svc.DeleteApp(msg.ResourceId)
		default:
			return fmt.Errorf("Unknown action (%s)", msg.Action)
		}

	case ResourceConfig:
		switch msg.Action {
		case ActionCreate, ActionUpdate:
			return app.svc.ApplyConfig(msg.ResourceId)
		case ActionDelete:
			return app.svc.DeleteConfig(msg.ResourceId)
		default:
			return fmt.Errorf("Unknown action (%s)", msg.Action)
		}

	case ResourceSecret:
		switch msg.Action {
		case ActionCreate, ActionUpdate:
			return app.svc.ApplySecret(msg.ResourceId)
		case ActionDelete:
			return app.svc.DeleteSecret(msg.ResourceId)
		default:
			return fmt.Errorf("Unknown action (%s)", msg.Action)
		}

	case ResourceManagedSvc:
		switch msg.Action {
		case ActionCreate:
			return fmt.Errorf("Not Implemented Yet")
		case ActionUpdate:
			return fmt.Errorf("Not Implemented Yet")
		case ActionDelete:
			return fmt.Errorf("Not Implemented Yet")
		default:
			return fmt.Errorf("Unknown action (%s)", msg.Action)
		}

	case ResourceManagedRes:
		switch msg.Action {
		case ActionCreate:
			return fmt.Errorf("Not Implemented Yet")
			// return app.svc.ApplySecret(msg.ResourceId)
		case ActionUpdate:
			return fmt.Errorf("Not Implemented Yet")
		case ActionDelete:
			return fmt.Errorf("Not Implemented Yet")
		default:
			return fmt.Errorf("Unknown action (%s)", msg.Action)
		}
	}

	return nil
}

func MakeApp(kApplier *domain.K8sApplier, gqlClient *GqlClient) App {
	return &appI{
		svc:     domain.MakeDomain(kApplier),
		httpCli: &http.Client{},
		gql:     gqlClient,
	}
}
