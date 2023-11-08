# kloudlite-autoscalers

[kloudlite-autoscalers](https://github.com/kloudlite.io/helm-charts/charts/kloudlite-autoscalers) A Helm Chart for installing autoscalers in kloudlite enabled kubernetes clusters

![Version: v1.0.5](https://img.shields.io/badge/Version-v1.0.5-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0.5](https://img.shields.io/badge/AppVersion-v1.0.5-informational?style=flat-square)

## Get Repo Info

```console
helm repo add kloudlite https://kloudlite.github.io/helm-charts
helm repo update
```

## Install Chart

**Important:** only helm3 is supported

**Important:** ensure kloudlite CRDs have been installed

```console
helm install [RELEASE_NAME] kloudlite/kloudlite-autoscalers --namespace [NAMESPACE] [--create-namespace]
```

The command deploys kloudlite-agent on the Kubernetes cluster in the default configuration.

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Installing Nightly Releases

To list all nightly versions (**NOTE**: nightly versions are suffixed by `-nightly`)

```console
helm search repo kloudlite/kloudlite-autoscalers --devel
```

To install
```console
helm install [RELEASE_NAME] kloudlite/kloudlite-autoscalers --version [NIGHTLY_VERSION] --namespace [NAMESPACE] --create-namespace
```

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME] -n [NAMESPACE]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] kloudlite/kloudlite-autoscalers --install --namespace [NAMESPACE]
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values kloudlite/kloudlite-autoscalers
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| IACStateStore | object | `{"bucketDir":"","bucketName":"","bucketRegion":""}` | infrastructure-as-code state store configuration |
| IACStateStore.bucketDir | string | `""` | bucket directory, state file will be stored in this directory |
| IACStateStore.bucketName | string | `""` | bucket name |
| IACStateStore.bucketRegion | string | `""` | bucket region |
| cloudprovider.secretName | string | `"kloudlite-cloud-config"` |  |
| cloudprovider.values.accessKey | string | `""` |  |
| cloudprovider.values.secretKey | string | `""` |  |
| clusterAutoscaler.enabled | bool | `true` |  |
| clusterAutoscaler.image.repository | string | `"ghcr.io/kloudlite/operators/cluster-autoscaler"` |  |
| clusterRegion | string | `"ap-south-1"` |  |
| defaults.imagePullPolicy | string | `"Always"` |  |
| defaults.imageTag | string | `"v1.0.5-nightly"` |  |
| k3sMasters.joinToken | string | `""` |  |
| k3sMasters.publicHost | string | `""` |  |
| nodepools.enabled | bool | `true` |  |
| nodepools.image.repository | string | `"ghcr.io/kloudlite/operators/nodepool"` |  |
| nodepools.image.tag | string | `""` |  |
| replicaCount | int | `1` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.nameSuffix | string | `"sa"` |  |
