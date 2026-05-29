package main

import (
	"context"
	"dagger/kloudlite/internal/dagger"
	"fmt"
)

// +---------------------------------------+
// |  Build + Deploy Pipelines             |
// +---------------------------------------+

// BuildApp builds the binary or web app for the given app component.
// Returns a directory containing the build artifact.
func (m *Kloudlite) BuildApp(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	app App,
	// +optional
	version string,
) *dagger.Directory {
	switch app {
	case AppAPIServer, AppWorkmachineManager, AppKloudlite:
		bin := m.BuildKloudlite(ctx, source, "linux", "amd64", version)
		return dag.Directory().WithFile("binary", bin)
	case AppKli:
		bin := m.BuildKli(ctx, source, "linux", "amd64", version)
		return dag.Directory().WithFile("binary", bin)
	case AppKltun:
		bin := m.BuildKltun(ctx, source, "linux", "amd64", version)
		return dag.Directory().WithFile("binary", bin)
	case AppOciInstaller:
		bin := m.BuildOciInstaller(ctx, source, "linux", "amd64")
		return dag.Directory().WithFile("binary", bin)
	case AppCodeAnalyzer:
		bin := m.BuildCodeAnalyzer(ctx, source, "linux", "amd64")
		return dag.Directory().WithFile("binary", bin)
	case AppDashboard:
		build := m.BuildDashboard(ctx, source)
		return dag.Directory().WithDirectory(".next", build)
	case AppConsole:
		build := m.BuildConsole(ctx, source)
		return dag.Directory().WithDirectory(".next", build)
	case AppWebsite:
		build := m.BuildWebsite(ctx, source)
		return dag.Directory().WithDirectory(".next", build)
	case AppK3sBackup:
		return dag.Directory()
	}
	return dag.Directory()
}

// ImageApp builds the Docker image for the given app component.
func (m *Kloudlite) ImageApp(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	app App,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
) *dagger.Container {
	switch app {
	case AppAPIServer:
		return m.ImageAPIServer(ctx, source, tag, registry)
	case AppWorkmachineManager:
		return m.ImageWorkmachineManager(ctx, source, tag, registry)
	case AppDashboard:
		return m.ImageDashboard(ctx, source, tag, registry, nil)
	case AppConsole:
		return m.ImageConsole(ctx, source, tag, registry, nil)
	case AppWebsite:
		return m.ImageWebsite(ctx, source, tag, registry, nil)
	case AppOciInstaller:
		return m.ImageOciInstaller(ctx, source, tag, registry)
	case AppCodeAnalyzer:
		return m.ImageCodeAnalyzer(ctx, source, tag, registry)
	case AppK3sBackup:
		return m.ImageK3sBackup(ctx, tag, registry)
	case AppTunnelServer:
		return m.ImageTunnelServer(ctx, source, tag, registry)
	case AppWorkspaceBase:
		return m.ImageWorkspaceBase(ctx, source, tag, registry)
	case AppWorkspaceComprehensive:
		return m.ImageWorkspaceComprehensive(ctx, source, tag, registry, nil)
	}
	return dag.Container().From("alpine:3.20")
}

// +---------------------------------------+
// |  Publish images to registry           |
// +---------------------------------------+

// PublishImage builds and pushes a specific Docker image to a container registry.
// Returns the published image reference string (e.g. ghcr.io/kloudlite/api-server:dev-local).
// Registry auth is handled via Dagger engine configuration (dagger.json env or CI secrets).
func (m *Kloudlite) PublishImage(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	app App,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
	// Optional registry username for basic auth.
	// +optional
	registryUsername string,
	// Optional registry password/token for basic auth.
	// +optional
	registryToken *dagger.Secret,
) (string, error) {
	img := m.ImageApp(ctx, source, app, tag, registry)
	imageRef := fmt.Sprintf("%s/%s:%s", registry, appImageName(app), tag)

	if registryUsername != "" && registryToken != nil {
		img = img.WithRegistryAuth(registry, registryUsername, registryToken)
	}

	published, err := img.Publish(ctx, imageRef)
	if err != nil {
		return "", fmt.Errorf("publish %s: %w", imageRef, err)
	}
	return published, nil
}

