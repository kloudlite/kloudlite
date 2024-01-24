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
| apps.accountsApi.image | string | `"ghcr.io/kloudlite/api/accounts:v1.0.5-nightly"` | image (with tag) for accounts api |
| apps.auditLoggingWorker.image | string | `"ghcr.io/kloudlite/api/worker-audit-logging:v1.0.5-nightly"` | image (with tag) for audit logging worker |
| apps.authApi.image | string | `"ghcr.io/kloudlite/api/auth:v1.0.5-nightly"` | image (with tag) for auth api |
| apps.authWeb.enabled | bool | `true` |  |
| apps.authWeb.image | string | `"ghcr.io/kloudlite/web/auth:v1.0.5-nightly"` | image (with tag) for auth web |
| apps.commsApi.image | string | `"ghcr.io/kloudlite/api/comms:v1.0.5-nightly"` | image (with tag) for comms api |
| apps.consoleApi.configuration.consoleVPNDeviceNamespace | string | `"kloudlite-console-devices"` |  |
| apps.consoleApi.configuration.logsAndMetricsHttpPort | int | `9100` |  |
| apps.consoleApi.image | string | `"ghcr.io/kloudlite/api/console:v1.0.5-nightly"` | image (with tag) for console api |
| apps.consoleWeb.enabled | bool | `true` |  |
| apps.consoleWeb.image | string | `"ghcr.io/kloudlite/web/console:v1.0.5-nightly"` | image (with tag) for console web |
| apps.containerRegistryApi.configuration.authorizerPort | int | `4000` |  |
| apps.containerRegistryApi.configuration.buildClusterAccountName | string | `""` |  |
| apps.containerRegistryApi.configuration.buildClusterName | string | `""` |  |
| apps.containerRegistryApi.configuration.eventListenerPort | number | `4001` | port on which container registry event listener should listen |
| apps.containerRegistryApi.configuration.jobBuildNamespace | string | `"kloudlite"` | namespace, in which build runs should be created |
| apps.containerRegistryApi.configuration.registrySecret | string | `""` |  |
| apps.containerRegistryApi.image | string | `"ghcr.io/kloudlite/api/container-registry:v1.0.5-nightly"` | image (with tag) for container registry api |
| apps.gatewayApi.image | string | `"ghcr.io/kloudlite/api/gateway:v1.0.5-nightly"` |  |
| apps.iamApi.image | string | `"ghcr.io/kloudlite/api/iam:v1.0.5-nightly"` | image (with tag) for iam api |
| apps.infraApi.configuration.infraVPNDeviceNamespace | string | `"kloudlite-infra-devices"` |  |
| apps.infraApi.image | string | `"ghcr.io/kloudlite/api/infra:v1.0.5-nightly"` | image (with tag) for infra api |
| apps.klInstaller.image | string | `"ghcr.io/kloudlite/kl/installer:v1.0.5-nightly"` | image (with tag) for comms api |
| apps.messageOfficeApi.configuration.platformAccessToken | string | `"sample"` |  |
| apps.messageOfficeApi.configuration.tokenHashingSecret | string | `""` | consider using 128 characters random string, you can use `python -c "import secrets; print(secrets.token_urlsafe(128))"` |
| apps.messageOfficeApi.image | string | `"ghcr.io/kloudlite/api/message-office:v1.0.5-nightly"` | image (with tag) for message office api |
| apps.webhooksApi.image | string | `"ghcr.io/kloudlite/api/webhook:v1.0.5-nightly"` | image (with tag) for webhooks api |
| apps.websocketApi.image | string | `"ghcr.io/kloudlite/api/websocket-server:v1.0.5-nightly"` | image (with tag) for websocket-server api |
| aws.accessKey | string | `""` |  |
| aws.cloudformation.instanceProfileNamePrefix | string | `"kloudlite-instance-profile"` |  |
| aws.cloudformation.params.trustedARN | string | `"arn:aws:iam::855999427630:root"` |  |
| aws.cloudformation.roleNamePrefix | string | `"kloudlite-tenant-role"` |  |
| aws.cloudformation.stackNamePrefix | string | `"kloudlite-tenant-stack"` |  |
| aws.cloudformation.stackS3URL | string | `"https://kloudlite-platform-production-assets.s3.ap-south-1.amazonaws.com/public/cloudformation-templates/cloudformation.yml"` |  |
| aws.secretKey | string | `""` |  |
| certManager.certIssuer.acmeEmail | string | `"sample@example.com"` | email that should be used for communicating with lets-encrypt services |
| certManager.certIssuer.name | string | `"kloudlite-cluster-issuer"` |  |
| certManager.configuration.nodeSelector | object | `{}` |  |
| certManager.configuration.tolerations | list | `[]` |  |
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
| distribution.s3.accessKey | string | `""` |  |
| distribution.s3.bucketName | string | `""` |  |
| distribution.s3.enabled | bool | `false` |  |
| distribution.s3.endpoint | string | `""` |  |
| distribution.s3.region | string | `""` |  |
| distribution.s3.secretKey | string | `""` |  |
| distribution.secret | string | `"<distribution-secret>"` |  |
| distribution.tls.enabled | bool | `true` |  |
| envVars.db.accountsDB | string | `"accounts-db"` |  |
| envVars.db.authDB | string | `"auth-db"` |  |
| envVars.db.consoleDB | string | `"console-db"` |  |
| envVars.db.eventsDB | string | `"events-db"` |  |
| envVars.db.iamDB | string | `"iam-db"` |  |
| envVars.db.infraDB | string | `"infra-db"` |  |
| envVars.db.messageOfficeDB | string | `"message-office-db"` |  |
| envVars.db.registryDB | string | `"registry-db"` |  |
| envVars.grpc.authGRPCAddr | string | `"auth-api:3001"` |  |
| envVars.nats.buckets.consoleCacheBucket.name | string | `"console-cache"` |  |
| envVars.nats.buckets.consoleCacheBucket.storage | string | `"file"` |  |
| envVars.nats.buckets.resetTokenBucket.name | string | `"reset-token"` |  |
| envVars.nats.buckets.resetTokenBucket.storage | string | `"file"` |  |
| envVars.nats.buckets.sessionKVBucket.name | string | `"auth-session"` |  |
| envVars.nats.buckets.sessionKVBucket.storage | string | `"file"` |  |
| envVars.nats.buckets.verifyTokenBucket.name | string | `"verify-token"` |  |
| envVars.nats.buckets.verifyTokenBucket.storage | string | `"file"` |  |
| envVars.nats.streams.events.maxMsgBytes | string | `"2MB"` |  |
| envVars.nats.streams.events.name | string | `"events"` |  |
| envVars.nats.streams.events.subjects | string | `"events.>"` |  |
| envVars.nats.streams.logs.maxAge | string | `"3h"` |  |
| envVars.nats.streams.logs.maxMsgBytes | string | `"2MB"` |  |
| envVars.nats.streams.logs.name | string | `"logs"` |  |
| envVars.nats.streams.logs.subjects | string | `"logs.>"` |  |
| envVars.nats.streams.resourceSync.maxMsgBytes | string | `"2MB"` |  |
| envVars.nats.streams.resourceSync.name | string | `"resource-sync"` |  |
| envVars.nats.streams.resourceSync.subjects | string | `"resource-sync.>"` |  |
| envVars.nats.url | string | `"nats://nats:4222"` |  |
| global.accountName | string | `"kloudlite"` | kloudlite account name, required only for labelling purposes, does not need to be a real kloudlite account name |
| global.baseDomain | string | `"platform.kloudlite.io"` | base domain for all routers exposed through this cluster |
| global.clusterInternalDNS | string | `"cluster.local"` | cluster internal DNS name |
| global.clusterName | string | `"platform"` | kloudlite cluster name, required only for labelling purposes, does not need to be a real kloudlite cluster name |
| global.clusterSvcAccount | string | `"kloudlite-cluster-svc-account"` | service account for privileged k8s operations, like creating namespaces, apps, routers etc. |
| global.cookieDomain | string | `".kloudlite.io"` | cookie domain dictates at what domain, the cookies should be set for auth or other purposes |
| global.defaultProjectWorkspaceName | string | `"default"` | default project workspace name, the one that should be auto created, whenever you create a project |
| global.imagePullPolicy | string | `"Always"` | image pull policies for kloudlite pods, belonging to this chart |
| global.ingressClassName | string | `"ingress-nginx"` |  |
| global.isDev | bool | `false` |  |
| global.kloudlite_release | string | `"v1.0.5-nightly"` |  |
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
| mongo.externalDB.authDBName | string | `""` |  |
| mongo.externalDB.dbURL | string | `""` |  |
| mongo.nodeSelector | object | `{}` |  |
| mongo.replicas | int | `1` |  |
| mongo.runAsCluster | bool | `false` |  |
| mongo.size | string | `"2Gi"` |  |
| nats.replicas | int | `3` |  |
| nats.runAsCluster | bool | `false` |  |
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
| operators.platformOperator.configuration.cluster.IACStateStore.accessKey | string | `""` |  |
| operators.platformOperator.configuration.cluster.IACStateStore.s3BucketDir | string | `"terraform-states"` |  |
| operators.platformOperator.configuration.cluster.IACStateStore.s3BucketName | string | `"kloudlite-dev-tf"` | s3 bucket name, to store kloudlite's infrastructure-as-code remote state |
| operators.platformOperator.configuration.cluster.IACStateStore.s3BucketRegion | string | `"ap-south-1"` | s3 bucket region, to store kloudlite's infrastructure-as-code remote state |
| operators.platformOperator.configuration.cluster.IACStateStore.secretKey | string | `""` |  |
| operators.platformOperator.configuration.cluster.cloudflare.baseDomain | string | `""` | cloudflare base domain, on top of which CNAMES and wildcard names will be created, defaults to `global.baseDomain` |
| operators.platformOperator.configuration.cluster.cloudflare.zoneId | string | `""` | cloudflare zone id, to manage CNAMEs and A records for managed clusters |
| operators.platformOperator.configuration.cluster.jobImage | string | `"ghcr.io/kloudlite/infrastructure-as-code:v1.0.5-nightly"` |  |
| operators.platformOperator.configuration.nodepool.cloudProviderName | string | `"aws"` |  |
| operators.platformOperator.configuration.nodepool.cloudProviderRegion | string | `"ap-south-1"` |  |
| operators.platformOperator.configuration.nodepool.k3sAgentJoinToken | string | `""` | k3s agent join token, as nodepools are effectively agent nodes |
| operators.platformOperator.configuration.nodepool.k3sServerPublicHost | string | `""` | k3s masters public DNS Host |
| operators.platformOperator.enabled | bool | `true` | whether to enable platform operator |
| operators.platformOperator.image | string | `"ghcr.io/kloudlite/operator/platform:v1.0.5-nightly"` | image (with tag) for platform operator |
| operators.preferOperatorsOnMasterNodes | bool | `true` |  |
| operators.wgOperator.configuration | object | `{"enableExamples":false,"podCIDR":"10.42.0.0/16","svcCIDR":"10.43.0.0/16"}` | wireguard configuration options |
| operators.wgOperator.configuration.podCIDR | string | `"10.42.0.0/16"` | cluster pods CIDR range |
| operators.wgOperator.configuration.svcCIDR | string | `"10.43.0.0/16"` | cluster services CIDR range |
| operators.wgOperator.enabled | bool | `false` |  |
| operators.wgOperator.image | string | `"ghcr.io/kloudlite/operator/wireguard:v1.0.5-nightly"` | wg operator image and tag |
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
| victoriaMetrics.configuration.nodeSelector | object | `{}` |  |
| victoriaMetrics.configuration.volumeSize | string | `"2Gi"` |  |
| victoriaMetrics.enabled | bool | `true` |  |
| victoriaMetrics.name | string | `"victoria-metrics"` |  |
| webhookSecrets.authzSecret | string | `""` |  |
| webhookSecrets.githubAuthzSecret | string | `""` |  |
| webhookSecrets.githubSecret | string | `""` | webhook authz secret for GitHub webhooks |
| webhookSecrets.gitlabSecret | string | `""` | webhook authz secret for gitlab webhooks |
| webhookSecrets.kloudliteSecret | string | `""` | webhook authz secret for kloudlite internal calls |
| webhookSecrets.name | string | `"webhook-secrets"` |  |
