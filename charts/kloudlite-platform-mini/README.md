# kloudlite-platform

[kloudlite-platform](https://github.com/kloudlite.io/helm-charts/charts/kloudlite-platform) Helm Chart for installing and setting up kloudlite platform on your own hosted Kubernetes clusters.

![Version: v1.0.7](https://img.shields.io/badge/Version-v1.0.7-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0.7](https://img.shields.io/badge/AppVersion-v1.0.7-informational?style=flat-square)

## Get Repo Info

```console
helm repo add kloudlite https://kloudlite.github.io/helm-charts
helm repo update kloudlite
```

## Installation

> [!NOTE]
> only helm3 is supported

### Installing Kloudlite CRDs
```console
kubectl apply -f https://github.com/kloudlite/helm-charts/releases/download/v1.0.7/crds-all.yml --server-side
```

### Installing chart
```console
helm install kloudlite-platform kloudlite/kloudlite-platform --namespace kloudlite --create-namespace --version v1.0.7 --set baseDomain="<base_domain>"
```

> [!TIP]
> To list all available chart versions, run `helm search repo kloudlite/kloudlite-platform`

The command deploys kloudlite-platform on your Kubernetes cluster in the default configuration.

_See [configuration](#configuration) below._

## Upgrading Chart

```console
helm upgrade kloudlite-platform kloudlite/kloudlite-platform --namespace kloudlite --version v1.0.7 --set baseDomain="<base_domain>"
```

## Configuration

To see all configurable options with detailed comments, visit the chart's [values.yaml](https://github.com/kloudlite/helm-charts/charts/kloudlite-platform/values.yaml), or run these commands:

```console
helm show values kloudlite/kloudlite-platform
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| baseDomain | string | `""` | base domain |

## Uninstalling Chart

```console
helm uninstall kloudlite-platform -n kloudlite
```

This removes all the Kubernetes components associated with the chart and deletes the release.

