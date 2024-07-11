# kloudlite-platform

[kloudlite-platform](https://github.com/kloudlite.io/helm-charts/charts/kloudlite-platform) Helm Chart for installing and setting up kloudlite platform on your own hosted Kubernetes clusters.

![Version: v1.0.5](https://img.shields.io/badge/Version-v1.0.5-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.0.5](https://img.shields.io/badge/AppVersion-v1.0.5-informational?style=flat-square)

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
| apps.accountsApi.configuration.grpcPort | int | `3001` |  |
| apps.accountsApi.configuration.httpPort | int | `3000` |  |
| apps.accountsApi.configuration.replicas | int | `1` |  |
| apps.accountsApi.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/api/accounts","tag":""}` | image (with tag) for accounts api |
| apps.auditLoggingWorker.configuration.replicas | int | `1` |  |
| apps.auditLoggingWorker.image.repository | string | `"ghcr.io/kloudlite/kloudlite/api/worker-audit-logging"` |  |
| apps.auditLoggingWorker.image.tag | string | `""` |  |
| apps.authApi.configuration.grpcPort | int | `3001` |  |
| apps.authApi.configuration.httpPort | int | `3000` |  |
| apps.authApi.configuration.replicas | int | `1` |  |
| apps.authApi.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/api/auth","tag":""}` | image (with tag) for auth api |
| apps.authWeb.configuration.httpPort | int | `3000` |  |
| apps.authWeb.configuration.replicas | int | `1` |  |
| apps.authWeb.enabled | bool | `true` |  |
| apps.authWeb.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/web/auth","tag":""}` | image (with tag) for auth web |
| apps.commsApi.configuration.grpcPort | int | `3001` |  |
| apps.commsApi.configuration.replicas | int | `1` |  |
| apps.commsApi.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/api/comms","tag":""}` | image (with tag) for comms api |
| apps.consoleApi.configuration.grpcPort | int | `3001` |  |
| apps.consoleApi.configuration.httpPort | int | `3000` |  |
| apps.consoleApi.configuration.logsAndMetricsHttpPort | int | `9100` |  |
| apps.consoleApi.configuration.replicas | int | `1` |  |
| apps.consoleApi.configuration.vpnDeviceNamespace | string | `"kl-vpn-devices"` |  |
| apps.consoleApi.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/api/console","tag":""}` | image (with tag) for console api |
| apps.consoleWeb.configuration.artifactHubKeyID | string | `""` |  |
| apps.consoleWeb.configuration.artifactHubKeySecret | string | `""` |  |
| apps.consoleWeb.configuration.httpPort | int | `3000` |  |
| apps.consoleWeb.configuration.replicas | int | `1` |  |
| apps.consoleWeb.enabled | bool | `true` |  |
| apps.consoleWeb.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/web/console","tag":""}` | image (with tag) for console web |
| apps.containerRegistryApi.configuration.authorizerPort | int | `4000` |  |
| apps.containerRegistryApi.configuration.buildClusterAccountName | string | `""` |  |
| apps.containerRegistryApi.configuration.buildClusterName | string | `""` |  |
| apps.containerRegistryApi.configuration.eventListenerPort | number | `4001` | port on which container registry event listener should listen |
| apps.containerRegistryApi.configuration.grpcPort | int | `3001` |  |
| apps.containerRegistryApi.configuration.httpPort | int | `3000` |  |
| apps.containerRegistryApi.configuration.jobBuildNamespace | string | `"kloudlite"` | namespace, in which build runs should be created |
| apps.containerRegistryApi.configuration.registrySecret | string | `""` |  |
| apps.containerRegistryApi.configuration.replicas | int | `1` |  |
| apps.containerRegistryApi.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/api/container-registry","tag":""}` | image (with tag) for container registry api |
| apps.gatewayApi.configuration.httpPort | int | `3000` |  |
| apps.gatewayApi.configuration.replicas | int | `1` |  |
| apps.gatewayApi.image.repository | string | `"ghcr.io/kloudlite/kloudlite/api/gateway"` |  |
| apps.gatewayApi.image.tag | string | `""` |  |
| apps.iamApi.configuration.grpcPort | int | `3001` |  |
| apps.iamApi.configuration.httpPort | int | `3000` |  |
| apps.iamApi.configuration.replicas | int | `1` |  |
| apps.iamApi.image.repository | string | `"ghcr.io/kloudlite/kloudlite/api/iam"` |  |
| apps.iamApi.image.tag | string | `""` |  |
| apps.infraApi.configuration.globalVpnKubeReverseProxyAuthzToken | string | `""` |  |
| apps.infraApi.configuration.grpcPort | int | `3001` |  |
| apps.infraApi.configuration.httpPort | int | `3000` |  |
| apps.infraApi.configuration.kloudliteRelease | string | `""` |  |
| apps.infraApi.configuration.replicas | int | `1` |  |
| apps.infraApi.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/api/infra","tag":""}` | image (with tag) for infra api |
| apps.iotConsoleApi.configuration.replicas | int | `1` |  |
| apps.iotConsoleApi.image.repository | string | `"ghcr.io/kloudlite/kloudlite/api/iot-console"` |  |
| apps.iotConsoleApi.image.tag | string | `""` |  |
| apps.klInstaller.configuration.replicas | int | `1` |  |
| apps.klInstaller.image.repository | string | `"ghcr.io/kloudlite/bin-installer"` |  |
| apps.klInstaller.image.tag | string | `""` |  |
| apps.messageOfficeApi.configuration.platformAccessToken | string | `"sample"` |  |
| apps.messageOfficeApi.configuration.replicas | int | `1` |  |
| apps.messageOfficeApi.configuration.tokenHashingSecret | string | `""` | consider using 128 characters random string, you can use `python -c "import secrets; print(secrets.token_urlsafe(128))"` |
| apps.messageOfficeApi.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/api/message-office","tag":""}` | image (with tag) for message office api |
| apps.observabilityApi.configuration.httpPort | string | `"3000"` |  |
| apps.observabilityApi.configuration.replicas | int | `1` |  |
| apps.observabilityApi.image.repository | string | `"ghcr.io/kloudlite/kloudlite/api/observability"` |  |
| apps.observabilityApi.image.tag | string | `""` |  |
| apps.webhooksApi.configuration.replicas | int | `1` |  |
| apps.webhooksApi.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/api/webhook","tag":""}` | image (with tag) for webhooks api |
| apps.websocketApi.configuration.replicas | int | `1` |  |
| apps.websocketApi.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/api/websocket-server","tag":""}` | image (with tag) for websocket-server api |
| aws.accessKey | string | `""` |  |
| aws.cloudformation.instanceProfileNamePrefix | string | `"kloudlite-instance-profile"` |  |
| aws.cloudformation.params.trustedARN | string | `"arn:aws:iam::855999427630:root"` |  |
| aws.cloudformation.roleNamePrefix | string | `"kloudlite-tenant-role"` |  |
| aws.cloudformation.stackNamePrefix | string | `"kloudlite-tenant-stack"` |  |
| aws.cloudformation.stackS3URL | string | `"https://kloudlite-platform-production-assets.s3.ap-south-1.amazonaws.com/public/cloudformation-templates/cloudformation.yml"` |  |
| aws.secretKey | string | `""` |  |
| certManager.certIssuer.acmeEmail | string | `"sample@example.com"` | email that should be used for communicating with lets-encrypt services |
| certManager.certIssuer.name | string | `"kloudlite-cluster-issuer"` |  |
| certManager.enabled | bool | `true` | whether to enable cert-manager |
| certManager.name | string | `"cert-manager"` |  |
| cloudflareWildCardCert.certificateName | string | `"cloudflare-wildcard-cert"` |  |
| cloudflareWildCardCert.cloudflareCreds | object | `{"email":"","secretToken":""}` | cloudflare authz credentials |
| cloudflareWildCardCert.cloudflareCreds.email | string | `""` | cloudflare authorized email |
| cloudflareWildCardCert.cloudflareCreds.secretToken | string | `""` | cloudflare authorized secret token |
| cloudflareWildCardCert.domains | list | `["*.platform.kloudlite.io"]` | list of all SANs (Subject Alternative Names) for which wildcard certs should be created |
| cloudflareWildCardCert.domains[0] | string | `"*.platform.kloudlite.io"` | should default to basedomain |
| cloudflareWildCardCert.enabled | bool | `true` |  |
| cloudflareWildCardCert.tlsSecretName | string | `"kl-cert-wildcard-tls"` |  |
| descheduler.enabled | bool | `true` |  |
| distribution.domain | string | `"cr.khost.dev"` |  |
| distribution.secret | string | `"<distribution-secret>"` |  |
| distribution.storage.driver | string | `""` | should be one of gcs or s3 |
| distribution.storage.gcs.bucket | string | `""` |  |
| distribution.storage.gcs.keyfileJson | string | `""` |  |
| distribution.storage.s3.accessKey | string | `""` |  |
| distribution.storage.s3.bucketName | string | `""` |  |
| distribution.storage.s3.enabled | bool | `false` |  |
| distribution.storage.s3.endpoint | string | `""` |  |
| distribution.storage.s3.region | string | `""` |  |
| distribution.storage.s3.secretKey | string | `""` |  |
| distribution.tls.enabled | bool | `true` |  |
| edgeGateways.secretKeyRef.key | string | `""` |  |
| edgeGateways.secretKeyRef.name | string | `""` | assumes, that the secret is in the same namespace as the chart |
| envVars.db.accountsDB | string | `"accounts-db"` |  |
| envVars.db.authDB | string | `"auth-db"` |  |
| envVars.db.commsDB | string | `"comms-db"` |  |
| envVars.db.consoleDB | string | `"console-db"` |  |
| envVars.db.eventsDB | string | `"events-db"` |  |
| envVars.db.iamDB | string | `"iam-db"` |  |
| envVars.db.infraDB | string | `"infra-db"` |  |
| envVars.db.iotConsoleDB | string | `"iot-console-db"` |  |
| envVars.db.messageOfficeDB | string | `"message-office-db"` |  |
| envVars.db.registryDB | string | `"registry-db"` |  |
| envVars.grpc.authGRPCAddr | string | `"auth-api:3001"` |  |
| envVars.nats.buckets.consoleCacheBucket.name | string | `"console-cache"` |  |
| envVars.nats.buckets.consoleCacheBucket.storage | string | `"file"` |  |
| envVars.nats.buckets.iotConsoleCacheBucket.name | string | `"iot-console-cache"` |  |
| envVars.nats.buckets.iotConsoleCacheBucket.storage | string | `"file"` |  |
| envVars.nats.buckets.resetTokenBucket.name | string | `"reset-token"` |  |
| envVars.nats.buckets.resetTokenBucket.storage | string | `"file"` |  |
| envVars.nats.buckets.sessionKVBucket.name | string | `"auth-session"` |  |
| envVars.nats.buckets.sessionKVBucket.storage | string | `"file"` |  |
| envVars.nats.buckets.verifyTokenBucket.name | string | `"verify-token"` |  |
| envVars.nats.buckets.verifyTokenBucket.storage | string | `"file"` |  |
| envVars.nats.streams.events.maxMsgBytes | string | `"500kB"` |  |
| envVars.nats.streams.events.maxMsgsPerSubject | int | `2` |  |
| envVars.nats.streams.events.name | string | `"events"` |  |
| envVars.nats.streams.events.subjects | string | `"events.>"` |  |
| envVars.nats.streams.logs.maxAge | string | `"3h"` |  |
| envVars.nats.streams.logs.maxMsgBytes | string | `"2MB"` |  |
| envVars.nats.streams.logs.name | string | `"logs"` |  |
| envVars.nats.streams.logs.subjects | string | `"logs.>"` |  |
| envVars.nats.streams.resourceSync.maxMsgBytes | string | `"500kB"` |  |
| envVars.nats.streams.resourceSync.name | string | `"resource-sync"` |  |
| envVars.nats.streams.resourceSync.subjects | string | `"resource-sync.>"` |  |
| envVars.nats.url | string | `"nats://nats:4222"` |  |
| global.accountName | string | `"kloudlite"` | kloudlite account name, required only for labelling purposes, does not need to be a real kloudlite account name |
| global.baseDomain | string | `"platform.kloudlite.io"` | base domain for all routers exposed through this culuster |
| global.clusterInternalDNS | string | `"cluster.local"` | cluster internal DNS name |
| global.clusterName | string | `"platform"` | kloudlite cluster name, required only for labelling purposes, does not need to be a real kloudlite cluster name |
| global.clusterSvcAccount | string | `"kloudlite-cluster-svc-account"` | service account for privileged k8s operations, like creating namespaces, apps, routers etc. |
| global.cookieDomain | string | `".kloudlite.io"` | cookie domain dictates at what domain, the cookies should be set for auth or other purposes |
| global.defaultProjectWorkspaceName | string | `"default"` | default project workspace name, the one that should be auto created, whenever you create a project |
| global.imagePullPolicy | string | `""` | image pull policies for kloudlite pods, belonging to this chart, could be Always | IfNotPresent |
| global.ingressClassName | string | `"ingress-nginx"` |  |
| global.isDev | bool | `false` |  |
| global.kloudlite_release | string | `""` | if not set, defaults to .Chart.AppVersion |
| global.nodeSelector | object | `{}` |  |
| global.normalSvcAccount | string | `"kloudlite-svc-account"` | service account for non k8s operations, just for specifying image pull secrets |
| global.podLabels | object | `{}` | podlabels for pods belonging to this release |
| global.routerDomain | string | `""` | router domain defaults to `global.baseDomain` |
| global.secondaryDomain | string | `"khost.dev"` |  |
| global.statefulPriorityClassName | string | `"stateful"` |  |
| global.tolerations | list | `[]` | tolerations for pods belonging to this release |
| grafana.configuration.nodeSelector | object | `{}` |  |
| grafana.configuration.volumeSize | string | `"2Gi"` |  |
| grafana.enabled | bool | `true` |  |
| grafana.name | string | `"grafana"` |  |
| ingressController.configuration.controllerKind | string | `"DaemonSet"` | can be DaemonSet or Deployment |
| ingressController.configuration.nodeSelector."node-role.kubernetes.io/control-plane" | string | `"true"` |  |
| ingressController.configuration.tolerations[0].effect | string | `"NoSchedule"` |  |
| ingressController.configuration.tolerations[0].key | string | `"node-role.kubernetes.io/master"` |  |
| ingressController.enabled | bool | `true` |  |
| ingressController.name | string | `"ingress-nginx"` |  |
| loki.configuration.s3credentials.awsAccessKeyId | string | `""` |  |
| loki.configuration.s3credentials.awsSecretAccessKey | string | `""` |  |
| loki.configuration.s3credentials.bucketName | string | `""` |  |
| loki.configuration.s3credentials.region | string | `""` |  |
| loki.enabled | bool | `false` |  |
| loki.name | string | `"loki-stack"` |  |
| longhorn.enabled | bool | `false` |  |
| mongo.configuration.nodeSelector | object | `{}` |  |
| mongo.configuration.volumeSize | string | `"2Gi"` |  |
| mongo.externalDB.authDBName | string | `""` |  |
| mongo.externalDB.dbURL | string | `""` |  |
| mongo.replicas | int | `1` |  |
| mongo.runAsCluster | bool | `false` |  |
| nats.configuration.password | string | `"sample"` |  |
| nats.configuration.user | string | `"sample"` |  |
| nats.configuration.volumeSize | string | `"10Gi"` |  |
| nats.replicas | int | `3` |  |
| nats.runAsCluster | bool | `false` |  |
| nodepools.iac.labels."kloudlite.io/nodepool.role" | string | `"iac"` |  |
| nodepools.iac.taints[0].effect | string | `"NoExecute"` |  |
| nodepools.iac.taints[0].key | string | `"kloudlite.io/nodepool.role"` |  |
| nodepools.iac.taints[0].value | string | `"iac"` |  |
| nodepools.iac.tolerations[0].effect | string | `"NoExecute"` |  |
| nodepools.iac.tolerations[0].key | string | `"kloudlite.io/nodepool.role"` |  |
| nodepools.iac.tolerations[0].operator | string | `"Equal"` |  |
| nodepools.iac.tolerations[0].value | string | `"iac"` |  |
| nodepools.stateful.labels."kloudlite.io/nodepool.role" | string | `"stateful"` |  |
| nodepools.stateful.taints[0].effect | string | `"NoExecute"` |  |
| nodepools.stateful.taints[0].key | string | `"kloudlite.io/nodepool.role"` |  |
| nodepools.stateful.taints[0].value | string | `"stateful"` |  |
| nodepools.stateful.tolerations[0].effect | string | `"NoExecute"` |  |
| nodepools.stateful.tolerations[0].key | string | `"kloudlite.io/nodepool.role"` |  |
| nodepools.stateful.tolerations[0].operator | string | `"Equal"` |  |
| nodepools.stateful.tolerations[0].value | string | `"stateful"` |  |
| nodepools.stateless.labels."kloudlite.io/nodepool.role" | string | `"stateless"` |  |
| nodepools.stateless.tolerations | list | `[]` |  |
| oAuth.enabled | bool | `true` |  |
| oAuth.providers.github.appId | string | `""` | GitHub app id |
| oAuth.providers.github.appPrivateKey | string | `""` | GitHub app private key (base64 encoded) |
| oAuth.providers.github.callbackUrl | string | `"https://auth.platform.kloudlite.io/oauth2/callback/github"` | GitHub oAuth2 callback url |
| oAuth.providers.github.clientId | string | `""` | (REQUIRED, if enabled) GitHub oAuth2 Client ID |
| oAuth.providers.github.clientSecret | string | `""` | (REQUIRED, if enabled) GitHub oAuth2 Client Secret |
| oAuth.providers.github.enabled | bool | `true` | whether to enable GitHub oAuth2 |
| oAuth.providers.github.githubAppName | string | `""` | GitHub app name, that we want to install on user's GitHub account |
| oAuth.providers.gitlab.callbackUrl | string | `"https://auth.platform.kloudlite.io/oauth2/callback/gitlab"` | gitlab oAuth2 callback url |
| oAuth.providers.gitlab.clientId | string | `""` | gitlab oAuth2 Client ID |
| oAuth.providers.gitlab.clientSecret | string | `""` | gitlab oAuth2 Client Secret |
| oAuth.providers.gitlab.enabled | bool | `true` | whether to enable gitlab oAuth2 |
| oAuth.providers.google.callbackUrl | string | `"https://auth.platform.kloudlite.io/oauth2/callback/google"` | google oAuth2 callback url |
| oAuth.providers.google.clientId | string | `""` | google oAuth2 Client ID |
| oAuth.providers.google.clientSecret | string | `""` | google oAuth2 Client Secret |
| oAuth.providers.google.enabled | bool | `true` | whether to enable google oAuth2 |
| oAuth.secretName | string | `"oauth-secrets"` | secret where all oauth credentials should be |
| operators.platformOperator.configuration.cluster.cloudflare.baseDomain | string | `""` | cloudflare base domain, on top of which CNAMES and wildcard names will be created |
| operators.platformOperator.configuration.cluster.cloudflare.zoneId | string | `""` | cloudflare zone id, to manage CNAMEs and A records for managed clusters |
| operators.platformOperator.configuration.cluster.jobImage.repository | string | `"ghcr.io/kloudlite/kloudlite/infrastructure-as-code/iac-job"` |  |
| operators.platformOperator.configuration.cluster.jobImage.tag | string | `""` |  |
| operators.platformOperator.configuration.helmCharts.jobImage.repository | string | `"ghcr.io/kloudlite/kloudlite/operator/workers/helm-job-runner"` |  |
| operators.platformOperator.configuration.helmCharts.jobImage.tag | string | `""` |  |
| operators.platformOperator.configuration.nodepools.aws.vpc_params.readFromCluster | bool | `true` |  |
| operators.platformOperator.configuration.nodepools.aws.vpc_params.secret.keys.vpcId | string | `"vpc_id"` |  |
| operators.platformOperator.configuration.nodepools.aws.vpc_params.secret.keys.vpcPublicSubnets | string | `"vpc_public_subnets"` |  |
| operators.platformOperator.configuration.nodepools.aws.vpc_params.secret.name | string | `"kloudlite-aws-settings"` |  |
| operators.platformOperator.configuration.nodepools.aws.vpc_params.secret.namespace | string | `"kube-system"` |  |
| operators.platformOperator.configuration.nodepools.cloudProviderName | string | `"aws"` |  |
| operators.platformOperator.configuration.nodepools.cloudProviderRegion | string | `"ap-south-1"` |  |
| operators.platformOperator.configuration.nodepools.enabled | bool | `true` |  |
| operators.platformOperator.configuration.nodepools.extractFromCluster | bool | `true` |  |
| operators.platformOperator.configuration.nodepools.k3sAgentJoinToken | string | `""` | k3s agent join token, as nodepools are effectively agent nodes |
| operators.platformOperator.configuration.nodepools.k3sServerPublicHost | string | `""` | k3s masters public DNS Host |
| operators.platformOperator.configuration.wireguard.enableExamples | bool | `true` |  |
| operators.platformOperator.configuration.wireguard.podCIDR | string | `"10.42.0.0/16"` | cluster pods CIDR range |
| operators.platformOperator.configuration.wireguard.svcCIDR | string | `"10.43.0.0/16"` | cluster services CIDR range |
| operators.platformOperator.enabled | bool | `true` | whether to enable platform operator |
| operators.platformOperator.image | object | `{"repository":"ghcr.io/kloudlite/kloudlite/operator/platform","tag":""}` | image (with tag) for platform operator |
| operators.preferOperatorsOnMasterNodes | bool | `true` |  |
| persistence.storageClasses.ext4 | string | `"sc-ext4"` | ext4 storage class name |
| persistence.storageClasses.xfs | string | `"sc-xfs"` | xfs storage class name |
| prometheus.configuration.alertmanager.volumeSize | string | `"2Gi"` |  |
| prometheus.configuration.prometheus.nodeSelector | object | `{}` |  |
| prometheus.configuration.prometheus.volumeSize | string | `"2Gi"` |  |
| prometheus.enabled | bool | `true` |  |
| prometheus.name | string | `"prometheus"` |  |
| sendGrid.apiKey | string | `""` | sendgrid api key for email communications, if (sendgrid.enabled) |
| sendGrid.supportEmail | string | `""` | email through which we should be sending emails to target users, if (sendgrid.enabled) |
| vector.enabled | bool | `true` |  |
| vector.name | string | `"vector"` |  |
| vectorAgent.description | string | `"vector agent for shipping logs to centralized vector aggregator"` |  |
| vectorAgent.enabled | bool | `true` |  |
| vectorAgent.name | string | `"vector-agent"` |  |
| victoriaMetrics.configuration.vmcluster.volumeSize | string | `"10Gi"` |  |
| victoriaMetrics.configuration.vmselect.volumeSize | string | `"2Gi"` |  |
| victoriaMetrics.enabled | bool | `true` |  |
| victoriaMetrics.name | string | `"victoria-metrics"` |  |
| webhookSecrets.authzSecret | string | `""` |  |
| webhookSecrets.githubAuthzSecret | string | `""` |  |
| webhookSecrets.githubSecret | string | `""` | webhook authz secret for GitHub webhooks |
| webhookSecrets.gitlabSecret | string | `""` | webhook authz secret for gitlab webhooks |
| webhookSecrets.kloudliteSecret | string | `""` | webhook authz secret for kloudlite internal calls |
| webhookSecrets.name | string | `"webhook-secrets"` |  |