// PublishAllImages builds and pushes all deployable Docker images to the container registry.
// This is the CI entry point for pushing all images after a build.
func (m *Kloudlite) PublishAllImages(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
	// Optional registry username for basic auth.
	// +optional
	registryUsername string,
	// Optional registry password/token for basic auth.
	// +optional
	registryToken *dagger.Secret,
) (string, error) {
	apps := []App{
		AppAPIServer,
		AppWorkmachineManager,
		AppDashboard,
		AppConsole,
		AppWebsite,
		AppOciInstaller,
		AppCodeAnalyzer,
		AppK3sBackup,
		AppTunnelServer,
		AppWorkspaceBase,
		AppWorkspaceComprehensive,
	}
	var lastRef string
	for _, app := range apps {
		ref, err := m.PublishImage(ctx, source, app, tag, registry, registryUsername, registryToken)
		if err != nil {
			return "", fmt.Errorf("publish %s: %w", app, err)
		}
		lastRef = ref
	}
	return lastRef, nil
}

// +---------------------------------------+
// |  Build + Deploy Pipelines             |
// +---------------------------------------+

// BuildAndDeployApp builds and optionally pushes the Docker image for an app,
// then deploys it to Kubernetes. This is the high-level CI/CD pipeline.
// For CLI binary apps (kloudlite, kli, kltun), only the build step runs.
func (m *Kloudlite) BuildAndDeployApp(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	app App,
	// +optional
	kubeConfig *dagger.File,
	// +default="dev-local"
	tag string,
	// +default="ghcr.io/kloudlite"
	registry string,
	// +default="kloudlite"
	namespace string,
	// +optional
	version string,
	// If true, push the built image to the container registry before deploying.
	// +optional
	// +default=true
	publish bool,
	// Optional registry username if publish is true.
	// +optional
	registryUsername string,
	// Optional registry password/token if publish is true.
	// +optional
	registryToken *dagger.Secret,
) *dagger.Container {
	imageRef := fmt.Sprintf("%s/%s:%s", registry, appImageName(app), tag)
	yamlKey := appYamlManifest(app)

	switch {
	case yamlKey != "":
		// 1. Build image
		img := m.ImageApp(ctx, source, app, tag, registry)

		// 2. Optionally push to registry
		if publish {
			pushImg := img
			if registryUsername != "" && registryToken != nil {
				pushImg = pushImg.WithRegistryAuth(registry, registryUsername, registryToken)
			}
			if _, err := pushImg.Publish(ctx, imageRef); err != nil {
				return dag.Container().From("alpine:3.20").
					WithExec([]string{"echo", fmt.Sprintf("ERROR publishing %s: %s", imageRef, err)})
			}
		}

		// 3. Deploy to Kubernetes
		return m.kubeDeploy(ctx, source, kubeConfig, yamlKey, imageRef, namespace, nil)

	case app == AppKloudlite || app == AppKli || app == AppKltun:
		// CLI-only: build binary
		m.BuildApp(ctx, source, app, version)
		return dag.Container().From("alpine:3.20").
			WithExec([]string{"echo", fmt.Sprintf("%s binary built (not deployable as K8s service)", app)})

	default:
		// Docker-image-only: build and optionally push
		img := m.ImageApp(ctx, source, app, tag, registry)
		if publish {
			pushImg := img
			if registryUsername != "" && registryToken != nil {
				pushImg = pushImg.WithRegistryAuth(registry, registryUsername, registryToken)
			}
			if _, err := pushImg.Publish(ctx, imageRef); err != nil {
				return dag.Container().From("alpine:3.20").
					WithExec([]string{"echo", fmt.Sprintf("ERROR publishing %s: %s", imageRef, err)})
			}
		}
		return dag.Container().From("alpine:3.20").
			WithExec([]string{"echo", fmt.Sprintf("App %s image built and published to %s", app, imageRef)})
	}
}

// +---------------------------------------+
// |  Binary Releases to GitHub            |
// +---------------------------------------+

