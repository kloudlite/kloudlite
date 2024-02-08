

# kloudlite-operators

[kloudlite-operators](https://github.com/kloudlite.io/helm-charts/charts/kloudlite-operators) K8s Operators for kloudlite CRDs

![Version: v1.0.5](https://img.shields.io/badge/Version-v1.0.5-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0.5](https://img.shields.io/badge/AppVersion-v1.0.5-informational?style=flat-square)

## Get Repo Info

```console
helm repo add kloudlite https://kloudlite.github.io/helm-charts
helm repo update
```

## Install Kloudlite CRDs
```console
curl -L0 https://github.com/kloudlite/helm-charts/releases/download/v1.0.5/crds/all.yml | kubectl apply -f -
```

## Install Chart

**Important:** only helm3 is supported
**Important:** ensure kloudlite CRDs have been installed

```console
helm install [RELEASE_NAME] kloudlite/kloudlite-operators --namespace [NAMESPACE] --create-namespace
```

The command deploys kloudlite-agent on the Kubernetes cluster in the default configuration.

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Installing Nightly Releases

To list all nightly versions (**NOTE**: nightly versions are suffixed by `-nightly`)

```console
helm search repo kloudlite/kloudlite-operators --devel
```
To install
```console
helm install [RELEASE_NAME] kloudlite/kloudlite-operators --version [NIGHTLY_VERSION] --namespace [NAMESPACE] --create-namespace
```

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME] [-n NAMESPACE]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] kloudlite/kloudlite-operators --install
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values kloudlite/kloudlite-operators
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| imagePullPolicy | string | `"Always"` | container image pull policy |
| kloudliteRelease | string | `"v1.0.5-nightly"` |  |
| nodeSelector | object | `{}` | node selectors for all pods in this chart |
| operators.app.enabled | bool | `true` | whether to enable app operator |
| operators.app.image | object | `{"repository":"ghcr.io/kloudlite/operators/app"}` | app operator image and tag image: ghcr.io/kloudlite/operators/app:v1.0.5-nightly |
| operators.app.name | string | `"kl-app"` | app operator workload name |
| operators.clusterAutoscaler.enabled | bool | `true` | whether to enable cluster-autoscaler |
| operators.clusterAutoscaler.image.pullPolicy | string | `"Always"` |  |
| operators.clusterAutoscaler.image.repository | string | `"ghcr.io/kloudlite/cluster-autoscaler-amd64"` |  |
| operators.clusterAutoscaler.image.tag | string | `"kloudlite-v1.0.5-nightly"` |  |
| operators.clusterAutoscaler.name | string | `"kl-cluster-autoscaler"` | cluster-autoscaler workload name |
| operators.csiDrivers.enabled | bool | `false` | whether to enable csi drivers operator |
| operators.csiDrivers.image | string | `"ghcr.io/kloudlite/operators/csi-drivers:v1.0.5-nightly"` | csi drivers operator image and tag |
| operators.csiDrivers.name | string | `"kl-csi-drivers"` | csi drivers operator workload name |
| operators.helmChartsOperator.enabled | bool | `true` | whether to enable helm-charts operator |
| operators.helmChartsOperator.image | string | `"ghcr.io/kloudlite/operators/helm-charts:v1.0.5-nightly"` | helm-charts operator image and tag |
| operators.helmChartsOperator.name | string | `"kl-helm-charts-operator"` | helm-charts operator workload name |
| operators.helmOperator.enabled | bool | `true` | whether to enable helm operator |
| operators.helmOperator.image | string | `"ghcr.io/kloudlite/operators/helm:v1.0.5-nightly"` | helm operator image and tag |
| operators.helmOperator.name | string | `"kl-helm-operator"` | helm operator workload name |
| operators.msvcElasticsearch.enabled | bool | `false` | whether to enable msvc-elasticsearch operator |
| operators.msvcElasticsearch.image | string | `"ghcr.io/kloudlite/operators/msvc-elasticsearch:v1.0.5-nightly"` | msvc elasticsearch operator image and tag |
| operators.msvcElasticsearch.name | string | `"kl-msvc-elasticsearch"` | msvc elasticsearch operator workload name |
| operators.msvcMongo.enabled | bool | `true` | whether to enable msvc-mongo operator |
| operators.msvcMongo.image | string | `"ghcr.io/kloudlite/operators/msvc-mongo:v1.0.5-nightly"` | name msvc mongo operator image and tag |
| operators.msvcMongo.name | string | `"kl-msvc-mongo"` | msvc mongo operator workload name |
| operators.msvcNMres.enabled | bool | `true` | whether to enable msvc-n-mres operator |
| operators.msvcNMres.image | string | `"ghcr.io/kloudlite/operators/msvc-n-mres:v1.0.5-nightly"` | msvc-n-mres operator image and tag |
| operators.msvcNMres.name | string | `"kl-msvc-n-mres"` | msvc-n-mres operator workload name |
| operators.msvcRedis.enabled | bool | `true` | whether to enable msvc-redis operator |
| operators.msvcRedis.image | string | `"ghcr.io/kloudlite/operators/msvc-redis:v1.0.5-nightly"` | msvc redis operator image and tag |
| operators.msvcRedis.name | string | `"kl-msvc-redis"` | msvc redis operator workload name |
| operators.msvcRedpanda.enabled | bool | `true` | whether to enable msvc-redpanda operator |
| operators.msvcRedpanda.image | string | `"ghcr.io/kloudlite/operators/msvc-redpanda:v1.0.5-nightly"` | msvc redpanda operator image and tag |
| operators.msvcRedpanda.name | string | `"kl-redpanda"` | msvc redpanda operator workload name |
| operators.project.enabled | bool | `true` | whether to enable project operator |
| operators.project.image | object | `{"repository":"ghcr.io/kloudlite/operators/project"}` | project operator image and tag |
| operators.project.name | string | `"kl-projects"` | project operator workload name |
| operators.routers.enabled | bool | `true` | whether to enable router operator |
| operators.routers.image | string | `"ghcr.io/kloudlite/operators/routers:v1.0.5-nightly"` | routers operator image and tag |
| operators.routers.name | string | `"kl-routers"` | router operator workload name |
| podLabels | object | `{}` | pod labels for all pods in this chart |
| preferOperatorsOnMasterNodes | bool | `true` | affine operator pods to master nodes |
| svcAccountName | string | `"kloudlite-cluster-svc-account"` | container image pull policy |
| tolerations | array | `[]` | tolerations for all pods in this chart |
