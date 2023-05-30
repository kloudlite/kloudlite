# kloudlite-platform

[kloudlite-platform](https://github.com/kloudlite.io/helm-charts/charts/kloudlite-platform) A Helm chart for Kubernetes

![Version: 1.0.5-nightly](https://img.shields.io/badge/Version-1.0.5--nightly-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.0.5-nightly](https://img.shields.io/badge/AppVersion-1.0.5--nightly-informational?style=flat-square)

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.vectorized.io | redpanda-operator | 22.1.6 |
| https://kubernetes.github.io/ingress-nginx | ingress-nginx | 4.6.0 |

## Get Repo Info

```console
helm repo add kloudlite https://kloudlite.github.io/helm-charts
helm repo update
```

## Install Chart

**Important:** only helm3 is supported
**Important:** [kloudlite-operators](../kloudlite-operators) must be installed beforehand
**Important:** ensure kloudlite CRDs have been installed

```console
helm install [RELEASE_NAME] kloudlite/kloudlite-platform --namespace kl-init-operators --create-namespace
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
helm install  [RELEASE_NAME] kloudlite/kloudlite-platform --version [NIGHTLY_VERSION] --namespace kl-init-operators --create-namespace
```

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] kloudlite/kloudlite-platform --install
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
| apps.accountsWeb.image | string | `"ghcr.io/kloudlite/platform/web/accounts-web:v1.0.5-nightly"` | image (with tag) for accounts web |
| apps.auditLoggingWorker.image | string | `"ghcr.io/kloudlite/platform/apis/audit-logging-worker:v1.0.5-nightly"` | image (with tag) for audit logging worker |
| apps.authApi.configuration.oAuth2.enabled | bool | `true` | whether to enable oAuth2 |
| apps.authApi.configuration.oAuth2.github.appId | string | `"<github-app-id>"` | github app Id |
| apps.authApi.configuration.oAuth2.github.appPrivateKey | string | `"<github-app-private-key>"` | github app private key (base64 encoded) |
| apps.authApi.configuration.oAuth2.github.callbackUrl | string | `"https://auth.dev.kloudlite.io/oauth2/callback/github"` | github oAuth2 callback url |
| apps.authApi.configuration.oAuth2.github.clientId | string | `"<github-client-id>"` | github oAuth2 Client ID |
| apps.authApi.configuration.oAuth2.github.clientSecret | string | `"<github-client-secret>"` | github oAuth2 Client Secret |
| apps.authApi.configuration.oAuth2.github.enabled | bool | `true` | whether to enable github oAuth2 |
| apps.authApi.configuration.oAuth2.github.githubAppName | string | `"kloudlite-dev"` | github app name, that we want to install on user's github account |
| apps.authApi.configuration.oAuth2.gitlab.callbackUrl | string | `"https://auth.dev.kloudlite.io/oauth2/callback/gitlab"` | gitlab oAuth2 callback url |
| apps.authApi.configuration.oAuth2.gitlab.clientId | string | `"<gitlab-client-id>"` | gitlab oAuth2 Client ID |
| apps.authApi.configuration.oAuth2.gitlab.clientSecret | string | `"<gitlab-client-secret>"` | gitlab oAuth2 Client Secret |
| apps.authApi.configuration.oAuth2.gitlab.enabled | bool | `true` | whether to enable gitlab oAuth2 |
| apps.authApi.configuration.oAuth2.google.callbackUrl | string | `"https://auth.dev.kloudlite.io/oauth2/callback/google"` | google oAuth2 callback url |
| apps.authApi.configuration.oAuth2.google.clientId | string | `"<google-client-id>"` | google oAuth2 Client ID |
| apps.authApi.configuration.oAuth2.google.clientSecret | string | `"<google-client-secret>"` | google oAuth2 Client Secret |
| apps.authApi.configuration.oAuth2.google.enabled | bool | `true` | whether to enable google oAuth2 |
| apps.authApi.image | string | `"ghcr.io/kloudlite/platform/apis/auth-api:v1.0.5-nightly"` | image (with tag) for auth api |
| apps.authWeb.image | string | `"ghcr.io/kloudlite/platform/web/accounts-web:v1.0.5-nightly"` | image (with tag) for auth web |
| apps.commsApi.configuration | object | `{"sendgridApiKey":"<sendgrid-api-key>","supportEmail":"<support-email>"}` | configurations for comms api |
| apps.commsApi.configuration.sendgridApiKey | string | `"<sendgrid-api-key>"` | sendgrid api key for email communications, if (sendgrid.enabled) |
| apps.commsApi.configuration.supportEmail | string | `"<support-email>"` | email through which we should be sending emails to target users, if (sendgrid.enabled) |
| apps.commsApi.enabled | bool | `true` | whether to enable communications api |
| apps.commsApi.image | string | `"ghcr.io/kloudlite/platform/apis/comms-api:v1.0.5-nightly"` | image (with tag) for comms api |
| apps.consoleApi.configuration | object | `{}` |  |
| apps.consoleApi.image | string | `"ghcr.io/kloudlite/platform/apis/console-api:v1.0.5-nightly"` | image (with tag) for console api |
| apps.consoleWeb.image | string | `"ghcr.io/kloudlite/platform/web/console-web:v1.0.5-nightly"` | image (with tag) for console web |
| apps.containerRegistryApi.configuration.harbor.adminPassword | string | `"<harbor-admin-password>"` | harbor api admin password |
| apps.containerRegistryApi.configuration.harbor.adminUsername | string | `"<harbor-admin-username>"` | harbor api admin username |
| apps.containerRegistryApi.configuration.harbor.apiVersion | string | `"v2.0"` | harbor api version |
| apps.containerRegistryApi.configuration.harbor.imageRegistryHost | string | `"<harbor-registry-host>"` | harbor image registry host |
| apps.containerRegistryApi.configuration.harbor.webhookAuthz | string | `"<harbor-webhook-authz>"` | harbor webhook authz secret |
| apps.containerRegistryApi.configuration.harbor.webhookEndpoint | string | `"https://webhooks.dev.kloudlite.io/harbor"` | harbor webhook endpoint, (for receiving webhooks for every images pushed) |
| apps.containerRegistryApi.configuration.harbor.webhookName | string | `"kloudlite-dev-webhook"` | harbor webhook name |
| apps.containerRegistryApi.enabled | bool | `true` |  |
| apps.containerRegistryApi.image | string | `"ghcr.io/kloudlite/platform/apis/container-registry-api:v1.0.5-nightly"` | image (with tag) for container registry api |
| apps.dnsApi.configuration | object | `{"dnsNames":["ns1.dev.kloudlite.io"],"edgeCNAME":"edge.dev.kloudlite.io"}` | configurations for dns api |
| apps.dnsApi.configuration.dnsNames | list | `["ns1.dev.kloudlite.io"]` | list of all dnsNames for which, you want wildcard certificate to be issued for |
| apps.dnsApi.configuration.edgeCNAME | string | `"edge.dev.kloudlite.io"` | base domain for CNAME for all the edges managed (or, to be managed) by this cluster |
| apps.dnsApi.image | string | `"ghcr.io/kloudlite/platform/apis/dns-api:v1.0.5-nightly"` | image (with tag) for dns api |
| apps.financeApi.image | string | `"ghcr.io/kloudlite/platform/apis/finance-api:v1.0.5-nightly"` | image (with tag) for finance api |
| apps.gatewayApi.image | string | `"ghcr.io/kloudlite/platform/apis/gateway-api:v1.0.5-nightly"` | image (with tag) for container registry api |
| apps.iamApi.configuration | object | `{}` |  |
| apps.iamApi.image | string | `"ghcr.io/kloudlite/platform/apis/iam-api:v1.0.5-nightly"` | image (with tag) for iam api |
| apps.infraApi.image | string | `"ghcr.io/kloudlite/platform/apis/infra-api:v1.0.5-nightly"` | image (with tag) for infra api |
| apps.messageOfficeApi.image | string | `"ghcr.io/kloudlite/platform/apis/message-office-api:v1.0.5-nightly"` | image (with tag) for message office api |
| apps.socketWeb.image | string | `"ghcr.io/kloudlite/platform/web/socket-web:v1.0.5-nightly"` | image (with tag) for socket web |
| apps.webhooksApi.configuration.webhookAuthz.githubSecret | string | `"<webhook-authz-github-secret>"` | webhook authz secret for github webhooks |
| apps.webhooksApi.configuration.webhookAuthz.gitlabSecret | string | `"<webhook-authz-gitlab-secret>"` | webhook authz secret for gitlab webhooks |
| apps.webhooksApi.configuration.webhookAuthz.harborSecret | string | `"<harbor-webhook-authz>"` | webhook authz secret for harbor webhooks |
| apps.webhooksApi.configuration.webhookAuthz.kloudliteSecret | string | `"<webhook-authz-kloudlite-secret>"` | webhook authz secret for kloudlite internal calls |
| apps.webhooksApi.enabled | bool | `true` |  |
| apps.webhooksApi.image | string | `"ghcr.io/kloudlite/platform/apis/webhooks-api:v1.0.5-nightly"` | image (with tag) for webhooks api |
| baseDomain | string | `"dev.kloudlite.io"` | base domain for all routers exposed through this cluster |
| clusterIssuer.acmeEmail | string | `"sample@example.com"` | email that should be used for communicating with letsencrypt services |
| clusterIssuer.cloudflareWildCardCert.cloudflareCreds | object | `{"email":"<cloudflare-email>","secretToken":"<cloudflare-secret-token>"}` | cloudflare authz credentials |
| clusterIssuer.cloudflareWildCardCert.cloudflareCreds.email | string | `"<cloudflare-email>"` | cloudflare authorized email |
| clusterIssuer.cloudflareWildCardCert.cloudflareCreds.secretToken | string | `"<cloudflare-secret-token>"` | cloudflare authorized secret token |
| clusterIssuer.cloudflareWildCardCert.create | bool | `true` |  |
| clusterIssuer.cloudflareWildCardCert.domains | list | `["*.dev.kloudlite.io"]` | list of all SANs (Subject Alternative Names) for which wildcard certs should be created |
| clusterIssuer.cloudflareWildCardCert.name | string | `"kl-cert-wildcard"` | name for wildcard cert |
| clusterIssuer.cloudflareWildCardCert.secretName | string | `"kl-cert-wildcard-tls"` | k8s secret where wildcard cert should be stored |
| clusterIssuer.create | bool | `true` |  |
| clusterIssuer.name | string | `"cluster-issuer"` | name of cluster issuer, to be used for issuing wildcard cert |
| clusterSvcAccount | string | `"kloudlite-cluster-svc-account"` | service account for privileged k8s operations, like creating namespaces, apps, routers etc. |
| cookieDomain | string | `".kloudlite.io"` | cookie domain dictates at what domain, the cookies should be set for auth or other purposes |
| defaultProjectWorkspaceName | string | `"default"` | default project workspace name, that should be auto created, whenever you create a project |
| imagePullPolicy | string | `"Always"` | image pull policies for kloudlite pods, belonging to this chart |
| ingress-nginx | object | `{"controller":{"admissionWebhooks":{"enabled":false,"failurePolicy":"Ignore"},"dnsPolicy":"ClusterFirstWithHostNet","electionID":"ingress-nginx","extraArgs":{"default-ssl-certificate":"kl-init-operators/kl-cert-wildcard-tls"},"hostNetwork":true,"hostPort":{"enabled":true,"ports":{"healthz":10254,"http":80,"https":443}},"ingressClass":"ingress-nginx","ingressClassByName":true,"ingressClassResource":{"controllerValue":"k8s.io/ingress-nginx","enabled":true,"name":"ingress-nginx"},"kind":"DaemonSet","podLabels":{},"resources":{"requests":{"cpu":"100m","memory":"200Mi"}},"service":{"type":"ClusterIP"},"watchIngressWithoutClass":false},"create":true,"nameOverride":"ingress-nginx","rbac":{"create":false},"serviceAccount":{"create":false,"name":"kloudlite-cluster-svc-account"}}` | ingress nginx configurations, read more at https://kubernetes.github.io/ingress-nginx/ |
| ingressClassName | string | `"ingress-nginx"` | ingress class name that should be used for all the ingresses, created by this chart |
| kafka.consumerGroupId | string | `"control-plane"` | consumer group ID for kafka consumers running with this helm chart |
| kafka.topicBYOCClientUpdates | string | `"kl-byoc-client-updates"` | kafka topic for messages where target clusters sends updates for cluster BYOC resource |
| kafka.topicBilling | string | `"kl-billing"` | kafka topic for dispatching billing events |
| kafka.topicErrorOnApply | string | `"kl-error-on-apply"` | kafka topic for messages when agent encounters an error while applying k8s resources |
| kafka.topicEvents | string | `"kl-events"` | kafka topic for dispatching audit log events |
| kafka.topicGitWebhooks | string | `"kl-git-webhooks"` | kafka topic for dispatching git webhook messages |
| kafka.topicHarborWebhooks | string | `"kl-harbor-webhooks"` | kafka topic for dispatching harbor webhook messages |
| kafka.topicInfraStatusUpdates | string | `"kl-infra-updates"` | kafka topic for messages regarding infra resources on target clusters |
| kafka.topicStatusUpdates | string | `"kl-status-updates"` | kafka topic for messages regarding kloudlite resources on target clusters |
| nodeSelector | object | `{}` | node selectors to apply on all the pods belonging to this release |
| normalSvcAccount | string | `"kloudlite-svc-account"` | service account for non k8s operations, just for specifying image pull secrets |
| operators.accountOperator | object | `{"enabled":true,"image":"ghcr.io/kloudlite/plaform/account-operator:v1.0.5-nightly"}` | kloudlite account operator |
| operators.accountOperator.enabled | bool | `true` | whether to enable account operator |
| operators.accountOperator.image | string | `"ghcr.io/kloudlite/plaform/account-operator:v1.0.5-nightly"` | image (with tag) for account operator |
| operators.artifactsHarbor.configuration.harbor.adminPassword | string | `"<harbor-admin-password>"` | harbor api admin password |
| operators.artifactsHarbor.configuration.harbor.adminUsername | string | `"<harbor-admin-username>"` | harbor api admin username |
| operators.artifactsHarbor.configuration.harbor.apiVersion | string | `"v2.0"` | harbor api version |
| operators.artifactsHarbor.configuration.harbor.imageRegistryHost | string | `"<harbor-registry-host>"` | harbor image registry host |
| operators.artifactsHarbor.configuration.harbor.webhookAuthz | string | `"<harbor-webhook-authz>"` | harbor webhook authz secret |
| operators.artifactsHarbor.configuration.harbor.webhookEndpoint | string | `"https://webhooks.dev.kloudlite.io/harbor"` | harbor webhook endpoint, (for receiving webhooks for every images pushed) |
| operators.artifactsHarbor.configuration.harbor.webhookName | string | `"kloudlite-dev-webhook"` | harbor webhook name |
| operators.artifactsHarbor.enabled | bool | `true` | whether to enable artifacts harbor operator |
| operators.artifactsHarbor.image | string | `"ghcr.io/kloudlite/plaform/artifacts-harbor-operator:v1.0.5-nightly"` | image (with tag) for artifacts harbor operator |
| operators.byocOperator.enabled | bool | `true` | whether to enable byoc operator |
| operators.byocOperator.image | string | `"ghcr.io/kloudlite/platform/byoc-operator:v1.0.5-nightly"` | image (with tag) for byoc operator |
| operatorsNamespace | string | `"kl-init-operators"` | namespace where chart kloudlite-operators have been installed |
| persistence.XfsStorageClassName | string | `"<xfs-sc>"` | xfs storage class name |
| persistence.storageClassName | string | `"<storage-class-name>"` | ext4 storage class name |
| podLabels | object | `{}` | podlabels for pods belonging to this release |
| redpanda-operator | object | `{"fullnameOverride":"redpanda-operator","nameOverride":"redpanda-operator","resources":{"limits":{"cpu":"60m","memory":"60Mi"},"requests":{"cpu":"40m","memory":"40Mi"}},"webhook":{"enabled":false}}` | redpanda operator configuration, read more at https://vectorized.io/docs/quick-start-kubernetes |
| redpandaCluster | object | `{"create":true,"name":"redpanda","replicas":1,"resources":{"limits":{"cpu":"300m","memory":"400Mi"},"requests":{"cpu":"200m","memory":"200Mi"}},"storage":{"capacity":"2Gi"},"version":"v22.1.6"}` | redpanda cluster configuration, read more at https://vectorized.io/docs/quick-start-kubernetes |
| routers.accountsWeb.domain | string | `"accounts.dev.kloudlite.io"` | domain for accounts web router |
| routers.authWeb.domain | string | `"auth.dev.kloudlite.io"` | domain for auth web router |
| routers.consoleWeb.domain | string | `"console.dev.kloudlite.io"` | domain for console web router |
| routers.dnsApi.domain | string | `"dns-api.dev.kloudlite.io"` | domain for dns api router |
| routers.gatewayApi.domain | string | `"gateway.dev.kloudlite.io"` | domain for gateway api router |
| routers.messageOfficeApi.domain | string | `"message-office-api.dev.kloudlite.io"` | router domain for message office api |
| routers.socketWeb.domain | string | `"socket-web.dev.kloudlite.io"` | domain for socket web router |
| routers.webhooksApi.domain | string | `"webhooks.dev.kloudlite.io"` | domain for gateway api router |
| routers.webhooksApi.enabled | bool | `true` |  |
| tolerations | list | `[]` | tolerations for pods belonging to this release |
