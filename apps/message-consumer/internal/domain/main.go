package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	batchv1 "k8s.io/api/batch/v1"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/functions"
	"sigs.k8s.io/yaml"
)

type domain struct {
	KubeApplyJob func(j *JobVars) error
	httpCli      *http.Client
	gql          *GqlClient
	Kube         *K8sApplier
}

func (d *domain) ApplyProject(projectId string) (e error) {
	defer errors.HandleErr(&e)
	errors.Assert(projectId != "", fmt.Errorf("project id is required"))

	j := JobVars{
		Name:            fmt.Sprintf("apply-project-%s", projectId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           JOB_IMAGE_PROJECT,
		ImagePullPolicy: "Always",
		Args:            []string{"apply", "--projectId", projectId},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) DeleteProject(projectId string) (e error) {
	defer errors.HandleErr(&e)

	errors.Assert(projectId != "", fmt.Errorf("project id is required"))

	j := JobVars{
		Name:            fmt.Sprintf("delete-project-%s", projectId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           JOB_IMAGE_PROJECT,
		ImagePullPolicy: "Always",
		Args:            []string{"delete", "--projectId", projectId},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) ApplyApp(appId string) (e error) {
	defer errors.HandleErr(&e)

	j := JobVars{
		Name:            fmt.Sprintf("apply-app-%s", appId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           JOB_IMAGE_APP,
		ImagePullPolicy: "Always",
		Args:            []string{"apply", "--appId", appId},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) DeleteApp(appId string) (e error) {
	defer errors.HandleErr(&e)

	j := JobVars{
		Name:            fmt.Sprintf("delete-app-%s", appId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           JOB_IMAGE_APP,
		ImagePullPolicy: "Always",
		Args:            []string{"delete", "--appId", appId},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) ApplyConfig(configId string) (e error) {
	defer errors.HandleErr(&e)

	j := JobVars{
		Name:            fmt.Sprintf("apply-config-%s", configId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           JOB_IMAGE_CONFIG,
		ImagePullPolicy: "Always",
		Args:            []string{"apply", "--configId", configId},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) DeleteConfig(configId string) (e error) {
	defer errors.HandleErr(&e)

	j := JobVars{
		Name:            fmt.Sprintf("delete-config-%s", configId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           JOB_IMAGE_CONFIG,
		ImagePullPolicy: "Always",
		Args: []string{
			"delete",
			"--configId", configId,
		},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) ApplySecret(secretId string) (e error) {
	defer errors.HandleErr(&e)

	j := JobVars{
		Name:            fmt.Sprintf("apply-secret-%s", secretId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           JOB_IMAGE_SECRET,
		ImagePullPolicy: "Always",
		Args: []string{
			"apply",
			"--secretId", secretId,
		},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) DeleteSecret(secretId string) (e error) {
	defer errors.HandleErr(&e)

	j := JobVars{
		Name:            fmt.Sprintf("delete-secret-%s", secretId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           JOB_IMAGE_SECRET,
		ImagePullPolicy: "Always",
		Args: []string{
			"delete",
			"--secretId", secretId,
		},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) InstallManagedSvc(installationId string, dockerImage string) (e error) {
	defer errors.HandleErr(&e)
	j := JobVars{
		Name:            fmt.Sprintf("install-msvc-%s", installationId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           dockerImage,
		ImagePullPolicy: "Always",
		Args: []string{
			"install",
			"--installationId", installationId,
		},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) UpdateManagedSvc(installationId string, dockerImage string) (e error) {
	defer errors.HandleErr(&e)
	j := JobVars{
		Name:            fmt.Sprintf("update-msvc-%s", installationId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           dockerImage,
		ImagePullPolicy: "Always",
		Args: []string{
			"update",
			"--installationId", installationId,
		},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) UninstallManagedSvc(installationId string, dockerImage string) (e error) {
	defer errors.HandleErr(&e)
	j := JobVars{
		Name:            fmt.Sprintf("uninstall-msvc-%s", installationId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           dockerImage,
		ImagePullPolicy: "Always",
		Args: []string{
			"uninstall",
			"--installationId", installationId,
		},
	}

	return d.KubeApplyJob(&j)
}

func (d *domain) CreateManagedRes(resId string, dockerImage string) (e error) {
	defer errors.HandleErr(&e)
	j := JobVars{
		Name:            fmt.Sprintf("create-mres-%s", resId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           dockerImage,
		ImagePullPolicy: "Always",
		Args: []string{
			"create",
			"--resId", resId,
		},
	}

	return d.KubeApplyJob(&j)
}
func (d *domain) UpdateManagedRes(resId string, dockerImage string) (e error) {
	defer errors.HandleErr(&e)
	if dockerImage == "" {
		panic(fmt.Errorf("dockerImage is empty, i.e there is no update operation on the resource"))
	}
	j := JobVars{
		Name:            fmt.Sprintf("update-mres-%s", resId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           dockerImage,
		ImagePullPolicy: "Always",
		Args: []string{
			"update",
			"--resId", resId,
		},
	}

	return d.KubeApplyJob(&j)
}
func (d *domain) DeleteManagedRes(resId string, dockerImage string) (e error) {
	defer errors.HandleErr(&e)
	j := JobVars{
		Name:            fmt.Sprintf("delete-mres-%s", resId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           dockerImage,
		ImagePullPolicy: "Always",
		Args: []string{
			"delete",
			"--resId", resId,
		},
	}

	return d.KubeApplyJob(&j)
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
	req, e := d.gql.Request(q, map[string]interface{}{
		"jobId": jobId,
	})

	errors.AssertNoError(e, fmt.Errorf("could not build graphql request"))
	resp, e := d.httpCli.Do(req)
	errors.AssertNoError(e, fmt.Errorf("failed while making graphql request"))

	respB, e := io.ReadAll(resp.Body)

	var j struct {
		Data struct {
			job struct {
				commit bool `json:"commit"`
			} `json:"job"`
		} `json:"data"`
	}

	e = json.Unmarshal(respB, &j)
	errors.AssertNoError(e, fmt.Errorf("failed while unmarshalling graphql response"))

	return j.Data.job.commit, nil
}

func (d *domain) undoJob(jobId string) (result bool, e error) {
	const q = `
	  mutation Commit($jobId: ID!) {
    job {
      undo(jobId: $jobId)
    }
  }
	`
	req, e := d.gql.Request(q, map[string]interface{}{
		"jobId": jobId,
	})

	errors.AssertNoError(e, fmt.Errorf("could not build graphql request"))
	resp, e := d.httpCli.Do(req)
	errors.AssertNoError(e, fmt.Errorf("failed while making graphql request"))

	respB, e := io.ReadAll(resp.Body)

	var j struct {
		Data struct {
			job struct {
				undo bool `json:"undo"`
			} `json:"job"`
		} `json:"data"`
	}

	e = json.Unmarshal(respB, &j)
	errors.AssertNoError(e, fmt.Errorf("failed while unmarshalling graphql response"))

	return j.Data.job.undo, nil
}

func (d *domain) ApplyJob(job *Job) (e error) {
	defer errors.HandleErr(&e)

	// TODO: iterate over all the actions, and make jobs for it, apply it all over the place, then
	// TODO: wait for their execution to complete, and if it fails, rolllback them

	type Memo struct {
		Data     string
		KubeData string
	}

	var memo map[string]Memo

	idx := 0
	for _, actionRef := range job.Actions {
		idx += 1
		b64Data, e := functions.ToBase64String(actionRef.Data)
		errors.AssertNoError(e, fmt.Errorf("could not encode into base64 string"))
		b64KubeData, e := functions.ToBase64String(actionRef.KubeData)
		errors.AssertNoError(e, fmt.Errorf("could not encode into base64 string"))
		memo[actionRef.Id] = Memo{Data: b64Data, KubeData: b64KubeData}

		j := JobVars{
			Name:            fmt.Sprintf("apply-job-%s", job.Id),
			ServiceAccount:  JOB_SERVICE_ACCOUNT,
			Image:           ResourceImageMap[actionRef.ResourceType],
			ImagePullPolicy: DEFAULT_IMAGE_PULL_POLICY,
			Args: []string{
				actionRef.Action,
				"--data", b64Data,
				"--kubeData", b64KubeData,
			},
		}

		e = d.KubeApplyJob(&j)
		if e != nil {
			fmt.Printf("Job at index (%d) failed\n", idx)
		}

		isUndoable := func() bool {
			canUndo := true
			for _, actionRef := range job.Actions {
				canUndo = canUndo && ResourceActionUndoMap[actionRef.ResourceType][actionRef.Action]
			}
			return canUndo
		}()

		if e != nil {
			if isUndoable {
				fmt.Println("isUndoable: TRUE")
				// STEP: commit job as can't undo
			}

			for _, actionRef := range job.Actions {
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
					b64Data, e = functions.ToBase64String(map[string]interface{}{
						"oldData": newData,
						"newData": oldData,
					})
				}

				b64Data, e := functions.ToBase64String(actionRef.Data)
				errors.AssertNoError(e, fmt.Errorf("could not encode into base64 string"))
				b64KubeData, e := functions.ToBase64String(actionRef.KubeData)
				errors.AssertNoError(e, fmt.Errorf("could not encode into base64 string"))

				j := JobVars{
					Name:            fmt.Sprintf("apply-job-%s", job.Id),
					ServiceAccount:  JOB_SERVICE_ACCOUNT,
					Image:           ResourceImageMap[actionRef.ResourceType],
					ImagePullPolicy: DEFAULT_IMAGE_PULL_POLICY,
					Args: []string{
						ReverseActionMap[actionRef.Action],
						"--data", b64Data,
						"--kubeData", b64KubeData,
					},
				}

				e = d.KubeApplyJob(&j)
				if e != nil {
					fmt.Printf("Job at index (%d) failed\n")
				}
			}

			// STEP: commit undo now
			status, e := d.undoJob(job.Id)
			fmt.Printf("Job commit status: %s\n", status)
			return e
		}
	}

	// STEP: job commit successfull
	status, e := d.commitJob(job.Id)
	fmt.Printf("Job commit status: %s\n", status)
	return e
}

func MakeDomain(kApplier *K8sApplier, gqlClient *GqlClient) DomainSvc {
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
		httpCli:      &http.Client{},
		Kube:         kApplier,
	}
}
