{{/* defaultRegion: master */}}

# -- image pull policies for kloudlite pods, belonging to this chart
imagePullPolicy: Always

# -- node selectors to apply on all the pods belonging to this release
nodeSelector: &nodeSelector {}

# -- tolerations for pods belonging to this release
tolerations: &tolerations []

# -- podlabels for pods belonging to this release
podLabels: &podLabels {}


# -- kloudlite account name for kloudlite resources, belonging to this chart
accountName: kloudlite-dev

envName: {{.EnvName}}

# -- cookie domain dictates at what domain, the cookies should be set for auth or other purposes
cookieDomain: "{{.CookieDomain}}"

# -- account cookie name, that console-api should expect, while any client communicates through it's graphql interface
accountCookieName: "kloudlite-account"
# -- cluster cookie name, that console-api should expect, while any client communicates through it's graphql interface
clusterCookieName: "kloudite-cluster"

# -- default project workspace name, that should be auto created, whenever you create a project
defaultProjectWorkspaceName: "{{.DefaultProjectWorkspaceName}}"

# -- sendgrid api key for email communications
sendgridApiKey: {{.SendgridAPIKey}}
# -- email through which we should be sending emails to target users
supportEmail: {{.SupportEmail}}

rbac:
  pullSecret:
    name: kl-docker-creds
    value: {{.ImagePullSecret}}

persistence:
  # -- ext4 storage class name
  storageClassName: {{.StorageClassName}}
  # -- xfs storage class name
  XfsStorageClassName: {{.XfsStorageClassName}}

secrets:
  names:
    # -- secret where all oauth credentials should be
    oAuthSecret: oauth-secrets
    # -- secret where all the webhook related should be
    webhookAuthzSecret: webhook-authz
    # -- secret where all the stripe related should be
    redpandaAdminAuthSecret: msvc-redpanda-admin-auth
    # -- harbor admin secret name
    harborAdminSecret: harbor-admin-creds
    # -- secret, where all the stripe related credentials should be
    stripeSecret: stripe-creds 

# -- redpanda operator configuration, read more at https://vectorized.io/docs/quick-start-kubernetes
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

# -- redpanda cluster configuration, read more at https://vectorized.io/docs/quick-start-kubernetes
redpandaCluster:
  name: "redpanda"
  {{/* version: v23.1.7 */}}
  version: v22.1.6
  replicas: 1
  storage:
    capacity: 2Gi
    storageClassName: {{.XfsStorageClassName}}
  resources:
    requests:
      cpu: 200m
      memory: 200Mi
    limits:
      cpu: 300m
      memory: 400Mi

ingressClassName: {{.IngressClassName}}

# -- ingress nginx configurations, read more at https://kubernetes.github.io/ingress-nginx/
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
    podLabels: *podLabels

    resources:
      requests:
        cpu: 100m
        memory: 200Mi

    admissionWebhooks:
      enabled: false
      failurePolicy: Ignore

# cloudflare configurations
cloudflareWildcardCert:
  # -- whether to create a wildcard cert for domains in this release
  enabled: true

  # -- name for wildcard cert
  name: {{.WildcardCertName}}
  # -- k8s secret where wildcard cert should be stored
  secretName: {{.WildcardCertName}}-tls

  # -- cloudflare authz credentials
  cloudflareCreds:
    # -- cloudflare authorized email
    email: {{.CloudflareEmail}}
    # -- cloudflare authorized secret token
    secretToken: {{.CloudflareSecretToken}}
  # -- list of all SANs (Subject Alternative Names) for which wildcard cert should be created
  domains: 
    - "*.{{.BaseDomain}}"

# -- namespace where operators have been installed
operatorsNamespace: {{.OperatorsNamespace}}

clusterIssuer:
  # -- name of cluster issuer, to be used for issuing wildcard cert
  name: "cluster-issuer"
  # -- email that should be used for communicating with letsencrypt services
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

# -- name of the secret that should contain all the oauth secrets
oAuthSecretName: oauth-secrets
# -- github app name, that we want to install on user's github account
githubAppName: kloudlite-dev

