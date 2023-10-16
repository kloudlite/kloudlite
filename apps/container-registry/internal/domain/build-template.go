package domain

import (
	"bytes"
	"fmt"
	"net/url"
	"os"

	"text/template"

	text_templates "kloudlite.io/pkg/text-templates"
)

func BuildUrl(repo, branch, pullToken string) (string, error) {
	parsedURL, err := url.Parse(repo)
	if err != nil {
		fmt.Println("Error parsing Repo URL:", err)
		return "", err
	}

	parsedURL.User = url.User(pullToken)
	parsedURL.Fragment = branch

	return parsedURL.String(), nil
}

type BuildJobTemplateObject struct {
	Registry         string
	Name             string
	Tag              string
	RegistryRepoName string
	DockerPassword   string
	Namespace        string
	PullUrl          string
	DockerHost       string
	Labels           map[string]string
}

func getTemplate(obj BuildJobTemplateObject) ([]byte, error) {

	b, err := os.ReadFile("./templates/build-job.yml.tpl")
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
