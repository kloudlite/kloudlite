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
| BaseDomain | string | `"dev.kloudlite.io"` | base domain for all routers exposed through this cluster |
| accountCookieName | string | `"kloudlite-account"` | account cookie name, that console-api should expect, while any client communicates through it's graphql interface |
| accountName | string | `"kloudlite-dev"` | kloudlite account name for kloudlite resources, belonging to this chart |
| apps.accountsWeb.image | string | `"docker.io/kloudlite/platform/web/accounts-web:v1.0.5-nightly"` | image (with tag) for accounts web |
| apps.accountsWeb.name | string | `"accounts-web"` | workload name for accounts web |
| apps.auditLoggingWorker.image | string | `"docker.io/kloudlite/platform/apis/audit-logging-worker:v1.0.5-nightly"` | image (with tag) for audit logging worker |
| apps.auditLoggingWorker.name | string | `"audit-logging-worker"` | workload name for dudit logging worker |
| apps.authApi.image | string | `"docker.io/kloudlite/platform/apis/auth-api:v1.0.5-nightly"` | image (with tag) for auth api |
| apps.authApi.name | string | `"auth-api"` | workload name for auth api |
| apps.authWeb.image | string | `"docker.io/kloudlite/platform/web/accounts-web:v1.0.5-nightly"` | image (with tag) for auth web |
| apps.authWeb.name | string | `"auth-web"` | workload name for auth web |
| apps.commsApi.image | string | `"docker.io/kloudlite/platform/apis/comms-api:v1.0.5-nightly"` | image (with tag) for comms api |
| apps.commsApi.name | string | `"comms-api"` | workload name for comms api |
| apps.consoleApi.image | string | `"docker.io/kloudlite/platform/apis/console-api:v1.0.5-nightly"` | image (with tag) for console api |
| apps.consoleApi.name | string | `"console-api"` | workload name for console api |
| apps.consoleWeb.image | string | `"docker.io/kloudlite/platform/web/console-web:v1.0.5-nightly"` | image (with tag) for console web |
| apps.consoleWeb.name | string | `"console-web"` | workload name for console web |
| apps.containerRegistryApi.image | string | `"docker.io/kloudlite/platform/apis/container-registry-api:v1.0.5-nightly"` | image (with tag) for container registry api |
| apps.containerRegistryApi.name | string | `"container-registry-api"` | workload name for container registry api |
| apps.dnsApi.image | string | `"docker.io/kloudlite/platform/apis/dns-api:v1.0.5-nightly"` | image (with tag) for dns api |
| apps.dnsApi.name | string | `"dns-api"` | workload name for dns api |
| apps.financeApi.image | string | `"docker.io/kloudlite/platform/apis/finance-api:v1.0.5-nightly"` | image (with tag) for finance api |
| apps.financeApi.name | string | `"finance-api"` | workload name for finance api |
| apps.gatewayApi.image | string | `"docker.io/kloudlite/platform/apis/gateway-api:v1.0.5-nightly"` | image (with tag) for container registry api |
| apps.gatewayApi.name | string | `"gateway-api"` | workload name for gateway api |
| apps.iamApi.image | string | `"docker.io/kloudlite/platform/apis/iam-api:v1.0.5-nightly"` | image (with tag) for iam api |
| apps.iamApi.name | string | `"iam-api"` | workload name for iam api |
| apps.infraApi.image | string | `"docker.io/kloudlite/platform/apis/infra-api:v1.0.5-nightly"` | image (with tag) for infra api |
| apps.infraApi.name | string | `"infra-api"` | workload name for infra api |
| apps.jsEvalApi.image | string | `"docker.io/kloudlite/platform/apis/js-eval-api:v1.0.5-nightly"` | image (with tag) for js-eval-api |
| apps.jsEvalApi.name | string | `"js-eval-api"` | workload name for js-eval-api |
| apps.messageOfficeApi.image | string | `"docker.io/kloudlite/platform/apis/message-office-api:v1.0.5-nightly"` | image (with tag) for message office api |
| apps.messageOfficeApi.name | string | `"message-office-api"` | workload name for message office api |
| apps.socketWeb.image | string | `"docker.io/kloudlite/platform/web/socket-web:v1.0.5-nightly"` | image (with tag) for socket web |
| apps.socketWeb.name | string | `"socket-web"` | workload name for socket web |
| apps.webhooksApi.image | string | `"docker.io/kloudlite/platform/apis/webhooks-api:v1.0.5-nightly"` | image (with tag) for webhooks api |
| apps.webhooksApi.name | string | `"webhooks-api"` | workload name for webhooks api |
| cloudflareWildcardCert.cloudflareCreds | object | `{"email":"<cloudflare-email>","secretToken":"<cloudflare-secret-token>"}` | cloudflare authz credentials |
| cloudflareWildcardCert.cloudflareCreds.email | string | `"<cloudflare-email>"` | cloudflare authorized email |
| cloudflareWildcardCert.cloudflareCreds.secretToken | string | `"<cloudflare-secret-token>"` | cloudflare authorized secret token |
| cloudflareWildcardCert.domains | list | `["*.dev.kloudlite.io"]` | list of all SANs (Subject Alternative Names) for which wildcard cert should be created |
| cloudflareWildcardCert.enabled | bool | `true` | whether to create a wildcard cert for domains in this release |
| cloudflareWildcardCert.name | string | `"kl-cert-wildcard"` | name for wildcard cert |
| cloudflareWildcardCert.secretName | string | `"kl-cert-wildcard-tls"` | k8s secret where wildcard cert should be stored |
| clusterCookieName | string | `"kloudite-cluster"` | cluster cookie name, that console-api should expect, while any client communicates through it's graphql interface |
| clusterIssuer.acmeEmail | string | `"sample@example.com"` | email that should be used for communicating with letsencrypt services |
| clusterIssuer.name | string | `"cluster-issuer"` | name of cluster issuer, to be used for issuing wildcard cert |
| clusterSvcAccount | string | `"kloudlite-cluster-svc-account"` | service account for privileged k8s operations, like creating namespaces, apps, routers etc. |
| cookieDomain | string | `".kloudlite.io"` | cookie domain dictates at what domain, the cookies should be set for auth or other purposes |
| defaultProjectWorkspaceName | string | `"default"` | default project workspace name, that should be auto created, whenever you create a project |
| envName | string | `"development"` |  |
| githubAppName | string | `"kloudlite-dev"` | github app name, that we want to install on user's github account |
| harbor.adminPassword | string | `"<harbor-admin-password>"` | harbor api admin password |
| harbor.adminUsername | string | `"<harbor-admin-username>"` | harbor api admin username |
| harbor.apiVersion | string | `"v2.0"` | harbor api version |
| harbor.imageRegistryHost | string | `"<harbor-registry-host>"` | harbor image registry host |
| harbor.webhookAuthz | string | `"<harbor-webhook-authz>"` | harbor webhook authz secret |
| harbor.webhookEndpoint | string | `"https://webhooks.dev.kloudlite.io/harbor"` | harbor webhook endpoint, (for receiving webhooks for every images pushed) |
| harbor.webhookName | string | `"kloudlite-dev-webhook"` | harbor webhook name |
| imagePullPolicy | string | `"Always"` | image pull policies for kloudlite pods, belonging to this chart |
| ingress-nginx | object | `{"controller":{"admissionWebhooks":{"enabled":false,"failurePolicy":"Ignore"},"dnsPolicy":"ClusterFirstWithHostNet","electionID":"ingress-nginx","extraArgs":{"default-ssl-certificate":"kl-init-operators/kl-cert-wildcard-tls"},"hostNetwork":true,"hostPort":{"enabled":true,"ports":{"healthz":10254,"http":80,"https":443}},"ingressClass":"ingress-nginx","ingressClassByName":true,"ingressClassResource":{"controllerValue":"k8s.io/ingress-nginx","enabled":true,"name":"ingress-nginx"},"kind":"DaemonSet","podLabels":{},"resources":{"requests":{"cpu":"100m","memory":"200Mi"}},"service":{"type":"ClusterIP"}},"enabled":true,"nameOverride":"ingress-nginx","rbac":{"create":false},"serviceAccount":{"create":false,"name":"kloudlite-cluster-svc-account"}}` | ingress nginx configurations, read more at https://kubernetes.github.io/ingress-nginx/ |
| ingressClassName | string | `"ingress-nginx"` |  |
| kafka.consumerGroupId | string | `"control-plane"` | consumer group ID for kafka consumers running with this helm chart |
| kafka.topicBYOCClientUpdates | string | `"kl-byoc-client-updates"` | kafka topic for messages where target clusters sends updates for cluster BYOC resource |
| kafka.topicBilling | string | `"kl-billing"` | kafka topic for dispatching billing events |
| kafka.topicErrorOnApply | string | `"kl-error-on-apply"` | kafka topic for messages when agent encounters an error while applying k8s resources |
| kafka.topicEvents | string | `"kl-events"` | kafka topic for dispatching audit log events |
| kafka.topicGitWebhooks | string | `"kl-git-webhooks"` | kafka topic for dispatching git webhook messages |
| kafka.topicHarborWebhooks | string | `"kl-harbor-webhooks"` | kafka topic for dispatching harbor webhook messages |
| kafka.topicInfraStatusUpdates | string | `"kl-infra-updates"` | kafka topic for messages regarding infra resources on target clusters |
| kafka.topicPipelineRunUpdates | string | `"kl-pipeline-run-updates"` | kafka topic for tekton pipeline run events |
| kafka.topicStatusUpdates | string | `"kl-status-updates"` | kafka topic for messages regarding kloudlite resources on target clusters |
| managedResources.authDb | string | `"auth-db"` |  |
| managedResources.authRedis | string | `"auth-redis"` |  |
| managedResources.ciDb | string | `"ci-db"` |  |
| managedResources.ciRedis | string | `"ci-redis"` |  |
| managedResources.consoleDb | string | `"console-db"` |  |
| managedResources.consoleRedis | string | `"console-redis"` |  |
| managedResources.containerRegistryDb | string | `"container-registry-db"` |  |
| managedResources.dnsDb | string | `"dns-db"` |  |
| managedResources.dnsRedis | string | `"dns-redis"` |  |
| managedResources.eventsDb | string | `"events-db"` |  |
| managedResources.financeDb | string | `"finance-db"` |  |
| managedResources.financeRedis | string | `"finance-redis"` |  |
| managedResources.iamDb | string | `"iam-db"` |  |
| managedResources.iamRedis | string | `"iam-redis"` |  |
| managedResources.infraDb | string | `"infra-db"` |  |
| managedResources.messageOfficeDb | string | `"message-office-db"` |  |
| managedResources.socketWebRedis | string | `"socket-web-redis"` |  |
| managedServices.mongoSvc | string | `"mongo-svc"` |  |
| managedServices.redisSvc | string | `"redis-svc"` |  |
| networking.dnsNames | list | `["ns1.dev.kloudlite.io"]` | list of all dnsNames for which, you want wildcard certificate to be issued for |
| networking.edgeCNAME | string | `"edge.dev.kloudlite.io"` | basedomain for CNAME for all the edges managed (or, to be managed) by this cluster |
| nodeSelector | object | `{}` | node selectors to apply on all the pods belonging to this release |
| normalSvcAccount | string | `"kloudlite-svc-account"` | service account for non k8s operations, just for specifying image pull secrets |
| oAuth2.github.appId | string | `"<github-app-id>"` |  |
| oAuth2.github.appPrivateKey | string | `"<github-app-private-key>"` |  |
| oAuth2.github.callbackUrl | string | `"https://%s/oauth2/callback/github"` |  |
| oAuth2.github.clientId | string | `"<github-client-id>"` |  |
| oAuth2.github.clientSecret | string | `"<github-client-secret>"` |  |
| oAuth2.github.webhookUrl | string | `"https://%s/git/github"` |  |
| oAuth2.gitlab.callbackUrl | string | `"https://%s/oauth2/callback/gitlab"` |  |
| oAuth2.gitlab.clientId | string | `"<gitlab-client-id>"` |  |
| oAuth2.gitlab.clientSecret | string | `"<gitlab-client-secret>"` |  |
| oAuth2.gitlab.webhookUrl | string | `"https://%s/git/gitlab"` |  |
| oAuth2.google.callbackUrl | string | `"https://%s/oauth2/callback/gitlab"` |  |
| oAuth2.google.clientId | string | `"<google-client-id>"` |  |
| oAuth2.google.clientSecret | string | `"<google-client-secret>"` |  |
| oAuthSecretName | string | `"oauth-secrets"` | name of the secret that should contain all the oauth secrets |
| operators.accountOperator | object | `{"enabled":true,"image":"docker.io/kloudlite/plaform/account-operator:v1.0.5-nightly","name":"kl-accounts-operator"}` | kloudlite account operator |
| operators.accountOperator.enabled | bool | `true` | whether to enable account operator |
| operators.accountOperator.image | string | `"docker.io/kloudlite/plaform/account-operator:v1.0.5-nightly"` | image (with tag) for account operator |
| operators.accountOperator.name | string | `"kl-accounts-operator"` | workload name for account operator |
| operators.artifactsHarbor.enabled | bool | `true` | whether to enable artifacts harbor operator |
| operators.artifactsHarbor.image | string | `"docker.io/kloudlite/plaform/artifacts-harbor-operator:v1.0.5-nightly"` | image (with tag) for artifacts harbor operator |
| operators.artifactsHarbor.name | string | `"kl-artifacts-harbor"` | workload name for artifacts harbor operator |
| operators.byocOperator.enabled | bool | `true` | whether to enable byoc operator |
| operators.byocOperator.image | string | `"docker.io/kloudlite/platform/byoc-operator:v1.0.5-nightly"` | image (with tag) for byoc operator |
| operators.byocOperator.name | string | `"kl-byoc-operator"` | workload name for byoc operator |
| operatorsNamespace | string | `"kl-init-operators"` | namespace where operators have been installed |
| persistence.XfsStorageClassName | string | `"<xfs-sc>"` | xfs storage class name |
| persistence.storageClassName | string | `"<storage-class-name>"` | ext4 storage class name |
| podLabels | object | `{}` | podlabels for pods belonging to this release |
| rbac.pullSecret.name | string | `"kl-docker-creds"` |  |
| rbac.pullSecret.value | string | `"<image-pull-secret>"` |  |
| redpanda-operator | object | `{"fullnameOverride":"redpanda-operator","nameOverride":"redpanda-operator","resources":{"limits":{"cpu":"60m","memory":"60Mi"},"requests":{"cpu":"40m","memory":"40Mi"}},"webhook":{"enabled":false}}` | redpanda operator configuration, read more at https://vectorized.io/docs/quick-start-kubernetes |
| redpandaCluster | object | `{"name":"redpanda","replicas":1,"resources":{"limits":{"cpu":"300m","memory":"400Mi"},"requests":{"cpu":"200m","memory":"200Mi"}},"storage":{"capacity":"2Gi","storageClassName":"<xfs-sc>"},"version":"v22.1.6"}` | redpanda cluster configuration, read more at https://vectorized.io/docs/quick-start-kubernetes |
| routers.accountsWeb.domain | string | `"accounts.dev.kloudlite.io"` | domain for accounts web router |
| routers.accountsWeb.name | string | `"accounts-web"` | router name for accounts web router |
| routers.authWeb.domain | string | `"auth.dev.kloudlite.io"` | domain for auth web router |
| routers.authWeb.name | string | `"auth-web"` | router name for auth web router |
| routers.consoleWeb.domain | string | `"console.dev.kloudlite.io"` | domain for console web router |
| routers.consoleWeb.name | string | `"console-web"` | router name for console web router |
| routers.dnsApi.domain | string | `"dns-api.dev.kloudlite.io"` | domain for dns api router |
| routers.dnsApi.name | string | `"dns-api"` | router name for dns api router |
| routers.gatewayApi.domain | string | `"gateway.dev.kloudlite.io"` | domain for gateway api router |
| routers.gatewayApi.name | string | `"gateway-api"` | router name for gateway api router |
| routers.messageOfficeApi.domain | string | `"message-office-api.dev.kloudlite.io"` | router domain for message office api |
| routers.messageOfficeApi.name | string | `"message-office-api"` | router name for message office api router |
| routers.socketWeb.domain | string | `"socket-web.dev.kloudlite.io"` | domain for socket web router |
| routers.socketWeb.name | string | `"socket-web"` | router name for socket web router |
| routers.webhooksApi.domain | string | `"webhooks.dev.kloudlite.io"` | domain for gateway api router |
| routers.webhooksApi.name | string | `"webhooks-api"` | router name for gateway api router |
| secrets.names.harborAdminSecret | string | `"harbor-admin-creds"` | harbor admin secret name |
| secrets.names.oAuthSecret | string | `"oauth-secrets"` | secret where all oauth credentials should be |
| secrets.names.redpandaAdminAuthSecret | string | `"msvc-redpanda-admin-auth"` | secret where all the stripe related should be |
| secrets.names.stripeSecret | string | `"stripe-creds"` | secret, where all the stripe related credentials should be |
| secrets.names.webhookAuthzSecret | string | `"webhook-authz"` | secret where all the webhook related should be |
| sendgridApiKey | string | `"<sendgrid-api-key>"` | sendgrid api key for email communications |
| stripe | object | `{"publicKey":"<stripe-public-key>","secretKey":"<stripe-private-key>"}` | stripe credentials |
| stripe.publicKey | string | `"<stripe-public-key>"` | stripe public key |
| stripe.secretKey | string | `"<stripe-private-key>"` | stripe secret key |
| supportEmail | string | `"<support-email>"` | email through which we should be sending emails to target users |
| tolerations | list | `[]` | tolerations for pods belonging to this release |
| webhookAuthz.githubSecret | string | `"<webhook-authz-github-secret>"` | webhook authz secret for github webhooks |
| webhookAuthz.gitlabSecret | string | `"<webhook-authz-gitlab-secret>"` | we |
| webhookAuthz.harborSecret | string | `"<harbor-webhook-authz>"` | webhook authz secret for harbor webhooks |
| webhookAuthz.kloudliteSecret | string | `"<webhook-authz-kloudlite-secret>"` | webhook authz secret for kloudlite internal calls |
