package main

import (
	"context"
	"dagger/kloudlite/internal/dagger"
)

// +---------------------------------------+
// |  CRD Generation                       |
// +---------------------------------------+

// GenerateCRDs generates CRD YAML manifests, embedded Go assets, and deepcopy files.
func (m *Kloudlite) GenerateCRDs(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.Directory {
	return dag.Container().
		From("golang:1.24-alpine").
		WithExec([]string{"apk", "add", "--no-cache", "git", "curl", "bash"}).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod-v1")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build-v1")).
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"go", "install", "sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0"}).
		WithExec([]string{"controller-gen", "crd", "paths=./types/...", "output:crd:artifacts:config=manifests"}).
		WithExec([]string{"go", "run", "api/crds/assets/generate.go", "-manifests", "manifests", "-output", "api/crds/assets/generated.go"}).
		WithExec([]string{"mkdir", "-p", "cli/kli/internal/manifests"}).
		WithExec([]string{"sh", "-c", "cat manifests/*.yaml > cli/kli/internal/manifests/crds.yaml"}).
		WithExec([]string{"controller-gen", "object", "paths=./...", "-w"}).
		Directory("/src")
}
