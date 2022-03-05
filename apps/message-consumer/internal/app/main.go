package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
	"kloudlite.io/apps/message-consumer/internal/domain"
	"kloudlite.io/pkg/errors"
)

const (
	ResourceProject = "project"
	ResourceAccount = "account"
	ResourceApp     = "app"
	ResourceConfig  = "config"
	ResourceSecret  = "secret"
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

//func HandleAccount(a *appI, msg *Message) error {
//	if msg.ResourceType != ResourceAccount {
//		return fmt.Errorf("invalid resource type: %s", msg.ResourceType)
//	}
//	switch msg.Action {
//	case ActionCreate, ActionUpdate:
//		//TODO: create account
//	case ActionDelete:
//		//TODO: delete account
//	}

//	return nil
//}

func HandleProject(app *appI, msg *Message) (e error) {
	defer errors.HandleErr(&e)
	errors.Assert(msg.ResourceType == ResourceProject, fmt.Errorf("invalid resource type: %s", msg.ResourceType))

	const projectQ = `
  query Project($projectId: ID!) {
    project(projectId: $projectId) {
      id
      name
			account {
      	id
      }
      displayName
      logo
      cluster
      description
    }
  }
	`

	switch msg.Action {
	case ActionCreate, ActionUpdate:
		req, e := app.gql.Request(projectQ, map[string]interface{}{
			"projectId": msg.ProjectId,
		})

		errors.AssertNoError(e, fmt.Errorf("failed to create request: %v", e))

		resp, e := app.httpCli.Do(req)
		errors.AssertNoError(e, fmt.Errorf("failed to query project because %v", e))

		respBytes, e := ioutil.ReadAll(resp.Body)
		errors.AssertNoError(e, fmt.Errorf("failed to read response body: %v", e))

		var d struct {
			Data struct {
				Project domain.Project `json:"project"`
			} `json:"data"`
		}
		e = json.Unmarshal(respBytes, &d)
		errors.AssertNoError(e, fmt.Errorf("failed to unmarshal response: %v", e))
		fmt.Println("project: ", d.Data.Project)

		e = app.svc.ApplyProject(&d.Data.Project)
		errors.AssertNoError(e, fmt.Errorf("failed to apply project: %v", e))

	case ActionDelete:
		return app.svc.DeleteProject(&domain.Project{
			Id: msg.Metadata["projectId"],
		})
	}

	return
}

func HandleApp(app *appI, msg *Message) (e error) {
	errors.Assert(msg.Metadata["resourceType"] == ResourceApp, fmt.Errorf("invalid resource type: %v", msg.Metadata["resourceType"]))
	const appQ = `
	  query App($appId: ID!) {
			app(appId: $appId) {
				name
				id
				containers {
					name
					imagePullPolicy
					image
					id
					env {
						key
						refKey
						type
						refName
						value
					}
				}
				namespace
				version
				project {
					id
					name
				}
				services {
					type
					targetPort
					port
				}
				replicas
			}
  }
	`
	switch msg.Action {
	case ActionCreate, ActionUpdate:
		req, e := app.gql.Request(appQ, map[string]interface{}{
			"appId": msg.Metadata["appId"],
		})

		errors.AssertNoError(e, fmt.Errorf("failed to create request: %v", e))

		resp, err := app.httpCli.Do(req)

		respB, e := ioutil.ReadAll(resp.Body)
		errors.AssertNoError(err, fmt.Errorf("failed to query app because %v", err))

		var d struct {
			data struct {
				App domain.App `json:"app"`
			} `json:"data"`
		}

		e = json.Unmarshal(respB, &d)
		errors.AssertNoError(e, fmt.Errorf("failed to unmarshal response: %v", respB))
		return app.svc.ApplyApp(&d.data.App)

	case ActionDelete:
		return app.svc.DeleteApp(&domain.App{
			Id: msg.Metadata["appId"],
		})

	default:
		return fmt.Errorf("invalid handle project action: %s", msg.Action)
	}
	return nil
}

func HandleConfig(app *appI, msg *Message) error {
	if msg.ResourceType != ResourceConfig {
		return fmt.Errorf("invalid resource type: %s", msg.ResourceType)
	}

	const configQ = `
  query Config($name: String!, $projectId: ID!) {
		project(projectId: $projectId) {
      name
		}

		query Config($name: String!, $projectId: ID!) {
			config(name: $name, projectId: $projectId) {
				name
				project {
					name
				}
				entries {
					key
					value
				}
			}
		}
  }`

	switch msg.Action {
	case ActionCreate, ActionUpdate:
		req, e := app.gql.Request(configQ, map[string]interface{}{
			"name":      msg.Metadata["name"],
			"projectId": msg.Metadata["projectId"],
		})

		log.Debugf("request: %v", map[string]interface{}{
			"query": configQ,
			"variables": map[string]interface{}{
				"name":      msg.Metadata["name"],
				"projectId": msg.Metadata["projectId"],
			},
		})

		resp, e := app.httpCli.Do(req)

		var d struct {
			Data struct {
				Project domain.Project `json:"project"`
				Config  domain.Config  `json:"config"`
			} `json:"data"`
		}

		respB, e := ioutil.ReadAll(resp.Body)
		errors.AssertNoError(e, fmt.Errorf("failed to query config because %v", e))
		fmt.Println("config response: ", string(respB))

		e = json.Unmarshal(respB, &d)
		errors.AssertNoError(e, fmt.Errorf("failed to unmarshal response: %v", e))
		fmt.Println("parsed config: ", d)

		return app.svc.ApplyConfig(&d.Data.Config, &d.Data.Project)
		return nil
	case ActionDelete:
		// TODO: delete config
	}
	return nil
}

func HandleSecret(app *appI, msg *Message) error {
	if msg.Metadata["resourceType"] != ResourceSecret {
		err := fmt.Errorf("invalid resource type: %s", msg.Metadata["resourceType"])
		fmt.Println("failed to handle secret: ", err)
		return err
	}
	switch msg.Action {
	case ActionCreate, ActionUpdate:
		return app.svc.ApplySecret(&domain.Secret{})
	case ActionDelete:
		return app.svc.ApplySecret(&domain.Secret{})
	}
	return nil
}

func (app *appI) Handle(msg *Message) (e error) {
	defer errors.HandleErr(&e)
	switch msg.ResourceType {
	case ResourceAccount:
		fmt.Println("Currently, does not suppport account resource")
		// return HandleAccount(app, msg)

	case ResourceProject:
		fmt.Println("project: ", msg.ProjectId)
		e = HandleProject(app, msg)
		errors.AssertNoError(e, fmt.Errorf("failed to handle project: %v", e))

	case ResourceApp:
		return HandleApp(app, msg)

	case ResourceConfig:
		return HandleConfig(app, msg)

	case ResourceSecret:
		return HandleSecret(app, msg)
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
