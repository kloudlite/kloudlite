# -- image pull policies for kloudlite pods, belonging to this chart
imagePullPolicy: Always

nodeSelector: &nodeSelector { }

# -- tolerations for pods belonging to this release
tolerations: &tolerations [ ]

# -- podlabels for pods belonging to this release
podLabels: &podLabels { }

# -- cookie domain dictates at what domain, the cookies should be set for auth or other purposes
cookieDomain: {{.CookieDomain | squote}}

# -- base domain for all routers exposed through this cluster
baseDomain: {{.BaseDomain | squote }}

# -- cluster internal DNS name
clusterInternalDNS: "svc.cluster.local"

# @ignored
# -- account cookie name, that console-api should expect, while any client communicates through it's graphql interface
accountCookieName: "kloudlite-account"

# -- cluster cookie name, that console-api should expect, while any client communicates through it's graphql interface
# @ignored
clusterCookieName: "kloudlite-cluster"

# -- service account for privileged k8s operations, like creating namespaces, apps, routers etc.
clusterSvcAccount: {{.ClusterSvcAccount}}

# -- service account for non k8s operations, just for specifying image pull secrets
normalSvcAccount: {{.NormalSvcAccount}}

# -- default project workspace name, the one that should be auto created, whenever you create a project
defaultProjectWorkspaceName: "{{.DefaultProjectWorkspaceName}}"

helmCharts:
  cert-manager:
    enabled: true
    name: cert-manager

  ingress-nginx:
    enabled: true
    name: ingress-nginx

    configuration:
      # -- can be DaemonSet or Deployment
      controllerKind: "{{.IngressControllerKind}}"
      ingressClassName: "{{.IngressClassName}}"


  loki-stack:
    enabled: true
    name: loki-stack
    configuration:
      s3credentials:
        awsAccessKeyId: {{.LokiS3AwsAccessKeyId}}
        awsSecretAccessKey: {{.LokiS3AwsSecretAccessKey}}
        region: {{.LokiS3BucketRegion}}
        bucketName: {{.LokiS3BucketName}}

  redpanda-operator:
    enabled: true
    name: redpanda-operator

    configuration:
      # -- cpu, and memory resources for redpanda operator
      resources:
        limits:
          cpu: 60m
          memory: 60Mi
        requests:
          cpu: 40m
          memory: 40Mi

  vector:
    enabled: true
    name: vector

  grafana:
    enabled: true
    name: grafana

    configuration:
      volumeSize: 2Gi

  kube-prometheus:
    enabled: true
    name: prometheus

    configuration:
      prometheus:
        volumeSize: 2Gi
      alertmanager:
        volumeSize: 2Gi

persistence:
  storageClasses:
    ext4: {{.Ext4StorageClassName}}
    xfs: {{.XfsStorageClassName}}

  # -- ext4 storage class name
{{/*  ext4StorageClassName: {{.StorageClassName}}*/}}
  # -- xfs storage class name
{{/*  xfsStorageClassName: {{.XfsStorageClassName}}*/}}

# @ignored
secretNames:
  # -- secret where all oauth credentials should be
  oAuthSecret: oauth-secrets
  # -- secret where all the webhook related should be
  webhookAuthzSecret: webhook-authz
  # -- secret where all the redpanda admin related creds should be
  redpandaAdminAuthSecret: msvc-redpanda-admin-auth
  # -- harbor admin secret name
  harborAdminSecret: harbor-admin-creds

# -- redpanda cluster configuration, read more at https://vectorized.io/docs/quick-start-kubernetes
redpandaCluster:
  create: true
  name: "redpanda"
  version: v22.1.6
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

clusterIssuer:
  # -- whether to install cluster issuer
  create: true

  # -- name of cluster issuer, to be used for issuing wildcard cert
  name: "cluster-issuer"
  # -- email that should be used for communicating with lets-encrypt services
  acmeEmail: {{.AcmeEmail}}

cloudflareWildCardCert:
  create: {{.WildcardCertEnabled}}

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

  # -- list of all SANs (Subject Alternative Names) for which wildcard certs should be created
  domains:
    - '*.{{.BaseDomain}}'

# @ignored
kafka:
  # -- consumer group ID for kafka consumers running with this helm chart
  consumerGroupId: control-plane

  # -- kafka topic for dispatching audit log events
  topicAuditEvents: kl-events

  # -- kafka topic for dispatching harbor webhook messages
  topicHarborWebhooks: kl-harbor-webhooks

  # -- kafka topic for dispatching git webhook messages
  topicGitWebhooks: kl-git-webhooks

  # -- kafka topic for dispatching billing events
  topicBilling: kl-billing

  # -- kafka topic for messages regarding kloudlite resources on target clusters
  topicStatusUpdates: kl-status-updates

  # -- kafka topic for messages regarding infra resources on target clusters
  topicInfraStatusUpdates: kl-infra-updates

  # -- kafka topic for messages where target cluster sends updates for cluster BYOC resource
  topicBYOCClientUpdates: kl-byoc-client-updates

  # -- kafka topic for messages when an agent encounters an error while applying k8s resources
  topicErrorOnApply: kl-error-on-apply

