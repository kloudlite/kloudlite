package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	"kloudlite.io/pkg/functions"
	"sigs.k8s.io/yaml"
	"kloudlite.io/pkg/errors"
)

type domain struct {
	KubeApplyJob func(j *JobVars) error
	httpClient   *http.Client
	gql          *GqlClient
	Kube         *K8sApplier
}

func (d *domain) commitJob(jobId string) (result bool, e error) {
	defer errors.HandleErr(&e)

	const q = `
	  mutation Commit($jobId: ID!) {
    job {
      commit(jobId: $jobId)
    }
  }
	`
	_, respB, e := d.gql.DoRequest(q, map[string]interface{}{
		"jobId": jobId,
	})

	errors.AssertNoError(e, fmt.Errorf("could not do graphql request"))

	var j struct {
		Data struct {
			Job struct {
				Commit bool `json:"commit"`
			} `json:"job"`
		} `json:"data"`
	}

	e = json.Unmarshal(respB, &j)
	errors.AssertNoError(e, fmt.Errorf("failed while unmarshalling graphql response"))

	return j.Data.Job.Commit, nil
}

func (d *domain) undoJob(jobId string) (result bool, e error) {
	const q = `
	  mutation Commit($jobId: ID!) {
    job {
      undo(jobId: $jobId)
    }
  }
	`
	_, respB, e := d.gql.DoRequest(q, map[string]interface{}{
		"jobId": jobId,
	})
	errors.AssertNoError(e, fmt.Errorf("could not do graphql request"))

	var j struct {
		Data struct {
			Job struct {
				Undo bool `json:"undo"`
			} `json:"job"`
		} `json:"data"`
	}

	e = json.Unmarshal(respB, &j)
	errors.AssertNoError(e, fmt.Errorf("failed while unmarshalling graphql response"))

	return j.Data.Job.Undo, nil
}

func getDockerImage(actionRef JobAction) (dockerImage string, e error) {
	defer errors.HandleErr(&e)
	dockerImage, ok := ResourceImageMap[actionRef.ResourceType]
	if !ok {
		switch actionRef.ResourceType {
		case RESOURCE_MANAGED_SERVICE:
			{
				j, e := json.Marshal(actionRef.KubeData["template"])
				errors.AssertNoError(e, fmt.Errorf("could not marshal kubeData.template into []byte"))

				var template MsvcTemplate
				e = json.Unmarshal(j, &template)
				errors.AssertNoError(e, fmt.Errorf("could not unmarshal kubeData.template into MsvcTemplate"))

				fmt.Println("Template: ", template)

				switch actionRef.Action {
				case ACTION_CREATE:
					dockerImage = template.Operations.Create
				case ACTION_UPDATE:
					dockerImage = template.Operations.Update
				case ACTION_DELETE:
					dockerImage = template.Operations.Delete
				default:
					errors.Assert(false, fmt.Errorf("unknown source type"))
				}
			}

		case RESOURCE_MANAGED_RESOURCE:
			{
				j, e := json.Marshal(actionRef.KubeData["template"])
				errors.AssertNoError(e, fmt.Errorf("could not marshal kubeData.template into []byte"))

				var template MsvcTemplateResource
				e = json.Unmarshal(j, &template)

				errors.AssertNoError(e, fmt.Errorf("could not unmarshal kubeData.template into MsvcResource"))

				switch actionRef.Action {
				case ACTION_CREATE:
					dockerImage = template.Operations.Create
				case ACTION_UPDATE:
					dockerImage = template.Operations.Update
				case ACTION_DELETE:
					dockerImage = template.Operations.Delete
				}
			}

		default:
			{
				errors.Assert(len(dockerImage) != 0, fmt.Errorf("unknown resourcetype (%s)", actionRef.ResourceType))
			}
		}
	}

	return dockerImage, nil
}

