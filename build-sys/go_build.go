package main

import (
	"context"
	"dagger/kloudlite/internal/dagger"
	"fmt"
)

// +---------------------------------------+
// |  Go Binary Builds                     |
// +---------------------------------------+

// BuildKloudlite builds the unified kloudlite backend binary.
func (m *Kloudlite) BuildKloudlite(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="linux"
	goos string,
	// +default="amd64"
	goarch string,
	// +optional
	version string,
) *dagger.File {
	return m.buildGoBinary(ctx, source, "kloudlite", goos, goarch, version, true)
}

// BuildKli builds the kli (infrastructure installer) CLI binary.
func (m *Kloudlite) BuildKli(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="linux"
	goos string,
	// +default="amd64"
	goarch string,
	// +optional
	version string,
) *dagger.File {
	return m.buildGoBinary(ctx, source, "kli", goos, goarch, version, false)
}

// BuildKltun builds the kltun (WireGuard tunnel) CLI binary.
func (m *Kloudlite) BuildKltun(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="linux"
	goos string,
	// +default="amd64"
	goarch string,
	// +optional
	version string,
) *dagger.File {
	return m.buildGoBinary(ctx, source, "kltun", goos, goarch, version, false)
}

// BuildOciInstaller builds the OCI installer binary.
func (m *Kloudlite) BuildOciInstaller(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="linux"
	goos string,
	// +default="amd64"
	goarch string,
) *dagger.File {
	return m.buildGoBinary(ctx, source, "kli/oci-installer", goos, goarch, "", false)
}

// BuildCodeAnalyzer builds the code-analyzer binary.
func (m *Kloudlite) BuildCodeAnalyzer(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="linux"
	goos string,
	// +default="amd64"
	goarch string,
) *dagger.File {
	return m.buildGoBinary(ctx, source, "code-analyzer", goos, goarch, "", false)
}

// BuildKlLinux builds the kl CLI binary for linux/amd64 (used by workspace-comprehensive image).
func (m *Kloudlite) BuildKlLinux(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
) *dagger.File {
	return m.buildGoBinary(ctx, source, "kl", "linux", "amd64", "", false)
}

// BuildKlTunProxy builds the kl-tun-proxy binary.
func (m *Kloudlite) BuildKlTunProxy(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="linux"
	goos string,
	// +default="amd64"
	goarch string,
) *dagger.File {
	return m.buildGoBinary(ctx, source, "kl-tun-proxy", goos, goarch, "", false)
}

// BuildAllBinaries cross-compiles Go binaries (kloudlite, kli, kltun) for all
// OS/arch combos (linux/darwin/windows x amd64/arm64).
func (m *Kloudlite) BuildAllBinaries(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +optional
	version string,
) *dagger.Directory {
	out := dag.Directory()
	type appDef struct{ name, suffix string; genCRD bool }
	apps := []appDef{
		{"kloudlite", "kloudlite", true},
		{"kli", "kli", false},
		{"kltun", "kltun", false},
	}
	for _, app := range apps {
		for _, goos := range []string{"linux", "darwin", "windows"} {
			for _, goarch := range []string{"amd64", "arm64"} {
				ext := ""
				if goos == "windows" {
					ext = ".exe"
				}
				bin := m.buildGoBinary(ctx, source, app.suffix, goos, goarch, version, app.genCRD)
				out = out.WithFile(fmt.Sprintf("%s-%s-%s%s", app.name, goos, goarch, ext), bin)
			}
		}
	}
	return out
}

// buildGoBinary is the shared helper for all Go binary builds.
func (m *Kloudlite) buildGoBinary(
	ctx context.Context,
	source *dagger.Directory,
	app string,
	goos, goarch, version string,
	genCRD bool,
) *dagger.File {
	// Persistent cache volumes
	goModCache := dag.CacheVolume("go-mod-v1")
	goBuildCache := dag.CacheVolume("go-build-v1")

	ctr := dag.Container().
		From("golang:1.24-alpine").
		WithExec([]string{"apk", "add", "--no-cache", "git", "curl", "bash", "build-base"}).
		WithMountedCache("/go/pkg/mod", goModCache).
		WithMountedCache("/root/.cache/go-build", goBuildCache).
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithEnvVariable("GOOS", goos).
		WithEnvVariable("GOARCH", goarch).
		WithEnvVariable("CGO_ENABLED", "0")

	if genCRD {
		ctr = ctr.
			WithExec([]string{"go", "install", "sigs.k8s.io/controller-tools/cmd/controller-gen@v0.19.0"}).
			WithExec([]string{"controller-gen", "crd", "paths=./types/...", "output:crd:artifacts:config=manifests"}).
			WithExec([]string{"go", "run", "api/crds/assets/generate.go", "-manifests", "manifests", "-output", "api/crds/assets/generated.go"}).
			WithExec([]string{"mkdir", "-p", "cli/kli/internal/manifests"}).
			WithExec([]string{"sh", "-c", "cat manifests/*.yaml > cli/kli/internal/manifests/crds.yaml"}).
			WithExec([]string{"controller-gen", "object", "paths=./...", "-w"})
	}

	if app == "kltun" && goos == "windows" {
		wtd := dag.Container().
			From("alpine:3.20").
			WithExec([]string{"apk", "add", "--no-cache", "curl", "unzip"}).
			WithExec([]string{"sh", "-c", "curl -LO https://www.wintun.net/builds/wintun-0.14.1.zip && unzip wintun-0.14.1.zip"}).
			Directory("/wintun")
		ctr = ctr.
			WithDirectory("cli/kltun/pkg/wintun/dll", wtd.Directory("bin/amd64")).
			WithDirectory("cli/kltun/pkg/wintun/dll", wtd.Directory("bin/arm64"))
	}

	ldflags := "-s -w"
	if version != "" {
		switch app {
		case "kli", "kloudlite", "kltun":
			ldflags = fmt.Sprintf("-s -w -X github.com/kloudlite/kloudlite/cli/%s/cmd.Version=%s", app, version)
		}
	}

	var buildPath, outName string
	switch app {
	case "kloudlite":
		buildPath, outName = "./cli/kloudlite/", "kloudlite"
	case "kli":
		buildPath, outName = "./cli/kli/", "kli"
	case "kli/oci-installer":
		buildPath, outName = "./cli/kli/oci-installer/", "oci-installer"
	case "code-analyzer":
		buildPath, outName = "./cli/code-analyzer/", "code-analyzer"
	case "kl-tun-proxy":
		buildPath, outName = "./cli/kl-tun-proxy/", "kl-tun-proxy"
	case "kl":
		buildPath, outName = "./cli/kl/", "kl"
	case "kltun":
		buildPath, outName = "./cli/kltun/", "kltun"
	}
	return ctr.
		WithExec([]string{"go", "build", "-ldflags=" + ldflags, "-o", "/out/" + outName, buildPath}).
		File("/out/" + outName)
}
