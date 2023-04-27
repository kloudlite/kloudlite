defaultRegion: master

accountName: kloudlite-dev
imagePullPolicy: Always

envName: {{.EnvName}}
imageTag: {{.ImageTag}}

cookieDomain: "{{.CookieDomain}}"

accountCookieName: "kloudlite-account"
clusterCookieName: "kloudite-cluster"

sendgridApiKey: {{.SendgridAPIKey}}
supportEmail: {{.SupportEmail}}

namespaces:
  klCore: kl-core

rbac:
  pullSecret:
    name: kl-docker-creds
    value: {{.ImagePullSecret}}

persistence:
  storageClassName: {{.StorageClassName}}

secrets:
  names:
    oAuthSecret: oauth-secrets
    webhookAuthzSecret: webhook-authz
    redpandaAdminAuthSecret: msvc-redpanda-admin-auth
    harborAdminSecret: harbor-admin-creds
    stripeSecret: stripe-creds 

# uses redpanda-operator chart
redpanda-operator:
  nameOverride: redpanda-operator
  fullnameOverride: redpanda-operator

  resources:
    limits:
      cpu: 60m
      memory: 60Mi
    requests:
      cpu: 40m
      memory: 40Mi

  webhook:
    enabled: false

redpandaCluster:
  name: "redpanda"
  version: v23.1.7
  replicas: 1
  storage:
    capacity: 2Gi
  resources:
    requests:
      cpu: 200m
      memory: 200Mi
    limits:
      cpu: 300m
      memory: 400Mi

ingressClassName: {{.IngressClassName}}
ingress-nginx:
  enabled: true
  nameOverride: "ingress-nginx"

  rbac:
    create: false

  serviceAccount:
    create: false
    name: {{.ClusterSvcAccount}}

  controller:
    kind: DaemonSet
    hostNetwork: true
    hostPort:
      enabled: true
      ports:
        http: 80
        https: 443
        healthz: 10254

    dnsPolicy: ClusterFirstWithHostNet

    {{/* watchIngressWithoutClass: false */}}
    ingressClassByName: true
    ingressClass: {{.IngressClassName}}
    electionID: {{.IngressClassName}}
    ingressClassResource:
      enabled: true
      name: {{.IngressClassName}}
      controllerValue: "k8s.io/{{.IngressClassName}}"
      {{/* controllerValue: "k8s.io/%s" {{.Values.ingressClassName}} */}}

    service:
      type: "ClusterIP"

    extraArgs:
      default-ssl-certificate: "{{.OperatorsNamespace}}/{{.WildcardCertName}}-tls"
    {{/* podLabels: {{$labels  | toYAML | nindent 6}} */}}

    resources:
      requests:
        cpu: 100m
        memory: 200Mi

    admissionWebhooks:
      enabled: false
      failurePolicy: Ignore

cloudflareWildcardCert:
  enabled: true

  name: {{.WildcardCertName}}
  secretName: {{.WildcardCertName}}-tls

  cloudflareCreds:
    email: {{.CloudflareEmail}}
    secretToken: {{.CloudflareSecretToken}}
  domains: 
    - "*.{{.BaseDomain}}"

operatorsNamespace: {{.OperatorsNamespace}}

clusterIssuer:
  name: "cluster-issuer"
  acmeEmail: {{.AcmeEmail}}

managedServices:
  mongoSvc: mongo-svc
  redisSvc: redis-svc

managedResources:
  authDb: auth-db
  authRedis: auth-redis

  dnsDb: dns-db
  dnsRedis: dns-redis

  ciDb: ci-db
  ciRedis: ci-redis

  financeDb: finance-db
  financeRedis: finance-redis

  iamDb: iam-db
  iamRedis: iam-redis

  infraDb: infra-db
  
  consoleDb: console-db
  consoleRedis: console-redis

  messageOfficeDb: message-office-db

  socketWebRedis: socket-web-redis
  eventsDb: events-db
  containerRegistryDb: container-registry-db

oAuth2:
  github:
    callbackUrl: https://%s/oauth2/callback/github
    clientId: {{.OAuthGithubClientId}}
    clientSecret: {{.OAuthGithubClientSecret}}
    webhookUrl: https://%s/git/github
    appId: {{.OAuthGithubAppId}}
    appPrivateKey: {{.OAuthGithubPrivateKey}}

  gitlab: 
    callbackUrl: https://%s/oauth2/callback/gitlab
    clientId: {{.OAuthGitlabClientId}}
    clientSecret: {{.OAuthGitlabClientSecret}}
    webhookUrl: https://%s/git/gitlab

  google:
    callbackUrl: https://%s/oauth2/callback/gitlab
    clientId: {{.OAuthGoogleClientId}}
    clientSecret: {{.OAuthGoogleClientSecret}}