# @ignored
managedServices:
  mongoSvc: mongo-svc
  redisSvc: redis-svc

# @ignored
managedResources:
  authDb: auth-db
  authRedis: auth-redis

  accountsDb: accounts-db

  dnsDb: dns-db
  dnsRedis: dns-redis

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

routers:
  authWeb:
    # @ignored
    # -- router name for auth web router
    name: auth

  accountsWeb:
    # @ignored
    # -- router name for accounts web router
    name: accounts

  consoleWeb:
    # @ignored
    # -- router name for console web router
    name: console

  socketWeb:
    # @ignored
    # -- router name for socket web router
    name: socket

  webhooksApi:
    enabled: true
    # @ignored
    # -- router name for gateway api router
    name: webhooks

  gatewayApi:
    # @ignored
    # -- router name for gateway api router
    name: gateway

  dnsApi:
    # @ignored
    # -- router name for dns api router
    name: dns-api

  messageOfficeApi:
    # @ignored
    # -- router name for message office api router
    name: message-office-api

  observabilityApi:
    # @ignored
    # -- router name for logs and metrics api
    name: observability

apps:
  authApi:
    # @ignored
    # -- workload name for auth api
    name: auth-api
    # -- image (with tag) for auth api
    image: {{.ImageAuthApi}}

    configuration:
      oAuth2:
        # -- whether to enable oAuth2
        enabled: {{.OAuth2Enabled}}
        github:
          # -- whether to enable GitHub oAuth2
          enabled: {{.OAuth2GithubEnabled}}
          # -- GitHub oAuth2 callback url
          callbackUrl: https://auth.{{.BaseDomain}}/oauth2/callback/github
          # -- GitHub oAuth2 Client ID
          clientId: {{.OAuthGithubClientId}}
          # -- GitHub oAuth2 Client Secret
          clientSecret: {{.OAuthGithubClientSecret}}
          {{/* webhookUrl: https://%s/git/github */}}
          # -- GitHub app id
          appId: {{.OAuthGithubAppId}}
          # -- GitHub app private key (base64 encoded)
          appPrivateKey: {{.OAuthGithubPrivateKey}}
          # -- GitHub app name, that we want to install on user's GitHub account
          githubAppName: kloudlite-dev

        gitlab:
          # -- whether to enable gitlab oAuth2
          enabled: {{.OAuth2GitlabEnabled}}
          # -- gitlab oAuth2 callback url
          callbackUrl: https://auth.{{.BaseDomain}}/oauth2/callback/gitlab
          # -- gitlab oAuth2 Client ID
          clientId: {{.OAuthGitlabClientId}}
          # -- gitlab oAuth2 Client Secret
          clientSecret: {{.OAuthGitlabClientSecret}}

          {{/* webhookUrl: https://%s/git/gitlab */}}

        google:
          # -- whether to enable google oAuth2
          enabled: {{.OAuth2GoogleEnabled}}
          # -- google oAuth2 callback url
          callbackUrl: https://auth.{{.BaseDomain}}/oauth2/callback/google
          # -- google oAuth2 Client ID
          clientId: {{.OAuthGoogleClientId}}
          # -- google oAuth2 Client Secret
          clientSecret: {{.OAuthGoogleClientSecret}}

  dnsApi:
    enabled: false
    # @ignored
    # -- workload name for dns api
    name: dns-api
    # -- image (with tag) for dns api
    image: {{.ImageDnsApi}}

    # -- configurations for dns api
    configuration:
      # -- list of all dnsNames for which, you want wildcard certificate to be issued for
      dnsNames:
        - "{{.DnsName}}"
      # -- base domain for CNAME for all the edges managed (or, to be managed) by this cluster
      edgeCNAME: "{{.EdgeCNAME}}"

  commsApi:
    # -- whether to enable communications api
    enabled: true

    # @ignored
    # -- workload name for comms api
    name: comms-api

    # -- image (with tag) for comms api
    image: {{.ImageCommsApi}}

    # -- configurations for comms api
    configuration:
      # -- sendgrid api key for email communications, if (sendgrid.enabled)
      sendgridApiKey: {{.SendgridAPIKey}}

      # -- email through which we should be sending emails to target users, if (sendgrid.enabled)
      supportEmail: {{.SupportEmail}}

      # @ignored
      grpcPort: 3001

      # -- account web invite url
      accountsWebInviteUrl: https://accounts.{{.BaseDomain}}/invite

      # -- project web invite url
      projectsWebInviteUrl: https://projects.{{.BaseDomain}}/invite

      # -- console web invite url
      kloudliteConsoleWebUrl: https://console.{{.BaseDomain}}

      # -- reset password web url
      resetPasswordWebUrl: https://auth.{{.BaseDomain}}/reset-password

      # -- verify email web url
      verifyEmailWebUrl: https://auth.{{.BaseDomain}}/verify-email

  consoleApi:
    # @ignored
    # -- workload name for console api
    name: console-api

    # -- image (with tag) for console api
    image: {{.ImageConsoleApi}}

    configuration:
      # @ignored
      httpPort: 3000
      # @ignored
      grpcPort: 3001
      # @ignored
      logsAndMetricsHttpPort: 9100

  accountsApi:
    # @ignored
    # -- workload name for accounts api
    name: accounts-api

    # -- image (with tag) for accounts api
    image: {{.ImageAccountsApi}}

    configuration:
      # @ignored
      httpPort: 3000

      # @ignored
      grpcPort: 3001

  iamApi:
    # @ignored
    # -- workload name for iam api
    name: iam-api

    # -- image (with tag) for iam api
    image: {{.ImageIAMApi}}

    configuration:
      # @ignored
      grpcPort: 3001

  infraApi:
    # @ignored
    # -- workload name for infra api
    name: infra-api

    # -- image (with tag) for infra api
    image: {{.ImageInfraApi}}

  gatewayApi:
    # @ignored
    # -- workload name for gateway api
    name: gateway-api
    # -- image (with tag) for container registry api
    image: {{.ImageGatewayApi}}

  containerRegistryApi:
    enabled: false
    {{- /* enabled: &containerRegistryEnabled {{.ContainerRegistryApiEnabled}} */}}

    # @ignored
    # -- workload name for container registry api
    name: container-registry-api

    # -- image (with tag) for container registry api
    image: {{.ImageContainerRegistryApi}}

    configuration:
      # @ignored
      # -- (number) port on which container registry api should listen
      httpPort: 3000
      # -- (number) port on which container registry grpc api should listen
      # @ignored
      grpcPort: 3001

      # -- harbor configuration, required only if .apps.containerRegistryApi.enabled
      harbor: &harborConfiguration
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

  consoleWeb:
    # @ignored
    # -- workload name for console web
    name: console-web
    # -- image (with tag) for console web
    image: {{.ImageConsoleWeb}}

  authWeb:
    # @ignored
    # -- workload name for auth web
    name: auth-web
    # -- image (with tag) for auth web
    image: {{.ImageAuthWeb}}

  accountsWeb:
    # @ignored
    # -- workload name for accounts web
    name: accounts-web
    # -- image (with tag) for accounts web
    image: {{.ImageAccountsWeb}}

  auditLoggingWorker:
    # @ignored
    # -- workload name for audit logging worker
    name: audit-logging-worker
    # -- image (with tag) for audit logging worker
    image: {{.ImageAuditLoggingWorker}}

  webhooksApi:
    enabled: true
    # @ignored
    # -- workload name for webhooks api
    name: webhooks-api
    # -- image (with tag) for webhooks api
    image: {{.ImageWebhooksApi}}

    configuration:
      webhookAuthz:
        # -- webhook authz secret for gitlab webhooks
        gitlabSecret: {{.WebhookAuthzGitlabSecret}}
        # -- webhook authz secret for GitHub webhooks
        githubSecret: {{.WebhookAuthzGithubSecret}}
        # -- webhook authz secret for harbor webhooks
        harborSecret: {{.HarborWebhookAuthz}}
        # -- webhook authz secret for kloudlite internal calls
        kloudliteSecret: {{.WebhookAuthzKloudliteSecret}}

  messageOfficeApi:
    # @ignored
    # -- workload name for message office api
    name: message-office-api

    # -- image (with tag) for message office api
    image: {{.ImageMessageOfficeApi}}

    configuration:
      # @ignored
      grpcPort: 3001

      # @ignored
      httpPort: 3000

      # -- token hashing secret, that is used to hash access tokens for kloudlite agents
      # -- consider using 128 characters random string, you can use `python -c "import secrets; print(secrets.token_urlsafe(128))"`
      tokenHashingSecret: {{.TokenHashingSecret}}

operators:
  # -- kloudlite account operator
  accountOperator:
    # -- whether to enable account operator
    enabled: true
    # @ignored
    # -- workload name for account operator
    name: kl-accounts-operator
    # -- image (with tag) for account operator
    image: {{.ImageAccountOperator}}

  byocOperator:
    # -- whether to enable byoc operator
    enabled: true

    # @ignored
    # -- workload name for byoc operator
    name: kl-byoc-operator

    # -- image (with tag) for byoc operator
    image: {{.ImageBYOCOperator}}

  wgOperator:
    # -- whether to enable wg operator
    enabled: true
    # -- wg operator workload name
    # @ignored
    name: kl-wg-operator
    # -- wg operator image and tag
    image: {{.ImageWgOperator}}

    # -- wireguard configuration options
    configuration:
      # -- cluster pods CIDR range
      podCIDR: {{.WgPodCIDR}}
      # -- cluster services CIDR range
      svcCIDR: {{.WgSvcCIDR}}
      # -- dns hosted zone, i.e., dns pointing to this cluster
      dnsHostedZone: {{.WgDnsHostedZone}}

      enableExamples: {{.EnableWgExamples}}

