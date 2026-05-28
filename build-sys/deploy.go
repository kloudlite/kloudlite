package main

import (
	"context"
	"dagger/kloudlite/internal/dagger"
	"fmt"
)

// +---------------------------------------+
// |  Kubernetes Deployment                |
// +---------------------------------------+

// KubeDeployAPIServer deploys the api-server image into a Kubernetes cluster.
func (m *Kloudlite) KubeDeployAPIServer(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +optional
	kubeConfig *dagger.File,
	imageRef string,
	// +default="kloudlite"
	namespace string,
	// +optional
	envFile *dagger.File,
) *dagger.Container {
	return m.kubeDeploy(ctx, source, kubeConfig, "api-server", imageRef, namespace, envFile)
}

// KubeDeployDashboard deploys the dashboard image into a Kubernetes cluster.
func (m *Kloudlite) KubeDeployDashboard(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +optional
	kubeConfig *dagger.File,
	imageRef string,
	// +default="kloudlite"
	namespace string,
	// +optional
	envFile *dagger.File,
) *dagger.Container {
	return m.kubeDeploy(ctx, source, kubeConfig, "dashboard", imageRef, namespace, envFile)
}

// KubeDeploy deploys any component by name into Kubernetes using kubectl apply.
func (m *Kloudlite) KubeDeploy(
	ctx context.Context,
	// +defaultPath="/"
	source *dagger.Directory,
	// +optional
	kubeConfig *dagger.File,
	component string,
	imageRef string,
	// +default="kloudlite"
	namespace string,
	// +optional
	envFile *dagger.File,
) *dagger.Container {
	return m.kubeDeploy(ctx, source, kubeConfig, component, imageRef, namespace, envFile)
}

// kubeDeploy is the shared helper for deploying components into Kubernetes.
func (m *Kloudlite) kubeDeploy(
	ctx context.Context,
	source *dagger.Directory,
	kubeConfig *dagger.File,
	component string,
	imageRef string,
	namespace string,
	envFile *dagger.File,
) *dagger.Container {
	ctr := dag.Container().
		From("alpine:3.20").
		WithMountedCache("/var/cache/apk", dag.CacheVolume("apk-v1")).
		WithExec([]string{"apk", "add", "--no-cache", "curl", "bash", "gettext"})

	ctr = ctr.
		WithMountedCache("/tmp/kubectl-dl", dag.CacheVolume("kubectl-v1-32")).
		WithExec([]string{"sh", "-c",
			"mkdir -p /tmp/kubectl-dl && " +
				"curl -fsSL https://dl.k8s.io/release/v1.32.0/bin/linux/amd64/kubectl -o /tmp/kubectl-dl/kubectl && " +
				"chmod +x /tmp/kubectl-dl/kubectl && " +
				"cp /tmp/kubectl-dl/kubectl /usr/local/bin/kubectl"})

	if kubeConfig != nil {
		ctr = ctr.WithMountedDirectory("/root/.kube",
			dag.Directory().WithFile("config", kubeConfig))
	}

	manifestsDir := source.Directory("manifests")
	ctr = ctr.WithMountedDirectory("/manifests", manifestsDir)

	ctr = ctr.
		WithExec([]string{"sh", "-c", fmt.Sprintf(`
set -e
COMP="%s"
IMAGE="%s"
NS="%s"

# Search multiple locations for manifest
for DIR in /manifests /deploy-manifests; do
  if [ -f "${DIR}/${COMP}.yaml" ]; then
    echo "Found manifest at ${DIR}/${COMP}.yaml"
    cat "${DIR}/${COMP}.yaml" | sed "s|image: .*|image: $IMAGE|g" > /tmp/deploy.yaml
    break
  fi
done

if [ ! -f /tmp/deploy.yaml ]; then
  echo "No manifest found for ${COMP}.yaml in /manifests/ or /deploy-manifests/"
  ls -la /manifests/ 2>/dev/null || true
  exit 1
fi

kubectl --kubeconfig=/root/.kube/config apply -f /tmp/deploy.yaml --namespace="$NS" 2>&1 || kubectl apply -f /tmp/deploy.yaml --namespace="$NS" 2>&1 || true
`, component, imageRef, namespace)})

	return ctr
}
