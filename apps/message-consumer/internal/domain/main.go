package domain

import (
	"bytes"
	"encoding/base64"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	"kloudlite.io/pkg/errors"
	"sigs.k8s.io/yaml"
)

type domain struct {
	Inject func(*JobVars) ([]byte, error)
	Kube   *K8sApplier
}

func (d *domain) ApplyProject(p *Project) (e error) {
	defer errors.HandleErr(&e)
	projectB, e := yaml.Marshal(&ProjectValues{
		Name:        p.Name,
		AccountId:   p.Account.Id,
		DisplayName: p.DisplayName,
		Logo:        p.Logo,
		Cluster:     p.Cluster,
		Description: p.Description,
	})

	errors.AssertNoError(e, fmt.Errorf("failed to marshal project into yaml, because %v", e))

	j := JobVars{
		Name:            fmt.Sprintf("apply-project-%s", p.Name),
		ServiceAccount:  "hotspot-cluster-svc-account",
		Image:           "registry.gitlab.com/madhouselabs/kloudlite/api/jobs/project:latest",
		ImagePullPolicy: "Always",
		Args:            []string{"apply"},
		Env: map[string]string{
			"RELEASE_NAME": p.Name,
			"NAMESPACE":    "hotspot",
			"VALUES":       base64.StdEncoding.EncodeToString(projectB),
		},
	}

	jobBytes, e := d.Inject(&j)
	errors.AssertNoError(e, fmt.Errorf("failed to inject job values in template, because %v", e))

	var kJob batchv1.Job
	e = yaml.UnmarshalStrict(jobBytes, &kJob)

	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal templated job into yaml, because %v", e))

	e = d.Kube.Apply(&kJob)
	errors.AssertNoError(e, fmt.Errorf("failed to apply job, because %v", e))

	return nil
}

func (d *domain) DeleteProject(p *Project) (e error) {
	defer errors.HandleErr(&e)
	errors.AssertNoError(e, fmt.Errorf("failed to marshal app into yaml, because %v", e))

	j := JobVars{
		Name:            fmt.Sprintf("delete-project-%s", p.Id),
		ServiceAccount:  "hotspot-cluster-svc-account",
		Image:           "registry.gitlab.com/madhouselabs/kloudlite/api/jobs/app:latest",
		ImagePullPolicy: "Always",
		Args:            []string{"delete"},
		Env: map[string]string{
			"RELEASE_NAME": p.Id,
			"NAMESPACE":    "hotspot",
		},
	}

	tData, e := d.Inject(&j)
	errors.AssertNoError(e, fmt.Errorf("failed to inject job values in template, because %v", e))

	var kJob batchv1.Job
	e = yaml.UnmarshalStrict(tData, &kJob)
	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal templated job into yaml, because %v", e))

	return d.Kube.Apply(&kJob)
}

func (d *domain) ApplyApp(app *App) (e error) {
	defer errors.HandleErr(&e)

	data, e := yaml.Marshal(&app)
	errors.AssertNoError(e, fmt.Errorf("failed to marshal app into yaml, because %v", e))

	j := JobVars{
		Name:            fmt.Sprintf("apply-app-%s", app.Name),
		ServiceAccount:  "hotspot-cluster-svc-account",
		Image:           "registry.gitlab.com/madhouselabs/kloudlite/api/jobs/app:latest",
		ImagePullPolicy: "Always",
		Args:            []string{"apply"},
		Env: map[string]string{
			"RELEASE_NAME": app.Id,
			"NAMESPACE":    "hotspot",
			"VALUES":       base64.StdEncoding.EncodeToString(data),
		},
	}

	tData, e := d.Inject(&j)
	errors.AssertNoError(e, fmt.Errorf("failed to inject job values in template, because %v", e))

	var kJob batchv1.Job
	e = yaml.UnmarshalStrict(tData, &kJob)
	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal templated job into yaml, because %v", e))

	return d.Kube.Apply(&kJob)
}

func (d *domain) DeleteApp(app *App) (e error) {
	defer errors.HandleErr(&e)
	errors.AssertNoError(e, fmt.Errorf("failed to marshal app into yaml, because %v", e))

	j := JobVars{
		Name:            fmt.Sprintf("delete-app-%s", app.Id),
		ServiceAccount:  "hotspot-cluster-svc-account",
		Image:           "registry.gitlab.com/madhouselabs/kloudlite/api/jobs/app:latest",
		ImagePullPolicy: "Always",
		Args:            []string{"delete"},
		Env: map[string]string{
			"RELEASE_NAME": app.Id,
			"NAMESPACE":    "hotspot",
		},
	}

	tData, e := d.Inject(&j)
	errors.AssertNoError(e, fmt.Errorf("failed to inject job values in template, because %v", e))

	var kJob batchv1.Job
	e = yaml.UnmarshalStrict(tData, &kJob)
	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal templated job into yaml, because %v", e))

	return d.Kube.Apply(&kJob)
}

