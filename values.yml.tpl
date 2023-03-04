region: master

accountName: kl-core
imagePullPolicy: Always

envName: {{.EnvName}}
imageTag: {{.ImageTag}}

cookieDomain: ".kloudlite.io"

sendgridApiKey: '***REMOVED***'
supportEmail: support@kloudlite.io

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
  financeRedis: finannce-redis

  iamDb: iam-db
  iamRedis: iam-redis

  infraDb: infra-db
  
  consoleDb: console-db
  consoleRedis: console-redis

  socketWebRedis: socket-web-redis

  eventsDb: events-db

oAuthSecretName: oauth-secrets
githubAppName: kloudlite-dev
redpandaAdminSecretName: msvc-redpanda-admin-creds

secrets:
  stripePublicKey: ***REMOVED***
  stripeSecretKey: ***REMOVED***

clusterSvcAccount: kloudlite-cluster-svc-account
normalSvcAccount: kloudlite-svc-account

nodeSelector: 
  k1: v1

tolerations: []

networking:
  dnsNames: 
    - ns1.dev.kloudlite.io
    - ns2.dev.kloudlite.io
  edgeCNAME: dev.edge.kloudlite.io

baseDomain: {{.baseDomain}}

kafka:
  consumerGroupId: control-plane
  topicEvents: kl-events
  topicHarborWebhooks: kl-harbor-webhooks

webhookAuthz:
  gitlabSecret: '***REMOVED***'
  githubSecret: '***REMOVED***'
  harborSecret: '***REMOVED***'
  kloudliteSecret: '***REMOVED***'

routers:
  authWeb: 
    name: auth-web
    domain: auth.{{.baseDomain}}
  accountsWeb: 
    name: accounts-web
    domain: accounts.{{.baseDomain}}
  consoleWeb:
    name: console-web
    domain: console.{{.baseDomain}}
  socketWeb:
    name: socket-web
    domain: console.{{.baseDomain}}
  webhooksApi:
    name: webhooks-api
    domain: webhooks.{{.baseDomain}}

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
    image: registry.kloudlite.io/kloudlite/{{.EnvName}}/console-api:{{.ImageTag}}
  
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

  # web
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
