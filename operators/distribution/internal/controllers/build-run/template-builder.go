package buildrun

import (
	"fmt"
	"strings"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	"github.com/kloudlite/operator/pkg/templates"
)

type BuildOptions struct {
	BuildArgs         string
	BuildContexts     string
	DockerfilePath    string
	DockerfileContent *string
	TargetPlatforms   string
	ContextDir        string
}

type Resource struct {
	Cpu    int
	Memory int
}

type BuildObj struct {
	Name             string
	Namespace        string
	Labels           map[string]string
	Annotations      map[string]string
	AccountName      string
	RegistryHost     string
	RegistryReponame string
	RegistryUsername string
	RegistryPassword string

	RegistryTags string

	GitRepoUrl    string
	GitRepoBranch string
	BuildOptions  BuildOptions
	BuildCacheKey *string

	ServerResource Resource
	ClientResource Resource
}

func getBuildTemplate(obj *dbv1.BuildRun) ([]byte, error) {
	if obj.Spec.Resource.Cpu < 500 {
		return nil, fmt.Errorf("cpu cannot be less than 500")
	}
	if obj.Spec.Resource.MemoryInMb < 1000 {
		return nil, fmt.Errorf("memory cannot be less than 1000")
	}

	o := &BuildObj{
		Name:             obj.Name,
		Namespace:        obj.Namespace,
		Labels:           obj.Labels,
		Annotations:      obj.Annotations,
		AccountName:      obj.Spec.AccountName,
		RegistryHost:     obj.Spec.Registry.Host,
		RegistryReponame: obj.Spec.Registry.Repo.Name,
		RegistryUsername: obj.Spec.Registry.Username,
		RegistryPassword: obj.Spec.Registry.Password,
		GitRepoUrl:       obj.Spec.GitRepo.Url,
		GitRepoBranch:    obj.Spec.GitRepo.Branch,
		BuildCacheKey:    obj.Spec.CacheKeyName,
		ClientResource: Resource{
			Cpu:    100,
			Memory: 200,
		},
	}

	o.ServerResource = Resource{
		Cpu:    obj.Spec.Resource.Cpu - o.ClientResource.Cpu,
		Memory: obj.Spec.Resource.MemoryInMb - o.ClientResource.Memory,
	}

	var err error
	if o.RegistryTags, err = func() (string, error) {
		var tags string
		for _, tag := range obj.Spec.Registry.Repo.Tags {
			if tag != "" {
				tags += fmt.Sprintf("--tag %q ", fmt.Sprintf("%s/%s:%s", obj.Spec.Registry.Host, obj.Spec.Registry.Repo.Name, tag))
			}
		}
		if tags == "" {
			return "", fmt.Errorf("tags cannot be empty")
		}
		return tags, nil
	}(); err != nil {
		return nil, err
	}

	if o.BuildOptions, err = func() (BuildOptions, error) {
		if obj.Spec.BuildOptions == nil {
			return BuildOptions{
				BuildArgs:         "",
				BuildContexts:     "",
				DockerfilePath:    "./Dockerfile",
				DockerfileContent: nil,
				TargetPlatforms:   "",
				ContextDir:        "",
			}, nil
		}

		bo := obj.Spec.BuildOptions

		var buildOptions BuildOptions

		if bo.TargetPlatforms != nil && len(bo.TargetPlatforms) > 0 {
			buildOptions.TargetPlatforms = fmt.Sprintf("--platform %q", strings.Join(bo.TargetPlatforms, ","))
		} else {
			buildOptions.TargetPlatforms = ""
		}

		if bo.ContextDir != nil && *bo.ContextDir != "" {
			buildOptions.ContextDir = fmt.Sprintf("%q", *bo.ContextDir)
		} else {
			buildOptions.ContextDir = ""
		}

		if bo.DockerfilePath != nil && *bo.DockerfilePath != "" {
			buildOptions.DockerfilePath = fmt.Sprintf("%q", *bo.DockerfilePath)
		} else {
			buildOptions.DockerfilePath = "./Dockerfile"
		}

		if bo.DockerfileContent != nil && *bo.DockerfileContent != "" {
			buildOptions.DockerfileContent = bo.DockerfileContent
		} else {
			buildOptions.DockerfileContent = nil
		}

		if bo.BuildArgs != nil && len(bo.BuildArgs) > 0 {
			var buildArgs string
			for k, v := range bo.BuildArgs {
				buildArgs += fmt.Sprintf("--build-arg %q=%q ", k, v)
			}
			buildOptions.BuildArgs = buildArgs
		} else {
			buildOptions.BuildArgs = ""
		}

		if bo.BuildContexts != nil && len(bo.BuildContexts) > 0 {
			var buildContexts string
			for k, v := range bo.BuildContexts {
				buildContexts += fmt.Sprintf("--build-arg %q=%q ", k, v)
			}
			buildOptions.BuildContexts = buildContexts
		} else {
			buildOptions.BuildContexts = ""
		}

		return buildOptions, nil
	}(); err != nil {
		return nil, err
	}

	b, err := templates.Parse(templates.Distribution.BuildJob, o)

	if err != nil {
		return nil, err
	}

	return b, nil
}
