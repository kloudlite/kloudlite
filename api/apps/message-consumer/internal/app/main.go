package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"kloudlite.io/apps/message-consumer/internal/domain"
	"kloudlite.io/pkg/errors"
)

type appI struct {
	svc     domain.DomainSvc
	httpCli *http.Client
	gql     *domain.GqlClient
}

func (app *appI) getJob(jobId string) (job *domain.Job, e error) {
	const q = `
	  query Job($jobId: ID!) {
    job(jobId: $jobId) {
      id
      project {
        name
      }
      actions {
        id
        action
        resourceId
        resourceType
        data
        kubeData
      }
      flows
      createdAt
      klStatus {
        state
        feed {
          type
          timestamp
          message
        }
      }
    }
  }
  `

	_, respB, e := app.gql.DoRequest(q, map[string]interface{}{"jobId": jobId})
	errors.AssertNoError(e, fmt.Errorf("could not do graphql request"))
	// fmt.Println("response body: ", string(respB))

	errors.AssertNoError(e, fmt.Errorf("could not read http response as %v", e))

	var j struct {
		Data struct {
			Job domain.Job `json:"job"`
		} `json:"data"`
	}

	e = json.Unmarshal(respB, &j)
	errors.AssertNoError(e, fmt.Errorf("could not unmarshal response as %v", e))

	return &j.Data.Job, nil
}

func (app *appI) Handle(msg *Message) (e error) {
	defer errors.HandleErr(&e)
	job, e := app.getJob(msg.JobId)
	errors.AssertNoError(e, fmt.Errorf("failed to get job because %v", e))

	// fmt.Printf("job action: length: %v, action: %v\n", len(job.Actions), job.Actions[0].Action)

	e = app.svc.ApplyJob(job)
	errors.AssertNoError(e, fmt.Errorf("could not apply job because %v", e))

	return nil
}

func MakeApp(kApplier *domain.K8sApplier, gqlClient *domain.GqlClient, httpClient *http.Client) App {
	return &appI{
		svc:     domain.MakeDomain(kApplier, gqlClient, httpClient),
		httpCli: &http.Client{},
		gql:     gqlClient,
	}
}
