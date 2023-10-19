package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"kloudlite.io/apps/container-registry/internal/domain"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	text_templates "kloudlite.io/pkg/text-templates"
)

func getTemplate(obj domain.BuildJobTemplateObject) ([]byte, error) {

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

func main() {
	obj := domain.BuildJobTemplateObject{
		KlAdmin:          "admin",
		Registry:         "cr.khost.dev",
		Name:             "sample",
		Tags:             []string{"latest", "v1"},
		RegistryRepoName: "kloudlite/sample",
		DockerPassword:   "token",
		Namespace:        "kl-core",
		GitRepoUrl:       "https://github.com/abdheshnayak/demo-env.git",
		Branch:           "sample",
		Labels: map[string]string{
			"kloudlite.io/build-id": "1",
		},
		Annotations: map[string]string{
			"kloudlite.io/build-id": "1",
		},
		AccountName: "kloudlite",
		BuildOptions: &entities.BuildOptions{
			BuildArgs: map[string]string{},
			BuildContexts: map[string]string{
				"sample":       ".",
				"project_root": "../project",
			},
			DockerfilePath: func(s string) *string {
				return &s
			}("./Dockerfile"),
			DockerfileContent: func(s string) *string {
				return &s
			}(`
FROM node
npm i -g pnpm
			`),
			TargetPlatforms: []string{
				"linux/amd64",
				"linux/arm64",
			},
			ContextDir: new(string),
		},
	}

	b, err := getTemplate(obj)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
