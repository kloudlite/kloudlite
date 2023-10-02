# aws-spot-termination-handler

[aws-spot-termination-handler](https://github.com/kloudlite.io/helm-charts/charts/aws-spot-termination-handler) A Helm chart for Kubernetes

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

## Get Repo Info

```console
helm repo add kloudlite https://kloudlite.github.io/helm-charts
helm repo update
```

## Install Chart
```console
helm install [RELEASE_NAME] kloudlite/aws-spot-termination-handler --namespace [NAMESPACE]
```

The command deploys kloudlite-agent on the Kubernetes cluster in the default configuration.

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Installing Nightly Releases

To list all nightly versions (**NOTE**: nightly versions are suffixed by `-nightly`)

```console
helm search repo kloudlite/aws-spot-termination-handler --devel
```

To install
```console
helm install  [RELEASE_NAME] kloudlite/aws-spot-termination-handler --version [NIGHTLY_VERSION] --namespace [NAMESPACE] --create-namespace
```

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME] -n [NAMESPACE]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] kloudlite/aws-spot-termination-handler
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values kloudlite/aws-spot-termination-handler
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| image.name | string | `"ghcr.io/kloudlite/platform/aws-spot-k3s-termination-handler"` | kloudlite image repository, tag will be dervied from {{.kloudliteRelease}} |
| kloudliteRelease | string | `"v1.0.5-nightly"` | kloudlite release identifier |
| name | string | `"aws-spot-termination-handler"` |  |
| nodeSelector | object | `{}` | node selector for the spot termination handler, it is required because it must be running only on aws spot instances |