# -- stripe credentials
stripe:
  # -- stripe public key
  publicKey: {{.StripePublicKey}}
  # -- stripe secret key
  secretKey: {{.StripeSecretKey}}

# -- service account for privileged k8s operations, like creating namespaces, apps, routers etc.
clusterSvcAccount: {{.ClusterSvcAccount}}
# -- service account for non k8s operations, just for specifying image pull secrets
normalSvcAccount: {{.NormalSvcAccount}}


networking:
  # -- list of all dnsNames for which, you want wildcard certificate to be issued for
  dnsNames: 
    - "{{.DnsName}}"
  # -- basedomain for CNAME for all the edges managed (or, to be managed) by this cluster
  edgeCNAME: "{{.EdgeCNAME}}"

# -- base domain for all routers exposed through this cluster
BaseDomain: {{.BaseDomain}}

kafka:
  # -- consumer group ID for kafka consumers running with this helm chart
  consumerGroupId: control-plane

  # -- kafka topic for dispatching audit log events
  topicEvents: kl-events
  # -- kafka topic for dispatching harbor webhook messages
  topicHarborWebhooks: kl-harbor-webhooks
  # -- kafka topic for dispatching git webhook messages
  topicGitWebhooks: kl-git-webhooks

  # -- kafka topic for tekton pipeline run events
  topicPipelineRunUpdates: kl-pipeline-run-updates
  # -- kafka topic for dispatching billing events
  topicBilling: kl-billing
  # -- kafka topic for messages regarding kloudlite resources on target clusters
  topicStatusUpdates: kl-status-updates
  # -- kafka topic for messages regarding infra resources on target clusters
  topicInfraStatusUpdates: kl-infra-updates
  # -- kafka topic for messages where target clusters sends updates for cluster BYOC resource
  topicBYOCClientUpdates: kl-byoc-client-updates
  # -- kafka topic for messages when agent encounters an error while applying k8s resources
  topicErrorOnApply: kl-error-on-apply

harbor:
  # -- harbor api version
  apiVersion: v2.0
  # -- harbor api admin username
  adminUsername: {{.HarborAdminUsername}}
  # -- harbor api admin password
  adminPassword: {{.HarborAdminPassword}}

  # -- harbor image registry host
  imageRegistryHost: {{.HarborRegistryHost}}

  # -- harbor webhook endpoint, (for receiving webhooks for every images pushed)
  webhookEndpoint: https://webhooks.{{.BaseDomain}}/harbor 
  # -- harbor webhook name
  webhookName: {{.HarborWebhookName}}
  # -- harbor webhook authz secret
  webhookAuthz: {{.HarborWebhookAuthz}}

webhookAuthz:
  # -- we
  gitlabSecret: {{.WebhookAuthzGitlabSecret}}
  # -- webhook authz secret for github webhooks
  githubSecret: {{.WebhookAuthzGithubSecret}}
  # -- webhook authz secret for harbor webhooks
  harborSecret: {{.HarborWebhookAuthz}}
  # -- webhook authz secret for kloudlite internal calls
  kloudliteSecret: {{.WebhookAuthzKloudliteSecret}}

routers:
  authWeb: 
    # -- router name for auth web router
    name: auth-web
    # -- domain for auth web router
    domain: auth.{{.BaseDomain}}

  accountsWeb: 
    # -- router name for accounts web router
    name: accounts-web
    # -- domain for accounts web router
    domain: accounts.{{.BaseDomain}}

  consoleWeb:
    # -- router name for console web router
    name: console-web
    # -- domain for console web router
    domain: console.{{.BaseDomain}}

  socketWeb:
    # -- router name for socket web router
    name: socket-web
    # -- domain for socket web router
    domain: socket-web.{{.BaseDomain}}

  webhooksApi:
    # -- router name for gateway api router
    name: webhooks-api
    # -- domain for gateway api router
    domain: webhooks.{{.BaseDomain}}

  gatewayApi:
    # -- router name for gateway api router
    name: gateway-api
    # -- domain for gateway api router
    domain: gateway.{{.BaseDomain}}

  dnsApi:
    # -- router name for dns api router
    name: dns-api
    # -- domain for dns api router
    domain: dns-api.{{.BaseDomain}}

  messageOfficeApi:
    # -- router name for message office api router
    name: message-office-api
    # -- router domain for message office api
    domain: message-office-api.{{.BaseDomain}}