// ReleaseBinaries builds all platforms binaries for the given app,
// generates SHA256 checksums, and creates a GitHub Release with assets.
func (m *Kloudlite) ReleaseBinaries(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// Version string e.g. "1.2.0" - determines the git tag: <app>/v<version>.
	version string,
	// GitHub token with contents:write permission.
	// Can be passed as env:GITHUB_TOKEN (from host env or CI) or file:/path/.env (from local .env).
	gitHubToken *dagger.Secret,
	// GitHub repository e.g. "kloudlite/kloudlite".
	// +default="kloudlite/kloudlite"
	repository string,
	// Which app to release. If empty, releases all three (kloudlite, kli, kltun).
	// +optional
	app App,
) *dagger.Container {
	apps := []App{app}
	if app == "" {
		apps = []App{AppKloudlite, AppKli, AppKltun}
	}

	result := dag.Container().From("alpine:3.20")

	for _, a := range apps {
		appName := appImageName(a)

		// Build all 6 binaries (3 OS × 2 arch) for this app only
		bins := dag.Directory()
		for _, goos := range []string{"linux", "darwin", "windows"} {
			for _, goarch := range []string{"amd64", "arm64"} {
				ext := ""
				if goos == "windows" {
					ext = ".exe"
				}
				name := fmt.Sprintf("%s-%s-%s%s", appName, goos, goarch, ext)

				var bin *dagger.File
				switch a {
				case AppKloudlite:
					bin = m.BuildKloudlite(ctx, source, goos, goarch, version)
				case AppKli:
					bin = m.BuildKli(ctx, source, goos, goarch, version)
				case AppKltun:
					bin = m.BuildKltun(ctx, source, goos, goarch, version)
				}
				bins = bins.WithFile(name, bin)
			}
		}

		releaseCtr := dag.Container().
			From("alpine:3.20").
			WithExec([]string{"apk", "add", "--no-cache", "curl", "bash", "jq"}).
			WithMountedDirectory("/bins", bins).
			WithMountedSecret("/run/secrets/GITHUB_TOKEN", gitHubToken).
			WithExec([]string{"sh", "-c", fmt.Sprintf(`
set -e
APP="%s"
VERSION="%s"
TAG="%s/v%s"
REPO="%s"

cd /bins

echo "=== Generating SHA256 checksums ==="
for f in *; do
  sha256sum "$f" > "$f.sha256"
done
ls -la

GITHUB_TOKEN=$(cat /run/secrets/GITHUB_TOKEN)

echo "=== Creating GitHub Release: $TAG ==="

# Delete existing release for this tag if it exists
EXISTING_ID=$(curl -sf -H "Authorization: token $GITHUB_TOKEN" \
  "https://api.github.com/repos/$REPO/releases/tags/$TAG" | jq -r '.id // empty' 2>/dev/null || echo "")

if [ -n "$EXISTING_ID" ]; then
  echo "Deleting existing release id=$EXISTING_ID"
  curl -sf -X DELETE -H "Authorization: token $GITHUB_TOKEN" \
    "https://api.github.com/repos/$REPO/releases/$EXISTING_ID" || true
fi

# Create release
RELEASE_JSON=$(cat <<EOJSON
{
  "tag_name": "$TAG",
  "name": "$APP v$VERSION",
  "body": "## $APP $VERSION\n\n### Installation\n\nSee attached binaries and checksums.",
  "draft": false,
  "prerelease": false
}
EOJSON
)

RESP=$(curl -sf -X POST -H "Authorization: token $GITHUB_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$RELEASE_JSON" \
  "https://api.github.com/repos/$REPO/releases")

RELEASE_ID=$(echo "$RESP" | jq -r '.id')
echo "Release created: id=$RELEASE_ID"

# Upload each file as an asset
echo "=== Uploading assets ==="
for f in *; do
  echo "Uploading $f..."
  curl -sf -X POST -H "Authorization: token $GITHUB_TOKEN" \
    -H "Content-Type: application/octet-stream" \
    --data-binary @"$f" \
    "https://uploads.github.com/repos/$REPO/releases/$RELEASE_ID/assets?name=$f" > /dev/null || echo "Failed to upload $f"
done

echo "=== Release complete: $TAG ==="
`, appName, version, appName, version, repository)})

		result = releaseCtr
	}

	return result
}
