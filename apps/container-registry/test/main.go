package main

import (
	"fmt"

	"kloudlite.io/apps/container-registry/internal/domain"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
)

func main() {
	obj := domain.BuildJobTemplateData{
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

	var d domain.Impl
	b, err := d.GetBuildTemplate(obj)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
