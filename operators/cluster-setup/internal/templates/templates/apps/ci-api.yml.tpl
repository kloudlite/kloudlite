{{- $namespace := get . "namespace" -}}
{{- $svcAccount := get . "svc-account" -}}
{{- $cookieDomain := get . "cookie-domain" -}}
{{- $sharedConstants := get . "shared-constants" -}}
{{- $ownerRefs := get . "owner-refs" | default list -}}

{{- $accountRef := get . "account-ref" | default "kl-core" -}}
{{- $region := get . "region" | default "master" -}}
{{- $imagePullPolicy := get . "image-pull-policy" | default "Always" -}}

{{ with $sharedConstants}}
{{/*gotype: github.com/kloudlite/operator/apis/cluster-setup/v1.SharedConstants*/}}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.AppCiApi}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
spec:
  region: {{$region}}
  serviceAccount: {{$svcAccount}}
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
      image: {{.ImageCiApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"
      env:
        - key: REDIS_HOSTS
          type: secret
          refName: mres-{{.CiRedisName}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.CiRedisName}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.CiRedisName}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.CiRedisName}}
          refKey: USERNAME

        - key: AUTH_REDIS_HOSTS
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: HOSTS

        - key: AUTH_REDIS_PASSWORD
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: PASSWORD

        - key: AUTH_REDIS_PREFIX
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: PREFIX

        - key: AUTH_REDIS_USERNAME
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: USERNAME

        - key: IAM_ADDR
          value: {{.AppIAMApi}}.{{$namespace}}.svc.cluster.local:3001

        - key: FINANCE_ADDR
          value: {{.AppFinanceApi}}.{{$namespace}}.svc.cluster.local:3001

        - key: AUTH_ADDR
          value: {{.AppAuthApi}}.{{$namespace}}.svc.cluster.local:3001

        - key: CONSOLE_ADDR
          value: {{.AppConsoleApi}}.{{$namespace}}.svc.cluster.local:3001

        - key: COMMS_ADDR
          value: {{.AppCommsApi}}.{{$namespace}}.svc.cluster.local:3001

        - key: MONGO_URI
          type: secret
          refName: mres-{{.CiDbName}}
          refKey: URI

        - key: MONGO_DB_NAME
          value: {{.CiDbName}}

        - key: KAFKA_USERNAME
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: PASSWORD

        - key: KAFKA_BROKERS
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: KAFKA_BROKERS

{{/*          HARBOR CREDS*/}}
{{/*  HARBOR_HOST:  $HARBOR_DOMAIN*/}}
{{/*  HARBOR_ADMIN_USERNAME: $HARBOR_ADMIN_USERNAME*/}}
{{/*  HARBOR_ADMIN_PASSWORD: $HARBOR_ADMIN_PASSWORD*/}}
{{/*  HARBOR_REGISTRY_HOST: $HARBOR_REGISTRY_HOST*/}}
        - key: HARBOR_HOST
          type: secret
          refName: {{.HarborAdminCredsSecretName}}
          refKey: HARBOR_IMAGE_REGISTRY_HOST

        - key: HARBOR_REGISTRY_HOST
          type: secret
          refName: {{.HarborAdminCredsSecretName}}
          refKey: HARBOR_IMAGE_REGISTRY_HOST

        - key: HARBOR_ADMIN_USERNAME
          type: secret
          refName: {{.HarborAdminCredsSecretName}}
          refKey: HARBOR_ADMIN_USERNAME

        - key: HARBOR_ADMIN_PASSWORD
          type: secret
          refName: {{.HarborAdminCredsSecretName}}
          refKey: HARBOR_ADMIN_PASSWORD

        - key: GITHUB_WEBHOOK_AUTHZ_SECRET
          type: secret
          refName: {{.OAuthSecretName}}
          refKey: GITHUB_WEBHOOK_AUTHZ_SECRET

        - key: GITLAB_WEBHOOK_AUTHZ_SECRET
          type: secret
          refName: {{.OAuthSecretName}}
          refKey: GITLAB_WEBHOOK_AUTHZ_SECRET

        - key: GITHUB_APP_PK_FILE
          value: /hotspot/github-app-pk.pem

      envFrom:
        - type: secret
          refName: ci-env
        - type: secret
          refName: oauth-secrets
      volumes:
        - mountPath: /hotspot
          type: secret
          refName: {{.OAuthSecretName}}
          items:
            - key: github-app-pk.pem
              fileName: github-app-pk.pem
---
apiVersion: v1
kind: Secret
metadata:
  name: ci-env
  namespace: {{$namespace}}
stringData:
  GRPC_PORT: "3001"
  PORT: "3000"

  ORIGINS: https://studio.apollographql.com

  COOKIE_DOMAIN: ".{{$cookieDomain}}"

  KAFKA_TOPIC_GIT_WEBHOOKS: {{.KafkaTopicGitWebhooks}}
  KAFKA_TOPIC_PIPELINE_RUN_UPDATES: {{.KafkaTopicPipelineRunUpdates}}

  KAFKA_GIT_WEBHOOKS_CONSUMER_ID: "kloudlite/ci-api"
  KL_HOOK_TRIGGER_AUTHZ_SECRET: '***REMOVED***'
{{end}}
