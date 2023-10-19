package domain

import (
	"bytes"
	"fmt"
	"net/url"
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

type BuildJobTemplateObject struct {
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

func (*Impl) GetBuildTemplate(obj BuildJobTemplateObject) ([]byte, error) {
	b, err := getTemplate(obj)
	if err != nil {
		return nil, err
	}

	return b, nil
}
