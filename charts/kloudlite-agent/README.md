# kloudlite-agent

[kloudlite-agent](https://github.com/kloudlite.io/helm-charts/charts/kloudlite-agent) Kloudlite Agent to make your kubernetes cluster communicate securely with kloudlite control plane

![Version: v1.0.5](https://img.shields.io/badge/Version-v1.0.5-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0.5](https://img.shields.io/badge/AppVersion-v1.0.5-informational?style=flat-square)

## Chart also installs these charts
- [ingress-nginx](https://kubernetes.github.io/ingress-nginx)
- [cert-manager](https://charts.jetstack.io)
- [vector](https://vector.dev/docs/setup/installation/package-managers/helm)

## Get Repo Info

```console
helm repo add kloudlite https://kloudlite.github.io/helm-charts
helm repo update
```

## Install Chart

**Important:** only helm3 is supported</br>
**Important:** [kloudlite-operators](../kloudlite-operators) must be installed beforehand</br>
**Important:** ensure kloudlite CRDs have been installed</br>

```console
helm install [RELEASE_NAME] kloudlite/kloudlite-agent --namespace [NAMESPACE]
```

The command deploys kloudlite-agent on the Kubernetes cluster in the default configuration.

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Installing Nightly Releases

To list all nightly versions (**NOTE**: nightly versions are suffixed by `-nightly`)

```console
helm search repo kloudlite/kloudlite-agent --devel
```

To install
```console
helm install  [RELEASE_NAME] kloudlite/kloudlite-agent --version [NIGHTLY_VERSION] --namespace [NAMESPACE] --create-namespace
```

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME] -n [NAMESPACE]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] kloudlite/kloudlite-agent --install
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values kloudlite/kloudlite-agent
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| accountName | string, if available | `""` | kloudlite account name |
| agent.nodeSelector | object | `{}` |  |
| agent.tolerations | list | `[]` |  |
| agentOperator.nodeAffinity | object | `{}` |  |
| agentOperator.nodeSelector | object | `{}` |  |
| agentOperator.tolerations | list | `[]` |  |
| clusterName | string, if available | `""` | kloudlite cluster name |
| clusterToken | string REQUIRED | `""` | kloudlite issued cluster token |
| helmCharts.vector.enabled | bool | `true` |  |
| helmCharts.vector.nodeSelector | object | `{}` |  |
| imagePullPolicy | string | `"Always"` | container image pull policy |
| kloudliteRelease | string | `""` | kloudlite release version, defaults to `Helm AppVersion` |
| messageOfficeGRPCAddr | string | `""` | kloudlite message office api grpc address, should be in the form of 'grpc-host:grcp-port', grpc-api.domain.com:443 |
