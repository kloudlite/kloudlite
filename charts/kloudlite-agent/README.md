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
| accountName | string REQUIRED | `""` | kloudlite account name |
| agent.enabled | bool | `true` | enable/disable kloudlite agent |
| agent.image | object | `{"pullPolicy":"","repository":"ghcr.io/kloudlite/agents/kl-agent","tag":""}` | kloudlite agent image name and tag |
| agent.image.pullPolicy | string | `""` | image pull policy for kloudlite agent, default is .imagePullPolicy |
| agent.image.tag | string | `""` | image tag for kloudlite agent, by default uses kloudlite_release |
| agent.nodeSelector | object | `{}` |  |
| agent.tolerations | list | `[]` |  |
| clusterIdentitySecretName | string | `"kl-cluster-identity"` | cluster identity secret name, which keeps cluster token and access token |
| clusterInternalDNS | string | `"cluster.local"` | cluster internal DNS, like 'cluster.local' |
| clusterName | string REQUIRED | `""` | kloudlite cluster name |
| clusterToken | string REQUIRED | `""` | kloudlite issued cluster token |
| helmCharts.certManager.affinity | object | `{}` |  |
| helmCharts.certManager.configuration.clusterIssuers[0].acme.email | string | `"support@kloudlite.io"` |  |
| helmCharts.certManager.configuration.clusterIssuers[0].acme.server | string | `"https://acme-v02.api.letsencrypt.org/directory"` |  |
| helmCharts.certManager.configuration.clusterIssuers[0].default | bool | `true` |  |
| helmCharts.certManager.configuration.clusterIssuers[0].name | string | `"letsencrypt-prod"` |  |
| helmCharts.certManager.configuration.defaultClusterIssuer | string | `"letsencrypt-prod"` |  |
| helmCharts.certManager.enabled | bool | `true` |  |
| helmCharts.certManager.name | string | `"cert-manager"` |  |
| helmCharts.certManager.nodeSelector | object | `{}` |  |
| helmCharts.certManager.tolerations | list | `[]` |  |
| helmCharts.clusterAutoscaler.enabled | bool | `true` |  |
| helmCharts.ingressNginx.configuration.controllerKind | string | `"DaemonSet"` |  |
| helmCharts.ingressNginx.configuration.ingressClassName | string | `"nginx"` |  |
| helmCharts.ingressNginx.enabled | bool | `true` |  |
| helmCharts.ingressNginx.name | string | `"ingress-nginx"` |  |
| helmCharts.ingressNginx.nodeSelector | object | `{}` |  |
| helmCharts.ingressNginx.tolerations | list | `[]` |  |
| helmCharts.vector.debugOnStdout | bool | `false` |  |
| helmCharts.vector.enabled | bool | `true` |  |
| helmCharts.vector.name | string | `"vector"` |  |
| helmCharts.vector.nodeSelector | object | `{}` |  |
| helmCharts.vector.tolerations | list | `[]` |  |
| imagePullPolicy | string | `"Always"` | container image pull policy |
| jobsNamespace | string | `"kloudlite-jobs"` |  |
| kloudliteRelease | string | `""` | kloudlite release version |
| messageOfficeGRPCAddr | string | `""` | kloudlite message office api grpc address, should be in the form of 'grpc-host:grcp-port', grpc-api.domain.com:443 |
| operators.agentOperator.configuration.helmCharts.jobImage.repository | string | `"ghcr.io/kloudlite/operator/workers/helm-runner"` |  |
| operators.agentOperator.configuration.helmCharts.jobImage.tag | string | `""` |  |
| operators.agentOperator.configuration.iacJobImage.repository | string | `"ghcr.io/kloudlite/infrastructure-as-code/iac-job"` |  |
| operators.agentOperator.configuration.iacJobImage.tag | string | `""` |  |
| operators.agentOperator.configuration.letsEncryptSupportEmail | string | `"support@kloudlite.io"` |  |
| operators.agentOperator.configuration.wireguard.deviceNamespace | string | `"kl-vpn-devices"` |  |
| operators.agentOperator.configuration.wireguard.podCIDR | string | `"10.42.0.0/16"` |  |
| operators.agentOperator.configuration.wireguard.svcCIDR | string | `"10.43.0.0/16"` |  |
| operators.agentOperator.enabled | bool | `true` | enable/disable kloudlite agent operator |
| operators.agentOperator.image | object | `{"pullPolicy":"","repository":"ghcr.io/kloudlite/operator/agent","tag":""}` | kloudlite resource watcher image name and tag |
| operators.agentOperator.image.pullPolicy | string | `""` | image pullPolicy for kloudlite resource watcher, by default uses .Chart.AppVersion |
| operators.agentOperator.image.tag | string | `""` | image tag for kloudlite resource watcher, by default uses .Chart.AppVersion |
| operators.agentOperator.name | string | `"kl-agent-operator"` | workload name for kloudlite agent operator |
| operators.agentOperator.nodeSelector | object | `{}` |  |
| operators.agentOperator.tolerations | list | `[]` |  |
| preferOperatorsOnMasterNodes | boolean | `true` | configuration for different kloudlite operators used in this chart |
| svcAccountName | string | `"sa"` | k8s service account name, which all the pods installed by this chart uses, will always be of format <.Release.Name>-<.Values.svcAccountName> |
