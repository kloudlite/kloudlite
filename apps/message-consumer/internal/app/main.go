package app

import (
	"encoding/json"
	"fmt"
	"io"
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
	ResourceJob        = "job"
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

func (app *appI) getMsvcInstallation(installationId string) (source *domain.MsvcSource, e error) {
	defer errors.HandleErr(&e)
	req, e := app.gql.Request(`
	  query GetInstallation($installationId: ID!) {
    managedSvc {
			installation: getInstallation(installationId: $installationId) {
        source {
          id
          name
					operations {
            install
            uninstall
            update
          }
          resources {
            name
            operations {
              create
              delete
              update
            }
          }
        }
      }
    }
  }`, map[string]interface{}{
		"installationId": installationId,
	})

	errors.AssertNoError(e, fmt.Errorf("could not make graphql request because %v", e))
	resp, e := app.httpCli.Do(req)
	errors.AssertNoError(e, fmt.Errorf("failed while performing graphql gateway request because %v", e))

	respB, e := io.ReadAll(resp.Body)
	fmt.Println("response body: ", string(respB))
	errors.AssertNoError(e, fmt.Errorf("failed to read response bytes because %v", e))

	var j struct {
		Data struct {
			ManagedSvc struct {
				Installation struct {
					Source domain.MsvcSource
				} `json:"installation"`
			} `json:"managedSvc"`
		} `json:"data"`
	}

	e = json.Unmarshal(respB, &j)
	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal response bytes because %v", e))

	return &j.Data.ManagedSvc.Installation.Source, nil
}

func (app *appI) getMsvcResource(resId string) (source *domain.MsvcResource, e error) {
	defer errors.HandleErr(&e)
	req, e := app.gql.Request(
		`query GetResource($resId: ID!, $nextVersion: Boolean) {
			managedRes {
				getResource(resId: $resId, nextVersion: $nextVersion) {
					id
					name
					resourceName
					version
					installation {
						name
						project {
							name
						}
						source {
							resources {
								name
								operations {
									create
									update
									delete
								}
							}
						}
					}
					artifacts {
						type
						refName
						refKey
					}
				}
			}
		}`, map[string]interface{}{
			"resId": resId,
		})

	errors.AssertNoError(e, fmt.Errorf("could not make graphql request because %v", e))
	resp, e := app.httpCli.Do(req)
	errors.AssertNoError(e, fmt.Errorf("failed while performing graphql gateway request because %v", e))

	respB, e := io.ReadAll(resp.Body)
	errors.AssertNoError(e, fmt.Errorf("failed to read response bytes because %v", e))

	var j struct {
		Data struct {
			ManagedRes struct {
				GetResource struct {
					ResourceName string `json:"resourceName"`
					Installation struct {
						Project domain.Project    `json:"project"`
						Source  domain.MsvcSource `json:"source"`
					} `json:"installation"`
				} `json:"getResource"`
			} `json:"managedRes"`
		} `json:"data"`
	}

	e = json.Unmarshal(respB, &j)
	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal response bytes because %v", e))

	// fmt.Println("response body: ", string(respB), j.Data.ManagedRes.GetResource.Installation.Source.Resources)
	fmt.Println("ResourceId: ", resId)
	fmt.Println("ResourceName: ", j.Data.ManagedRes.GetResource.ResourceName)
	fmt.Println("Namespace: ", j.Data.ManagedRes.GetResource.Installation.Project.Name)

	for _, resource := range j.Data.ManagedRes.GetResource.Installation.Source.Resources {
		fmt.Println("RESOURCE: ", resource.Operations)
		fmt.Println("RESOURCE.name", resource.Name)
		if resource.Name == j.Data.ManagedRes.GetResource.ResourceName {
			return &resource, nil
		}
	}

	panic(fmt.Errorf("could not find resource %s in installation %s", j.Data.ManagedRes.GetResource.ResourceName, j.Data.ManagedRes.GetResource.Installation.Source.Name))
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
			source, e := app.getMsvcInstallation(msg.ResourceId)
			errors.AssertNoError(e, fmt.Errorf("failed to get managed service installation because %v", e))
			fmt.Println("docker image:", source)
			return app.svc.InstallManagedSvc(msg.ResourceId, source.Operations.Install)
		case ActionUpdate:
			source, e := app.getMsvcInstallation(msg.ResourceId)
			errors.AssertNoError(e, fmt.Errorf("failed to get managed service installation because %v", e))
			return app.svc.UpdateManagedSvc(msg.ResourceId, source.Operations.Update)
		case ActionDelete:
			source, e := app.getMsvcInstallation(msg.ResourceId)
			errors.AssertNoError(e, fmt.Errorf("failed to get managed service installation because %v", e))
			return app.svc.UninstallManagedSvc(msg.ResourceId, source.Operations.Uninstall)
		default:
			return fmt.Errorf("Unknown action (%s)", msg.Action)
		}

	case ResourceManagedRes:
		switch msg.Action {
		case ActionCreate:
			resource, e := app.getMsvcResource(msg.ResourceId)
			errors.AssertNoError(e, fmt.Errorf("failed to get managed resource source because %v", e))
			return app.svc.CreateManagedRes(msg.ResourceId, resource.Operations.Create)
		case ActionUpdate:
			resource, e := app.getMsvcResource(msg.ResourceId)
			errors.AssertNoError(e, fmt.Errorf("failed to get managed resource source because %v", e))
			return app.svc.UpdateManagedRes(msg.ResourceId, resource.Operations.Update)
		case ActionDelete:
			resource, e := app.getMsvcResource(msg.ResourceId)
			errors.AssertNoError(e, fmt.Errorf("failed to get managed resource source because %v", e))
			return app.svc.DeleteManagedRes(msg.ResourceId, resource.Operations.Delete)
		default:
			return fmt.Errorf("Unknown action (%s)", msg.Action)
		}

	case ResourceJob:

	return nil
}

func MakeApp(kApplier *domain.K8sApplier, gqlClient *GqlClient) App {
	return &appI{
		svc:     domain.MakeDomain(kApplier),
		httpCli: &http.Client{},
		gql:     gqlClient,
	}
}
