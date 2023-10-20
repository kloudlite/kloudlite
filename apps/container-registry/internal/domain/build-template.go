package domain

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/templates"
	text_templates "kloudlite.io/pkg/text-templates"
)

func BuildUrl(repo, pullToken string) (string, error) {
	parsedURL, err := url.Parse(repo)
	if err != nil {
		fmt.Println("Error parsing Repo URL:", err)
		return "", err
	}

	parsedURL.User = url.User(pullToken)

	return parsedURL.String(), nil
}

type BuildJobTemplateData struct {
	KlAdmin          string
	Registry         string
	Name             string
	Tags             []string
	RegistryRepoName string
	DockerPassword   string
	Namespace        string
	GitRepoUrl       string
	Labels           map[string]string
	Annotations      map[string]string
	AccountName      string
	BuildOptions     *entities.BuildOptions
	Branch           string
}

type BuildOptions struct {
	BuildArgs         string
	BuildContexts     string
	DockerfilePath    string
	DockerfileContent *string
	TargetPlatforms   string
	ContextDir        string
}

type BuildJobTemplateObject struct {
	KlAdmin          string
	Registry         string
	Name             string
	Tags             string
	RegistryRepoName string
	DockerPassword   string
	Namespace        string
	GitRepoUrl       string
	Labels           map[string]string
	Annotations      map[string]string
	AccountName      string
	BuildOptions     BuildOptions
	Branch           string
}

func getTemplate(obj BuildJobTemplateObject) ([]byte, error) {
	b, err := templates.ReadBuildJobTemplate()
	if err != nil {
		return nil, err
	}

	tmpl := text_templates.WithFunctions(template.New("build-job-template"))

	tmpl, err = tmpl.Parse(string(b))
	if err != nil {
		return nil, err
	}

	out := new(bytes.Buffer)
	err = tmpl.Execute(out, obj)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (*Impl) GetBuildTemplate(obj BuildJobTemplateData) ([]byte, error) {

	var build BuildJobTemplateObject
	var err error

	build.KlAdmin = obj.KlAdmin
	build.Registry = obj.Registry
	build.Name = obj.Name

	if build.Tags, err = func() (string, error) {
		var tags string
		for _, tag := range obj.Tags {
			if tag != "" {
				tags += fmt.Sprintf("--tag %q ", fmt.Sprintf("%s/%s:%s", obj.Registry, obj.RegistryRepoName, tag))
			}
		}
		if tags == "" {
			return "", fmt.Errorf("tags cannot be empty")
		}
		return tags, nil
	}(); err != nil {
		return nil, err
	}

	build.RegistryRepoName = obj.RegistryRepoName
	build.DockerPassword = obj.DockerPassword
	build.Namespace = obj.Namespace
	build.GitRepoUrl = obj.GitRepoUrl
	build.Labels = obj.Labels
	build.Annotations = obj.Annotations
	build.AccountName = obj.AccountName
	build.Branch = obj.Branch

	build.BuildOptions, err = func() (BuildOptions, error) {
		if obj.BuildOptions == nil {
			return BuildOptions{
				BuildArgs:         "",
				BuildContexts:     "",
				DockerfilePath:    "./Dockerfile",
				DockerfileContent: nil,
				TargetPlatforms:   "",
				ContextDir:        "",
			}, nil
		}

		bo := obj.BuildOptions

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
	}()

	if err != nil {
		return nil, err
	}

	b, err := getTemplate(build)
	if err != nil {
		return nil, err
	}

	return b, nil
}
