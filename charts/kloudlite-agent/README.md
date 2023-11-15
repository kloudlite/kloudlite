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
| accessToken | string | `""` | kloudlite issued access token (if already have) |
| accountName | string ⚠️  **Required** | `""` | kloudlite account name |
| agent.enabled | bool | `true` | enable/disable kloudlite agent |
| agent.image | object | `{"pullPolicy":"","repository":"ghcr.io/kloudlite/agents/kl-agent","tag":""}` | kloudlite agent image name and tag |
| agent.image.pullPolicy | string | `""` | image pull policy for kloudlite agent, default is .imagePullPolicy |
| agent.image.tag | string | `""` | image tag for kloudlite agent, by default uses kloudlite_release |
| clusterIdentitySecretName | string | `"kl-cluster-identity"` | cluster identity secret name, which keeps cluster token and access token |
| clusterInternalDNS | string | `"cluster.local"` | cluster internal DNS, like 'cluster.local' |
| clusterName | string ⚠️  **Required** | `""` | kloudlite cluster name |
| clusterToken | string ⚠️  **Required** | `""` | kloudlite issued cluster token |
| helmCharts.cert-manager.affinity | object | `{}` |  |
| helmCharts.cert-manager.enabled | bool | `true` |  |
| helmCharts.cert-manager.name | string | `"cert-manager"` |  |
| helmCharts.cert-manager.nodeSelector | object | `{}` |  |
| helmCharts.cert-manager.tolerations | list | `[]` |  |
| helmCharts.ingress-nginx.controllerKind | string | `"DaemonSet"` |  |
| helmCharts.ingress-nginx.enabled | bool | `true` |  |
| helmCharts.ingress-nginx.ingressClassName | string | `"nginx"` |  |
| helmCharts.ingress-nginx.name | string | `"ingress-nginx"` |  |
| helmCharts.vector.debugOnStdout | bool | `false` |  |
| helmCharts.vector.enabled | bool | `true` |  |
| helmCharts.vector.name | string | `"vector"` |  |
| helmCharts.vector.nodeSelector | object | `{}` |  |
| helmCharts.vector.tolerations | list | `[]` |  |
| imagePullPolicy | string | `"Always"` | container image pull policy |
| messageOfficeGRPCAddr | string | `""` | kloudlite message office api grpc address, should be in the form of 'grpc-host:grcp-port', grpc-api.domain.com:443 |
| operators.resourceWatcher.enabled | bool | `true` | enable/disable kloudlite resource watcher |
| operators.resourceWatcher.image | object | `{"pullPolicy":"","repository":"ghcr.io/kloudlite/agents/resource-watcher","tag":""}` | kloudlite resource watcher image name and tag |
| operators.resourceWatcher.image.pullPolicy | string | `""` | image pullPolicy for kloudlite resource watcher, by default uses .Chart.AppVersion |
| operators.resourceWatcher.image.tag | string | `""` | image tag for kloudlite resource watcher, by default uses .Chart.AppVersion |
| operators.wgOperator.configuration | object | `{"dnsHostedZone":"","podCIDR":"10.42.0.0/16","svcCIDR":"10.43.0.0/16"}` | wireguard configuration options |
| operators.wgOperator.configuration.dnsHostedZone | string | `""` | dns hosted zone, i.e., dns pointing to this cluster, like 'wireguard.domain.com' |
| operators.wgOperator.configuration.podCIDR | string | `"10.42.0.0/16"` | cluster pods CIDR range |
| operators.wgOperator.configuration.svcCIDR | string | `"10.43.0.0/16"` | cluster services CIDR range |
| operators.wgOperator.enabled | bool | `true` | whether to enable wg operator |
| operators.wgOperator.image | object | `{"pullPolicy":"","repository":"ghcr.io/kloudlite/operators/wireguard","tag":""}` | wg operator image and tag |
| operators.wgOperator.image.pullPolicy | string | `""` | image pull policy for kloudlite wireguard operator, default is .imagePullPolicy |
| operators.wgOperator.image.tag | string | `""` | image tag for kloudlite wireguard operator, by default uses .Chart.AppVersion |
| preferOperatorsOnMasterNodes | boolean | `true` | configuration for different kloudlite operators used in this chart |
| svcAccountName | string | `"sa"` | k8s service account name, which all the pods installed by this chart uses, will always be of format <.Release.Name>-<.Values.svcAccountName> |
