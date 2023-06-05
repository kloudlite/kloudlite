

# kloudlite-operators

[kloudlite-operators](https://github.com/kloudlite.io/helm-charts/charts/kloudlite-operators) K8s Operators for kloudlite CRDs

![Version: 1.0.5-nightly](https://img.shields.io/badge/Version-1.0.5--nightly-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.0.5-nightly](https://img.shields.io/badge/AppVersion-1.0.5--nightly-informational?style=flat-square)

## Get Repo Info

```console
helm repo add kloudlite https://kloudlite.github.io/helm-charts
helm repo update
```

## Install Kloudlite CRDs
```console
curl -L0 https://github.com/kloudlite/helm-charts/releases/download/kloudlite-crds-1.0.5-nightly/crds.yml | kubectl apply -f -
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
| cert-manager.cainjector.podLabels | object | `{}` |  |
| cert-manager.cainjector.resources | object | `{"limits":{"cpu":"120m","memory":"200Mi"},"requests":{"cpu":"80m","memory":"200Mi"}}` | resource limits for cert-manager cainjector pods |
| cert-manager.cainjector.resources.limits | object | `{"cpu":"120m","memory":"200Mi"}` | resource limits for cert-manager webhook pods |
| cert-manager.cainjector.resources.limits.cpu | string | `"120m"` | cpu limit for cert-manager cainjector pods |
| cert-manager.cainjector.resources.limits.memory | string | `"200Mi"` | memory limit for cert-manager cainjector pods |
| cert-manager.cainjector.resources.requests.cpu | string | `"80m"` | cpu requests for cert-manager cainjector pods |
| cert-manager.cainjector.resources.requests.memory | string | `"200Mi"` | memory requests for cert-manager cainjector pods |
| cert-manager.extraArgs[0] | string | `"--dns01-recursive-nameservers-only"` |  |
| cert-manager.extraArgs[1] | string | `"--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53"` |  |
| cert-manager.installCRDs | bool | `true` | cert-manager whether to install CRDs |
| cert-manager.nodeSelector | object | `{}` |  |
| cert-manager.podLabels | object | `{}` |  |
| cert-manager.resources.limits | object | `{"cpu":"80m","memory":"120Mi"}` | resource limits for cert-manager controller pods |
| cert-manager.resources.limits.cpu | string | `"80m"` | cpu limit for cert-manager controller pods |
| cert-manager.resources.limits.memory | string | `"120Mi"` | memory limit for cert-manager controller pods |
| cert-manager.resources.requests.cpu | string | `"40m"` | cpu request for cert-manager controller pods |
| cert-manager.resources.requests.memory | string | `"120Mi"` | memory request for cert-manager controller pods |
| cert-manager.startupapicheck.enabled | bool | `false` |  |
| cert-manager.tolerations | list | `[]` |  |
| cert-manager.webhook.podLabels | object | `{}` |  |
| cert-manager.webhook.resources | object | `{"limits":{"cpu":"60m","memory":"60Mi"},"requests":{"cpu":"30m","memory":"60Mi"}}` | resource limits for cert-manager webhook pods |
| cert-manager.webhook.resources.limits | object | `{"cpu":"60m","memory":"60Mi"}` | resource limits for cert-manager webhook pods |
| cert-manager.webhook.resources.limits.cpu | string | `"60m"` | cpu limit for cert-manager webhook pods |
| cert-manager.webhook.resources.limits.memory | string | `"60Mi"` | memory limit for cert-manager webhook pods |
| cert-manager.webhook.resources.requests.cpu | string | `"30m"` | cpu limit for cert-manager webhook pods |
| cert-manager.webhook.resources.requests.memory | string | `"60Mi"` | memory limit for cert-manager webhook pods |
| defaultImagePullSecretName | string | `"kl-image-pull-creds"` | default image pull secret name |
| imagePullPolicy | string | `"Always"` | container image pull policy |
| nodeSelector | object | `{}` | node selectors for all pods in this chart |
| operators.app.enabled | bool | `true` | whether to enable app operator |
| operators.app.image | string | `"ghcr.io/kloudlite/operators/app:v1.0.5-nightly"` | app operator image and tag |
| operators.app.name | string | `"kl-app"` | app operator workload name |
| operators.csiDrivers.enabled | bool | `false` | whether to enable csi drivers operator |
| operators.csiDrivers.image | string | `"ghcr.io/kloudlite/operators/csi-drivers:v1.0.5-nightly"` | csi drivers operator image and tag |
| operators.csiDrivers.name | string | `"kl-csi-drivers"` | csi drivers operator workload name |
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
| operators.msvcRedpanda.enabled | bool | `false` | whether to enable msvc-redpanda operator |
| operators.msvcRedpanda.image | string | `"ghcr.io/kloudlite/operators/msvc-redpanda:v1.0.5-nightly"` | msvc redpanda operator image and tag |
| operators.msvcRedpanda.name | string | `"kl-redpanda"` | msvc redpanda operator workload name |
| operators.project.enabled | bool | `true` | whether to enable project operator |
| operators.project.image | string | `"ghcr.io/kloudlite/operators/project:v1.0.5-nightly"` | project operator image and tag |
| operators.project.name | string | `"kl-projects"` | project operator workload name |
| operators.routers.enabled | bool | `true` | whether to enable routers operator |
| operators.routers.image | string | `"ghcr.io/kloudlite/operators/routers:v1.0.5-nightly"` | routers operator image and tag |
| operators.routers.name | string | `"kl-routers"` | routers operator workload name |
| operators.wgOperator.enabled | bool | `false` | whether to enable wg operator |
| operators.wgOperator.image | string | `"ghcr.io/kloudlite/operators/wg-operator:v1.0.5-nightly"` | wg operator image and tag |
| operators.wgOperator.name | string | `"kl-wg-operator"` | wg operator workload name |
| podLabels | object | `{}` | pod labels for all pods in this chart |
| svcAccountName | string | `"kloudlite-cluster-svc-account"` | container image pull policy |
| tolerations | array | `[]` | tolerations for all pods in this chart |
| wg | object | `{"baseDomain":"<wg-base-domain>","nameserver":{"basicAuth":{"enabled":true,"password":"<dns-api-basic-auth-password>","username":"<dns-api-basic-auth-username>"},"endpoint":"<https://dns-api-endpoint>"},"podCidr":"10.42.0.0/16","svcCidr":"10.43.0.0/16"}` | wireguard configuration options |
| wg.baseDomain | string | `"<wg-base-domain>"` | baseDomain for wireguard service, to be exposed |
| wg.nameserver | object | `{"basicAuth":{"enabled":true,"password":"<dns-api-basic-auth-password>","username":"<dns-api-basic-auth-username>"},"endpoint":"<https://dns-api-endpoint>"}` | dns nameserver http endpoint |
| wg.nameserver.basicAuth | object | `{"enabled":true,"password":"<dns-api-basic-auth-password>","username":"<dns-api-basic-auth-username>"}` | basic auth configurations for dns nameserver http endpoint |
| wg.nameserver.basicAuth.enabled | bool | `true` | whether to enable basic auth for dns nameserver http endpoint |
| wg.nameserver.basicAuth.password | string | `"<dns-api-basic-auth-password>"` | if enabled, basic auth password for dns nameserver http endpoint |
| wg.nameserver.basicAuth.username | string | `"<dns-api-basic-auth-username>"` | if enabled, basic auth username for dns nameserver http endpoint |
| wg.podCidr | string | `"10.42.0.0/16"` | cluster pods CIDR range |
| wg.svcCidr | string | `"10.43.0.0/16"` | cluster services CIDR range |
