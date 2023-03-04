apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.ciApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region}}
  serviceAccount: {{.Values.clusterSvcAccount}}
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.ciApi.name}}
      imagePullPolicy: {{.Values.apps.ciApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "30m"
        max: "50m"
      resourceMemory:
        min: "50Mi"
        max: "80Mi"
      env:
        - key: REDIS_HOSTS
          type: secret
          refName: mres-{{.Values.managedResources.ciRedis}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.Values.managedResources.ciRedis}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.Values.managedResources.ciRedis}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.Values.managedResources.ciRedis}}
          refKey: USERNAME

        - key: AUTH_REDIS_HOSTS
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: HOSTS

        - key: AUTH_REDIS_PASSWORD
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: PASSWORD

        - key: AUTH_REDIS_PREFIX
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: PREFIX

        - key: AUTH_REDIS_USERNAME
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: USERNAME

        - key: IAM_ADDR
          value: {{.Values.apps.iamApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: FINANCE_ADDR
          value: {{.Values.apps.financeApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: AUTH_ADDR
          value: {{.Values.apps.authApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: CONSOLE_ADDR
          value: {{.Values.apps.consoleApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: COMMS_ADDR
          value: {{.Values.apps.commsApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001

        - key: MONGO_URI
          type: secret
          refName: mres-{{.Values.managedResources.ciDb}}
          refKey: URI

        - key: MONGO_DB_NAME
          value: {{.Values.managedResources.ciDb}}

        - key: KAFKA_USERNAME
          type: secret
          refName: {{.Values.redpandaAdminSecretName}}
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: {{.Values.redpandaAdminSecretName}}
          refKey: PASSWORD

        - key: KAFKA_BROKERS
          type: secret
          refName: {{.Values.redpandaAdminSecretName}}
          refKey: KAFKA_BROKERS

        {{/* - key: HARBOR_HOST */}}
        {{/*   type: secret */}}
        {{/*   refName: {{.HarborAdminCredsSecretName}} */}}
        {{/*   refKey: HARBOR_IMAGE_REGISTRY_HOST */}}
        {{/**/}}
        {{/* - key: HARBOR_REGISTRY_HOST */}}
        {{/*   type: secret */}}
        {{/*   refName: {{.HarborAdminCredsSecretName}} */}}
        {{/*   refKey: HARBOR_IMAGE_REGISTRY_HOST */}}
        {{/**/}}
        {{/* - key: HARBOR_ADMIN_USERNAME */}}
        {{/*   type: secret */}}
        {{/*   refName: {{.HarborAdminCredsSecretName}} */}}
        {{/*   refKey: HARBOR_ADMIN_USERNAME */}}
        {{/**/}}
        {{/* - key: HARBOR_ADMIN_PASSWORD */}}
        {{/*   type: secret */}}
        {{/*   refName: {{.HarborAdminCredsSecretName}} */}}
        {{/*   refKey: HARBOR_ADMIN_PASSWORD */}}

        - key: GITHUB_WEBHOOK_AUTHZ_SECRET
          type: secret
          refName: {{.Values.oAuthSecretName}}
          refKey: GITHUB_WEBHOOK_AUTHZ_SECRET

        - key: GITLAB_WEBHOOK_AUTHZ_SECRET
          type: secret
          refName: {{.Values.oAuthSecretName}}
          refKey: GITLAB_WEBHOOK_AUTHZ_SECRET

        - key: GITHUB_APP_PK_FILE
          value: /hotspot/github-app-pk.pem

      envFrom:
        - type: secret
          refName: ci-env

        - type: secret
          refName: {{.Values.oAuthSecretName}}

      volumes:
        - mountPath: /hotspot
          type: secret
          refName: {{.Values.oAuthSecretName}}
          items:
            - key: github-app-pk.pem
              fileName: github-app-pk.pem
---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.apps.ciApi.name}}-env
  namespace: {{.Release.Namespace}}
stringData:
  GRPC_PORT: "3001"
  PORT: "3000"

  ORIGINS: https://studio.apollographql.com

  COOKIE_DOMAIN: "{{.Values.cookieDomain}}"

  {{/* KAFKA_TOPIC_GIT_WEBHOOKS: {{.KafkaTopicGitWebhooks}} */}}
  {{/* KAFKA_TOPIC_PIPELINE_RUN_UPDATES: {{.KafkaTopicPipelineRunUpdates}} */}}

  KAFKA_GIT_WEBHOOKS_CONSUMER_ID: "kloudlite/ci-api"
  KL_HOOK_TRIGGER_AUTHZ_SECRET: '***REMOVED***'
