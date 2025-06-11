{{- $appName := "comms" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: "{{$appName}}"
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4 }}
  annotations: {{ include "common.pod-annotations" . | nindent 4 }}
spec:
  serviceAccount: {{ .Values.serviceAccounts.normal.name }}

  nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 4}}

  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.commsApi.replicas }}

  services:
    - port: {{ include "apps.commsApi.httpPort" . }}
    - port: {{ include "apps.commsApi.grpcPort" . }}

  hpa:
    enabled: {{.Values.apps.commsApi.hpa.enabled}}
    minReplicas: {{.Values.apps.commsApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.commsApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.commsApi.image.repository}}:{{.Values.apps.commsApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "80Mi"
        max: "120Mi"

      env:
        - key: HTTP_PORT
          value: {{ include "apps.commsApi.httpPort" . | squote }}

        - key: GRPC_PORT
          value: {{ include "apps.commsApi.grpcPort" .  | squote}}

        - key: SUPPORT_EMAIL
          value: {{.Values.sendgrid.sender}}

        - key: SENDGRID_API_KEY
          value: {{.Values.sendgrid.apiKey}}

        - key: ACCOUNTS_WEB_INVITE_URL
          value: https://auth.{{ .Values.baseDomain }}/invite

        - key: PROJECTS_WEB_INVITE_URL
          value: https://auth.{{ .Values.baseDomain }}/invite

        - key: KLOUDLITE_CONSOLE_WEB_URL
          value: https://console.{{.Values.baseDomain}}

        - key: RESET_PASSWORD_WEB_URL
          value: https://auth.{{.Values.baseDomain}}/reset-password

        - key: VERIFY_EMAIL_WEB_URL
          value: https://auth.{{.Values.baseDomain }}/verify-email
        
        {{/* TODO: url should definitely NOT be auth.Values.baseDomain */}}
        - key: EMAIL_LINKS_BASE_URL
          value: https://auth.{{.Values.baseDomain}}/

        {{- /* notifications params */}}
        - key: ACCOUNT_COOKIE_NAME
          value: {{ include "kloudlite.account-cookie-name" . }}

        - key: NATS_URL
          value: {{ include "nats.url" . }}

        - key: NOTIFICATION_NATS_STREAM
          value: {{.Values.nats.streams.events.name}}

        - key: SESSION_KV_BUCKET
          value: {{.Values.nats.buckets.sessionKVBucket.name}}
        
        - key: IAM_GRPC_ADDR
          value: 'iam:{{ include "apps.iamApi.grpcPort" . }}'

        - key: MONGO_URI
          type: secret
          refName: mres-comms-db-creds
          refKey: .CLUSTER_LOCAL_URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-comms-db-creds
          refKey: DB_NAME