apps:
  authApi:
    # -- workload name for auth api
    name: auth-api
    # -- image (with tag) for auth api
    image: {{.ImageAuthApi}}
  
  dnsApi:
    # -- workload name for dns api
    name: dns-api 
    # -- image (with tag) for dns api
    image: {{.ImageDnsApi}}

  commsApi:
    # -- workload name for comms api
    name: comms-api
    # -- image (with tag) for comms api
    image: {{.ImageCommsApi}}

  consoleApi:
    # -- workload name for console api
    name: console-api
    # -- image (with tag) for console api
    image: {{.ImageConsoleApi}}

  financeApi:
    # -- workload name for finance api
    name: finance-api
    # -- image (with tag) for finance api
    image: {{.ImageFinanceApi}}

  iamApi:
    # -- workload name for iam api
    name: iam-api
    # -- image (with tag) for iam api
    image: {{.ImageIAMApi}}
  
  infraApi:
    # -- workload name for infra api
    name: infra-api
    # -- image (with tag) for infra api
    image: {{.ImageInfraApi}}

  jsEvalApi:
    # -- workload name for js-eval-api
    name: js-eval-api
    # -- image (with tag) for js-eval-api
    image: {{.ImageJsEvalApi}}

  gatewayApi:
    # -- workload name for gateway api
    name: gateway-api
    # -- image (with tag) for container registry api
    image: {{.ImageGatewayApi}}

  containerRegistryApi:
    # -- workload name for container registry api
    name: container-registry-api
    # -- image (with tag) for container registry api
    image: {{.ImageContainerRegistryApi}}

  {{/* # web */}}
  socketWeb:
    # -- workload name for socket web
    name: socket-web
    # -- image (with tag) for socket web
    image: {{.ImageSocketWeb}}

  consoleWeb:
    # -- workload name for console web
    name: console-web
    # -- image (with tag) for console web
    image: {{.ImageConsoleWeb}}

  authWeb:
    # -- workload name for auth web
    name: auth-web
    # -- image (with tag) for auth web
    image: {{.ImageAuthWeb}}

  accountsWeb:
    # -- workload name for accounts web
    name: accounts-web
    # -- image (with tag) for accounts web
    image: {{.ImageAccountsWeb}}

  auditLoggingWorker:
    # -- workload name for dudit logging worker
    name: audit-logging-worker
    # -- image (with tag) for audit logging worker
    image: {{.ImageAuditLoggingWorker}}

  webhooksApi:
    # -- workload name for webhooks api
    name: webhooks-api
    # -- image (with tag) for webhooks api
    image: {{.ImageWebhooksApi}}

  messageOfficeApi:
    # -- workload name for message office api
    name: message-office-api
    # -- image (with tag) for message office api
    image: {{.ImageMessageOfficeApi}}

operators:
  # -- kloudlite account operator
  accountOperator:
    # -- whether to enable account operator
    enabled: true
    # -- workload name for account operator
    name: kl-accounts-operator
    # -- image (with tag) for account operator
    image: {{.ImageAccountOperator}}

  artifactsHarbor:
    # -- whether to enable artifacts harbor operator
    enabled: true
    # -- workload name for artifacts harbor operator
    name: kl-artifacts-harbor
    # -- image (with tag) for artifacts harbor operator
    image: {{.ImageArtifactsHarborOperator }}

  byocOperator:
    # -- whether to enable byoc operator
    enabled: true
    # -- workload name for byoc operator
    name: kl-byoc-operator
    # -- image (with tag) for byoc operator
    image: {{.ImageBYOCOperator}}
