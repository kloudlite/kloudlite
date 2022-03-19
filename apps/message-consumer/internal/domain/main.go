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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
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

	return d.ApplyJob(&j)
}

func (d *domain) ApplyKlJob(jobId string) (e error) {
	defer errors.HandleErr(&e)
	j := JobVars{
		Name:            fmt.Sprintf("apply-job-%s", jobId),
		ServiceAccount:  JOB_SERVICE_ACCOUNT,
		Image:           JOB_IMAGE,
		ImagePullPolicy: "Always",
		Args: []string{
			"apply",
			"--jobId", jobId,
		},
	}

	return d.ApplyJob(&j)
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
