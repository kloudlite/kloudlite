package buildrun

import (
	"fmt"
	"net/url"
	"strings"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	"github.com/kloudlite/operator/pkg/functions"
	rApi "github.com/kloudlite/operator/pkg/operator"
	"github.com/kloudlite/operator/pkg/templates"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	ServerResource  Resource
	ClientResource  Resource
	OwnerReferences []metav1.OwnerReference
}

func (r *Reconciler) getCreds(req *rApi.Request[*dbv1.BuildRun]) (err error, ra []byte, rp []byte, rh []byte, gp []byte) {
	ctx, obj := req.Context(), req.Object

	if err := func() error {
		s, err := rApi.Get(ctx, r.Client, functions.NN(obj.Spec.CredentialsRef.Namespace, obj.Spec.CredentialsRef.Name), &corev1.Secret{})
		if err != nil {
			return err
		}

		commonError := "please ensure the secret has the following keys: registry-admin, registry-token, registry-host, github-token"

		var ok bool

		if ra, ok = s.Data["registry-admin"]; !ok || len(ra) == 0 {
			return fmt.Errorf("registry-admin key not found in secret %s, %s", obj.Spec.CredentialsRef.Name, commonError)
		}

		if rp, ok = s.Data["registry-token"]; !ok || len(rp) == 0 {
			return fmt.Errorf("registry-token key not found in secret %s, %s", obj.Spec.CredentialsRef.Name, commonError)
		}

		if rh, ok = s.Data["registry-host"]; !ok || len(rh) == 0 {
			return fmt.Errorf("registry-host key not found in secret %s, %s", obj.Spec.CredentialsRef.Name, commonError)
		}

		if gp, ok = s.Data["github-token"]; !ok || len(gp) == 0 {
			return fmt.Errorf("github-token key not found in secret %s, %s", obj.Spec.CredentialsRef.Name, commonError)
		}

		return nil
	}(); err != nil {
		return err, nil, nil, nil, nil
	}

	return nil, ra, rp, rh, gp
}

func BuildUrl(repo string, pullToken []byte) (string, error) {
	parsedURL, err := url.Parse(repo)
	if err != nil {
		fmt.Println("Error parsing Repo URL:", err)
		return "", err
	}

	parsedURL.User = url.User(string(pullToken))

	return parsedURL.String(), nil
}

func (r *Reconciler) getBuildTemplate(req *rApi.Request[*dbv1.BuildRun]) ([]byte, error) {
	obj := req.Object

	if obj.Spec.Resource.Cpu < 500 {
		return nil, fmt.Errorf("cpu cannot be less than 500")
	}
	if obj.Spec.Resource.MemoryInMb < 1000 {
		return nil, fmt.Errorf("memory cannot be less than 1000")
	}

	var err error

	err, ra, rp, rh, gp := r.getCreds(req)
	if err != nil {
		return nil, err
	}

	gitRepoUrl, err := BuildUrl(obj.Spec.GitRepo.Url, gp)
	if err != nil {
		return nil, err
	}

	o := &BuildObj{
		Name:             obj.Name,
		Namespace:        obj.Namespace,
		Labels:           obj.Labels,
		Annotations:      obj.Annotations,
		AccountName:      obj.Spec.AccountName,
		RegistryHost:     string(rh),
		RegistryReponame: obj.Spec.Registry.Repo.Name,
		RegistryUsername: string(ra),
		RegistryPassword: string(rp),
		GitRepoUrl:       gitRepoUrl,
		GitRepoBranch:    obj.Spec.GitRepo.Branch,
		BuildCacheKey:    obj.Spec.CacheKeyName,
		ClientResource: Resource{
			Cpu:    200,
			Memory: 200,
		},
		OwnerReferences: []metav1.OwnerReference{functions.AsOwner(obj, true)},
	}

	o.ServerResource = Resource{
		Cpu:    obj.Spec.Resource.Cpu - o.ClientResource.Cpu,
		Memory: obj.Spec.Resource.MemoryInMb - o.ClientResource.Memory,
	}

	if o.RegistryTags, err = func() (string, error) {
		var tags string
		for _, tag := range obj.Spec.Registry.Repo.Tags {
			if tag != "" {
				tags += fmt.Sprintf("--tag %q ", fmt.Sprintf("%s/%s/%s:%s", rh, obj.Spec.AccountName, obj.Spec.Registry.Repo.Name, tag))
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
