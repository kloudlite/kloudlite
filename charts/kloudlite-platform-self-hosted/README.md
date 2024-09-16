# kloudlite-platform

[kloudlite-platform](https://github.com/kloudlite.io/helm-charts/charts/kloudlite-platform) Helm Chart for installing and setting up kloudlite platform on your own hosted Kubernetes clusters.

![Version: v1.0.7](https://img.shields.io/badge/Version-v1.0.7-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0.7](https://img.shields.io/badge/AppVersion-v1.0.7-informational?style=flat-square)

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
helm install [RELEASE_NAME] kloudlite/kloudlite-platform --namespace [NAMESPACE] [--create-namespace]
```

The command deploys kloudlite-agent on the Kubernetes cluster in the default configuration.

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Installing Nightly Releases

To list all nightly versions (**NOTE**: nightly versions are suffixed by `-nightly`)

```console
helm search repo kloudlite/kloudlite-platform --devel
```

To install
```console
helm install [RELEASE_NAME] kloudlite/kloudlite-platform --version [NIGHTLY_VERSION] --namespace [NAMESPACE] --create-namespace
```

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME] -n [NAMESPACE]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] kloudlite/kloudlite-platform --install --namespace [NAMESPACE]
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values kloudlite/kloudlite-platform
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| baseDomain | string | `"a-client.kloudlite.io"` | base domain |
| kloudliteDNSSuffix | string | `"khost.dev"` | kloudlite DNS suffix |
| kloudliteRelease | string | `""` | kloudlite release version, defaults to `Chart.AppVersion` |