oAuthSecretName: oauth-secrets
githubAppName: kloudlite-dev

stripe:
  publicKey: {{.StripePublicKey}}
  secretKey: {{.StripeSecretKey}}

clusterSvcAccount: {{.ClusterSvcAccount}}
normalSvcAccount: {{.NormalSvcAccount}}

nodeSelector: {}
tolerations: []

networking:
  dnsNames: 
    - "{{.DnsName}}"
  edgeCNAME: "{{.EdgeCNAME}}"

BaseDomain: {{.BaseDomain}}

kafka:
  consumerGroupId: control-plane

  topicEvents: kl-events
  topicHarborWebhooks: kl-harbor-webhooks
  topicGitWebhooks: kl-git-webhooks
  topicPipelineRunUpdates: kl-pipeline-run-updates
  topicBilling: kl-billing
  topicStatusUpdates: kl-status-updates
  topicInfraStatusUpdates: kl-infra-updates
  topicBYOCClientUpdates: kl-byoc-client-updates
  topicErrorOnApply: kl-error-on-apply

harbor:
  apiVersion: v2.0
  adminUsername: {{.HarborAdminUsername}}
  adminPassword: {{.HarborAdminPassword}}

  imageRegistryHost: {{.HarborRegistryHost}}

  webhookEndpoint: https://webhooks.{{.BaseDomain}}/harbor 
  webhookName: {{.HarborWebhookName}}
  webhookAuthz: {{.HarborWebhookAuthz}}

webhookAuthz:
  gitlabSecret: {{.WebhookAuthzGitlabSecret}}
  githubSecret: {{.WebhookAuthzGithubSecret}}
  harborSecret: {{.HarborWebhookAuthz}}
  kloudliteSecret: {{.WebhookAuthzKloudliteSecret}}

routers:
  authWeb: 
    name: auth-web
    domain: auth.{{.BaseDomain}}

  accountsWeb: 
    name: accounts-web
    domain: accounts.{{.BaseDomain}}

  consoleWeb:
    name: console-web
    domain: console.{{.BaseDomain}}

  socketWeb:
    name: socket-web
    domain: socket-web.{{.BaseDomain}}

  webhooksApi:
    name: webhooks-api
    domain: webhooks.{{.BaseDomain}}

  gatewayApi:
    name: gateway-api
    domain: gateway.{{.BaseDomain}}

  dnsApi:
    name: dns-api
    domain: dns-api.{{.BaseDomain}}

  messageOfficeApi:
    name: message-office-api
    domain: message-office-api.{{.BaseDomain}}

apps:
  authApi:
    name: auth-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/auth-api:{{.ImageTag}}
  
  dnsApi:
    name: dns-api 
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/dns-api:{{.ImageTag}}

  commsApi:
    name: comms-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/comms-api:{{.ImageTag}}

  consoleApi:
    name: console-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/console-api:{{.ImageTag}}

  ciApi:
    name: ci-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/ci-api:{{.ImageTag}}
  
  financeApi:
    name: finance-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/finance-api:{{.ImageTag}}

  iamApi:
    name: iam-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/iam-api:{{.ImageTag}}
  
  infraApi:
    name: infra-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/infra-api:{{.ImageTag}}

  jsEvalApi:
    name: js-eval-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/js-eval-api:{{.ImageTag}}

  gatewayApi:
    name: gateway-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/gateway-api:{{.ImageTag}}

  containerRegistryApi:
    name: container-registry-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/container-registry-api:{{.ImageTag}}

  {{/* # web */}}
  socketWeb:
    name: socket-web
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/socket-web:{{.ImageTag}}

  consoleWeb:
    name: console-web
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/console-web:{{.ImageTag}}

  authWeb:
    name: auth-web
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/auth-web:{{.ImageTag}}

  accountsWeb:
    name: accounts-web
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/accounts-web:{{.ImageTag}}

  auditLoggingWorker:
    name: audit-logging-worker
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/audit-logging-worker:{{.ImageTag}}

  webhooksApi:
    name: webhooks-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/webhooks-api:{{.ImageTag}}

  messageOfficeApi:
    name: message-office-api
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/message-office-api:{{.ImageTag}}

operators:
  accountOperator:
    enabled: true
    name: kl-accounts-operator
    image: registry.kloudlite.io/kloudlite/operators/{{.EnvName}}/account:{{.ImageTag}}

  artifactsHarbor:
    enabled: true
    name: kl-artifacts-harbor
    image: registry.kloudlite.io/kloudlite/operators/{{.EnvName}}/artifacts-harbor:{{.ImageTag}}

