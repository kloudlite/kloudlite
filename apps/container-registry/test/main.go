package main

import (
	"fmt"

	dbv1 "github.com/kloudlite/operator/apis/distribution/v1"
	"kloudlite.io/apps/container-registry/internal/domain"
)

func main() {
	obj := domain.BuildJobTemplateData{
		AccountName: "sample",
		Name:        "sample",
		Namespace:   "kl-core",
		Labels: map[string]string{
			"kloudlite.io/build-id": "1",
		},
		Annotations: map[string]string{
			"kloudlite.io/build-id": "1",
		},
		Registry: dbv1.Registry{
			Username: "sldfkjs",
			Password: "sldkfjs",
			Host:     "sldjfl",
			Repo: dbv1.Repo{
				Name: "slkdfj",
				Tags: []string{"sfkjs"},
			},
		},
		CacheKeyName: func(s string) *string { return &s }("sample"),
		Resource:     dbv1.Resource{},
		GitRepo:      dbv1.GitRepo{},
		BuildOptions: &dbv1.BuildOptions{
			BuildArgs: map[string]string{
				"linux/amd64": "sldkfj",
			},
			BuildContexts: map[string]string{
				"linux/amd64": "sldkfj",
			},
			DockerfilePath:    new(string),
			DockerfileContent: new(string),
			TargetPlatforms:   []string{"linux/amd64"},
			ContextDir:        new(string),
		},
	}

	var d domain.Impl
	b, err := d.GetBuildTemplate(obj)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
