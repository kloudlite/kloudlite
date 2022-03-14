package domain

import (
	"bytes"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	"kloudlite.io/pkg/errors"
	"sigs.k8s.io/yaml"
)

type domain struct {
	ApplyJob func(j *JobVars) error
	Kube     *K8sApplier
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
}

func (d *domain) CreateManagedRes(resId string, version int) (e error) {
	defer errors.HandleErr(&e)

	j := JobVars{
		Name: fmt.Sprintf("create-mres-%s", resId),
		ServiceAccount: JOB_IMAGE_SECRET,
		Image: JOB_IMAGE_MANAGED_SVC,
	}
}

func MakeDomain(kApplier *K8sApplier) DomainSvc {
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
		ApplyJob: applyJob,
		Kube:     kApplier,
	}
}
