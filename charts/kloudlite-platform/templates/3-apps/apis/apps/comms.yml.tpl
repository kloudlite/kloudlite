{{- $appName := "accounts-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: comms
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{ .Values.global.clusterSvcAccount }}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.commsApi.configuration.replicas }}

  services:
    - port: 3001

  hpa:
    enabled: true
    minReplicas: {{.Values.apps.commsApi.minReplicas}}
    maxReplicas: {{.Values.apps.commsApi.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: {{.Values.apps.commsApi.image.repository}}:{{.Values.apps.commsApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"

      env:
        - key: GRPC_PORT
          value: "3001"

        - key: SUPPORT_EMAIL
          value: {{.Values.sendGrid.supportEmail}}

        - key: SENDGRID_API_KEY
          value: {{.Values.sendGrid.apiKey}}

        - key: ACCOUNTS_WEB_INVITE_URL
          value: https://auth.{{ include "router-domain" . }}/invite

        - key: PROJECTS_WEB_INVITE_URL
          value: https://auth.{{ include "router-domain" . }}/invite

        - key: KLOUDLITE_CONSOLE_WEB_URL
          value: https://console.{{include "router-domain" .}}

        - key: RESET_PASSWORD_WEB_URL
          value: https://auth.{{include "router-domain" .}}/reset-password

        - key: VERIFY_EMAIL_WEB_URL
          value: https://auth.{{include "router-domain" . }}/verify-email
        
        {{/* TODO: url should definitely NOT be auth.{{.Values.baseDomain}} */}}
        - key: EMAIL_LINKS_BASE_URL
          value: https://auth.{{include "router-domain" .}}/

        {{- /* notifications params */}}
        - key: ACCOUNT_COOKIE_NAME
          value: {{.Values.global.accountCookieName}}

        - key: NATS_URL
          value: {{.Values.envVars.nats.url}}

        - key: NOTIFICATION_NATS_STREAM
          value: {{.Values.envVars.nats.streams.logs.name}}

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}
        
        - key: IAM_GRPC_ADDR
          value: "iam:3001"

        - key: MONGO_URI
          type: secret
          refName: mres-comms-db-creds
          refKey: .CLUSTER_LOCAL_URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-comms-db-creds
          refKey: DB_NAME

        - key: HTTP_PORT
          value: "3000"
