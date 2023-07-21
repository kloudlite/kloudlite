# kloudlite-agent

[kloudlite-agent](https://github.com/kloudlite.io/helm-charts/charts/kloudlite-agent) Kloudlite Agent to make your kubernetes cluster communicate securely with kloudlite control plane

![Version: 1.0.5-nightly](https://img.shields.io/badge/Version-1.0.5--nightly-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.0.5-nightly](https://img.shields.io/badge/AppVersion-1.0.5--nightly-informational?style=flat-square)

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://helm.vector.dev | vector | 0.23.0 |

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
| accountName | string | `"‼️ Required"` | kloudlite account name |
| agent.enabled | bool | `true` | enable/disable kloudlite agent |
| agent.image | string | `"ghcr.io/kloudlite/agents/kl-agent:v1.0.5-nightly"` | kloudlite agent image name and tag |
| clusterIdentitySecretName | string | `"kl-cluster-identity"` | cluster identity secret name, which keeps cluster token and access token |
| clusterName | string | `"‼️ Required"` | kloudlite cluster name |
| clusterToken | string | `"‼️ Required"` | kloudlite issued cluster token |
| defaultImagePullSecretName | string | `"kl-image-pull-creds"` | default image pull secret name, defaults to kl-image-pull-creds |
| imagePullPolicy | string | `"Always"` | container image pull policy |
| messageOfficeGRPCAddr | string | `"message-office-api.dev.kloudlite.io:443"` | kloudlite message office api grpc address, should be in the form of 'grpc-host:grcp-port' |
| operators | object | `{"resourceWatcher":{"enabled":true,"image":"ghcr.io/kloudlite/agents/resource-watcher:v1.0.5-nightly"},"wgOperator":{"configuration":{"dnsHostedZone":"<dns-hosted-zone>","podCIDR":"10.42.0.0/16","svcCIDR":"10.43.0.0/16"},"enabled":true,"image":"ghcr.io/kloudlite/agent/operator/wg:v1.0.5-nightly"}}` | configuration for different kloudlite operators used in this chart |
| operators.resourceWatcher.enabled | bool | `true` | enable/disable kloudlite resource watcher |
| operators.resourceWatcher.image | string | `"ghcr.io/kloudlite/agents/resource-watcher:v1.0.5-nightly"` | kloudlite resource watcher image name and tag |
| operators.wgOperator.configuration | object | `{"dnsHostedZone":"<dns-hosted-zone>","podCIDR":"10.42.0.0/16","svcCIDR":"10.43.0.0/16"}` | wireguard configuration options |
| operators.wgOperator.configuration.dnsHostedZone | string | `"<dns-hosted-zone>"` | dns hosted zone, i.e. dns pointing to this cluster |
| operators.wgOperator.configuration.podCIDR | string | `"10.42.0.0/16"` | cluster pods CIDR range |
| operators.wgOperator.configuration.svcCIDR | string | `"10.43.0.0/16"` | cluster services CIDR range |
| operators.wgOperator.enabled | bool | `true` | whether to enable wg operator |
| operators.wgOperator.image | string | `"ghcr.io/kloudlite/agent/operator/wg:v1.0.5-nightly"` | wg operator image and tag |
| svcAccountName | string | `"cluster-svc-account"` | k8s service account name, which all the pods installed by this chart uses |
| vector.containerPorts[0].containerPort | int | `6000` |  |
| vector.customConfig.api.address | string | `"127.0.0.1:8686"` |  |
| vector.customConfig.api.enabled | bool | `true` |  |
| vector.customConfig.api.playground | bool | `false` |  |
| vector.customConfig.data_dir | string | `"/vector-data-dir"` |  |
| vector.customConfig.sinks.kloudlite_hosted_vector | object | `{"address":"kl-agent.kl-init-operators.svc.cluster.local:6000","inputs":["kubernetes_logs","kubelet_metrics_exporter"],"type":"vector"}` | custom configuration |
| vector.customConfig.sinks.stdout | string | `nil` |  |
| vector.customConfig.sources.host_metrics | string | `nil` |  |
| vector.customConfig.sources.internal_metrics | string | `nil` |  |
| vector.customConfig.sources.kubelet_metrics_exporter.endpoints[0] | string | `"http://localhost:9999/metrics/resource"` |  |
| vector.customConfig.sources.kubelet_metrics_exporter.type | string | `"prometheus_scrape"` |  |
| vector.customConfig.sources.kubernetes_logs.type | string | `"kubernetes_logs"` |  |
| vector.extraContainers[0].args[0] | string | `"--addr"` |  |
| vector.extraContainers[0].args[10] | string | `"kloudlite.io/=kl_"` |  |
| vector.extraContainers[0].args[1] | string | `"0.0.0.0:9999"` |  |
| vector.extraContainers[0].args[2] | string | `"--enrich-from-annotations"` |  |
| vector.extraContainers[0].args[3] | string | `"--enrich-tag"` |  |
| vector.extraContainers[0].args[4] | string | `"kl_account_name=‼️ Required"` |  |
| vector.extraContainers[0].args[5] | string | `"--enrich-tag"` |  |
| vector.extraContainers[0].args[6] | string | `"kl_cluster_name=‼️ Required"` |  |
| vector.extraContainers[0].args[7] | string | `"--filter-prefix"` |  |
| vector.extraContainers[0].args[8] | string | `"kloudlite.io/"` |  |
| vector.extraContainers[0].args[9] | string | `"--replace-prefix"` |  |
| vector.extraContainers[0].env[0].name | string | `"NODE_NAME"` |  |
| vector.extraContainers[0].env[0].valueFrom.fieldRef.fieldPath | string | `"spec.nodeName"` |  |
| vector.extraContainers[0].image | string | `"ghcr.io/nxtcoder17/kubelet-metrics-reexporter:v1.0.0"` |  |
| vector.extraContainers[0].name | string | `"kubelet-metrics-reexporter"` |  |
| vector.install | bool | `true` |  |
| vector.role | string | `"Agent"` |  |
| vector.service.enabled | bool | `false` |  |
| vector.serviceAccount.create | bool | `false` |  |
| vector.serviceAccount.name | string | `"vector-svc-account"` |  |
| vector.serviceHeadless.enabled | bool | `false` |  |
| vectorSvcAccountName | string | `"vector-svc-account"` | vector service account name, which all the vector pods will use |