func (d *domain) ApplyJob(job *Job) (e error) {
	defer errors.HandleErr(&e)
	// TODO: iterate over all the actions, and make jobs for it, apply it all over the place, then
	// TODO: wait for their execution to complete, and if it fails, rolllback them

	type Memo struct {
		Data     string
		KubeData string
	}

	memo := make(map[string]Memo)

	idx := 0
	for _, actionRef := range job.Actions {
		idx += 1
		b64Data, e := functions.ToBase64StringFromJson(actionRef.Data)
		errors.AssertNoError(e, fmt.Errorf("could not encode into base64 string"))
		b64KubeData, e := functions.ToBase64StringFromJson(actionRef.KubeData)
		errors.AssertNoError(e, fmt.Errorf("could not encode into base64 string"))

		memo[actionRef.Id] = Memo{Data: b64Data, KubeData: b64KubeData}

		dockerImage, e := getDockerImage(actionRef)
		fmt.Printf("ACTION %s | DI %s", actionRef.Action, dockerImage)
		errors.AssertNoError(e, fmt.Errorf("could not get docker image for resource (type=%s, id=%s)", actionRef.ResourceType, actionRef.ResourceId))

		j := JobVars{
			Name:            fmt.Sprintf("apply-job-%s", job.Id),
			ServiceAccount:  JOB_SERVICE_ACCOUNT,
			Image:           dockerImage,
			ImagePullPolicy: DEFAULT_IMAGE_PULL_POLICY,
			Args: []string{
				actionRef.Action,
				"--data", b64Data,
				"--kubeData", b64KubeData,
			},
		}

		fmt.Printf("Applying JOB for [ResourceType] %s [ResourceId] %s\n", actionRef.ResourceType, actionRef.ResourceId)
		e = d.KubeApplyJob(&j)
		if e != nil {
			fmt.Printf("Job at index (%d) failed because %v\n", idx, e)
		}

		isUndoable := func() bool {
			canUndo := true
			for _, actionRef := range job.Actions {
				op := ResourceActionUndoMap[actionRef.ResourceType][actionRef.Action]
				// fmt.Println("can op: ", op, actionRef.ResourceType, actionRef.Action)
				canUndo = canUndo && op
			}
			return canUndo
		}()

		if e != nil {
			fmt.Println("Undo Started ...")
			if !isUndoable {
				fmt.Println("isUndoable: TRUE")
				// STEP: commit job as can't undo
				return nil
			}

			for k := 0; k < idx; k++ {
				actionRef = job.Actions[k]
				canUndo := ResourceActionUndoMap[actionRef.ResourceType][actionRef.Action]
				if !canUndo {
					return nil
				}

				b64Data = memo[actionRef.Id].Data
				b64KubeData = memo[actionRef.Id].KubeData

				if actionRef.Action == ACTION_UPDATE {
					oldData, ok := actionRef.Data["oldData"].(map[string]interface{})
					newData, ok := actionRef.Data["newData"].(map[string]interface{})

					errors.Assert(ok, fmt.Errorf("update request should have had 'oldData' and 'newData' in their data body"))
					b64Data, e = functions.ToBase64StringFromJson(map[string]interface{}{
						"oldData": newData,
						"newData": oldData,
					})

					errors.AssertNoError(e, fmt.Errorf("could not encode into base64 string"))
				}

				dockerImage, e := getDockerImage(actionRef)
				errors.AssertNoError(e, fmt.Errorf("could not get docker image for resource (type=%s, id=%s)", actionRef.ResourceType, actionRef.ResourceId))

				j := JobVars{
					Name:            fmt.Sprintf("apply-job-%s", job.Id),
					ServiceAccount:  JOB_SERVICE_ACCOUNT,
					Image:           dockerImage,
					ImagePullPolicy: DEFAULT_IMAGE_PULL_POLICY,
					Args: []string{
						ReverseActionMap[actionRef.Action],
						"--data", b64Data,
						"--kubeData", b64KubeData,
					},
				}

				e = d.KubeApplyJob(&j)
				if e != nil {
					fmt.Printf("Job at index (%d) failed\n", idx)
				}
			}

			// STEP: commit undo now
			status, e := d.undoJob(job.Id)
			fmt.Printf("Job undo status: %v\n", status)
			return e
		}
	}

	// STEP: job commit successfull
	status, e := d.commitJob(job.Id)
	fmt.Printf("Job commit status: %v\n", status)
	return e
}

func MakeDomain(kApplier *K8sApplier, gqlClient *GqlClient, httpClient *http.Client) DomainSvc {
	t := readJobTemplate()

	applyJob := func(values *JobVars) (e error) {
		defer errors.HandleErr(&e)
		w := new(bytes.Buffer)
		e = t.ExecuteTemplate(w, "job-template.yml", values)
		errors.AssertNoError(e, fmt.Errorf("failed to inject values into template, because %v", e))

		var kJob batchv1.Job
		e = yaml.UnmarshalStrict(w.Bytes(), &kJob)
		errors.AssertNoError(e, fmt.Errorf("failed to unmarshal templated job into yaml, because %v", e))

		e = kApplier.Apply(&kJob)
		errors.AssertNoError(e, fmt.Errorf("failed to apply job, because %v", e))

		return nil
	}

	return &domain{
		KubeApplyJob: applyJob,
		httpClient:   httpClient,
		gql:          gqlClient,
		Kube:         kApplier,
	}
}