func (d *domain) ApplyConfig(cfg *Config, project *Project) (e error) {
	defer errors.HandleErr(&e)

	data, e := yaml.Marshal(&cfg)
	errors.AssertNoError(e, fmt.Errorf("failed to marshal app into yaml, because %v", e))

	j := JobVars{
		Name:            fmt.Sprintf("apply-config-%s", cfg.Name),
		ServiceAccount:  "hotspot-cluster-svc-account",
		Image:           "registry.gitlab.com/madhouselabs/kloudlite/api/jobs/config:latest",
		ImagePullPolicy: "Always",
		Args:            []string{"apply"},
		Env: map[string]string{
			"RELEASE_NAME": fmt.Sprintf("config-%s", cfg.Name),
			"NAMESPACE":    project.Name,
			"VALUES":       base64.StdEncoding.EncodeToString(data),
		},
	}

	tData, e := d.Inject(&j)
	errors.AssertNoError(e, fmt.Errorf("failed to inject job values in template, because %v", e))

	var kJob batchv1.Job
	e = yaml.UnmarshalStrict(tData, &kJob)
	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal templated job into yaml, because %v", e))

	return d.Kube.Apply(&kJob)
}

func (d *domain) DeleteConfig(cfg *Config) (e error) {
	defer errors.HandleErr(&e)

	j := JobVars{
		Name:            fmt.Sprintf("delete-config-%s", cfg.Name),
		ServiceAccount:  "hotspot-cluster-svc-account",
		Image:           "registry.gitlab.com/madhouselabs/kloudlite/api/jobs/config:latest",
		ImagePullPolicy: "Always",
		Args:            []string{"delete"},
		Env: map[string]string{
			"RELEASE_NAME": fmt.Sprintf("config-%s", cfg.Name),
			"NAMESPACE":    "hotspot",
		},
	}

	tData, e := d.Inject(&j)
	errors.AssertNoError(e, fmt.Errorf("failed to inject job values in template, because %v", e))

	var kJob batchv1.Job
	e = yaml.UnmarshalStrict(tData, &kJob)
	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal templated job into yaml, because %v", e))

	return d.Kube.Apply(&kJob)
}

func (d *domain) ApplySecret(secret *Secret) (e error) {
	defer errors.HandleErr(&e)

	data, e := yaml.Marshal(&secret)
	errors.AssertNoError(e, fmt.Errorf("failed to marshal app into yaml, because %v", e))

	j := JobVars{
		Name:            fmt.Sprintf("apply-config-%s", secret.Name),
		ServiceAccount:  "hotspot-cluster-svc-account",
		Image:           "registry.gitlab.com/madhouselabs/kloudlite/api/jobs/secret:latest",
		ImagePullPolicy: "Always",
		Args:            []string{"apply"},
		Env: map[string]string{
			"RELEASE_NAME": fmt.Sprintf("%s-%s", secret.Name, secret.Project.Id),
			"NAMESPACE":    "hotspot",
			"VALUES":       base64.StdEncoding.EncodeToString(data),
		},
	}

	tData, e := d.Inject(&j)
	errors.AssertNoError(e, fmt.Errorf("failed to inject job values in template, because %v", e))

	var kJob batchv1.Job
	e = yaml.UnmarshalStrict(tData, &kJob)
	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal templated job into yaml, because %v", e))

	return d.Kube.Apply(&kJob)
}

func (d *domain) DeleteSecret(secret *Secret) (e error) {
	defer errors.HandleErr(&e)

	j := JobVars{
		Name:            fmt.Sprintf("delete-config-%s", secret.Name),
		ServiceAccount:  "hotspot-cluster-svc-account",
		Image:           "registry.gitlab.com/madhouselabs/kloudlite/api/jobs/secret:latest",
		ImagePullPolicy: "Always",
		Args:            []string{"delete"},
		Env: map[string]string{
			"RELEASE_NAME": fmt.Sprintf("%s-%s", secret.Name, secret.Project.Id),
			"NAMESPACE":    "hotspot",
		},
	}

	tData, e := d.Inject(&j)
	errors.AssertNoError(e, fmt.Errorf("failed to inject job values in template, because %v", e))

	var kJob batchv1.Job
	e = yaml.UnmarshalStrict(tData, &kJob)
	errors.AssertNoError(e, fmt.Errorf("failed to unmarshal templated job into yaml, because %v", e))

	return d.Kube.Apply(&kJob)
}

func MakeDomain(kApplier *K8sApplier) DomainSvc {
	t := readJobTemplate()

	injecValues := func(values *JobVars) ([]byte, error) {
		w := new(bytes.Buffer)
		err := t.ExecuteTemplate(w, "job-template.yml", values)
		return w.Bytes(), err
	}

	return &domain{
		Inject: injecValues,
		Kube:   kApplier,
	}
}
